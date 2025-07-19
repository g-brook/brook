package run

import (
	"fmt"
	"github.com/brook/client/cli"
	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"os"
)

var (
	config  configs.ClientConfig
	cfgPath string
)

func init() {
	cmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "./client.json", "brook client config")
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
	cli.Page.RemoteAddress = fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
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
