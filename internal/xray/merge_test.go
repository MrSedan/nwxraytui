package xray_test

import (
	"encoding/json"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/xray"
)

var baseConfig = json.RawMessage(`{
	"remarks": "Test",
	"inbounds": [
		{"tag":"socks","port":10808,"protocol":"socks"},
		{"tag":"http","port":10809,"protocol":"http"}
	],
	"outbounds": [{"tag":"proxy","protocol":"vless"}],
	"routing": {}
}`)

func TestMerge_SocksMode_NoChange(t *testing.T) {
	out, err := xray.MergeConfig(baseConfig, "socks")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Inbounds []struct {
			Tag string `json:"tag"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Inbounds) != 2 {
		t.Fatalf("socks mode: want 2 inbounds, got %d", len(cfg.Inbounds))
	}
}

func TestMerge_TUNMode_InjectsInbound(t *testing.T) {
	out, err := xray.MergeConfig(baseConfig, "tun")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Inbounds []struct {
			Tag      string `json:"tag"`
			Protocol string `json:"protocol"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	hasTun := false
	for _, ib := range cfg.Inbounds {
		if ib.Tag == "tun" {
			hasTun = true
		}
	}
	if !hasTun {
		t.Fatal("tun mode: expected tun inbound, not found")
	}
	if len(cfg.Inbounds) != 3 {
		t.Fatalf("tun mode: want 3 inbounds (socks+http+tun), got %d", len(cfg.Inbounds))
	}
}

func TestMerge_TUNMode_PreservesExisting(t *testing.T) {
	out, err := xray.MergeConfig(baseConfig, "tun")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Inbounds []struct {
			Tag string `json:"tag"`
		} `json:"inbounds"`
	}
	json.Unmarshal(out, &cfg)
	tags := map[string]bool{}
	for _, ib := range cfg.Inbounds {
		tags[ib.Tag] = true
	}
	if !tags["socks"] || !tags["http"] {
		t.Fatal("tun mode: original inbounds not preserved")
	}
}
