package tcp

import (
	"github.com/brook/common/exchange"
	trp "github.com/brook/common/transport"
	. "github.com/brook/server/tunnel"
)

type Resources struct {
	pool       *TunnelPool
	proxyId    string
	remotePort int
	getManager func() trp.Channel
}

// NewResources creates and returns a new Resources instance
// This is a constructor function that initializes a Resources struct
func NewResources(size int, proxyId string, remotePort int, getManager func() trp.Channel) *Resources {
	p := &Resources{
		proxyId:    proxyId,
		remotePort: remotePort,
		getManager: getManager,
	}
	p.pool = NewTunnelPool(p.createConnection, size)
	return p
}

func (htl *Resources) createConnection() error {
	manager := htl.getManager()
	if manager != nil {
		req := &exchange.WorkConnReqByServer{
			ProxyId:    htl.proxyId,
			RemotePort: htl.remotePort,
		}
		request, _ := exchange.NewRequest(req)
		manager.Write(request.Bytes())
	}
	return nil
}

func (htl *Resources) get() (trp.Channel, error) {
	return htl.pool.Get()
}

func (htl *Resources) put(ch trp.Channel) error {
	return htl.pool.Put(ch)
}
