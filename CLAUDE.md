# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**nwxraytui** — a TUI application for managing xray proxy subscriptions and TUN connections.

- Go module: `github.com/mrsedan/nwxraytui`
- Go 1.26+
- Dev environment managed via Nix flakes

## Development Environment

```sh
nix develop
```

The shell sets `GOPATH=$PWD/.gopath` automatically and provides Go, gopls, golangci-lint, delve, xray, just.

## Build & Test

```sh
go build ./...
go test ./...
go test ./internal/pkg -run TestName
golangci-lint run
```

## Architecture

Two-process model: a background **daemon** and a **TUI** frontend connected via a Unix domain socket.

**Entry point** (`cmd/nwxraytui/main.go`): starts as daemon with `--daemon`, otherwise starts the TUI and auto-spawns the daemon if not already running.

**Daemon** (`internal/daemon`): manages the xray subprocess, subscription refresh, latency pings, and all connected TUI clients. On connect, sends current status; broadcasts events (status, server list, latency, logs) to all connected clients via newline-delimited JSON.

**IPC** (`internal/ipc`): newline-delimited JSON envelopes with a `type` field (Go struct name) and `payload`. Commands flow TUI→daemon; events flow daemon→TUI. All message types are in `messages.go`.

**Subscription** (`internal/subscription`): fetches subscription URLs; parses either a base64-encoded newline-separated URI list or a JSON array of xray configs. Servers are cached at `~/.cache/nwxraytui/servers.json`.

**xray** (`internal/xray`): `Runner` manages the xray subprocess lifecycle and captures stdout/stderr into a 500-line ring buffer streamed via `LogCh`. `merge.go` injects inbound proxy/TUN config into a server's outbound config JSON before writing to `~/.cache/nwxraytui/active-config.json`.

**TUI** (`internal/tui`): Bubbletea app with three panels — server list (left), server details (right), log output (bottom). IPC events drive model updates via `tea.Cmd` wrappers.

**Config** (`internal/config`): TOML at `~/.config/nwxraytui/config.toml`. Socket path is `/run/user/$UID/nwxraytui/daemon.sock` on Linux, `$TMPDIR/nwxraytui/daemon.sock` on macOS.

**Proxy** (`internal/proxy`): sets/unsets system proxy env vars and checks TUN capability. **Latency** (`internal/latency`): TCP dial to measure server ping.

## Rules

Don't add md files to git. Ask if you want to create it.

Write briefly, clearly, without elaborating on details unless I ask for them.

No coauthor commits, just commit without Co-authored by.
