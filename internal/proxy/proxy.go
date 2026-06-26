package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func Set(mode string, socksPort, httpPort int) error {
	if mode != "system" {
		return nil
	}
	switch runtime.GOOS {
	case "darwin":
		return setMacOS(socksPort, httpPort)
	case "linux":
		return setLinux(socksPort, httpPort)
	default:
		return fmt.Errorf("system proxy not supported on %s", runtime.GOOS)
	}
}

func Unset() error {
	switch runtime.GOOS {
	case "darwin":
		return unsetMacOS()
	case "linux":
		return unsetLinux()
	default:
		return nil
	}
}

func setMacOS(socksPort, httpPort int) error {
	svcs, err := macNetworkServices()
	if err != nil {
		return err
	}
	p := strconv.Itoa(socksPort)
	hp := strconv.Itoa(httpPort)
	for _, svc := range svcs {
		exec.Command("networksetup", "-setsocksfirewallproxy", svc, "127.0.0.1", p).Run()
		exec.Command("networksetup", "-setsocksfirewallproxystate", svc, "on").Run()
		exec.Command("networksetup", "-setwebproxy", svc, "127.0.0.1", hp).Run()
		exec.Command("networksetup", "-setsecurewebproxy", svc, "127.0.0.1", hp).Run()
	}
	return nil
}

func unsetMacOS() error {
	svcs, _ := macNetworkServices()
	for _, svc := range svcs {
		exec.Command("networksetup", "-setsocksfirewallproxystate", svc, "off").Run()
		exec.Command("networksetup", "-setwebproxystate", svc, "off").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", svc, "off").Run()
	}
	return nil
}

func macNetworkServices() ([]string, error) {
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil, err
	}
	var svcs []string
	for _, line := range strings.Split(string(out), "\n")[1:] {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") {
			svcs = append(svcs, line)
		}
	}
	return svcs, nil
}

func setLinux(socksPort, httpPort int) error {
	sp := strconv.Itoa(socksPort)
	hp := strconv.Itoa(httpPort)
	cmds := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "manual"},
		{"gsettings", "set", "org.gnome.system.proxy.socks", "host", "127.0.0.1"},
		{"gsettings", "set", "org.gnome.system.proxy.socks", "port", sp},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", "127.0.0.1"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", hp},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", "127.0.0.1"},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", hp},
	}
	for _, c := range cmds {
		exec.Command(c[0], c[1:]...).Run()
	}
	return writeEnvFile(socksPort, httpPort)
}

func unsetLinux() error {
	exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "nwxraytui", "proxy-env.sh")
	return os.Remove(path)
}

func writeEnvFile(socksPort, httpPort int) error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "nwxraytui", "proxy-env.sh")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := fmt.Sprintf(
		"export HTTP_PROXY=\"http://127.0.0.1:%d\"\n"+
			"export HTTPS_PROXY=\"http://127.0.0.1:%d\"\n"+
			"export ALL_PROXY=\"socks5://127.0.0.1:%d\"\n"+
			"export http_proxy=\"$HTTP_PROXY\"\n"+
			"export https_proxy=\"$HTTPS_PROXY\"\n"+
			"export all_proxy=\"$ALL_PROXY\"\n",
		httpPort, httpPort, socksPort,
	)
	return os.WriteFile(path, []byte(content), 0o644)
}
