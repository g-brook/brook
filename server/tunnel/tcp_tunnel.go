package tunnel

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	server "github.com/brook/server/remote"
	srv2 "github.com/brook/server/srv"
	"io"
)

// TcpTunnel
// @Description: Tcp Tunnel is manager.
type TcpTunnel struct {
	srv2.BaseServerHandler

	cfg *configs.TunnelConfig

	server *srv2.Server

	inServer *server.InServer

	refChannels map[string]transport.Channel

	fromChannels map[string]transport.Channel
}

func NewTcpTunnel(cfg *configs.TunnelConfig, server *server.InServer) *TcpTunnel {
	return &TcpTunnel{
		cfg:          cfg,
		inServer:     server,
		refChannels:  make(map[string]transport.Channel),
		fromChannels: make(map[string]transport.Channel),
	}
}

func (t *TcpTunnel) Start() {
	go t.doStart()
}

func (t *TcpTunnel) doStart() {
	newServer := srv2.NewServer(t.Port())
	newServer.AddHandler(t)
	newServer.AddInitConnHandler(func(conn *srv2.GChannel) {
		conn.AddHandler(t)
	})
	t.server = newServer
	//defin.AddTunnel(t)
	err := t.server.Start(srv2.WithServerSmux(srv2.DefaultServerSmux()))
	if err != nil {
		log.Error("Start tunnel fail %v: %s", err, t.cfg.Port)
	} else {
		log.Info("Start tunnel success:[%d]", t.cfg.Port)
	}
}

func (t *TcpTunnel) RegisterConn(v2 transport.Channel, request exchange.RegisterReq) {
	//t.refChannels = append(t.refChannels, v2)
	//t.refChannels[v2.GetContext().Id] = v2
	//log.Info("Bind tcp tunnel conn t(tunnel/server): %d c(client): %d", t.Port(), v2.RemoteAddr())
}

func (t *TcpTunnel) Receiver(conn *srv2.GChannel) {
	id := conn.GetContext().Id
	toConn, ok := t.fromChannels[id]
	if ok {
		_, err := io.Copy(toConn.GetWriter(), conn.GetReader())
		if err != nil {
			log.Error("Copy to transport fail %v", err)
		}
	} else {
		log.Warn("Not found tunnel conn,%s", id)
	}
}

func (t *TcpTunnel) Port() int {
	return t.cfg.Port
}

func (t *TcpTunnel) Reader(conn transport.Channel, traverse srv2.TraverseBy) {
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

func (t *TcpTunnel) DoOpen(conn *srv2.GChannel) {
	//not function.
}

func (t *TcpTunnel) DoClose(conn *srv2.GChannel) {
	delete(t.refChannels, conn.GetContext().Id)
}
