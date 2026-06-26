package daemon_test

import (
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

	d := daemon.New(cfg, "/bin/false")
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
