//go:build !linux && !windows

package notify

func notifyReadiness() error { return nil }
func notifyReloading() error { return nil }
func notifyStopping() error  { return nil }
