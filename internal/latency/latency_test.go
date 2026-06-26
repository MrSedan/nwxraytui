package latency_test

import (
	"net"
	"testing"
	"time"

	"github.com/mrsedan/nwxraytui/internal/latency"
)

func TestPing_Reachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	host := "127.0.0.1"
	port := ln.Addr().(*net.TCPAddr).Port

	ms := latency.Ping(host, port, 2*time.Second)
	if ms < 0 {
		t.Fatalf("expected positive latency, got %d", ms)
	}
}

func TestPing_Unreachable(t *testing.T) {
	ms := latency.Ping("127.0.0.1", 1, 500*time.Millisecond)
	if ms != -1 {
		t.Fatalf("expected -1 for unreachable, got %d", ms)
	}
}
