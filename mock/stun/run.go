package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun/v3"
)

func main() {
	// 连接 STUN 服务器
	c, err := net.Dial("udp", "stun.miwifi.com:3478")
	if err != nil {
		panic(err)
	}

	// 创建 STUN Client
	client, err := stun.NewClient(c)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var xorAddr stun.XORMappedAddress

	// 发送 Binding 请求
	build := stun.MustBuild(stun.TransactionID, stun.BindingRequest,
		stun.Fingerprint,
	)
	err = client.Do(build, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}

		// 解析 STUN 结果
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}
		fmt.Println("Your public IP:", xorAddr.IP)
		fmt.Println("Your public Port:", xorAddr.Port)
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("STUN test completed")
	//go func() {
	//	for {
	//		buf := make([]byte, 1500)
	//		n, err2 := c.Read(buf)
	//		if err2 != nil {
	//			return
	//		}
	//		s := &stun.Message{
	//			Raw: make([]byte, n),
	//		}
	//		copy(s.Raw, buf[:n])
	//		_, err2 = client.Start(s)
	//	}
	//}()
	time.Sleep(111111 * time.Second)
}
