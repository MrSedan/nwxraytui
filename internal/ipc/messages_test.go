package ipc_test

import (
	"encoding/json"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/ipc"
)

func TestEncodeDecode_CmdStart(t *testing.T) {
	cmd := ipc.CmdStart{ServerIdx: 2, Mode: "tun"}
	data, err := ipc.Encode(cmd)
	if err != nil {
		t.Fatal(err)
	}
	env, err := ipc.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != ipc.TypeCmdStart {
		t.Fatalf("type: got %q want %q", env.Type, ipc.TypeCmdStart)
	}
	got, err := ipc.UnmarshalPayload[ipc.CmdStart](env)
	if err != nil {
		t.Fatal(err)
	}
	if got.ServerIdx != 2 || got.Mode != "tun" {
		t.Fatalf("payload: got %+v", got)
	}
}

func TestEncodeDecode_EventStatus(t *testing.T) {
	ev := ipc.EventStatus{Running: true, ServerIdx: 1, Mode: "socks", TunAvailable: false}
	data, err := ipc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	env, err := ipc.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ipc.UnmarshalPayload[ipc.EventStatus](env)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Running || got.Mode != "socks" || got.ServerIdx != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestEventSubscriptionList_RoundTrip(t *testing.T) {
	ev := ipc.EventSubscriptionList{
		Groups: []ipc.SubscriptionGroup{
			{
				URL: "https://example.com/sub",
				Meta: ipc.SubscriptionMeta{
					Title:          "VPN Pro",
					Upload:         1024,
					Download:       2048,
					Total:          10737418240,
					Expire:         1785600000,
					UpdateInterval: 24,
				},
				Servers: []ipc.ServerInfo{
					{Remarks: "Server A", Config: json.RawMessage(`{}`)},
				},
			},
		},
	}
	data, err := ipc.Encode(ev)
	if err != nil {
		t.Fatal(err)
	}
	env, err := ipc.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != ipc.TypeEventSubscriptionList {
		t.Fatalf("got type %q, want %q", env.Type, ipc.TypeEventSubscriptionList)
	}
	got, err := ipc.UnmarshalPayload[ipc.EventSubscriptionList](env)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Groups) != 1 || got.Groups[0].Meta.Title != "VPN Pro" {
		t.Fatalf("unexpected: %+v", got.Groups)
	}
	if got.Groups[0].Meta.Total != 10737418240 {
		t.Fatalf("total mismatch: got %d", got.Groups[0].Meta.Total)
	}
}
