package xray_test

import (
	"encoding/json"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/xray"
)

func TestServerHost(t *testing.T) {
	tests := []struct {
		name string
		cfg  string
		want string
	}{
		{
			name: "vless vnext",
			cfg:  `{"outbounds":[{"protocol":"vless","settings":{"vnext":[{"address":"v.example.com"}]}}]}`,
			want: "v.example.com",
		},
		{
			name: "shadowsocks servers",
			cfg:  `{"outbounds":[{"protocol":"shadowsocks","settings":{"servers":[{"address":"s.example.com"}]}}]}`,
			want: "s.example.com",
		},
		{
			name: "hysteria address",
			cfg:  `{"outbounds":[{"protocol":"hysteria","settings":{"address":"h.example.com","port":443}}]}`,
			want: "h.example.com",
		},
		{
			name: "skips empty freedom outbound",
			cfg:  `{"outbounds":[{"protocol":"hysteria","settings":{"address":"h.example.com"}},{"protocol":"freedom"}]}`,
			want: "h.example.com",
		},
		{
			name: "no outbounds",
			cfg:  `{"outbounds":[]}`,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xray.ServerHost(json.RawMessage(tt.cfg))
			if got != tt.want {
				t.Errorf("ServerHost() = %q, want %q", got, tt.want)
			}
		})
	}
}
