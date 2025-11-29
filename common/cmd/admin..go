package cmd

func Start(service string) {
	start(service)
}

func Stop(service string) {
	stop(service)
}

func Status(service string) {
	status(service)
}

func Restart(service string) {
	restart(service)
}
