package tcp

import (
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

type TunnelUdpServer struct {
	*tunnel.BaseTunnelServer
	registerLock sync.Mutex
	resources    *Resources
}

// NewUdpTunnelServer creates a new TCP tunnel server instance
func NewUdpTunnelServer(server *tunnel.BaseTunnelServer) *TunnelUdpServer {
	tunnelServer := &TunnelUdpServer{
		BaseTunnelServer: server,
		resources:        NewResources(100, server.Cfg.Id, server.Cfg.Port, server.GetManager),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *TunnelUdpServer) RegisterConn(ch trp.Channel, request exchange.TRegister) {
	if request.GetProxyId() == "" {
		log.Warn("Register udp tunnel, but It' proxyId is nil")
		return
	}
	switch sch := ch.(type) {
	case *trp.SChannel:
		htl.registerLock.Lock()
		defer htl.registerLock.Unlock()
		htl.BaseTunnelServer.RegisterConn(ch, request)
		_ = htl.resources.put(NewUdpChannel(sch))
		log.Info("Register udp tunnel, proxyId: %s", request.GetProxyId())
	}
}

func (htl *TunnelUdpServer) Reader(ch trp.Channel, tb srv.TraverseBy) {
	switch workConn := ch.(type) {
	case srv.GContext:
		userConn, _ := htl.resources.get()
		if userConn == nil {
			_ = ch.Close()
			return
		}
		data, _ := workConn.Next(-1)
		userConn.(*UdpChannel).AsyncWriter(data, ch)
		_ = htl.resources.put(userConn)
		return
	}
	tb()
}

func (htl *TunnelUdpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("udp tunnel server started:%v", htl.Port())
	return nil
}
