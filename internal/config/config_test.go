package config_test

import (
	"path/filepath"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Proxy.SocksPort != 10808 {
		t.Fatalf("socks port: got %d", cfg.Proxy.SocksPort)
	}
	if cfg.Proxy.HTTPPort != 10809 {
		t.Fatalf("http port: got %d", cfg.Proxy.HTTPPort)
	}
	if cfg.Proxy.Mode != "socks" {
		t.Fatalf("mode: got %q", cfg.Proxy.Mode)
	}
}

func TestSaveLoad_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := &config.Config{}
	cfg.Subscriptions.URLs = []string{"https://example.com/sub"}
	cfg.Proxy.SocksPort = 10808
	cfg.Proxy.HTTPPort = 10809
	cfg.Proxy.Mode = "system"
	if err := config.Save(cfg, path); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Subscriptions.URLs) != 1 || loaded.Subscriptions.URLs[0] != "https://example.com/sub" {
		t.Fatalf("urls: got %v", loaded.Subscriptions.URLs)
	}
	if loaded.Proxy.Mode != "system" {
		t.Fatalf("mode: got %q", loaded.Proxy.Mode)
	}
}
