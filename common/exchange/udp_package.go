package exchange

import (
	"net"
)

type UdpPackage struct {
	Data []byte `json:"data"`

	LocalAddress *net.UDPAddr `json:"local_address"`

	RemoteAddress *net.UDPAddr `json:"remote_address"`
}

func NewUdpPackage(data []byte, localAddr, remoteAddr *net.UDPAddr) *UdpPackage {
	return &UdpPackage{
		Data:          data,
		LocalAddress:  localAddr,
		RemoteAddress: remoteAddr,
	}
}

func (p *UdpPackage) GetRemoteAddress() *net.UDPAddr {
	return p.RemoteAddress
}

func (p *UdpPackage) GetLocalAddress() *net.UDPAddr {
	return p.LocalAddress
}

func (p *UdpPackage) GetData() []byte {
	return p.Data
}
