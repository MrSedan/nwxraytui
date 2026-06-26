package xray

import "encoding/json"

// ServerHost returns the remote host from the first outbound in cfg.
func ServerHost(cfg json.RawMessage) string {
	var parsed struct {
		Outbounds []struct {
			Settings struct {
				Vnext []struct {
					Address string `json:"address"`
				} `json:"vnext"`
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
		if len(ob.Settings.Vnext) > 0 {
			return ob.Settings.Vnext[0].Address
		}
		if len(ob.Settings.Servers) > 0 {
			return ob.Settings.Servers[0].Address
		}
	}
	return ""
}
