package proxy

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

func HasTunCapability() bool {
	switch runtime.GOOS {
	case "linux":
		return hasNetAdmin()
	case "darwin":
		return os.Getuid() == 0
	default:
		return false
	}
}

func hasNetAdmin() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CapAmb:") {
			continue
		}
		hex := strings.TrimSpace(strings.TrimPrefix(line, "CapAmb:"))
		caps, err := strconv.ParseUint(hex, 16, 64)
		if err != nil {
			return false
		}
		const capNetAdmin = 12
		return caps&(1<<capNetAdmin) != 0
	}
	return false
}
