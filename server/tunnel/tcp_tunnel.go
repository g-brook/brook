package tunnel

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	server "github.com/brook/server/remote"
	"io"
)

// TcpTunnel
// @Description: Tcp Tunnel is manager.
type TcpTunnel struct {
	srv.BaseServerHandler

	cfg *configs.TunnelConfig

	server *srv.Server

	inServer *server.InServer

	refChannels map[string]srv.Channel

	fromChannels map[string]srv.Channel
}

func NewTcpTunnel(cfg *configs.TunnelConfig, server *server.InServer) *TcpTunnel {
	return &TcpTunnel{
		cfg:          cfg,
		inServer:     server,
		refChannels:  make(map[string]srv.Channel),
		fromChannels: make(map[string]srv.Channel),
	}
}

func (t *TcpTunnel) Start() {
	go t.doStart()
}

func (t *TcpTunnel) doStart() {
	newServer := srv.NewServer(t.Port())
	newServer.AddHandler(t)
	newServer.AddInitConnHandler(func(conn *srv.GChannel) {
		conn.AddHandler(t)
	})
	t.server = newServer
	//defin.AddTunnel(t)
	err := t.server.Start(srv.WithServerSmux(srv.DefaultServerSmux()))
	if err != nil {
		log.Error("Start tunnel fail %v: %s", err, t.cfg.Port)
	} else {
		log.Info("Start tunnel success:[%d]", t.cfg.Port)
	}
}

func (t *TcpTunnel) RegisterConn(v2 srv.Channel, request exchange.RegisterReq) {
	//t.refChannels = append(t.refChannels, v2)
	//t.refChannels[v2.GetContext().Id] = v2
	//log.Info("Bind tcp tunnel conn t(tunnel/server): %d c(client): %d", t.Port(), v2.RemoteAddr())
}

func (t *TcpTunnel) Receiver(conn *srv.GChannel) {
	id := conn.GetContext().Id
	toConn, ok := t.fromChannels[id]
	if ok {
		_, err := io.Copy(toConn.GetWriter(), conn.GetReader())
		if err != nil {
			log.Error("Copy to srv fail %v", err)
		}
	} else {
		log.Warn("Not found tunnel conn,%s", id)
	}
}

func (t *TcpTunnel) Port() int32 {
	return t.cfg.Port
}

func (t *TcpTunnel) Reader(conn srv.Channel, traverse srv.TraverseBy) {
	length := len(t.refChannels)
	if length > 0 {
		var keys = make([]string, 0, length)
		for key := range t.refChannels {
			keys = append(keys, key)
		}
		firstKey := keys[0]
		target := t.refChannels[firstKey]
		t.fromChannels[firstKey] = conn
		_, err := io.Copy(target.GetWriter(), conn.GetReader())
		if err != nil {
			log.Warn("Error....")
		}
	}
	traverse()
}

func (t *TcpTunnel) DoOpen(conn *srv.GChannel) {
	//not function.
}

func (t *TcpTunnel) DoClose(conn *srv.GChannel) {
	delete(t.refChannels, conn.GetContext().Id)
}
