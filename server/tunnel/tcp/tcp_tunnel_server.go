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
	unId         string
	proxyId      string
	pool         *tunnel.TunnelPool
	manner       trp.Channel
	localAddress string
}

// NewTcpTunnelServer creates a new TCP tunnel server instance
func NewTcpTunnelServer(server *tunnel.BaseTunnelServer,
	openReq exchange.OpenTunnelReq,
	ch trp.Channel) *TcpTunnelServer {
	tunnelServer := &TcpTunnelServer{
		BaseTunnelServer: server,
		unId:             openReq.UnId,
		proxyId:          openReq.ProxyId,
		localAddress:     openReq.LocalAddress,
		manner:           ch,
	}
	tunnelServer.pool = tunnel.NewTunnelPool(tunnelServer.createConn, 1)
	server.DoStart = tunnelServer.startAfter
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
	_ = htl.pool.Put(ch)
	log.Info("Register tcp tunnel, proxyId: %s", request.ProxyId)

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
	userConn := htl.getUserConn()
	if userConn == nil {
		_ = ch.Close()
		return
	}
	switch workConn := ch.(type) {
	case *srv.GChannel:
		workConn.GetContext().AddAttr(defin.ToSChannelId, userConn.GetId())
		go func() {
			err := aio.SignPipe(userConn, workConn)
			log.Error("aio.SignPipe error %v", err)
		}()
	}
}

func (htl *TcpTunnelServer) createConn() (err error) {
	req := &exchange.ReqWorkConn{
		ProxyId:      htl.proxyId,
		Port:         htl.Port(),
		TunnelType:   htl.Cfg.Type,
		LocalAddress: htl.localAddress,
		UnId:         htl.unId,
	}
	err = htl.writeMsg(req)
	return
}

func (htl *TcpTunnelServer) getUserConn() trp.Channel {
	sch, _ := htl.pool.Get()
	return sch
}

func (htl *TcpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("TCP tunnel server started:%v", htl.Port())
	htl.background()
	return nil
}

func (htl *TcpTunnelServer) writeMsg(request exchange.InBound) (err error) {
	rt, _ := exchange.NewRequest(request)
	_, err = htl.manner.Write(rt.Bytes())
	return
}

func (htl *TcpTunnelServer) GetUnId() string {
	return htl.unId
}

func (htl *TcpTunnelServer) background() {
	go func() {
		<-htl.manner.Done()
		htl.Shutdown()
		CloseTunnelServer(htl.Port())
	}()
}
