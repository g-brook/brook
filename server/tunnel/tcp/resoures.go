package tcp

import (
	"github.com/brook/common/exchange"
	trp "github.com/brook/common/transport"
	"github.com/brook/common/utils"
	. "github.com/brook/server/tunnel"
)

type Resources struct {
	pool         *TunnelPool
	unId         string
	proxyId      string
	manner       trp.Channel
	network      utils.Network
	localAddress string
	port         int
	tunnelType   utils.TunnelType
}

// NewResources creates and returns a new Resources instance
// This is a constructor function that initializes a Resources struct
func NewResources(manner trp.Channel,
	openReq exchange.OpenTunnelReq, size int) *Resources {
	p := &Resources{
		manner:       manner,
		unId:         openReq.UnId,
		proxyId:      openReq.ProxyId,
		localAddress: openReq.LocalAddress,
		port:         openReq.TunnelPort,
		tunnelType:   openReq.TunnelType,
		network:      utils.Network(openReq.TunnelType),
	}
	p.pool = NewTunnelPool(p.createConnection, size)
	return p
}

func (htl *Resources) createConnection() error {
	req := &exchange.ReqWorkConn{
		ProxyId:      htl.proxyId,
		Port:         htl.port,
		TunnelType:   htl.tunnelType,
		LocalAddress: htl.localAddress,
		UnId:         htl.unId,
		Network:      htl.network,
	}
	request, _ := exchange.NewRequest(req)
	_, err := htl.manner.Write(request.Bytes())
	return err
}

func (htl *Resources) get() (trp.Channel, error) {
	return htl.pool.Get()
}

func (htl *Resources) put(ch trp.Channel) error {
	return htl.pool.Put(ch)
}
