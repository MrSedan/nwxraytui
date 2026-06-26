package xray

import (
	"encoding/json"
	"fmt"
	"runtime"
)

type tunInboundSpec struct {
	Tag      string      `json:"tag"`
	Protocol string      `json:"protocol"`
	Settings tunSettings `json:"settings"`
}

// tunSettings mirrors xray-core's proxy/tun Config. That inbound only brings
// the interface up; it deliberately does NOT assign addresses or configure
// routing/rules (see its README). Addressing and routing are the OS's job and
// are handled by the daemon via proxy/routes_*.go.
type tunSettings struct {
	Name string `json:"name"`
	MTU  int    `json:"MTU"`
}

func buildTunInbound() (json.RawMessage, error) {
	name := "tun0"
	if runtime.GOOS == "darwin" {
		name = "utun9"
	}
	spec := tunInboundSpec{
		Tag:      "tun",
		Protocol: "tun",
		Settings: tunSettings{
			Name: name,
			MTU:  1500,
		},
	}
	return json.Marshal(spec)
}

func MergeConfig(serverConfig json.RawMessage, mode string) (json.RawMessage, error) {
	if mode != "tun" {
		return serverConfig, nil
	}

	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(serverConfig, &cfg); err != nil {
		return nil, fmt.Errorf("MergeConfig: unmarshal: %w", err)
	}

	var inbounds []json.RawMessage
	if raw, ok := cfg["inbounds"]; ok {
		if err := json.Unmarshal(raw, &inbounds); err != nil {
			return nil, fmt.Errorf("MergeConfig: parse inbounds: %w", err)
		}
	}

	filtered := inbounds[:0]
	for _, ib := range inbounds {
		var meta struct {
			Tag string `json:"tag"`
		}
		json.Unmarshal(ib, &meta)
		if meta.Tag != "tun" {
			filtered = append(filtered, ib)
		}
	}

	tun, err := buildTunInbound()
	if err != nil {
		return nil, fmt.Errorf("MergeConfig: build tun inbound: %w", err)
	}
	filtered = append(filtered, tun)

	newInbounds, err := json.Marshal(filtered)
	if err != nil {
		return nil, err
	}
	cfg["inbounds"] = newInbounds

	return json.Marshal(cfg)
}
