package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

type TabsPanel struct {
	Groups       []ipc.SubscriptionGroup
	Latencies    map[int]int
	tabCursor    int
	serverCursor int
}

func NewTabsPanel() TabsPanel {
	return TabsPanel{Latencies: map[int]int{}}
}

func (m TabsPanel) SelectedAbsIdx() int {
	if len(m.Groups) == 0 {
		return -1
	}
	if m.tabCursor < 0 || m.tabCursor >= len(m.Groups) {
		return -1
	}
	g := m.Groups[m.tabCursor]
	if m.serverCursor < 0 || m.serverCursor >= len(g.Servers) {
		return -1
	}
	cur := 0
	for i := 0; i < m.tabCursor; i++ {
		cur += len(m.Groups[i].Servers)
	}
	return cur + m.serverCursor
}

func (m TabsPanel) SelectedServer() *ipc.ServerInfo {
	if m.tabCursor < 0 || m.tabCursor >= len(m.Groups) {
		return nil
	}
	g := m.Groups[m.tabCursor]
	if m.serverCursor < 0 || m.serverCursor >= len(g.Servers) {
		return nil
	}
	s := g.Servers[m.serverCursor]
	return &s
}

func (m TabsPanel) CurrentGroup() *ipc.SubscriptionGroup {
	if m.tabCursor < 0 || m.tabCursor >= len(m.Groups) {
		return nil
	}
	g := m.Groups[m.tabCursor]
	return &g
}

func (m TabsPanel) ServerAtAbsIdx(idx int) *ipc.ServerInfo {
	cur := 0
	for _, g := range m.Groups {
		if idx < cur+len(g.Servers) {
			s := g.Servers[idx-cur]
			return &s
		}
		cur += len(g.Servers)
	}
	return nil
}

func (m *TabsPanel) SetCursorByAbsIdx(idx int) bool {
	cur := 0
	for i, g := range m.Groups {
		if idx < cur+len(g.Servers) {
			m.tabCursor = i
			m.serverCursor = idx - cur
			return true
		}
		cur += len(g.Servers)
	}
	return false
}

func (m *TabsPanel) SetCursorByName(name string) bool {
	for i, g := range m.Groups {
		for j, s := range g.Servers {
			if s.Remarks == name {
				m.tabCursor = i
				m.serverCursor = j
				return true
			}
		}
	}
	return false
}

func (m TabsPanel) Update(msg tea.Msg) (TabsPanel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.tabCursor > 0 {
				m.tabCursor--
				m.serverCursor = 0
			}
		case "right", "l":
			if m.tabCursor < len(m.Groups)-1 {
				m.tabCursor++
				m.serverCursor = 0
			}
		case "up", "k":
			if m.serverCursor > 0 {
				m.serverCursor--
			}
		case "down", "j":
			if m.tabCursor >= 0 && m.tabCursor < len(m.Groups) {
				maxIdx := len(m.Groups[m.tabCursor].Servers) - 1
				if m.serverCursor < maxIdx {
					m.serverCursor++
				}
			}
		}
	}
	return m, nil
}

func (m TabsPanel) TabBarView(totalWidth int) string {
	if len(m.Groups) == 0 {
		return styles.DimText.Render("[No subscriptions]")
	}
	var parts []string
	for i, g := range m.Groups {
		title := g.Meta.Title
		if title == "" {
			title = g.URL
			if len(title) > 20 {
				title = title[:17] + "..."
			}
		}
		label := "[ " + title + " ]"
		if i == m.tabCursor {
			parts = append(parts, styles.ActiveItem.Render(label))
		} else {
			parts = append(parts, styles.DimText.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func (m TabsPanel) View(width, height int) string {
	if len(m.Groups) == 0 {
		return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(
			styles.DimText.Render("No servers. Press [A] to add a subscription."),
		)
	}
	if m.tabCursor < 0 || m.tabCursor >= len(m.Groups) {
		return styles.PanelBorder.Width(width - 2).Height(height - 2).Render("")
	}
	g := m.Groups[m.tabCursor]
	if len(g.Servers) == 0 {
		return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(
			styles.DimText.Render("No servers in this subscription."),
		)
	}

	base := 0
	for i := 0; i < m.tabCursor; i++ {
		base += len(m.Groups[i].Servers)
	}

	var sb strings.Builder
	for i, s := range g.Servers {
		absIdx := base + i
		ms, ok := m.Latencies[absIdx]
		var lat string
		if !ok {
			lat = styles.DimText.Render("— ms")
		} else {
			lat = styles.LatencyStyle(ms).Render(fmt.Sprintf("%d ms", ms))
		}
		if i == m.serverCursor {
			sb.WriteString(styles.ActiveItem.Render(fmt.Sprintf("> %-30s %s", s.Remarks, lat)) + "\n")
		} else {
			sb.WriteString(fmt.Sprintf("  %-30s %s\n", s.Remarks, lat))
		}
	}
	return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(sb.String())
}
