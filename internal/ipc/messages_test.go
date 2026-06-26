package ipc_test

import (
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
