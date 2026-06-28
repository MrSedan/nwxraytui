package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/panels"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

type ipcMsg struct{ env ipc.Envelope }
type ipcErrMsg struct{ err error }
type spinTickMsg struct{}

type App struct {
	client      *ipc.Client
	tabs        panels.TabsPanel
	info        panels.InfoPanel
	detail      panels.DetailPanel
	logPanel    panels.LogPanel
	status      ipc.EventStatus
	width       int
	height      int
	errMsg      string
	inputMode   bool
	inputCmd    string
	inputText   string
	proxyMode   string
	showDetails bool
	spinning    bool
}

func New(client *ipc.Client) *App {
	return &App{
		client:    client,
		tabs:      panels.NewTabsPanel(),
		proxyMode: "socks",
	}
}

func (a *App) Start() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) Init() tea.Cmd {
	a.spinning = true
	a.info.Refreshing = true
	return tea.Batch(
		recvIPC(a.client),
		sendCmd(a.client, ipc.CmdRefresh{}),
		tickSpinner(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case spinTickMsg:
		if a.spinning {
			a.info.SpinTick()
			cmds = append(cmds, tickSpinner())
		}

	case ipcMsg:
		cmds = append(cmds, recvIPC(a.client))
		a.handleIPC(msg.env)

	case ipcErrMsg:
		a.errMsg = msg.err.Error()

	case tea.KeyMsg:
		if a.inputMode {
			switch msg.String() {
			case "enter":
				if a.inputText != "" {
					switch a.inputCmd {
					case "add":
						cmds = append(cmds, sendCmd(a.client, ipc.CmdAddSub{URL: a.inputText}))
					case "del":
						cmds = append(cmds, sendCmd(a.client, ipc.CmdRemoveSub{URL: a.inputText}))
					}
				}
				a.inputMode = false
				a.inputText = ""
			case "esc", "ctrl+c":
				a.inputMode = false
				a.inputText = ""
			case "backspace", "ctrl+h":
				if len(a.inputText) > 0 {
					a.inputText = a.inputText[:len(a.inputText)-1]
				}
			default:
				if msg.Type == tea.KeyRunes {
					a.inputText += string(msg.Runes)
				}
			}
			return a, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "a":
			a.inputMode = true
			a.inputCmd = "add"
			a.inputText = ""
		case "d":
			a.inputMode = true
			a.inputCmd = "del"
			a.inputText = ""
		case " ":
			if a.status.Running {
				cmds = append(cmds, sendCmd(a.client, ipc.CmdStop{}))
			} else if a.tabs.SelectedAbsIdx() >= 0 {
				cmds = append(cmds, sendCmd(a.client, ipc.CmdStart{
					ServerIdx: a.tabs.SelectedAbsIdx(),
					Mode:      a.status.Mode,
				}))
			}
		case "t":
			if a.status.TunAvailable {
				mode := "tun"
				if a.status.Running && a.status.Mode == "tun" {
					mode = a.proxyMode
				}
				cmds = append(cmds, sendCmd(a.client, ipc.CmdSwitch{
					ServerIdx: a.tabs.SelectedAbsIdx(),
					Mode:      mode,
				}))
			}
		case "r":
			if !a.spinning {
				cmds = append(cmds, tickSpinner())
			}
			a.spinning = true
			a.info.Refreshing = true
			cmds = append(cmds, sendCmd(a.client, ipc.CmdRefresh{}))
		case "enter":
			a.showDetails = !a.showDetails
			if a.showDetails {
				a.detail.Server = a.tabs.SelectedServer()
			}
		case "esc":
			a.showDetails = false
		}
	}

	var newTabs panels.TabsPanel
	newTabs, _ = a.tabs.Update(msg)
	a.tabs = newTabs

	// Keep detail server in sync when browsing with showDetails open
	if a.showDetails {
		a.detail.Server = a.tabs.SelectedServer()
	}

	a.info.Status = a.status
	a.info.Group = a.tabs.CurrentGroup()

	lp, _ := (&a.logPanel).Update(msg)
	a.logPanel = *lp

	return a, tea.Batch(cmds...)
}

func (a *App) handleIPC(env ipc.Envelope) {
	switch env.Type {
	case ipc.TypeEventStatus:
		ev, _ := ipc.UnmarshalPayload[ipc.EventStatus](env)
		a.status = ev
		a.detail.Status = ev
		a.detail.TunAvailable = ev.TunAvailable
		if ev.Mode != "" && ev.Mode != "tun" {
			a.proxyMode = ev.Mode
		}
	case ipc.TypeEventSubscriptionList:
		ev, _ := ipc.UnmarshalPayload[ipc.EventSubscriptionList](env)
		a.tabs.Groups = ev.Groups
		a.spinning = false
		a.info.Refreshing = false
		a.info.LastRefresh = time.Now()
		a.info.Group = a.tabs.CurrentGroup()
	case ipc.TypeEventLatency:
		ev, _ := ipc.UnmarshalPayload[ipc.EventLatency](env)
		a.tabs.Latencies[ev.ServerIdx] = ev.Ms
	case ipc.TypeEventLog:
		ev, _ := ipc.UnmarshalPayload[ipc.EventLog](env)
		a.logPanel.Push(ev.Line)
	}
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	tabsW := a.width * 2 / 3
	infoW := a.width - tabsW
	topH := (a.height - 3) * 2 / 3
	botH := a.height - topH - 3

	tabBar := a.tabs.TabBarView(a.width)
	serverPane := a.tabs.View(tabsW, topH)

	var rightPane string
	if a.showDetails {
		rightPane = a.detail.View(infoW, topH)
	} else {
		rightPane = a.info.View(infoW, topH)
	}

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, serverPane, rightPane)
	bottom := a.logPanel.View(a.width, botH)
	statusLine := a.renderStatus()

	var helpLine string
	if a.inputMode {
		prompt := "Add subscription URL"
		if a.inputCmd == "del" {
			prompt = "Remove subscription URL"
		}
		helpLine = styles.DimText.Render(prompt+": ") + a.inputText + "█"
	} else {
		helpLine = styles.DimText.Render("←→ Tab  ↑↓ Srv  [Space] Connect  [T] TUN  [R] Refresh  [Enter] Details  [A/D] Sub  [Q] Quit")
	}

	return tabBar + "\n" + mainRow + "\n" + bottom + "\n" + statusLine + "\n" + helpLine
}

func (a *App) renderStatus() string {
	mode := a.status.Mode
	if mode == "" {
		mode = "socks"
	}
	running := "○ stopped"
	if a.status.Running {
		running = "● running"
	}
	server := "—"
	if a.status.ServerIdx >= 0 {
		if s := a.tabs.ServerAtAbsIdx(a.status.ServerIdx); s != nil {
			server = s.Remarks
		}
	}
	line := fmt.Sprintf("Mode: %s  Server: %s  Status: %s", mode, server, running)
	if a.errMsg != "" {
		line += "  " + styles.ErrorStyle.Render("Error: "+a.errMsg)
	}
	return styles.StatusBar.Width(a.width).Render(line)
}

func tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinTickMsg{}
	})
}

func recvIPC(client *ipc.Client) tea.Cmd {
	return func() tea.Msg {
		env, err := client.Recv()
		if err != nil {
			return ipcErrMsg{err}
		}
		return ipcMsg{env}
	}
}

func sendCmd(client *ipc.Client, v any) tea.Cmd {
	return func() tea.Msg {
		client.Send(v)
		return nil
	}
}
