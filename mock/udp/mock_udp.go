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
)

type MockUDP struct {
}

func main() {
	addr := ":9000"
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		panic(err)
	}
	defer func(conn net.PacketConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	fmt.Println("UDP server listening on", addr)

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}
		fmt.Printf("Received from %v: %s\n", remoteAddr, string(buf[:n]))
		_, _ = conn.WriteTo([]byte("Echo: "+string(buf[:n])), remoteAddr)
	}
}
