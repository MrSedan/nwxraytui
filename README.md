# NWXrayTUI - TUI for Xray

> **Disclaimer:** This is a vibe-coded application built for a specific personal setup. It may not work perfectly (or at all) for your configuration. Use at your own risk.

A terminal UI for managing [xray-core](https://github.com/XTLS/Xray-core) proxy subscriptions and TUN connections.

```
┌─[Sub1]──[Sub2]──────────────────────────┐ ┌─ Info ──────────────────────────────────┐
│ > server-hk-01          12 ms           │ │ Status:  ● running (tun)               │
│   server-sg-02          45 ms           │ │ Server:  server-hk-01                  │
│   server-us-03         120 ms           │ │ Mode:    tun                            │
│   server-de-04         210 ms           │ │ Ping:    12 ms                         │
│                                         │ │                                        │
└─────────────────────────────────────────┘ └────────────────────────────────────────┘
┌─ Log ──────────────────────────────────────────────────────────────────────────────────┐
│ 2026/01/01 12:00:00 [Info] Xray 24.x.x started                                       │
│ 2026/01/01 12:00:01 [Info] [tun] tun0 started                                        │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

## Features

- Subscription management (fetch, refresh, add/remove URLs)
- Per-server latency pings
- Proxy mode (SOCKS5 / HTTP) and TUN mode switching
- Daemon auto-starts in the background; TUI reconnects automatically
- State persistence (last selected server survives restart)

## Configuration

Config lives at `~/.config/nwxraytui/config.toml`. Subscription URLs are managed via the TUI (`a` to add, `d` to remove) — no need to edit the file manually.

```toml
[subscriptions]
urls = [
  "https://example.com/sub1",
  "https://example.com/sub2",
]
refresh_interval = "1h"

[proxy]
socks_port = 10808
http_port  = 10809
mode       = "socks"   # "socks" or "http"

[daemon]
autostart = true
```

## Keybindings

| Key         | Action                                      |
|-------------|---------------------------------------------|
| `↑`/`↓`     | Navigate servers                            |
| `←`/`→`     | Switch subscription tabs                    |
| `Space`     | Connect to selected server (or switch)      |
| `s`         | Stop xray                                   |
| `t`         | Toggle TUN mode (requires CAP_NET_ADMIN)    |
| `p`         | Ping all servers                            |
| `r`         | Refresh subscriptions                       |
| `Enter`     | Show server details                         |
| `Esc`       | Close details                               |
| `a`         | Add subscription URL                        |
| `d`         | Remove subscription URL                     |
| `q`/`Ctrl+C`| Quit TUI (daemon keeps running)            |

## Installation

### NixOS — system service (recommended)

Add to your `flake.nix` inputs:

```nix
inputs = {
  nwxraytui = {
    url = "github:mrsedan/nwxraytui";
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```

Pass the module to your system configuration:

```nix
nixosSystem {
  modules = [
    nwxraytui.nixosModules.nwxraytui
    # ...
  ];
};
```

Enable the service (e.g. in a dedicated module):

```nix
{ inputs, ... }: {
  services.nwxraytui = {
    enable    = true;
    enableTun = true;   # grants CAP_NET_ADMIN for TUN mode
    user      = "youruser";
    package   = inputs.nwxraytui.packages.${pkgs.system}.default;
  };

  environment.systemPackages = [
    inputs.nwxraytui.packages.${pkgs.system}.default
  ];
}
```

`enableTun = true` also configures the firewall (`trustedInterfaces = ["tun0"]`), relaxes reverse-path filtering, enables `systemd-resolved`, and adds a polkit rule so the daemon can set DNS through the tunnel without sudo.

### Home Manager — user profile only

If you only want the binary in your profile without a system service:

```nix
{ inputs, pkgs, ... }: {
  home.packages = [
    inputs.nwxraytui.packages.${pkgs.system}.default
  ];
}
```

The TUI will auto-spawn the daemon as a background process on first launch. TUN mode will not be available without the capability granted by the system service.

### macOS (nix-darwin)

```nix
inputs.nixpkgs.follows = "nixpkgs";

darwinSystem {
  modules = [
    nwxraytui.darwinModules.nwxraytui
    {
      services.nwxraytui = {
        enable    = true;
        enableTun = true;  # runs daemon as root for utun access
      };
    }
  ];
};
```

### Other systems

The binary should work on any Linux or macOS system with `xray` in `PATH`. Build from source:

```sh
git clone https://github.com/mrsedan/nwxraytui
cd nwxraytui
go build -o nwxraytui ./cmd/nwxraytui
```

TUN mode requires `CAP_NET_ADMIN` (Linux) or root (macOS). Proxy-only mode works without elevated privileges.

## Running

```sh
nwxraytui          # start TUI (daemon auto-starts if not running)
nwxraytui --daemon # start daemon only
```
