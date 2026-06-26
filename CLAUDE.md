# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**nixxray** — a TUI application for managing xray proxy subscriptions and TUN connections.

- Go module: `github.com/mrsedan/nixxray`
- Go 1.26+
- Dev environment managed via Nix flakes

## Development Environment

Enter the dev shell (provides Go, gopls, golangci-lint, delve, xray, just):

```sh
nix develop
```

The shell sets `GOPATH=$PWD/.gopath` automatically.

## Build & Test

```sh
go build ./...
go test ./...
go test ./path/to/pkg -run TestName   # single test
golangci-lint run                      # lint
```

## Architecture

No source code exists yet. The planned shape from the flake:

- **TUI** — interactive terminal interface for subscription management
- **xray integration** — delegates proxy/TUN work to the `xray` binary
- **TUN support** — Linux needs `iproute2`, `iptables`/`nftables`; macOS uses `utun` natively

When Go packages are added, prefer placing CLI entry point at `cmd/nixxray/main.go` and library code under `internal/`.
