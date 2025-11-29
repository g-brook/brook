package cmd

import (
	"context"

	"github.com/brook/common/log"
	"github.com/coreos/go-systemd/v22/dbus"
)

func start(service string) {
	conn := getConn()
	defer conn.Close()
	job := make(chan string, 1)
	_, err := conn.StartUnitContext(context.Background(), service, "replace", job)
	if err != nil {
		log.Error("Failed to start %s: %v", service, err)
		return
	}
	log.Info("Started success %s", service)
}

func getConn() *dbus.Conn {
	conn, err := dbus.NewSystemConnectionContext(context.Background())
	if err != nil {
		log.Error("Failed to connect to system bus:", err)
	}
	return conn
}
