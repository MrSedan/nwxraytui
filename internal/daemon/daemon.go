package daemon

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/mrsedan/nwxraytui/internal/config"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/latency"
	"github.com/mrsedan/nwxraytui/internal/proxy"
	"github.com/mrsedan/nwxraytui/internal/subscription"
	"github.com/mrsedan/nwxraytui/internal/xray"
)

type Daemon struct {
	cfg              *config.Config
	cfgPath          string
	runner           *xray.Runner
	groups           []subscription.Group
	activeIdx        int
	mode             string
	activeServerHost string
	tunAvailable     bool
	clients          []net.Conn
	mu               sync.RWMutex
}

func New(cfg *config.Config, xrayBin string, cfgPath string) *Daemon {
	d := &Daemon{
		cfg:          cfg,
		cfgPath:      cfgPath,
		runner:       xray.NewRunner(xrayBin),
		activeIdx:    -1,
		mode:         cfg.Proxy.Mode,
		tunAvailable: proxy.HasTunCapability(),
	}
	cachePath := filepath.Join(config.CacheDir(), "groups.json")
	if cached, err := subscription.LoadCachedGroups(cachePath); err == nil && len(cached) > 0 {
		d.groups = cached
	}
	return d
}

func (d *Daemon) Run(socketPath string) error {
	srv, err := ipc.NewServer(socketPath)
	if err != nil {
		return err
	}
	defer srv.Close()

	go d.forwardLogs()
	go d.autostart()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		d.disconnect()
		srv.Close()
		os.Exit(0)
	}()

	for {
		conn, err := srv.Accept()
		if err != nil {
			return err
		}
		go d.handleConn(conn)
	}
}

func (d *Daemon) handleConn(conn net.Conn) {
	d.addClient(conn)
	defer func() {
		d.removeClient(conn)
		conn.Close()
	}()

	d.sendTo(conn, d.statusEvent())
	d.mu.RLock()
	groups := d.toIPCGroups()
	d.mu.RUnlock()
	if len(groups) > 0 {
		d.sendTo(conn, ipc.EventSubscriptionList{Groups: groups})
	}

	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for sc.Scan() {
		env, err := ipc.Decode(sc.Bytes())
		if err != nil {
			continue
		}
		d.handleEnvelope(env)
	}
}

func (d *Daemon) handleEnvelope(env ipc.Envelope) {
	switch env.Type {
	case ipc.TypeCmdStart:
		cmd, _ := ipc.UnmarshalPayload[ipc.CmdStart](env)
		d.connect(cmd.ServerIdx, cmd.Mode)
	case ipc.TypeCmdStop:
		d.disconnect()
	case ipc.TypeCmdSwitch:
		cmd, _ := ipc.UnmarshalPayload[ipc.CmdSwitch](env)
		d.disconnect()
		d.connect(cmd.ServerIdx, cmd.Mode)
	case ipc.TypeCmdRefresh:
		go d.refresh()
	case ipc.TypeCmdPing:
		go d.pingAll()
	case ipc.TypeCmdSetAutostart:
		cmd, _ := ipc.UnmarshalPayload[ipc.CmdSetAutostart](env)
		d.cfg.Daemon.Autostart = cmd.Enabled
	case ipc.TypeCmdAddSub:
		cmd, _ := ipc.UnmarshalPayload[ipc.CmdAddSub](env)
		d.cfg.Subscriptions.URLs = append(d.cfg.Subscriptions.URLs, cmd.URL)
		if err := config.Save(d.cfg, d.cfgPath); err != nil {
			log.Printf("save config: %v", err)
		}
		go d.refresh()
	case ipc.TypeCmdRemoveSub:
		cmd, _ := ipc.UnmarshalPayload[ipc.CmdRemoveSub](env)
		d.removeSub(cmd.URL)
		if err := config.Save(d.cfg, d.cfgPath); err != nil {
			log.Printf("save config: %v", err)
		}
	}
}

