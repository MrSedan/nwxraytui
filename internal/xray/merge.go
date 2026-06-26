package xray

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// TunFwmark is the SO_MARK value applied to sockets created by xray's direct
// (freedom) outbound when TUN mode is active. A matching ip rule routes those
// packets via the real interface instead of looping back into tun0.
const TunFwmark = 255

type tunSniffingSpec struct {
	Enabled      bool     `json:"enabled"`
	RouteOnly    bool     `json:"routeOnly"`
	DestOverride []string `json:"destOverride"`
}

type tunInboundSpec struct {
	Tag      string         `json:"tag"`
	Protocol string         `json:"protocol"`
	Settings tunSettings    `json:"settings"`
	Sniffing tunSniffingSpec `json:"sniffing"`
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
		Sniffing: tunSniffingSpec{
			Enabled:      true,
			RouteOnly:    true,
			DestOverride: []string{"http", "tls", "quic"},
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

	// Mark all freedom outbounds so the kernel can route their sockets via the
	// real interface (via ip rule fwmark) instead of looping back into tun0.
	var outbounds []json.RawMessage
	if raw, ok := cfg["outbounds"]; ok {
		json.Unmarshal(raw, &outbounds)
	}
	for i, ob := range outbounds {
		outbounds[i] = markFreedomOutbound(ob)
	}
	if newOutbounds, err := json.Marshal(outbounds); err == nil {
		cfg["outbounds"] = newOutbounds
	}

	return json.Marshal(cfg)
}

// markFreedomOutbound sets sockopt.mark = TunFwmark on freedom-protocol
// outbounds so Linux policy routing can bypass the TUN interface for direct
// connections. Non-freedom outbounds are returned unchanged.
func markFreedomOutbound(ob json.RawMessage) json.RawMessage {
	var meta struct {
		Protocol string `json:"protocol"`
	}
	if json.Unmarshal(ob, &meta) != nil || meta.Protocol != "freedom" {
		return ob
	}

	var m map[string]json.RawMessage
	if json.Unmarshal(ob, &m) != nil {
		return ob
	}

	var ss map[string]json.RawMessage
	if raw, ok := m["streamSettings"]; ok {
		json.Unmarshal(raw, &ss)
	}
	if ss == nil {
		ss = make(map[string]json.RawMessage)
	}

	var sockopt map[string]json.RawMessage
	if raw, ok := ss["sockopt"]; ok {
		json.Unmarshal(raw, &sockopt)
	}
	if sockopt == nil {
		sockopt = make(map[string]json.RawMessage)
	}

	sockopt["mark"] = json.RawMessage(`255`)
	ssRaw, _ := json.Marshal(sockopt)
	ss["sockopt"] = ssRaw
	mRaw, _ := json.Marshal(ss)
	m["streamSettings"] = mRaw

	result, _ := json.Marshal(m)
	return result
}
