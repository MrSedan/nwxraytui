package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrsedan/nwxraytui/internal/ipc"
	"github.com/mrsedan/nwxraytui/internal/tui/panels"
	"github.com/mrsedan/nwxraytui/internal/tui/styles"
)

type ipcMsg struct{ env ipc.Envelope }
type ipcErrMsg struct{ err error }

type App struct {
	client     *ipc.Client
	serverList panels.ServerList
	detail     panels.DetailPanel
	logPanel   panels.LogPanel
	status     ipc.EventStatus
	width      int
	height     int
	errMsg     string
}

func New(client *ipc.Client) *App {
	return &App{
		client:     client,
		serverList: panels.NewServerList(),
	}
}

func (a *App) Start() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		recvIPC(a.client),
		sendCmd(a.client, ipc.CmdRefresh{}),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case ipcMsg:
		cmds = append(cmds, recvIPC(a.client))
		a.handleIPC(msg.env)

	case ipcErrMsg:
		a.errMsg = msg.err.Error()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case " ":
			if len(a.serverList.Servers) > 0 {
				cmds = append(cmds, sendCmd(a.client, ipc.CmdStart{ServerIdx: a.serverList.SelectedIdx(), Mode: a.status.Mode}))
			}
		case "s":
			cmds = append(cmds, sendCmd(a.client, ipc.CmdSwitch{ServerIdx: a.serverList.SelectedIdx(), Mode: "socks"}))
		case "t":
			if a.status.TunAvailable {
				cmds = append(cmds, sendCmd(a.client, ipc.CmdSwitch{ServerIdx: a.serverList.SelectedIdx(), Mode: "tun"}))
			}
		case "p":
			cmds = append(cmds, sendCmd(a.client, ipc.CmdSwitch{ServerIdx: a.serverList.SelectedIdx(), Mode: "system"}))
		case "r":
			cmds = append(cmds, sendCmd(a.client, ipc.CmdRefresh{}))
		}
	}

	var sl panels.ServerList
	sl, _ = a.serverList.Update(msg)
	a.serverList = sl

	if idx := a.serverList.SelectedIdx(); idx >= 0 && idx < len(a.serverList.Servers) {
		s := a.serverList.Servers[idx]
		a.detail.Server = &s
	}

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
	case ipc.TypeEventServerList:
		ev, _ := ipc.UnmarshalPayload[ipc.EventServerList](env)
		a.serverList.Servers = ev.Servers
	case ipc.TypeEventLatency:
		ev, _ := ipc.UnmarshalPayload[ipc.EventLatency](env)
		a.serverList.Latencies[ev.ServerIdx] = ev.Ms
	case ipc.TypeEventLog:
		ev, _ := ipc.UnmarshalPayload[ipc.EventLog](env)
		a.logPanel.Push(ev.Line)
	}
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	leftW := a.width / 3
	rightW := a.width - leftW
	topH := (a.height - 3) * 2 / 3
	botH := a.height - topH - 3

	left := a.serverList.View(leftW, topH)
	right := a.detail.View(rightW, topH)
	top := left + right
	bottom := a.logPanel.View(a.width, botH)

	statusLine := a.renderStatus()
	helpLine := styles.DimText.Render("[R] Refresh  [A] Add sub  [D] Del sub  [space] Connect  [Q] Quit")

	return top + "\n" + bottom + "\n" + statusLine + "\n" + helpLine
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
	if a.status.ServerIdx >= 0 && a.status.ServerIdx < len(a.serverList.Servers) {
		server = a.serverList.Servers[a.status.ServerIdx].Remarks
	}
	line := fmt.Sprintf("Mode: %s  Server: %s  Status: %s", mode, server, running)
	if a.errMsg != "" {
		line += "  " + styles.ErrorStyle.Render("Error: "+a.errMsg)
	}
	return styles.StatusBar.Width(a.width).Render(line)
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
