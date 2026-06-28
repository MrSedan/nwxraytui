//go:build linux

package latency

import "syscall"

func fwmarkControl(mark int) func(string, string, syscall.RawConn) error {
	return func(_, _ string, c syscall.RawConn) error {
		c.Control(func(fd uintptr) {
			syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, mark)
		})
		return nil
	}
}
