package panels_test

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/panels"
)

func makeGroups() []ipc.SubscriptionGroup {
	return []ipc.SubscriptionGroup{
		{
			URL:  "https://a.example.com/sub",
			Meta: ipc.SubscriptionMeta{Title: "Sub A"},
			Servers: []ipc.ServerInfo{
				{Remarks: "A1", Config: json.RawMessage(`{}`)},
				{Remarks: "A2", Config: json.RawMessage(`{}`)},
			},
		},
		{
			URL:  "https://b.example.com/sub",
			Meta: ipc.SubscriptionMeta{Title: "Sub B"},
			Servers: []ipc.ServerInfo{
				{Remarks: "B1", Config: json.RawMessage(`{}`)},
			},
		},
	}
}

func navigateDown(tp panels.TabsPanel, n int) panels.TabsPanel {
	for i := 0; i < n; i++ {
		tp, _ = tp.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	return tp
}

func navigateRight(tp panels.TabsPanel, n int) panels.TabsPanel {
	for i := 0; i < n; i++ {
		tp, _ = tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	}
	return tp
}

func navigateLeft(tp panels.TabsPanel, n int) panels.TabsPanel {
	for i := 0; i < n; i++ {
		tp, _ = tp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	}
	return tp
}

func TestTabsPanel_SelectedAbsIdx(t *testing.T) {
	tp := panels.NewTabsPanel()
	tp.Groups = makeGroups()

	// Tab 0, server 0 → abs 0
	if got := tp.SelectedAbsIdx(); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}

	// Navigate to tab 0, server 1 → abs 1
	tp = panels.NewTabsPanel()
	tp.Groups = makeGroups()
	tp = navigateDown(tp, 1)
	if got := tp.SelectedAbsIdx(); got != 1 {
		t.Fatalf("want 1, got %d", got)
	}

	// Navigate to tab 1, server 0 → abs 2
	tp = panels.NewTabsPanel()
	tp.Groups = makeGroups()
	tp = navigateRight(tp, 1)
	if got := tp.SelectedAbsIdx(); got != 2 {
		t.Fatalf("want 2, got %d", got)
	}
}

func TestTabsPanel_SelectedServer(t *testing.T) {
	tp := panels.NewTabsPanel()
	tp.Groups = makeGroups()
	tp = navigateRight(tp, 1) // tab 1, server 0 = B1

	s := tp.SelectedServer()
	if s == nil {
		t.Fatal("SelectedServer returned nil")
	}
	if s.Remarks != "B1" {
		t.Fatalf("want B1, got %q", s.Remarks)
	}
}

func TestTabsPanel_CurrentGroup(t *testing.T) {
	tp := panels.NewTabsPanel()
	tp.Groups = makeGroups()
	tp = navigateRight(tp, 1)

	g := tp.CurrentGroup()
	if g == nil {
		t.Fatal("CurrentGroup returned nil")
	}
	if g.Meta.Title != "Sub B" {
		t.Fatalf("want Sub B, got %q", g.Meta.Title)
	}
}

func TestTabsPanel_CursorClamping(t *testing.T) {
	tp := panels.NewTabsPanel()
	tp.Groups = makeGroups()

	// Can't go above first tab
	tp = navigateLeft(tp, 5)
	if tp.SelectedAbsIdx() != 0 {
		t.Fatalf("should stay at abs 0, got %d", tp.SelectedAbsIdx())
	}

	// Can't go below last server in tab 0
	tp = navigateDown(tp, 10)
	if tp.SelectedAbsIdx() != 1 {
		t.Fatalf("should clamp at last server (abs 1), got %d", tp.SelectedAbsIdx())
	}
}

func TestTabsPanel_TabSwitchResetsCursor(t *testing.T) {
	tp := panels.NewTabsPanel()
	tp.Groups = makeGroups()
	tp = navigateDown(tp, 1)  // server 1 in tab 0
	tp = navigateRight(tp, 1) // switch to tab 1, cursor resets
	if got := tp.SelectedAbsIdx(); got != 2 {
		t.Fatalf("want abs 2 (tab1,srv0), got %d", got)
	}
}

func TestTabsPanel_EmptyGroups(t *testing.T) {
	tp := panels.NewTabsPanel()
	if got := tp.SelectedAbsIdx(); got != -1 {
		t.Fatalf("empty: want -1, got %d", got)
	}
	if tp.SelectedServer() != nil {
		t.Fatal("empty: SelectedServer should be nil")
	}
	if tp.CurrentGroup() != nil {
		t.Fatal("empty: CurrentGroup should be nil")
	}
}
