package subscription_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/subscription"
)

func TestFetch_JSONArray(t *testing.T) {
	payload := `[
		{"remarks":"Server A","inbounds":[],"outbounds":[],"routing":{},"dns":{}},
		{"remarks":"Server B","inbounds":[],"outbounds":[],"routing":{},"dns":{}}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	}))
	defer srv.Close()

	servers, _, err := subscription.NewFetcher(srv.Client()).Fetch(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}
	if servers[0].Remarks != "Server A" {
		t.Fatalf("remarks[0]: got %q", servers[0].Remarks)
	}
}

func TestCacheRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "servers.json")
	in := []subscription.Server{
		{Remarks: "X", Config: json.RawMessage(`{"remarks":"X"}`)},
	}
	if err := subscription.CacheServers(in, path); err != nil {
		t.Fatal(err)
	}
	out, err := subscription.LoadCachedServers(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Remarks != "X" {
		t.Fatalf("got %+v", out)
	}
}

func TestFetch_WithHeaders(t *testing.T) {
	payload := `[{"remarks":"S1","inbounds":[],"outbounds":[],"routing":{},"dns":{}}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("subscription-userinfo", "upload=1024; download=2048; total=10737418240; expire=1785600000")
		w.Header().Set("profile-title", "VPN Pro")
		w.Header().Set("profile-update-interval", "24")
		w.Write([]byte(payload))
	}))
	defer srv.Close()

	servers, meta, err := subscription.NewFetcher(srv.Client()).Fetch(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 {
		t.Fatalf("got %d servers, want 1", len(servers))
	}
	if meta.Title != "VPN Pro" {
		t.Fatalf("title: got %q", meta.Title)
	}
	if meta.Upload != 1024 || meta.Download != 2048 {
		t.Fatalf("traffic: upload=%d download=%d", meta.Upload, meta.Download)
	}
	if meta.Total != 10737418240 {
		t.Fatalf("total: got %d", meta.Total)
	}
	if meta.Expire != 1785600000 {
		t.Fatalf("expire: got %d", meta.Expire)
	}
	if meta.UpdateInterval != 24 {
		t.Fatalf("updateInterval: got %d", meta.UpdateInterval)
	}
}

func TestFetch_Base64ProfileTitle(t *testing.T) {
	// "My VPN" base64-encoded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("profile-title", "base64:TXkgVlBO")
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	_, meta, err := subscription.NewFetcher(srv.Client()).Fetch(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "My VPN" {
		t.Fatalf("decoded title: got %q", meta.Title)
	}
}

func TestCacheGroupsRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "groups.json")
	in := []subscription.Group{
		{
			URL:  "https://example.com/sub",
			Meta: subscription.Meta{Title: "Test", Total: 1024},
			Servers: []subscription.Server{
				{Remarks: "S1", Config: json.RawMessage(`{}`)},
			},
		},
	}
	if err := subscription.CacheGroups(in, path); err != nil {
		t.Fatal(err)
	}
	out, err := subscription.LoadCachedGroups(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].URL != "https://example.com/sub" || out[0].Meta.Title != "Test" {
		t.Fatalf("got %+v", out)
	}
	if len(out[0].Servers) != 1 || out[0].Servers[0].Remarks != "S1" {
		t.Fatalf("servers: %+v", out[0].Servers)
	}
}
