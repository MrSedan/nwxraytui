package latency

import (
	"net"
	"strconv"
	"time"
)

// Ping measures TCP round-trip time to host:port.
func Ping(host string, port int, timeout time.Duration) int {
	return pingDial(host, port, timeout, 0)
}

// PingDirect is like Ping but sets SO_MARK on the socket on Linux so the
// connection bypasses any active TUN routing rules and goes via the real
// network interface.
func PingDirect(host string, port int, timeout time.Duration, fwmark int) int {
	return pingDial(host, port, timeout, fwmark)
}

func pingDial(host string, port int, timeout time.Duration, fwmark int) int {
	d := &net.Dialer{Timeout: timeout}
	if fwmark != 0 {
		d.Control = fwmarkControl(fwmark)
	}
	start := time.Now()
	conn, err := d.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return -1
	}
	conn.Close()
	return int(time.Since(start).Milliseconds())
}
