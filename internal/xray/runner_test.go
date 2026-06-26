package xray_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrsedan/nwxraytui/internal/xray"
)

func mockXrayBin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "xray")
	script := "#!/bin/sh\necho 'Xray started'\nsleep 60\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestRingBuffer(t *testing.T) {
	rb := xray.NewRingBuffer(3)
	rb.Push("a")
	rb.Push("b")
	rb.Push("c")
	rb.Push("d")
	lines := rb.Lines()
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d", len(lines))
	}
	if lines[0] != "b" || lines[2] != "d" {
		t.Fatalf("unexpected order: %v", lines)
	}
}

func TestRunner_StartStop(t *testing.T) {
	bin := mockXrayBin(t)
	cfg := filepath.Join(t.TempDir(), "config.json")
	os.WriteFile(cfg, []byte("{}"), 0o644)

	r := xray.NewRunner(bin)
	if r.IsRunning() {
		t.Fatal("should not be running before Start")
	}
	if err := r.Start(cfg); err != nil {
		t.Fatal(err)
	}
	if !r.IsRunning() {
		t.Fatal("should be running after Start")
	}

	select {
	case line := <-r.LogCh:
		if line != "Xray started" {
			t.Fatalf("got log %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for log")
	}

	if err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	if r.IsRunning() {
		t.Fatal("should not be running after Stop")
	}
}

func TestRunner_DoubleStart(t *testing.T) {
	bin := mockXrayBin(t)
	cfg := filepath.Join(t.TempDir(), "config.json")
	os.WriteFile(cfg, []byte("{}"), 0o644)

	r := xray.NewRunner(bin)
	r.Start(cfg)
	defer r.Stop()

	if err := r.Start(cfg); err == nil {
		t.Fatal("expected error on double Start")
	}
}
