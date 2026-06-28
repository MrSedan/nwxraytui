package panels_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/panels"
)

func TestInfoPanel_RunningStatus(t *testing.T) {
	p := panels.InfoPanel{
		Status: ipc.EventStatus{Running: true, Mode: "tun"},
	}
	out := p.View(40, 20)
	if !strings.Contains(out, "running") {
		t.Errorf("expected 'running' in output:\n%s", out)
	}
	if !strings.Contains(out, "tun") {
		t.Errorf("expected 'tun' in output:\n%s", out)
	}
}

func TestInfoPanel_SubscriptionMeta(t *testing.T) {
	g := ipc.SubscriptionGroup{
		URL: "https://example.com/sub",
		Meta: ipc.SubscriptionMeta{
			Title:          "VPN Pro",
			Upload:         1024 * 1024 * 1024,      // 1 GB
			Download:       2 * 1024 * 1024 * 1024,  // 2 GB
			Total:          10 * 1024 * 1024 * 1024, // 10 GB
			Expire:         1785600000,
			UpdateInterval: 24,
		},
	}
	p := panels.InfoPanel{Group: &g}
	out := p.View(40, 20)
	if !strings.Contains(out, "VPN Pro") {
		t.Errorf("expected title:\n%s", out)
	}
	if !strings.Contains(out, "3.0 GB / 10.0 GB") {
		t.Errorf("expected traffic:\n%s", out)
	}
	if !strings.Contains(out, "Auto-refresh: 24h") {
		t.Errorf("expected auto-refresh:\n%s", out)
	}
}

func TestInfoPanel_LastRefresh(t *testing.T) {
	p := panels.InfoPanel{
		LastRefresh: time.Date(2026, 6, 26, 14, 32, 0, 0, time.UTC),
	}
	out := p.View(40, 20)
	if !strings.Contains(out, "Last refresh: 14:32") {
		t.Errorf("expected last refresh:\n%s", out)
	}
}

func TestInfoPanel_SpinTick(t *testing.T) {
	p := panels.InfoPanel{Refreshing: true, SpinFrame: 0}
	p.SpinTick()
	if p.SpinFrame != 1 {
		t.Fatalf("want 1, got %d", p.SpinFrame)
	}
	// Wraps at 4
	p.SpinFrame = 3
	p.SpinTick()
	if p.SpinFrame != 0 {
		t.Fatalf("wrap: want 0, got %d", p.SpinFrame)
	}
	// Assert Refreshing branch rendered
	out := p.View(40, 20)
	if !strings.Contains(out, "refreshing") {
		t.Errorf("expected 'refreshing' in output:\n%s", out)
	}
}

func TestInfoPanel_NoExpiryWhenZero(t *testing.T) {
	g := ipc.SubscriptionGroup{
		Meta: ipc.SubscriptionMeta{Expire: 0},
	}
	p := panels.InfoPanel{Group: &g}
	out := p.View(40, 20)
	if strings.Contains(out, "Expires") {
		t.Errorf("should not show Expires when expire=0:\n%s", out)
	}
}
