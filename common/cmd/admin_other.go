//go:build !linux

package cmd

import (
	"github.com/brook/common/log"
)

func start(service string) {
	log.Error("Not supported command:start")
}

func stop(service string) {
	log.Error("Not supported command:stop")
}

func status(service string) {
	log.Error("Not supported command:status")
}

func restart(service string) {
	log.Error("Not supported command:restart")
}
