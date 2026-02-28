/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/g-brook/brook/common/cmd"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/notify"
	"github.com/g-brook/brook/common/pid"
	"github.com/g-brook/brook/common/version"
	"github.com/g-brook/brook/scmd/standard"
	"github.com/g-brook/brook/scmd/web"
	"github.com/g-brook/brook/scmd/web/service"
	"github.com/g-brook/brook/server/defin"
	"github.com/g-brook/brook/server/remote"
	"github.com/spf13/cobra"
)

var (
	serverConfig configs.ServerConfig
	cmdValue     = cmd.NewSevCmdValue()
)

// init function is called automatically when the package is initialized
// It sets up command line flags and registers server-specific flags
func init() {
	// Add a persistent string flag for configs file path
	// The flag can be referenced as "--configs" or "-c"
	// Default value is "./server.json"
	// The flag stores the configs file path in cfgPath variable
	rootCmd.PersistentFlags().StringVarP(&cmdValue.ConfigPath, "configs", "c", "./server.json", "configs file path")
	rootCmd.PersistentFlags().BoolVarP(&cmdValue.IsContainer, "container", "", false, "use container client")
	cmd.InitServerCmd(rootCmd)
}

var rootCmd = &cobra.Command{
	Use:     "start",
	Version: version.GetBuildVersion(),
	Long:    version.Banner(version.GetBuildVersion()) + "\nBrook is a cross-platform, high-performance network tunneling and proxy toolkit implemented in Go.\nIt supports a wide range of transport protocols, including TCP, UDP, HTTP(S), and WebSocket, ensuring compatibility with popular application protocols such as SSH, HTTP, Redis, and MySQL.\nA built-in web UI simplifies configuration.",
	Run:     rootRun,
}

func rootRun(_ *cobra.Command, _ []string) {
	fmt.Println("brook starting; hello world!! 👋")
	version.ShowBanner(version.GetBuildVersion())
	// Create a context that can be cancelled by interrupt signals (SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // Ensure the signal notification is stopped when the function returns
	if cmdValue.ConfigPath != "" {
		config, err := configs.GetServerConfig(cmdValue.ConfigPath)
		if err != nil {
			log.Error("get configs error %v", err)
			os.Exit(1)
		}
		serverConfig = config
	}
	initLogger(&serverConfig)
	configCheck(&serverConfig)
	run()
	<-ctx.Done()
	shutdown()
}

func configCheck(config *configs.ServerConfig) {

}

func initLogger(svf *configs.ServerConfig) {
	log.NewLogger(&svf.Logger)
}

func Start() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

// run is the main entry point for the server application
func run() {
	err := notify.NotifyReloading()
	if err != nil {
		log.Error("notify reloading error: %v", err)
	}
	defer func() {
		err = notify.NotifyReadiness()
		if err != nil {
			log.Error("notify readiness error: %v", err)
		}
	}()
	if serverConfig.EnableWeb {
		web.NewWebServer(serverConfig.WebPort)
	}
	//Start In-Server.
	remote.Inserver = remote.New().Start(&serverConfig)
	// Get tunnelServer infos.
	standard.InitTunnelConfig(&serverConfig)
	afterRun(&serverConfig)
	pid.CreatePidFile()
	defer func() {
		_ = pid.DeletePidFile()
	}()
}

// afterRun is a function that sets the authentication token based on the server configuration
// It takes a ServerConfig pointer as parameter and sets the token in the defin package
func afterRun(config *configs.ServerConfig) {
	var token string
	if config.EnableWeb {
		token = service.GetToken()
	} else {
		token = config.Token
	}
	defin.Set(defin.TokenKey, token)
	defin.Set(defin.ServerPort, config.ServerPort)
}

func shutdown() {
	log.Info("brook exiting; bye bye!! 👋")
	if remote.Inserver != nil {
		remote.Inserver.Shutdown()
	}
	if serverConfig.EnableWeb {
		web.Close()
	}
}
