package daemon_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/daemon"
)

func TestInstallRemoveAutostart(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("autostart only on linux/darwin")
	}
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", orig)

	if err := daemon.InstallAutostart("/usr/bin/nwxraytui"); err != nil {
		t.Fatal(err)
	}

	var expected string
	switch runtime.GOOS {
	case "linux":
		expected = filepath.Join(tmp, ".config", "systemd", "user", "nwxraytui.service")
	case "darwin":
		expected = filepath.Join(tmp, "Library", "LaunchAgents", "dev.nwxraytui.daemon.plist")
	}
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("service file not created at %s: %v", expected, err)
	}

	if err := daemon.RemoveAutostart(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(expected); !os.IsNotExist(err) {
		t.Fatal("service file should be removed")
	}
}
