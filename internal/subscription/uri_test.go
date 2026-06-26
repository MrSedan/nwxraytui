package subscription_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/subscription"
)

func TestParseURI_Vless(t *testing.T) {
	uri := "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443" +
		"?security=tls&sni=example.com&type=tcp&encryption=none#TestServer"
	srv, err := subscription.ParseURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Remarks != "TestServer" {
		t.Fatalf("remarks: got %q", srv.Remarks)
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(srv.Config, &cfg); err != nil {
		t.Fatal("config not valid JSON:", err)
	}
	if _, ok := cfg["outbounds"]; !ok {
		t.Fatal("config missing outbounds")
	}
	if _, ok := cfg["inbounds"]; !ok {
		t.Fatal("config missing inbounds")
	}
}

func TestParseURI_Trojan(t *testing.T) {
	uri := "trojan://secretpass@proxy.example.com:443?sni=proxy.example.com#TrojanSrv"
	srv, err := subscription.ParseURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Remarks != "TrojanSrv" {
		t.Fatalf("remarks: got %q", srv.Remarks)
	}
}

func TestParseURI_SS(t *testing.T) {
	creds := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:mypassword"))
	uri := "ss://" + creds + "@proxy.example.com:8388#SSServer"
	srv, err := subscription.ParseURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Remarks != "SSServer" {
		t.Fatalf("remarks: got %q", srv.Remarks)
	}
}

func TestParseURI_UnknownScheme(t *testing.T) {
	_, err := subscription.ParseURI("unknown://foo@bar:1234#test")
	if err == nil {
		t.Fatal("expected error for unknown scheme")
	}
}
