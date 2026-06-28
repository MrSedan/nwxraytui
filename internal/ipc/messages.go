package ipc

import (
	"encoding/json"
	"reflect"
)

type MsgType string

const (
	TypeCmdStart        MsgType = "CmdStart"
	TypeCmdStop         MsgType = "CmdStop"
	TypeCmdSwitch       MsgType = "CmdSwitch"
	TypeCmdRefresh      MsgType = "CmdRefresh"
	TypeCmdSetAutostart MsgType = "CmdSetAutostart"
	TypeCmdPing         MsgType = "CmdPing"
	TypeCmdAddSub       MsgType = "CmdAddSub"
	TypeCmdRemoveSub    MsgType = "CmdRemoveSub"
	TypeEventStatus          MsgType = "EventStatus"
	TypeEventServerList      MsgType = "EventServerList"
	TypeEventLatency         MsgType = "EventLatency"
	TypeEventLog             MsgType = "EventLog"
	TypeEventSubscriptionList MsgType = "EventSubscriptionList"
)

type Envelope struct {
	Type    MsgType         `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CmdStart struct {
	ServerIdx int    `json:"serverIdx"`
	Mode      string `json:"mode"`
}

type CmdStop struct{}

type CmdSwitch struct {
	ServerIdx int    `json:"serverIdx"`
	Mode      string `json:"mode"`
}

type CmdRefresh struct{}

type CmdPing struct{}

type CmdSetAutostart struct {
	Enabled bool `json:"enabled"`
}

type CmdAddSub struct {
	URL string `json:"url"`
}

type CmdRemoveSub struct {
	URL string `json:"url"`
}

type EventStatus struct {
	Running      bool   `json:"running"`
	ServerIdx    int    `json:"serverIdx"`
	Mode         string `json:"mode"`
	TunAvailable bool   `json:"tunAvailable"`
	Error        string `json:"error,omitempty"`
}

type ServerInfo struct {
	Remarks string          `json:"remarks"`
	Config  json.RawMessage `json:"config"`
}

type EventServerList struct {
	Servers []ServerInfo `json:"servers"`
}

type EventLatency struct {
	ServerIdx int `json:"serverIdx"`
	Ms        int `json:"ms"`
}

type EventLog struct {
	Line string `json:"line"`
}

type SubscriptionMeta struct {
	Title          string `json:"title"`
	Announce       string `json:"announce,omitempty"`
	Upload         int64  `json:"upload"`
	Download       int64  `json:"download"`
	Total          int64  `json:"total"`
	Expire         int64  `json:"expire"`
	UpdateInterval int    `json:"updateInterval"`
}

type SubscriptionGroup struct {
	URL     string           `json:"url"`
	Meta    SubscriptionMeta `json:"meta"`
	Servers []ServerInfo     `json:"servers"`
}

type EventSubscriptionList struct {
	Groups []SubscriptionGroup `json:"groups"`
}

func Encode(v any) ([]byte, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: MsgType(t.Name()), Payload: payload})
}

func Decode(data []byte) (Envelope, error) {
	var env Envelope
	return env, json.Unmarshal(data, &env)
}

func UnmarshalPayload[T any](env Envelope) (T, error) {
	var v T
	return v, json.Unmarshal(env.Payload, &v)
}
