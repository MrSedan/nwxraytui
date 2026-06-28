//go:build !linux

package latency

import "syscall"

func fwmarkControl(_ int) func(string, string, syscall.RawConn) error {
	return nil
}
