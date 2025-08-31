package tcp

import (
	"encoding/json"
	"net"

	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
)

type UdpChannel struct {
	*transport.SChannel
	bucket     *exchange.TunnelBucket
	udpConnMap hash.SyncMap[string, transport.Channel]
}

func NewUdpChannel(src *transport.SChannel) *UdpChannel {
	bucket := exchange.NewTunnelBucket(src, src.Ctx()).Run()
	channel := &UdpChannel{
		SChannel: src,
		bucket:   bucket,
	}
	bucket.DefaultRead(channel.read)
	return channel
}

func (r *UdpChannel) read(p *exchange.TunnelProtocol) {
	var udpPackage exchange.UdpPackage
	err := json.Unmarshal(p.Data, &udpPackage)
	if err != nil {
		return
	}
	s := udpPackage.RemoteAddress.String()
	ct, ok := r.udpConnMap.Load(s)
	if ok {
		_, _ = ct.Write(udpPackage.Data)
	}
}

func (r *UdpChannel) AsyncWriter(data []byte, ct transport.Channel) {
	remoteAddress, ok := ct.RemoteAddr().(*net.UDPAddr)
	if !ok {
		log.Warn("It not is udp addr %s", remoteAddress.String())
		return
	}
	udpPackage := exchange.NewUdpPackage(data, nil, remoteAddress)
	jsonData, _ := json.Marshal(udpPackage)
	s := remoteAddress.String()
	_, b := r.udpConnMap.Load(s)
	if !b {
		r.udpConnMap.Store(s, ct)
	}
	_ = r.bucket.Push(jsonData, nil)
}
