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

	tea "charm.land/bubbletea/v2"
	"github.com/g-brook/brook/client/cli"
	"github.com/g-brook/brook/client/run"
	"github.com/g-brook/brook/common/cmd"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/notify"
	"github.com/g-brook/brook/common/pid"
	"github.com/g-brook/brook/common/version"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	config   *configs.ClientConfig
	cmdValue = cmd.NewCliCmdValue()
	service  *run.Service
	name     = "Brook Tunnel Client(brook-cli)"
)

func init() {
	config = &configs.ClientConfig{}
	rootCmd.PersistentFlags().StringVarP(&cmdValue.ConfigPath, "configs", "c", "./client.json", "brook client configs")
	rootCmd.PersistentFlags().BoolVarP(&cmdValue.IsContainer, "container", "", false, "use container client")
	cmd.InitClientCmd(rootCmd)
}

var rootCmd = &cobra.Command{
	Use:     "start",
	Version: version.GetBuildVersion(),
	Long:    version.Banner(version.BuildVersion, name) + "\nBrook is a cross-platform, high-performance network tunneling and proxy toolkit implemented in Go.\nIt supports a wide range of transport protocols, including TCP, UDP, HTTP(S), and WebSocket, ensuring compatibility with popular application protocols such as SSH, HTTP, Redis, and MySQL.\nA built-in web UI simplifies configuration.",
	Run:     rootRun,
}

func rootRun(cmd *cobra.Command, args []string) {
	log.Info("brook starting; hello world!! 👋")
	version.ShowBanner(version.GetBuildVersion(), name)
	sysCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if cmdValue.ConfigPath == "" {
		log.Error("configs is null, please use -c or --configs to set configs file")
		os.Exit(1)
	}
	exist := configs.IsExist(cmdValue.ConfigPath)
	if exist {
		if err := configs.WriterConfig(cmdValue.ConfigPath, config); err != nil {
			log.Error(err.Error())
			_ = notify.NotifyStopping()
			os.Exit(1)
		}
	} else {
		_ = notify.NotifyStopping()
		os.Exit(1)
		return
	}
	loggerInit()
	run.LoadTunnel()
	verilyBaseConfig(config)
	startServer(config)
	// wait for exit
	<-sysCtx.Done()
	shutdown()
}

func loggerInit() {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		if config.Logger == nil {
			config.Logger = &configs.LoggerConfig{}
			config.Logger.LoggLevel = "info"
			config.Logger.Outs = "cli,file"
		} else {
			config.Logger.Outs = "cli,file"
		}
		log.AddWriter("cli", cli.NewCLIWriteSyncer(config.Logger.LoggLevel))
	}
	log.NewLogger(config.Logger)
}

func verilyBaseConfig(c *configs.ClientConfig) {
	if c.ServerHost == "" {
		panic("ServerHost is null, system exit")
	}
	if c.ServerPort <= 0 {
		panic("ServerPort is 0, system exit")
	}
	if c.Token == "" {
		panic("Token is nil, system exit")
	}
	if c.PingTime <= lang.DefaultPingTime {
		c.PingTime = lang.DefaultPingTime
	}
	i := len(c.Tunnels)
	if i <= 0 {
		panic("Tunnels is null, system exit")
	}
	for _, it := range c.Tunnels {
		if it.ProxyId == "" {
			panic("Tunnels ProxyId is null, system exit")
		}
		if it.TunnelType == "" {
			panic("Tunnels TunnelType（tcp、udp、http(s)） is null, system exit")
		}
	}
}

func Start() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func startServer(config *configs.ClientConfig) {
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
	go OpenCli()
	service = run.NewService()
	service.Run(config)
	pid.CreatePidFile()
	defer func() {
		_ = pid.DeletePidFile()
	}()
}

func shutdown() {
	log.Info("brook exiting; bye bye!! 👋")
	_ = notify.NotifyStopping()
	os.Exit(0)
}

func OpenCli() {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return
	}
	model := cli.NewTUIModel(fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort))
	program := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithoutSignals(),
		tea.WithOutput(os.Stdout))
	cli.SetGlobalProgram(program)
	_, err := program.Run()
	if err != nil {
		log.Error(err.Error())
		return
	}
}
