package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func InstallAutostart(binPath string) error {
	switch runtime.GOOS {
	case "linux":
		return installSystemd(binPath)
	case "darwin":
		return installLaunchd(binPath)
	default:
		return fmt.Errorf("autostart not supported on %s", runtime.GOOS)
	}
}

func RemoveAutostart() error {
	switch runtime.GOOS {
	case "linux":
		return os.Remove(systemdServicePath())
	case "darwin":
		return os.Remove(launchdPlistPath())
	default:
		return nil
	}
}

func installSystemd(binPath string) error {
	path := systemdServicePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf(`[Unit]
Description=nwxraytui proxy daemon
After=network.target

[Service]
ExecStart=%s --daemon
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, binPath)
	return os.WriteFile(path, []byte(content), 0o644)
}

func installLaunchd(binPath string) error {
	path := launchdPlistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.nwxraytui.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
`, binPath)
	return os.WriteFile(path, []byte(content), 0o644)
}

func systemdServicePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", "nwxraytui.service")
}

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "dev.nwxraytui.daemon.plist")
}
