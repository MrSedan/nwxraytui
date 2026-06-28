package daemon_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrsedan/nwxraytui/internal/config"
	"github.com/mrsedan/nwxraytui/internal/daemon"
	"github.com/mrsedan/nwxraytui/internal/ipc"
)

func TestDaemon_ConnectAndStatus(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "daemon.sock")
	cfg := &config.Config{}
	cfg.Proxy.SocksPort = 10808
	cfg.Proxy.HTTPPort = 10809
	cfg.Proxy.Mode = "socks"

	d := daemon.New(cfg, "/bin/false", "")
	go d.Run(sock)
	time.Sleep(100 * time.Millisecond)

	client, err := ipc.NewClient(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	env, err := client.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != ipc.TypeEventStatus {
		t.Fatalf("first event: got %q, want EventStatus", env.Type)
	}
	ev, err := ipc.UnmarshalPayload[ipc.EventStatus](env)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Running {
		t.Fatal("daemon should report not running before any CmdStart")
	}
}

func TestDaemon_RefreshBroadcastsSubscriptionList(t *testing.T) {
	// Serve a fake subscription endpoint
	payload := `[{"remarks":"S1","inbounds":[],"outbounds":[],"routing":{},"dns":{}}]`
	subSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("profile-title", "Test Sub")
		w.Write([]byte(payload))
	}))
	defer subSrv.Close()

	sock := filepath.Join(t.TempDir(), "daemon.sock")
	cfg := &config.Config{}
	cfg.Proxy.SocksPort = 10810
	cfg.Proxy.HTTPPort = 10811
	cfg.Proxy.Mode = "socks"
	cfg.Subscriptions.URLs = []string{subSrv.URL}

	d := daemon.New(cfg, "/bin/false", "")
	go d.Run(sock)
	time.Sleep(100 * time.Millisecond)

	client, err := ipc.NewClient(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// First event is always status
	env, err := client.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != ipc.TypeEventStatus {
		t.Fatalf("first event: got %q, want EventStatus", env.Type)
	}

	// Send refresh
	client.Send(ipc.CmdRefresh{})

	// Wait for EventSubscriptionList
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		env, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if env.Type != ipc.TypeEventSubscriptionList {
			continue
		}
		ev, err := ipc.UnmarshalPayload[ipc.EventSubscriptionList](env)
		if err != nil {
			t.Fatal(err)
		}
		if len(ev.Groups) != 1 {
			t.Fatalf("want 1 group, got %d", len(ev.Groups))
		}
		if ev.Groups[0].Meta.Title != "Test Sub" {
			t.Fatalf("title: got %q", ev.Groups[0].Meta.Title)
		}
		if len(ev.Groups[0].Servers) != 1 || ev.Groups[0].Servers[0].Remarks != "S1" {
			t.Fatalf("servers: %+v", ev.Groups[0].Servers)
		}
		return
	}
	t.Fatal("timed out waiting for EventSubscriptionList")
}
