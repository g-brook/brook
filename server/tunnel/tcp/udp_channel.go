package tcp

import (
	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
)

type UdpChannel struct {
	*transport.SChannel
	bucket *exchange.TunnelBucket
}

func NewUdpChannel(src *transport.SChannel) *UdpChannel {
	return &UdpChannel{
		SChannel: src,
		bucket:   exchange.NewTunnelBucket(src, src.Ctx()),
	}
}

func (r *UdpChannel) AsyncWriter(data []byte, rvc transport.Channel) {
	_ = r.bucket.Push(data, func(p *exchange.TunnelProtocol) {
		rvc.Write(p.Data)
	})
}
