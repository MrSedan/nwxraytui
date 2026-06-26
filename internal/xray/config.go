package xray

import (
	"encoding/json"
	"net"
	"strings"
)

// ServerHost returns the remote host from the first outbound in cfg.
func ServerHost(cfg json.RawMessage) string {
	var parsed struct {
		Outbounds []struct {
			Settings struct {
				// hysteria, wireguard, socks, http and similar keep the
				// remote host directly on settings.
				Address string `json:"address"`
				// vless / vmess
				Vnext []struct {
					Address string `json:"address"`
				} `json:"vnext"`
				// shadowsocks / trojan
				Servers []struct {
					Address string `json:"address"`
				} `json:"servers"`
			} `json:"settings"`
		} `json:"outbounds"`
	}
	if err := json.Unmarshal(cfg, &parsed); err != nil {
		return ""
	}
	for _, ob := range parsed.Outbounds {
		if ob.Settings.Address != "" {
			return ob.Settings.Address
		}
		if len(ob.Settings.Vnext) > 0 && ob.Settings.Vnext[0].Address != "" {
			return ob.Settings.Vnext[0].Address
		}
		if len(ob.Settings.Servers) > 0 && ob.Settings.Servers[0].Address != "" {
			return ob.Settings.Servers[0].Address
		}
	}
	return ""
}

// DNSServerIP returns the first plain public IP address found in cfg's
// dns.servers list. DoH URLs, domain names, "fakedns" and private/loopback
// IPs are skipped. Returns "" if none found.
func DNSServerIP(cfg json.RawMessage) string {
	var parsed struct {
		DNS struct {
			Servers []json.RawMessage `json:"servers"`
		} `json:"dns"`
	}
	if json.Unmarshal(cfg, &parsed) != nil {
		return ""
	}
	for _, raw := range parsed.DNS.Servers {
		// Each entry is either a plain string or an object with "address".
		var addr string
		if json.Unmarshal(raw, &addr) != nil {
			var obj struct {
				Address string `json:"address"`
			}
			if json.Unmarshal(raw, &obj) != nil {
				continue
			}
			addr = obj.Address
		}
		if addr == "" || strings.HasPrefix(addr, "https://") || addr == "fakedns" {
			continue
		}
		ip := net.ParseIP(addr)
		if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			continue
		}
		return addr
	}
	return ""
}
