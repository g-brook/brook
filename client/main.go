package main

import (
	"github.com/brook/client/run"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
)

func init() {
	config := configs.LoggerConfig{LogPath: "./", LoggLevel: "debug"}
	log.InitFunc(config.LoggLevel)
}
func main() {
	run.Start()
}
