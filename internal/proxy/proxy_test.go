package proxy_test

import (
	"testing"

	"github.com/mrsedan/nwxraytui/internal/proxy"
)

func TestHasTunCapability_ReturnsBool(t *testing.T) {
	_ = proxy.HasTunCapability()
}

func TestUnset_NoError(t *testing.T) {
	if err := proxy.Unset(); err != nil {
		t.Fatalf("Unset: %v", err)
	}
}
