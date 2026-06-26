package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/mrsedan/nwxraytui/internal/config"
	"github.com/mrsedan/nwxraytui/internal/daemon"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui"
)

func main() {
	isDaemon := flag.Bool("daemon", false, "run as background daemon")
	flag.Parse()

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if *isDaemon {
		runDaemon(cfg, cfgPath)
		return
	}
	runTUI(cfg)
}

func runDaemon(cfg *config.Config, cfgPath string) {
	xrayBin, err := exec.LookPath("xray")
	if err != nil {
		log.Fatalf("xray not found in PATH: %v", err)
	}
	sock := config.SocketPath()
	d := daemon.New(cfg, xrayBin, cfgPath)
	if err := d.Run(sock); err != nil {
		log.Fatalf("daemon: %v", err)
	}
}

func runTUI(cfg *config.Config) {
	sock := config.SocketPath()

	client, err := ipc.NewClient(sock)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Daemon not running. Starting...")
		go func() {
			cmd := exec.Command(os.Args[0], "--daemon")
			cmd.Start()
		}()
		time.Sleep(500 * time.Millisecond)
		client, err = ipc.NewClient(sock)
		if err != nil {
			log.Fatalf("cannot connect to daemon: %v", err)
		}
	}
	defer client.Close()

	app := tui.New(client)
	if err := app.Start(); err != nil {
		log.Fatalf("TUI: %v", err)
	}
}