func (d *Daemon) serverAtIdx(idx int) (subscription.Server, bool) {
	if idx < 0 {
		return subscription.Server{}, false
	}
	cur := 0
	for _, g := range d.groups {
		if idx < cur+len(g.Servers) {
			return g.Servers[idx-cur], true
		}
		cur += len(g.Servers)
	}
	return subscription.Server{}, false
}

func (d *Daemon) connect(idx int, mode string) {
	d.mu.Lock()
	srv, ok := d.serverAtIdx(idx)
	d.mu.Unlock()
	if !ok {
		return
	}

	merged, err := xray.MergeConfig(srv.Config, mode)
	if err != nil {
		d.broadcast(ipc.EventStatus{Error: err.Error()})
		return
	}

	cacheDir := config.CacheDir()
	os.MkdirAll(cacheDir, 0o700)
	cfgPath := filepath.Join(cacheDir, "active-config.json")
	if err := os.WriteFile(cfgPath, merged, 0o600); err != nil {
		d.broadcast(ipc.EventStatus{Error: err.Error()})
		return
	}

	if err := d.runner.Start(cfgPath); err != nil {
		d.broadcast(ipc.EventStatus{Error: err.Error()})
		return
	}

	serverHost := xray.ServerHost(srv.Config)
	if mode == "tun" {
		if err := proxy.SetTunRoutes(serverHost, xray.DNSServerIP(merged)); err != nil {
			log.Printf("SetTunRoutes: %v", err)
		}
	}

	d.mu.Lock()
	d.activeIdx = idx
	d.mode = mode
	d.activeServerHost = serverHost
	d.mu.Unlock()

	d.broadcast(d.statusEvent())
}

func (d *Daemon) disconnect() {
	d.mu.RLock()
	mode := d.mode
	serverHost := d.activeServerHost
	d.mu.RUnlock()

	d.runner.Stop()

	if mode == "tun" {
		proxy.UnsetTunRoutes(serverHost)
	}

	d.mu.Lock()
	d.activeIdx = -1
	d.activeServerHost = ""
	d.mu.Unlock()

	d.broadcast(d.statusEvent())
}

func (d *Daemon) refresh() {
	fetcher := subscription.NewFetcher(nil)
	var groups []subscription.Group
	for _, url := range d.cfg.Subscriptions.URLs {
		servers, meta, err := fetcher.Fetch(url)
		if err != nil {
			log.Printf("refresh %s: %v", url, err)
			continue
		}
		groups = append(groups, subscription.Group{URL: url, Meta: meta, Servers: servers})
	}

	if len(groups) > 0 {
		d.mu.Lock()
		d.groups = groups
		d.mu.Unlock()

		cachePath := filepath.Join(config.CacheDir(), "groups.json")
		os.MkdirAll(config.CacheDir(), 0o700)
		subscription.CacheGroups(groups, cachePath)
	}

	d.mu.RLock()
	ipcGroups := d.toIPCGroups()
	d.mu.RUnlock()

	d.broadcast(ipc.EventSubscriptionList{Groups: ipcGroups})
	go d.pingAll()
}

func (d *Daemon) toIPCGroups() []ipc.SubscriptionGroup {
	groups := make([]ipc.SubscriptionGroup, len(d.groups))
	for i, g := range d.groups {
		servers := make([]ipc.ServerInfo, len(g.Servers))
		for j, s := range g.Servers {
			servers[j] = ipc.ServerInfo{Remarks: s.Remarks, Config: s.Config}
		}
		groups[i] = ipc.SubscriptionGroup{
			URL: g.URL,
			Meta: ipc.SubscriptionMeta{
				Title:          g.Meta.Title,
				Upload:         g.Meta.Upload,
				Download:       g.Meta.Download,
				Total:          g.Meta.Total,
				Expire:         g.Meta.Expire,
				UpdateInterval: g.Meta.UpdateInterval,
			},
			Servers: servers,
		}
	}
	return groups
}

