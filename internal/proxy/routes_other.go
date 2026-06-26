//go:build !darwin

package proxy

func SetTunRoutes(serverHost string) error { return nil }
func UnsetTunRoutes(serverHost string)     {}
