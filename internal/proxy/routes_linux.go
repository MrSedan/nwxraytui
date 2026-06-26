//go:build linux

package proxy

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

const tunIface = "tun0"

// tunAddr is the point-to-point address assigned to the TUN device. xray-core's
// tun inbound brings the interface up but assigns no layer-3 address, so we do
// it here: source-address selection and on-link routing need it.
const tunAddr = "198.18.0.1/30"

// tunFwmark is the SO_MARK value set on xray's freedom-outbound sockets (see
// xray.TunFwmark). The matching ip rule below sends those packets through the
// real interface rather than back into tun0.
const tunFwmark = 255

// tunRouteTable is the auxiliary routing table used for fwmark bypass.
const tunRouteTable = "100"

// localBypass are networks kept off the tunnel: loopback, private LANs,
// link-local, CGNAT and multicast. They are routed via the real gateway so
// LAN hosts, the router (and any DNS it serves) stay reachable.
var localBypass = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"169.254.0.0/16",
	"100.64.0.0/10",
	"224.0.0.0/4",
}

func SetTunRoutes(serverHost, dnsServer string) error {
	if err := waitForIface(tunIface, 5*time.Second); err != nil {
		return err
	}
	gw, dev, err := defaultGateway()
	if err != nil {
		return fmt.Errorf("default gateway: %w", err)
	}
	log.Printf("routes: gateway %s dev %s, adding routes via %s", gw, dev, tunIface)

	// xray-core only creates the interface; assign its address and ensure it is
	// up before routing. Address may already exist on reconnect — ignore that.
	if out, err := exec.Command("ip", "address", "add", tunAddr, "dev", tunIface).CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "File exists") {
			log.Printf("routes: assign %s to %s: %v: %s", tunAddr, tunIface, err, out)
		}
	}
	if out, err := exec.Command("ip", "link", "set", tunIface, "up").CombinedOutput(); err != nil {
		log.Printf("routes: bring %s up: %v: %s", tunIface, err, out)
	}

	if serverHost != "" {
		ip, err := resolveToIP(serverHost)
		if err != nil {
			log.Printf("routes: cannot resolve %s for bypass: %v", serverHost, err)
		} else {
			out, err := exec.Command("ip", "route", "add", ip+"/32", "via", gw, "dev", dev).CombinedOutput()
			if err != nil {
				log.Printf("routes: bypass %s via %s: %v: %s", ip, gw, err, out)
			}
		}
	}

	for _, cidr := range localBypass {
		if out, err := exec.Command("ip", "route", "add", cidr, "via", gw, "dev", dev).CombinedOutput(); err != nil {
			log.Printf("routes: bypass %s via %s: %v: %s", cidr, gw, err, out)
		}
	}

	// Policy routing rule: packets marked by xray's direct outbound bypass tun0
	// and are forwarded via the real interface using table tunRouteTable.
	if out, err := exec.Command("ip", "route", "add", "default", "via", gw, "dev", dev, "table", tunRouteTable).CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "File exists") {
			log.Printf("routes: fwmark table default: %v: %s", err, out)
		}
	}
	if out, err := exec.Command("ip", "rule", "add", "fwmark", fmt.Sprintf("%d", tunFwmark), "table", tunRouteTable, "priority", "100").CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "File exists") {
			log.Printf("routes: fwmark rule: %v: %s", err, out)
		}
	}

	if out, err := exec.Command("ip", "route", "add", "0.0.0.0/1", "dev", tunIface).CombinedOutput(); err != nil {
		return fmt.Errorf("route 0/1: %w: %s", err, out)
	}
	if out, err := exec.Command("ip", "route", "add", "128.0.0.0/1", "dev", tunIface).CombinedOutput(); err != nil {
		exec.Command("ip", "route", "del", "0.0.0.0/1").Run()
		return fmt.Errorf("route 128/1: %w: %s", err, out)
	}
	log.Printf("routes: 0/1 and 128/1 via %s", tunIface)
	setDNS(dnsServer)
	return nil
}

const dnsFallback = "1.1.1.1"

func setDNS(dnsServer string) {
	target := dnsServer
	if target == "" {
		target = dnsFallback
	}
	if out, err := exec.Command("resolvectl", "dns", tunIface, target).CombinedOutput(); err != nil {
		log.Printf("dns: resolvectl dns: %v: %s", err, out)
		return
	}
	// "~." makes tun0 the default resolver for all domains.
	if out, err := exec.Command("resolvectl", "domain", tunIface, "~.").CombinedOutput(); err != nil {
		log.Printf("dns: resolvectl domain: %v: %s", err, out)
	}
	log.Printf("dns: set %s via resolvectl for %s", target, tunIface)
}

func unsetDNS() {
	if out, err := exec.Command("resolvectl", "revert", tunIface).CombinedOutput(); err != nil {
		log.Printf("dns: resolvectl revert: %v: %s", err, out)
		return
	}
	log.Printf("dns: reverted resolvectl for %s", tunIface)
}

func UnsetTunRoutes(serverHost string) {
	unsetDNS()
	exec.Command("ip", "route", "del", "0.0.0.0/1", "dev", tunIface).Run()
	exec.Command("ip", "route", "del", "128.0.0.0/1", "dev", tunIface).Run()
	for _, cidr := range localBypass {
		exec.Command("ip", "route", "del", cidr).Run()
	}
	if serverHost != "" {
		if ip, err := resolveToIP(serverHost); err == nil {
			exec.Command("ip", "route", "del", ip+"/32").Run()
		}
	}
	exec.Command("ip", "rule", "del", "fwmark", fmt.Sprintf("%d", tunFwmark), "table", tunRouteTable).Run()
	exec.Command("ip", "route", "flush", "table", tunRouteTable).Run()
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
		if exec.Command("ip", "link", "show", name).Run() == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("interface %s did not appear within %v", name, timeout)
}

func defaultGateway() (gw, dev string, err error) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", "", err
	}
	// "default via 192.168.1.1 dev eth0 proto dhcp ..."
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			gw = fields[i+1]
		}
		if f == "dev" && i+1 < len(fields) {
			dev = fields[i+1]
		}
	}
	if gw == "" {
		return "", "", fmt.Errorf("no default gateway in: %s", strings.TrimSpace(string(out)))
	}
	return gw, dev, nil
}
