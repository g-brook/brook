package tcp

import (
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

type UdpTunnelServer struct {
	*tunnel.BaseTunnelServer
	registerLock  sync.Mutex
	poolResources *Resources
}

// NewUdpTunnelServer creates a new TCP tunnel server instance
func NewUdpTunnelServer(server *tunnel.BaseTunnelServer,
	openReq exchange.OpenTunnelReq,
	ch trp.Channel) *UdpTunnelServer {
	tunnelServer := &UdpTunnelServer{
		BaseTunnelServer: server,
		poolResources:    NewResources(ch, openReq, 2),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *UdpTunnelServer) RegisterConn(ch trp.Channel, request exchange.TRegister) {
	if request.GetProxyId() == "" {
		log.Warn("Register udp tunnel, but It' proxyId is nil")
		return
	}
	switch sch := ch.(type) {
	case *trp.SChannel:
		htl.registerLock.Lock()
		defer htl.registerLock.Unlock()
		htl.BaseTunnelServer.RegisterConn(ch, request)
		_ = htl.poolResources.put(NewUdpChannel(sch))
		log.Info("Register udp tunnel, proxyId: %s", request.GetProxyId())
	}
}

func (htl *UdpTunnelServer) Reader(ch trp.Channel, tb srv.TraverseBy) {
	switch workConn := ch.(type) {
	case *srv.GChannel:
		userConn, _ := htl.poolResources.get()
		if userConn == nil {
			_ = ch.Close()
			return
		}
		data, _ := workConn.Next(-1)
		userConn.(*UdpChannel).AsyncWriter(data, ch)
		_ = htl.poolResources.put(userConn)
		return
	}
	tb()
}

func (htl *UdpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("udp tunnel server started:%v", htl.Port())
	htl.background()
	return nil
}

func (htl *UdpTunnelServer) GetUnId() string {
	return htl.poolResources.unId
}

func (htl *UdpTunnelServer) background() {
	go func() {
		<-htl.poolResources.manner.Done()
		htl.Shutdown()
		CloseTunnelServer(htl.Port())
	}()
}
