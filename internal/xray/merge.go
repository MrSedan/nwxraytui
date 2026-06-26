package xray

import (
	"encoding/json"
	"fmt"
)

var tunInbound = json.RawMessage(`{
	"tag": "tun",
	"protocol": "tun",
	"settings": {
		"name": "tun0",
		"address": ["198.18.0.1/30"],
		"mtu": 9000,
		"autoRoute": true,
		"strictRoute": false
	}
}`)

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
	filtered = append(filtered, tunInbound)

	newInbounds, err := json.Marshal(filtered)
	if err != nil {
		return nil, err
	}
	cfg["inbounds"] = newInbounds

	return json.Marshal(cfg)
}
