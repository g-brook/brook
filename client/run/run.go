package run

import (
	"bufio"
	"fmt"
	"github.com/brook/common/command"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/remote"
	"github.com/spf13/cobra"
	"net"
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
	Use: "brook",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgPath != "" {
			if newConfig, err := configs.GetClientConfig(cfgPath); err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				config = newConfig
			}
		}
		run(&config)
		return nil
	},
}

func Start() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(config *configs.ClientConfig) {
	service := NewService()
	err := service.Run(config)
	if err != nil {
		log.Error("Start client brook error", err)
		return
	}
}

func reader(conn net.Conn, ch chan remote.Protocol) {
	// 从服务器读取一行消息
	for {
		reader := bufio.NewReader(conn)
		decoder, err := remote.Decoder(reader)
		if err != nil {
			fmt.Println("读取失败:", err)
			return
		}
		if decoder.Cmd == remote.Communication {
			ch <- decoder
		} else {
			fmt.Println("收到消息,", decoder.RspCode)
		}
	}

}
