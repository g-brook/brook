package tcp

import (
	"github.com/brook/common/aio"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	defin "github.com/brook/server/define"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
	"sync"
)

type TcpTunnelServer struct {
	*tunnel.BaseTunnelServer
	registerLock sync.Mutex
	allConnects  []string
}

// NewTcpTunnelServer creates a new TCP tunnel server instance
func NewTcpTunnelServer(server *tunnel.BaseTunnelServer) *TcpTunnelServer {
	tunnelServer := &TcpTunnelServer{
		BaseTunnelServer: server,
		allConnects:      make([]string, 0),
	}
	server.DoStart = tunnelServer.startAfter
	server.AddEvent(tunnel.Unregister, tunnelServer.unRegisterConn)

	return tunnelServer
}

func (htl *TcpTunnelServer) RegisterConn(ch trp.Channel, request exchange.RegisterReqAndRsp) {
	if request.ProxyId == "" {
		log.Warn("Register tcp tunnel, but It' proxyId is nil")
		return
	}
	htl.registerLock.Lock()
	defer htl.registerLock.Unlock()
	htl.BaseTunnelServer.RegisterConn(ch, request)
	htl.allConnects = append(htl.allConnects, ch.GetId())
	log.Info("Register tcp tunnel, proxyId: %s", request.ProxyId)
	go func() {
		newRequest, _ := exchange.NewRequest(&exchange.ReqWorkConn{})
		_, _ = ch.Write(newRequest.Bytes())
	}()

}

func (htl *TcpTunnelServer) Reader(ch trp.Channel, _ srv.TraverseBy) {
	switch workConn := ch.(type) {
	case *srv.GChannel:
		chId, ok := workConn.GetContext().GetAttr(defin.ToSChannelId)
		if ok && chId != "" {
			dest, ok := htl.Managers[chId.(string)]
			if ok {
				err := aio.Copy(ch, dest)
				if err != nil {
					log.Error("aio.copy error %v", err)
				}
			}
		}
	}
}

func (htl *TcpTunnelServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	i := len(htl.allConnects)
	if i <= 0 {
		return
	}
	var rw trp.Channel
	bytes := make([]byte, 0)
	for _, chId := range htl.allConnects {
		if newRw, ok := htl.Managers[chId]; ok {
			_, err := newRw.Write(bytes)
			if err == nil {
				rw = newRw
				break
			}
		}
	}
	switch workConn := ch.(type) {
	case *srv.GChannel:
		workConn.GetContext().AddAttr(defin.ToSChannelId, rw.GetId())
		go func() {
			err := aio.SignPipe(rw, ch)
			if err != nil {
				log.Info("aio.SignPipe error %v", err)
				workConn.GetContext().AddAttr(defin.ToSChannelId, "")
			}
		}()
	}
}

func (htl *TcpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("TCP tunnel server started:%v", htl.Port())
	return nil
}

func (htl *TcpTunnelServer) unRegisterConn(ch trp.Channel) {
	log.Info("Remove tcp tunnel session.")
}
