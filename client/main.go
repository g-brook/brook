package main

import (
	"github.com/brook/client/run"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
)

func init() {
	log.InitFunc(configs.LoggerConfig{LogPath: "./", LoggLevel: "debug"})
}
func main() {
	run.Start()
}
