package run

import (
	"bufio"
	"common/command"
	"common/configs"
	"common/remote"
	"fmt"
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
	//Connection to server.
	transport := remote.NewTransport(1, config)
	transport.Connection(config.ServerHost, config.ServerPort)
}

func connectionServer(host string, port int32) {
	//client := remote.NewClient(host, port)
	//ch := make(chan remote.Protocol)
	//address := fmt.Sprintf("%s:%d", host, port)
	//dial, err := net.Dial("tcp", address)
	//go reader(dial, ch)
	//if err != nil {
	//	return "", errors.New("connection error")
	//}
	//registerRequest := remote.CommunicationInfo{}
	//bytes, _ := json.Marshal(registerRequest)
	//request := remote.NewRequest(remote.Communication, bytes)
	//byts := remote.Encoder(request)
	//_, _ = dial.Write(byts)
	//m := <-ch
	//if m.RspCode == remote.Rsp_success {
	//	fmt.Println("建立通道成功.")
	//	_ = json.Unmarshal(m.Data, &registerRequest)
	//	return registerRequest.BindId, nil
	//}
	//return "", errors.New("bind error")
}

func reader(conn net.Conn, ch chan remote.Protocol) {
	// 从服务器读取一行消息
	for true {
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
