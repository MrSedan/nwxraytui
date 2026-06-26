//go:build !darwin && !linux

package proxy

func SetTunRoutes(serverHost, _ string) error { return nil }
func UnsetTunRoutes(serverHost string)         {}
