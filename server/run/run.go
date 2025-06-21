package run

import (
	"fmt"
	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/remote"
	"github.com/brook/server/tunnel"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

var (
	serverConfig configs.ServerConfig
	cfgPath      string
)

// init
//
//	@Description: init.
func init() {
	cmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "./server.json", "config file path")
	command.RegisterServerFlags(cmd, serverConfig)
}

var cmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgPath != "" {
			config, err := configs.GetServerConfig(cfgPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			serverConfig = config
		}
		initLogger()
		run()
		return nil
	},
}

func initLogger() {
	log.InitFunc(serverConfig.Logger)
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func run() {
	group := sync.WaitGroup{}
	group.Add(1)
	//Start In-Server.
	_ = remote.New().Start(&serverConfig)
	for _, config := range serverConfig.Tunnel {
		baseServer := tunnel.NewBaseTunnelServer(&config)
		switch config.Type {
		case utils.Http:
			if err := tunnel.NewHttpTunnelServer(baseServer).Start(); err != nil {
				log.Error("HttpTunnelServer", "err", err)
				return
			}
		case utils.Https:
		case utils.Tcp:
			//tunnel.NewTcpTunnel(&config, server).Start()
		case utils.Udp:
			log.Error("没有实现当前的协议 %s", config.Type)
		}
	}
	group.Wait()
}
