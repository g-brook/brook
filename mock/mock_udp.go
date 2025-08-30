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
	defer conn.Close()
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
