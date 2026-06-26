package panels

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

type DetailPanel struct {
	Server       *ipc.ServerInfo
	Status       ipc.EventStatus
	TunAvailable bool
}

func (m DetailPanel) View(width, height int) string {
	var sb strings.Builder

	if m.Server == nil {
		sb.WriteString(styles.DimText.Render("No server selected."))
	} else {
		sb.WriteString(fmt.Sprintf("%-12s %s\n", "Name:", m.Server.Remarks))

		var cfg struct {
			Outbounds []struct {
				Tag      string `json:"tag"`
				Protocol string `json:"protocol"`
				Settings struct {
					Vnext []struct {
						Address string `json:"address"`
						Port    int    `json:"port"`
					} `json:"vnext"`
				} `json:"settings"`
				Stream struct {
					Network  string `json:"network"`
					Security string `json:"security"`
				} `json:"streamSettings"`
			} `json:"outbounds"`
			Inbounds []struct {
				Tag  string `json:"tag"`
				Port int    `json:"port"`
			} `json:"inbounds"`
		}
		if err := json.Unmarshal(m.Server.Config, &cfg); err == nil {
			for _, ob := range cfg.Outbounds {
				if ob.Tag != "proxy" {
					continue
				}
				sb.WriteString(fmt.Sprintf("%-12s %s/%s\n", "Protocol:", ob.Protocol, ob.Stream.Network))
				if len(ob.Settings.Vnext) > 0 {
					sb.WriteString(fmt.Sprintf("%-12s %s:%d\n", "Server:", ob.Settings.Vnext[0].Address, ob.Settings.Vnext[0].Port))
				}
				if ob.Stream.Security != "" {
					sb.WriteString(fmt.Sprintf("%-12s %s\n", "Security:", ob.Stream.Security))
				}
			}
			sb.WriteString("\nInbounds:\n")
			for _, ib := range cfg.Inbounds {
				sb.WriteString(fmt.Sprintf("  %-6s 127.0.0.1:%d\n", ib.Tag, ib.Port))
			}
		}

		sb.WriteString("\n")
	}

	tunLabel := "[T] TUN"
	if !m.TunAvailable {
		tunLabel = styles.DimText.Render("[T] TUN (set enableTun=true in NixOS config)")
	}
	sb.WriteString("[S] Socks/HTTP  " + tunLabel + "\n")
	sb.WriteString("[P] System proxy  [Q] Stop\n")

	return styles.PanelBorder.Width(width - 2).Height(height - 2).Render(sb.String())
}
