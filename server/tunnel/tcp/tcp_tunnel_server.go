package tcp

import (
	"sync"

	"github.com/brook/common/aio"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/defin"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

type TcpTunnelServer struct {
	*tunnel.BaseTunnelServer
	registerLock  sync.Mutex
	poolResources *Resources
}

// NewTcpTunnelServer creates a new TCP tunnel server instance
func NewTcpTunnelServer(server *tunnel.BaseTunnelServer, openReq exchange.OpenTunnelReq, ch trp.Channel) *TcpTunnelServer {
	tunnelServer := &TcpTunnelServer{
		BaseTunnelServer: server,
		poolResources:    NewResources(ch, openReq, 100),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *TcpTunnelServer) RegisterConn(ch trp.Channel, request exchange.TRegister) {
	if request.GetProxyId() == "" {
		log.Warn("Register tcp tunnel, but It' proxyId is nil")
		return
	}
	htl.registerLock.Lock()
	defer htl.registerLock.Unlock()
	htl.BaseTunnelServer.RegisterConn(ch, request)
	_ = htl.poolResources.put(ch)
	log.Info("Register tcp tunnel, proxyId: %s", request.GetProxyId())

}

func (htl *TcpTunnelServer) Reader(ch trp.Channel, _ srv.TraverseBy) {
	switch workConn := ch.(type) {
	case srv.GContext:
		chId, ok := workConn.GetContext().GetAttr(defin.ToSChannelId)
		if ok && chId != "" {
			dest, ok := htl.Managers[chId.(string)]
			if ok {
				err := aio.Copy(ch, dest)
				if err != nil {
					log.Debug("aio.copy error %v", err)
				}
			}
		}
	}
}

func (htl *TcpTunnelServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	userConn, _ := htl.poolResources.get()
	if userConn == nil {
		_ = ch.Close()
		return
	}
	switch workConn := ch.(type) {
	case *srv.GChannel:
		workConn.GetContext().AddAttr(defin.ToSChannelId, userConn.GetId())
		go func() {
			err := aio.SinglePipe(userConn, workConn)
			log.Debug("aio.SinglePipe error %v", err)
		}()
	}
}

func (htl *TcpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("TCP tunnel server started:%v", htl.Port())
	htl.background()
	return nil
}

func (htl *TcpTunnelServer) GetUnId() string {
	return htl.poolResources.unId
}

func (htl *TcpTunnelServer) background() {
	if htl.poolResources.manner != nil {
		go func() {
			<-htl.poolResources.manner.Done()
			htl.Shutdown()
			CloseTunnelServer(htl.Port())
		}()
	}
}
