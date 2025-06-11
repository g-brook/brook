package tunnel

import (
	"common/configs"
	"common/log"
	"common/remote"
	"io"
	defin "server/define"
	remote2 "server/remote"
)

// TcpTunnel
// @Description: Tcp Tunnel is manager.
type TcpTunnel struct {
	remote.BaseServerHandler

	cfg *configs.TunnelConfig

	server *remote.Server

	inServer *remote2.InServer

	refChannels map[string]*remote.ConnV2

	fromChannels map[string]*remote.ConnV2
}

func NewTcpTunnel(cfg *configs.TunnelConfig, server *remote2.InServer) *TcpTunnel {
	return &TcpTunnel{
		cfg:          cfg,
		inServer:     server,
		refChannels:  make(map[string]*remote.ConnV2),
		fromChannels: make(map[string]*remote.ConnV2),
	}
}

func (t *TcpTunnel) Start() {
	go t.doStart()
}

func (t *TcpTunnel) doStart() {
	server := remote.NewServer(t.Port())
	server.AddHandler(t)
	server.AddInitConnHandler(func(conn *remote.ConnV2) {
		conn.AddHandler(t)
	})
	t.server = server
	defin.AddTunnel(t)
	err := t.server.Start()
	if err != nil {
		log.Error("Start tunnel fail %v: %s", err, t.cfg.Port)
	} else {
		log.Info("Start tunnel success:[%d]", t.cfg.Port)
	}
}

func (t *TcpTunnel) RegisterConn(v2 *remote.ConnV2, request remote.RegisterReq) {
	//t.refChannels = append(t.refChannels, v2)
	t.refChannels[v2.GetContext().Id] = v2
	log.Info("Bind tcp tunnel conn t(tunnel/server): %d c(client): %d", t.Port(), v2.RemoteAddr())
}

func (t *TcpTunnel) Receiver(conn *remote.ConnV2) {
	id := conn.GetContext().Id
	toConn, ok := t.fromChannels[id]
	if ok {
		_, err := io.Copy(toConn.GetWriter(), conn.GetReader())
		if err != nil {
			log.Error("Copy to remote fail %v", err)
		}
	} else {
		log.Warn("Not found tunnel conn,%s", id)
	}
}

func (t *TcpTunnel) Port() int32 {
	return t.cfg.Port
}

func (t *TcpTunnel) Reader(conn *remote.ConnV2, traverse remote.TraverseBy) {
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

func (t *TcpTunnel) DoOpen(conn *remote.ConnV2) {
	//not function.
}

func (t *TcpTunnel) DoClose(conn *remote.ConnV2) {
	delete(t.refChannels, conn.GetContext().Id)
}
