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

	"github.com/brook/client/cli"
	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	config  configs.ClientConfig
	cfgPath string
)

func init() {
	cmd.PersistentFlags().StringVarP(&cfgPath, "configs", "c", "./client.json", "brook client configs")
	command.RegisterClientFlags(cmd, config)
}

var cmd = &cobra.Command{
	Use:   "Brook",
	Short: "Brook is a cross-platform(Linux/Mac/Windows) proxy software",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgPath != "" {
			if newConfig, err := configs.GetClientConfig(cfgPath); err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				config = newConfig
			}
		}
		log.NewLogger(&config.Logger)
		LoadTunnel()
		verilyBaseConfig(&config)
		run(&config)
		return nil
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
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func run(config *configs.ClientConfig) {
	//go OpenCli()
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
