package main

import (
	"net"
)

func main() {
	dial, _ := net.Dial("tcp", "14.212.114.55:26552")
	addr := dial.LocalAddr()
	println(addr.String())
}
