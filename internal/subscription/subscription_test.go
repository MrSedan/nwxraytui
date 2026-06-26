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

	servers, err := subscription.NewFetcher(srv.Client()).Fetch(srv.URL)
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
