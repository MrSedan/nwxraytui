//go:build darwin

package proxy

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

const tunIface = "utun9"

func SetTunRoutes(serverHost, _ string) error {
	if err := waitForIface(tunIface, 5*time.Second); err != nil {
		return err
	}
	gw, err := defaultGateway()
	if err != nil {
		return fmt.Errorf("default gateway: %w", err)
	}
	log.Printf("routes: default gateway %s, adding routes via %s", gw, tunIface)

	if serverHost != "" {
		ip, err := resolveToIP(serverHost)
		if err != nil {
			log.Printf("routes: cannot resolve %s for bypass route: %v", serverHost, err)
		} else {
			out, err := exec.Command("route", "add", "-host", ip, gw).CombinedOutput()
			if err != nil {
				log.Printf("routes: bypass route for %s (%s): %v: %s", serverHost, ip, err, out)
			} else {
				log.Printf("routes: bypass %s -> %s via %s", ip, tunIface, gw)
			}
		}
	}

	if out, err := exec.Command("route", "add", "0.0.0.0/1", "-interface", tunIface).CombinedOutput(); err != nil {
		return fmt.Errorf("route 0/1: %w: %s", err, out)
	}
	if out, err := exec.Command("route", "add", "128.0.0.0/1", "-interface", tunIface).CombinedOutput(); err != nil {
		exec.Command("route", "delete", "0.0.0.0/1").Run()
		return fmt.Errorf("route 128/1: %w: %s", err, out)
	}
	log.Printf("routes: 0/1 and 128/1 now via %s", tunIface)
	return nil
}

func UnsetTunRoutes(serverHost string) {
	exec.Command("route", "delete", "0.0.0.0/1").Run()
	exec.Command("route", "delete", "128.0.0.0/1").Run()
	if serverHost != "" {
		if ip, err := resolveToIP(serverHost); err == nil {
			exec.Command("route", "delete", "-host", ip).Run()
		}
	}
	log.Printf("routes: TUN routes removed")
}

func resolveToIP(host string) (string, error) {
	if net.ParseIP(host) != nil {
		return host, nil
	}
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("lookup %s: %w", host, err)
	}
	return addrs[0], nil
}

func waitForIface(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if exec.Command("ifconfig", name).Run() == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("interface %s did not appear within %v", name, timeout)
}

func defaultGateway() (string, error) {
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if gw, ok := strings.CutPrefix(line, "gateway:"); ok {
			return strings.TrimSpace(gw), nil
		}
	}
	return "", fmt.Errorf("no default gateway found")
}
