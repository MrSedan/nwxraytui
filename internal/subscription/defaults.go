package subscription

import "encoding/json"

func DefaultConfig(outbound json.RawMessage, remarks string) json.RawMessage {
	cfg := map[string]any{
		"remarks": remarks,
		"inbounds": []any{
			map[string]any{
				"tag": "socks", "port": 10808, "listen": "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]any{"udp": true, "auth": "noauth"},
				"sniffing": map[string]any{
					"enabled": true, "routeOnly": true,
					"destOverride": []string{"http", "tls", "quic"},
				},
			},
			map[string]any{
				"tag": "http", "port": 10809, "listen": "127.0.0.1",
				"protocol": "http",
				"settings": map[string]any{"allowTransparent": false},
				"sniffing": map[string]any{
					"enabled": true, "routeOnly": true,
					"destOverride": []string{"http", "tls", "quic"},
				},
			},
		},
		"outbounds": []json.RawMessage{
			outbound,
			json.RawMessage(`{"tag":"direct","protocol":"freedom","streamSettings":{"sockopt":{"domainStrategy":"ForceIPv4"}}}`),
			json.RawMessage(`{"tag":"block","protocol":"blackhole","settings":{"response":{"type":"http"}}}`),
		},
		"routing": map[string]any{
			"domainStrategy": "IPIfNonMatch",
			"rules": []any{
				map[string]any{"type": "field", "ip": []string{"geoip:private"}, "outboundTag": "direct"},
				map[string]any{"type": "field", "network": "tcp,udp", "outboundTag": "proxy"},
			},
		},
		"dns": map[string]any{
			"servers":       []string{"https://cloudflare-dns.com/dns-query", "1.1.1.1"},
			"queryStrategy": "UseIPv4",
		},
	}
	data, _ := json.Marshal(cfg)
	return data
}