func (d *Daemon) pingAll() {
	d.mu.RLock()
	groups := make([]subscription.Group, len(d.groups))
	copy(groups, d.groups)
	d.mu.RUnlock()

	abs := 0
	for _, g := range groups {
		for _, s := range g.Servers {
			var meta struct {
				Outbounds []struct {
					Settings struct {
						Address string `json:"address"`
						Port    int    `json:"port"`
						Vnext   []struct {
							Address string `json:"address"`
							Port    int    `json:"port"`
						} `json:"vnext"`
						Servers []struct {
							Address string `json:"address"`
							Port    int    `json:"port"`
						} `json:"servers"`
					} `json:"settings"`
				} `json:"outbounds"`
			}
			host, port := "", 0
			if err := json.Unmarshal(s.Config, &meta); err == nil {
				for _, ob := range meta.Outbounds {
					if ob.Settings.Address != "" {
						host = ob.Settings.Address
						port = ob.Settings.Port
						break
					}
					if len(ob.Settings.Vnext) > 0 {
						host = ob.Settings.Vnext[0].Address
						port = ob.Settings.Vnext[0].Port
						break
					}
					if len(ob.Settings.Servers) > 0 {
						host = ob.Settings.Servers[0].Address
						port = ob.Settings.Servers[0].Port
						break
					}
				}
			}
			if host != "" {
				ms := latency.PingDirect(host, port, 3e9, xray.TunFwmark)
				d.broadcast(ipc.EventLatency{ServerIdx: abs, Ms: ms})
			}
			abs++
		}
	}
}

func (d *Daemon) autostart() {
	state, err := config.LoadState()
	if err != nil || state.LastServerIdx < 0 {
		return
	}
	d.mu.RLock()
	hasGroups := len(d.groups) > 0
	d.mu.RUnlock()
	if !hasGroups {
		return
	}
	if _, ok := d.serverAtIdx(state.LastServerIdx); !ok {
		return
	}
	mode := d.cfg.Proxy.Mode
	if mode == "tun" || mode == "" {
		mode = "socks"
	}
	d.connect(state.LastServerIdx, mode)
}

func (d *Daemon) removeSub(url string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	urls := d.cfg.Subscriptions.URLs[:0]
	for _, u := range d.cfg.Subscriptions.URLs {
		if u != url {
			urls = append(urls, u)
		}
	}
	d.cfg.Subscriptions.URLs = urls
}

func (d *Daemon) forwardLogs() {
	for line := range d.runner.LogCh {
		d.broadcast(ipc.EventLog{Line: line})
	}
}

func (d *Daemon) statusEvent() ipc.EventStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return ipc.EventStatus{
		Running:      d.runner.IsRunning(),
		ServerIdx:    d.activeIdx,
		Mode:         d.mode,
		TunAvailable: d.tunAvailable,
	}
}

func (d *Daemon) broadcast(v any) {
	data, err := ipc.Encode(v)
	if err != nil {
		return
	}
	data = append(data, '\n')
	d.mu.RLock()
	clients := make([]net.Conn, len(d.clients))
	copy(clients, d.clients)
	d.mu.RUnlock()
	for _, c := range clients {
		c.Write(data)
	}
}

func (d *Daemon) sendTo(conn net.Conn, v any) {
	data, err := ipc.Encode(v)
	if err != nil {
		return
	}
	conn.Write(append(data, '\n'))
}

func (d *Daemon) addClient(conn net.Conn) {
	d.mu.Lock()
	d.clients = append(d.clients, conn)
	d.mu.Unlock()
}

func (d *Daemon) removeClient(conn net.Conn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, c := range d.clients {
		if c == conn {
			d.clients = append(d.clients[:i], d.clients[i+1:]...)
			return
		}
	}
}
