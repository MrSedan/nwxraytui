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

type tunSettings struct {
	Name        string   `json:"name"`
	Address     []string `json:"address"`
	MTU         int      `json:"mtu"`
	AutoRoute   bool     `json:"autoRoute"`
	StrictRoute bool     `json:"strictRoute"`
}

func buildTunInbound() (json.RawMessage, error) {
	spec := tunInboundSpec{
		Tag:      "tun",
		Protocol: "tun",
		Settings: tunSettings{
			Name:        "tun0",
			Address:     []string{"198.18.0.1/30"},
			MTU:         1500,
			AutoRoute:   true,
			StrictRoute: false,
		},
	}
	switch runtime.GOOS {
	case "darwin":
		spec.Settings.Name = "utun9"
		spec.Settings.AutoRoute = false
	case "linux":
		// daemon manages routes manually via ip(8).
		spec.Settings.AutoRoute = false
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
