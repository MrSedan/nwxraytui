package panels

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

const maxLogLines = 500

type LogPanel struct {
	lines  []string
	offset int
}

func (m *LogPanel) Push(line string) {
	m.lines = append(m.lines, line)
	if len(m.lines) > maxLogLines {
		m.lines = m.lines[len(m.lines)-maxLogLines:]
	}
	if m.offset > 0 {
		m.offset++
	}
}

func (m *LogPanel) Update(msg tea.Msg) (*LogPanel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "pgup":
			m.offset += 5
			if m.offset > len(m.lines) {
				m.offset = len(m.lines)
			}
		case "pgdown":
			m.offset -= 5
			if m.offset < 0 {
				m.offset = 0
			}
		}
	}
	return m, nil
}

func (m *LogPanel) View(width, height int) string {
	inner := height - 4
	if inner < 1 {
		inner = 1
	}
	end := len(m.lines) - m.offset
	if end < 0 {
		end = 0
	}
	start := end - inner
	if start < 0 {
		start = 0
	}
	visible := m.lines[start:end]
	var sb strings.Builder
	for _, l := range visible {
		if len(l) > width-4 {
			l = l[:width-4]
		}
		sb.WriteString(l + "\n")
	}
	return styles.PanelBorder.Width(width - 2).Render(sb.String())
}
