package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brook/server/tunnel/tcp"
	"github.com/brook/server/web"
	"github.com/brook/server/web/db"

	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/remote"
	"github.com/brook/server/tunnel"
	"github.com/brook/server/tunnel/http"
	"github.com/spf13/cobra"
)

var (
	serverConfig configs.ServerConfig
	cfgPath      string
)

// init function is called automatically when the package is initialized
// It sets up command line flags and registers server-specific flags
func init() {
	// Add a persistent string flag for config file path
	// The flag can be referenced as "--config" or "-c"
	// Default value is "./server.json"
	// The flag stores the config file path in cfgPath variable
	cmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "./server.json", "config file path")
	// Register server-specific flags with the command
	// This function likely adds flags related to server configuration
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
		initLogger(serverConfig)
		configCheck(serverConfig)
		run()
		return nil
	},
}

func configCheck(config configs.ServerConfig) {

}

func initLogger(svf configs.ServerConfig) {
	log.NewLogger(&svf.Logger)
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

// run is the main entry point for the server application
func run() {
	// Create a context that can be cancelled by interrupt signals (SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // Ensure the signal notification is stopped when the function returns
	if serverConfig.EnableWeb {
		web.NewWebServer(serverConfig.WebPort)
	}
	//Start In-Server.
	remote.Inserver = remote.New().Start(&serverConfig)

	// Get tunnelServer infos.
	tunnelConfig := GetTunnelConfig(serverConfig)
	tunnelServers := make([]tunnel.TunnelServer, len(serverConfig.Tunnel))
	for _, config := range tunnelConfig {
		baseServer := tunnel.NewBaseTunnelServer(config)
		var ts tunnel.TunnelServer
		switch config.Type {
		case utils.Http, utils.Https:
			ts = http.NewHttpTunnelServer(baseServer)
			if err := ts.Start(utils.NetworkTcp); err != nil {
				log.Error("HttpTunnelServer", "err", err)
				return
			}
		case utils.Tcp:
			tcp.AcceptTcpListener()
			break
		}
		if ts != nil {
			tunnelServers = append(tunnelServers, ts)
		}
	}
	<-ctx.Done()
	shutdown(remote.Inserver, tunnelServers)
}

func shutdown(inServer *remote.InServer, tunnelServers []tunnel.TunnelServer) {
	inServer.Shutdown()
	for _, t := range tunnelServers {
		if t != nil {
			t.Shutdown()
		}
	}
	db.Close()
}
