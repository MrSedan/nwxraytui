package styles

import "github.com/charmbracelet/lipgloss"

var (
	BorderColor   = lipgloss.Color("240")
	ActiveColor   = lipgloss.Color("86")
	InactiveColor = lipgloss.Color("241")
	ErrorColor    = lipgloss.Color("196")
	LatencyGood   = lipgloss.Color("82")
	LatencyMed    = lipgloss.Color("226")
	LatencyBad    = lipgloss.Color("196")

	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor)

	ActiveItem = lipgloss.NewStyle().
			Foreground(ActiveColor).
			Bold(true)

	DimText = lipgloss.NewStyle().
		Foreground(InactiveColor)

	StatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().Foreground(ErrorColor)
)

func LatencyStyle(ms int) lipgloss.Style {
	switch {
	case ms < 0:
		return DimText
	case ms < 100:
		return lipgloss.NewStyle().Foreground(LatencyGood)
	case ms < 300:
		return lipgloss.NewStyle().Foreground(LatencyMed)
	default:
		return lipgloss.NewStyle().Foreground(LatencyBad)
	}
}
