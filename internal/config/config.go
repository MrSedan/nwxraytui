package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Subscriptions SubscriptionConfig `toml:"subscriptions"`
	Proxy         ProxyConfig        `toml:"proxy"`
	Daemon        DaemonConfig       `toml:"daemon"`
}

type SubscriptionConfig struct {
	URLs            []string      `toml:"urls"`
	RefreshInterval time.Duration `toml:"refresh_interval"`
}

type ProxyConfig struct {
	SocksPort int    `toml:"socks_port"`
	HTTPPort  int    `toml:"http_port"`
	Mode      string `toml:"mode"`
}

type DaemonConfig struct {
	Autostart bool `toml:"autostart"`
}

func defaults() *Config {
	return &Config{
		Subscriptions: SubscriptionConfig{RefreshInterval: time.Hour},
		Proxy:         ProxyConfig{SocksPort: 10808, HTTPPort: 10809, Mode: "socks"},
	}
}

func Load(path string) (*Config, error) {
	cfg := defaults()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nwxraytui", "config.toml")
}

func CacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "nwxraytui")
}

func SocketPath() string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("/run/user/%d/nwxraytui/daemon.sock", os.Getuid())
	default:
		tmp := os.Getenv("TMPDIR")
		if tmp == "" {
			tmp = "/tmp"
		}
		return filepath.Join(tmp, "nwxraytui", "daemon.sock")
	}
}
