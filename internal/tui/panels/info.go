package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

var spinFrames = []string{"|", "/", "—", "\\"}

type InfoPanel struct {
	Status      ipc.EventStatus
	Group       *ipc.SubscriptionGroup
	LastRefresh time.Time
	Refreshing  bool
	SpinFrame   int
}

func (m *InfoPanel) SpinTick() {
	m.SpinFrame = (m.SpinFrame + 1) % len(spinFrames)
}

func (m InfoPanel) View(width, height int) string {
	var sb strings.Builder

	// Connection status line
	icon, statusText := "○", "stopped"
	if m.Status.Running {
		icon, statusText = "●", "running"
	}
	mode := m.Status.Mode
	if mode == "" {
		mode = "socks"
	}
	sb.WriteString(fmt.Sprintf("%s %s  %s\n\n", icon, statusText, mode))

	// Subscription info
	if m.Group != nil {
		title := m.Group.Meta.Title
		if title == "" {
			title = m.Group.URL
			if len(title) > 30 {
				title = title[:27] + "..."
			}
		}
		sb.WriteString(title + "\n")

		used := m.Group.Meta.Upload + m.Group.Meta.Download
		total := m.Group.Meta.Total
		if total > 0 || used > 0 {
			if total == 0 {
				sb.WriteString(fmt.Sprintf("Used: %s (unlimited)\n", formatBytes(used)))
			} else {
				sb.WriteString(fmt.Sprintf("Used: %s / %s\n", formatBytes(used), formatBytes(total)))
			}
		}

		if m.Group.Meta.Expire > 0 {
			exp := time.Unix(m.Group.Meta.Expire, 0)
			sb.WriteString(fmt.Sprintf("Expires: %s\n", exp.Format("2006-01-02")))
		}

		if m.Group.Meta.UpdateInterval > 0 {
			sb.WriteString(fmt.Sprintf("Auto-refresh: %dh\n", m.Group.Meta.UpdateInterval))
		}

		if m.Group.Meta.Announce != "" {
			sb.WriteString("\n" + m.Group.Meta.Announce + "\n")
		}
	}

	sb.WriteString("\n")

	// Refresh indicator
	if m.Refreshing {
		sb.WriteString(fmt.Sprintf("%s refreshing...\n", spinFrames[m.SpinFrame]))
	} else if !m.LastRefresh.IsZero() {
		sb.WriteString(fmt.Sprintf("Last refresh: %s\n", m.LastRefresh.Format("15:04")))
	}

	return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(sb.String())
}

func formatBytes(b int64) string {
	if b == 0 {
		return "0 B"
	}
	gb := float64(b) / (1024 * 1024 * 1024)
	if gb >= 1 {
		return fmt.Sprintf("%.1f GB", gb)
	}
	mb := float64(b) / (1024 * 1024)
	if mb >= 1 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	return fmt.Sprintf("%d B", b)
}
