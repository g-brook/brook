package tunnel

import (
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/remote"
	defin "github.com/brook/server/define"
	server "github.com/brook/server/remote"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
)

type TcpListener struct {
	conns chan net.Conn
}

func NewTcpListener() *TcpListener {
	return &TcpListener{conns: make(chan net.Conn, 128)}
}

func (t *TcpListener) Accept() (net.Conn, error) {
	if conn, ok := <-t.conns; ok {
		return conn, nil
	}
	return nil, errors.New("listener close")
}

func (t *TcpListener) Close() error {
	fmt.Println("Close")
	return nil
}

func (t *TcpListener) Addr() net.Addr {
	return (*net.TCPAddr)(nil)
}

func (t *TcpListener) PutConn(conn net.Conn) {
	t.conns <- conn
}

type HttpTunnel struct {
	remote.BaseServerHandler
	config      *configs.TunnelConfig
	server      *server.InServer
	tl          *TcpListener
	tc          sync.WaitGroup
	refChannels map[string]*remote.ConnV2

	fromChannels map[string]*remote.ConnV2
}

func (h *HttpTunnel) Open(conn *remote.ConnV2, traverse remote.TraverseBy) {
	h.tl.PutConn(conn.GetNetConn())
}

func (h *HttpTunnel) Boot(conn *remote.Server, traverse remote.TraverseBy) {
	h.tc.Done()
}

func (h *HttpTunnel) Reader(conn *remote.ConnV2, traverse remote.TraverseBy) {
	length := len(h.refChannels)
	if length > 0 {
		var keys = make([]string, 0, length)
		for key := range h.refChannels {
			keys = append(keys, key)
		}
		firstKey := keys[0]
		target := h.refChannels[firstKey]
		h.fromChannels[firstKey] = conn
		_, err := io.Copy(target.GetWriter(), conn.GetReader())
		if err != nil {
			log.Warn("Error....")
		}
	}
	traverse()
}

func NewHttpTunnel(config *configs.TunnelConfig, server *server.InServer) *HttpTunnel {
	return &HttpTunnel{config: config, server: server, tc: sync.WaitGroup{}, refChannels: make(map[string]*remote.ConnV2), fromChannels: make(map[string]*remote.ConnV2)}
}

func (h *HttpTunnel) Start() {
	h.tc.Add(1)
	go func() {
		server := remote.NewServer(h.Port())
		server.AddHandler(h)
		defin.AddTunnel(h)
		err := server.Start()
		if err != nil {
			log.Error("Started Http server error: %s", h.Port())
		}
	}()
	go func() {
		h.tc.Wait()
		if h.server == nil {
			log.Warn("Server is nil")
			return
		}
		h.tl = NewTcpListener()
		log.Info("Started Http tunnel success %d", h.Port())
		rp := &httputil.ReverseProxy{
			Rewrite: func(request *httputil.ProxyRequest) {
				out := request.Out
				out.URL.Scheme = "http"
			},
		}
		server := http.Server{Handler: rp, ReadHeaderTimeout: 0}
		err := server.Serve(h.tl)
		if err != nil {
			log.Info("HttpTunnel server stop")
		}
	}()
}

func (h *HttpTunnel) Port() int32 {
	return h.config.Port
}

func (h *HttpTunnel) RegisterConn(v2 *remote.ConnV2, request remote.RegisterReq) {
	//t.refChannels = append(t.refChannels, v2)
	h.refChannels[v2.GetContext().Id] = v2
	log.Info("Bind tcp tunnel conn t(tunnel/server): %d c(client): %d", h.Port(), v2.RemoteAddr())
}

func (h *HttpTunnel) Receiver(conn *remote.ConnV2) {
	id := conn.GetContext().Id
	toConn, ok := h.fromChannels[id]
	if ok {
		_, err := io.Copy(toConn.GetWriter(), conn.GetReader())
		if err != nil {
			log.Error("Copy to remote fail %v", err)
		}
	} else {
		log.Warn("Not found tunnel conn,%s", id)
	}
}
