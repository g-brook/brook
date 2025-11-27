package notify

import "golang.org/x/sys/windows/svc"

var globalStatus chan<- svc.Status

func SetGlobalStatus(status chan<- svc.Status) {
	globalStatus = status
}

func notifyReadiness() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{
			State:   svc.Running,
			Accepts: svc.AcceptStop | svc.AcceptShutdown,
		}
	}
	return nil
}

func notifyReloading() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{State: svc.StartPending}
	}
	return nil
}

func notifyStopping() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{State: svc.StopPending}
	}
	return nil
}
