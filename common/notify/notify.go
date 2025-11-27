package notify

func NotifyReloading() error {
	return notifyReloading()
}

func NotifyStopping() error {
	return notifyStopping()
}

func NotifyReadiness() error {
	return notifyReadiness()
}
