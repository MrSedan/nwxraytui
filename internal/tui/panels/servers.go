package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

type ServerList struct {
	Servers   []ipc.ServerInfo
	Latencies map[int]int
	cursor    int
}

func NewServerList() ServerList {
	return ServerList{Latencies: map[int]int{}}
}

func (m ServerList) SelectedIdx() int { return m.cursor }

func (m ServerList) Update(msg tea.Msg) (ServerList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Servers)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m ServerList) View(width, height int) string {
	if len(m.Servers) == 0 {
		return styles.DimText.Render("No servers. Press [A] to add a subscription.")
	}
	lines := make([]string, 0, len(m.Servers))
	for i, s := range m.Servers {
		ms, ok := m.Latencies[i]
		var lat string
		if !ok {
			lat = styles.DimText.Render("— ms")
		} else {
			lat = styles.LatencyStyle(ms).Render(fmt.Sprintf("%d ms", ms))
		}
		line := fmt.Sprintf("  %s  %s", s.Remarks, lat)
		if i == m.cursor {
			line = styles.ActiveItem.Render("> " + s.Remarks + "  " + lat)
		}
		lines = append(lines, line)
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(content)
}
