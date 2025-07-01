package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/remote"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
	"github.com/brook/server/tunnel/http"
	"github.com/spf13/cobra"
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
	//serverConfig.Logger.LogLevel
	log.InitFunc("debug")
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	//Start In-Server.
	inServer := remote.New().Start(&serverConfig)
	tunnelServers := make([]srv.TunnelServer, len(serverConfig.Tunnel))
	for _, config := range serverConfig.Tunnel {
		baseServer := tunnel.NewBaseTunnelServer(&config)
		var ts srv.TunnelServer
		switch config.Type {
		case utils.Http:
			ts = http.NewHttpTunnelServer(baseServer)
			if err := ts.Start(); err != nil {
				log.Error("HttpTunnelServer", "err", err)
				return
			}
		case utils.Https:
		case utils.Tcp:
			//tunnel.NewTcpTunnel(&config, server).Start()
		case utils.Udp:
			log.Error("没有实现当前的协议 %s", config.Type)
		}
		if ts != nil {
			tunnelServers = append(tunnelServers, ts)
		}
	}
	<-ctx.Done()
	shutdown(inServer, tunnelServers)
}

func shutdown(inServer *remote.InServer, tunnelServers []srv.TunnelServer) {
	inServer.Shutdown()
	for _, t := range tunnelServers {
		if t != nil {
			t.Shutdown()
		}
	}
}
