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
	"fmt"
	"os"
	"strings"

	"github.com/brook/client/cli"
	"github.com/brook/common/configs"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	config  *configs.ClientConfig
	cfgPath string
)

func init() {
	config = &configs.ClientConfig{}
	cmd.PersistentFlags().StringVarP(&cfgPath, "configs", "c", "./client.json", "brook client configs")
}

var cmd = &cobra.Command{
	Use:   "Brook-Cli-" + version.GetBuildVersion(),
	Short: "Brook is a cross-platform(Linux/Mac/Windows) proxy software",
	Run: func(cmd *cobra.Command, args []string) {
		if cfgPath == "" {
			fmt.Println("configs is null, please use -c or --configs to set configs file")
			os.Exit(1)
		}
		if err := configs.WriterConfig(cfgPath, config); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.NewLogger(config.Logger)
		LoadTunnel()
		verilyBaseConfig(config)
		run(config)
	},
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
	cli.Page.RemoteAddress = fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
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
		if it.Destination == "" {
			panic("Tunnels Destination is null, system exit")
		}
		split := strings.Split(it.Destination, ":")
		if len(split) != 2 {
			panic("Tunnels Destination is error:" + it.Destination + "correct is ip:port")
		}
	}
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func run(config *configs.ClientConfig) {
	go OpenCli()
	service := NewService()
	ctx := service.Run(config)
	<-ctx.Done()
	log.Info("Brook client exit...")
}

func OpenCli() {
	program := tea.NewProgram(cli.InitModel(), tea.WithInput(os.Stdin), tea.WithoutSignals(),
		tea.WithOutput(os.Stdout))
	_, err := program.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
