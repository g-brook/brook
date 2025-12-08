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
