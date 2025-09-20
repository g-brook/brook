package metrics

import (
	"github.com/brook/server/srv"
)

type Channel struct {
	*srv.GChannel
	traffic *TunnelTraffic
}

func NewMetricsChannel(src *srv.GChannel, traffic *TunnelTraffic) *Channel {
	return &Channel{
		GChannel: src,
		traffic:  traffic,
	}
}

func (c *Channel) Next(pos int) ([]byte, error) {
	next, err := c.GChannel.Next(pos)
	if c.traffic != nil && err == nil {
		c.traffic.AddInBytes(len(next))
	}
	return next, err
}

func (c *Channel) Write(p []byte) (n int, err error) {
	write, err := c.GChannel.Write(p)
	if c.traffic != nil && err == nil {
		c.traffic.AddOutBytes(write)
	}
	return write, err
}

func (c *Channel) Read(p []byte) (n int, err error) {
	read, err := c.GChannel.Read(p)
	if c.traffic != nil && err == nil {
		c.traffic.AddInBytes(read)
	}
	return read, err
}
