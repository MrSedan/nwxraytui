package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func ParseURI(uri string) (Server, error) {
	switch {
	case strings.HasPrefix(uri, "vless://"):
		return parseVless(uri)
	case strings.HasPrefix(uri, "vmess://"):
		return parseVmess(uri)
	case strings.HasPrefix(uri, "trojan://"):
		return parseTrojan(uri)
	case strings.HasPrefix(uri, "ss://"):
		return parseSS(uri)
	default:
		return Server{}, fmt.Errorf("unsupported URI scheme: %s", uri)
	}
}

func parseVless(uri string) (Server, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return Server{}, err
	}
	remarks, _ := url.PathUnescape(u.Fragment)
	if remarks == "" {
		remarks = u.Host
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return Server{}, fmt.Errorf("vless: %w", err)
	}
	port, _ := strconv.Atoi(portStr)
	q := u.Query()

	outbound, err := json.Marshal(map[string]any{
		"tag":      "proxy",
		"protocol": "vless",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": host,
				"port":    port,
				"users": []any{map[string]any{
					"id":         u.User.Username(),
					"encryption": q.Get("encryption"),
					"flow":       q.Get("flow"),
				}},
			}},
		},
		"streamSettings": buildStream(q.Get("type"), q.Get("host"), q.Get("path"),
			q.Get("security"), q.Get("sni"), q.Get("fp"), q.Get("alpn")),
	})
	if err != nil {
		return Server{}, err
	}
	return Server{Remarks: remarks, Config: DefaultConfig(outbound, remarks)}, nil
}

func parseVmess(uri string) (Server, error) {
	encoded := strings.TrimPrefix(uri, "vmess://")
	if i := strings.Index(encoded, "#"); i >= 0 {
		encoded = encoded[:i]
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return Server{}, fmt.Errorf("vmess: base64: %w", err)
		}
	}
	var v struct {
		PS   string      `json:"ps"`
		Add  string      `json:"add"`
		Port json.Number `json:"port"`
		ID   string      `json:"id"`
		Aid  json.Number `json:"aid"`
		Scy  string      `json:"scy"`
		Net  string      `json:"net"`
		Host string      `json:"host"`
		Path string      `json:"path"`
		TLS  string      `json:"tls"`
		SNI  string      `json:"sni"`
		FP   string      `json:"fp"`
		ALPN string      `json:"alpn"`
	}
	if err := json.Unmarshal(decoded, &v); err != nil {
		return Server{}, fmt.Errorf("vmess: %w", err)
	}
	port, _ := v.Port.Int64()
	aid, _ := v.Aid.Int64()
	scy := v.Scy
	if scy == "" {
		scy = "auto"
	}
	sec := ""
	if v.TLS == "tls" {
		sec = "tls"
	}
	outbound, err := json.Marshal(map[string]any{
		"tag":      "proxy",
		"protocol": "vmess",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": v.Add,
				"port":    port,
				"users": []any{map[string]any{
					"id": v.ID, "alterId": aid, "security": scy,
				}},
			}},
		},
		"streamSettings": buildStream(v.Net, v.Host, v.Path, sec, v.SNI, v.FP, v.ALPN),
	})
	if err != nil {
		return Server{}, err
	}
	remarks := v.PS
	if remarks == "" {
		remarks = v.Add
	}
	return Server{Remarks: remarks, Config: DefaultConfig(outbound, remarks)}, nil
}

func parseTrojan(uri string) (Server, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return Server{}, err
	}
	remarks, _ := url.PathUnescape(u.Fragment)
	if remarks == "" {
		remarks = u.Host
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return Server{}, fmt.Errorf("trojan: %w", err)
	}
	port, _ := strconv.Atoi(portStr)
	q := u.Query()
	sni := q.Get("sni")
	if sni == "" {
		sni = host
	}
	outbound, err := json.Marshal(map[string]any{
		"tag":      "proxy",
		"protocol": "trojan",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address": host, "port": port, "password": u.User.Username(),
			}},
		},
		"streamSettings": buildStream(q.Get("type"), q.Get("host"), q.Get("path"),
			"tls", sni, q.Get("fp"), ""),
	})
	if err != nil {
		return Server{}, err
	}
	return Server{Remarks: remarks, Config: DefaultConfig(outbound, remarks)}, nil
}

func parseSS(uri string) (Server, error) {
	rest := strings.TrimPrefix(uri, "ss://")
	remarks := ""
	if i := strings.Index(rest, "#"); i >= 0 {
		remarks, _ = url.PathUnescape(rest[i+1:])
		rest = rest[:i]
	}
	// SIP002: base64(method:pass)@host:port
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		userInfo := rest[:at]
		hostPort := rest[at+1:]
		dec, err := tryBase64(userInfo)
		if err == nil {
			parts := strings.SplitN(string(dec), ":", 2)
			if len(parts) == 2 {
				host, portStr, err := net.SplitHostPort(hostPort)
				if err == nil {
					port, _ := strconv.Atoi(portStr)
					return buildSSServer(parts[0], parts[1], host, port, remarks)
				}
			}
		}
	}
	// Legacy: base64(method:pass@host:port)
	dec, err := tryBase64(rest)
	if err != nil {
		return Server{}, fmt.Errorf("ss: cannot decode: %w", err)
	}
	at := strings.LastIndex(string(dec), "@")
	if at < 0 {
		return Server{}, fmt.Errorf("ss: invalid format")
	}
	parts := strings.SplitN(string(dec)[:at], ":", 2)
	if len(parts) != 2 {
		return Server{}, fmt.Errorf("ss: invalid method:pass")
	}
	host, portStr, err := net.SplitHostPort(string(dec)[at+1:])
	if err != nil {
		return Server{}, err
	}
	port, _ := strconv.Atoi(portStr)
	return buildSSServer(parts[0], parts[1], host, port, remarks)
}

func buildSSServer(method, password, host string, port int, remarks string) (Server, error) {
	outbound, err := json.Marshal(map[string]any{
		"tag":      "proxy",
		"protocol": "shadowsocks",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address": host, "port": port, "method": method, "password": password,
			}},
		},
	})
	if err != nil {
		return Server{}, err
	}
	return Server{Remarks: remarks, Config: DefaultConfig(outbound, remarks)}, nil
}

func buildStream(network, host, path, security, sni, fp, alpn string) map[string]any {
	if network == "" {
		network = "tcp"
	}
	ss := map[string]any{"network": network}
	if security == "tls" {
		tls := map[string]any{"serverName": sni}
		if fp != "" {
			tls["fingerprint"] = fp
		}
		if alpn != "" {
			tls["alpn"] = strings.Split(alpn, ",")
		}
		ss["security"] = "tls"
		ss["tlsSettings"] = tls
	}
	switch network {
	case "ws":
		ss["wsSettings"] = map[string]any{"path": path, "host": host}
	case "grpc":
		ss["grpcSettings"] = map[string]any{"serviceName": path}
	case "h2", "http":
		ss["httpSettings"] = map[string]any{"path": path, "host": []string{host}}
	case "xhttp":
		ss["xhttpSettings"] = map[string]any{"path": path, "host": host}
	}
	return ss
}

func tryBase64(s string) ([]byte, error) {
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}
