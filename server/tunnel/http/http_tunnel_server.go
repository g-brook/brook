package http

import (
	"errors"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
	"net"
	"net/http"
	"strconv"
	"time"
)

type HttpTunnelServer struct {
	*tunnel.BaseTunnelServer

	listener *TcpListener

	proxyToConn map[string]string
}

// NewHttpTunnelServer is a constructor function for HttpTunnelServer. It takes a pointer to BaseTunnelServer as input
// and returns a pointer to HttpTunnelServer. The constructor sets the DoStart field of BaseTunnelServer to the startAfter
// method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter the server
// processes the request. The constructor also returns a pointer to HttpTunnelServer.
func NewHttpTunnelServer(server *tunnel.BaseTunnelServer) *HttpTunnelServer {
	if server.Cfg == nil {
		panic("start http tunnel server error, cfg is nil")
	}
	if err := verifyCfg(server.Cfg); err != nil {
		panic("http tunnel server cfg verify is false")
	}
	tunnelServer := HttpTunnelServer{
		BaseTunnelServer: server,
		proxyToConn:      make(map[string]string),
	}
	server.DoStart = tunnelServer.startAfter
	addRoute(server.Cfg)
	return &tunnelServer
}

func addRoute(cfg *configs.ServerTunnelConfig) {
	for _, proxy := range cfg.Proxy {
		AddRouteInfo(proxy.Id, proxy.Paths)
	}
}

func verifyCfg(cfg *configs.ServerTunnelConfig) error {
	if cfg.Proxy == nil {
		log.Fatal("proxy is nil")
		return nil
	}
	for _, proxy := range cfg.Proxy {
		if proxy.Id == "" {
			log.Fatal("proxy id is nil")
			return nil
		}
		if proxy.Paths == nil {
			log.Fatal("proxy paths is nil")
			return nil
		}
		for _, path := range proxy.Paths {
			if path == "" {
				log.Fatal("proxy path is empty")
				return nil
			}
		}
	}
	return nil
}

//	func (b *BaseTunnelServer) Reader(ch transport.Channel, _ srv.TraverseBy) {
//		buffer := defin.NewDuplexBuffer()
//		i := len(b.Managers)
//		if i > 0 {
//			for key := range b.Managers {
//				channel := b.Managers[key]
//				buffer.Copy(ch, channel)
//				break
//			}
//		} else {
//			log.Warn("Not found register tunnel channel.")
//		}
//	}

func (t *HttpTunnelServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	t.listener.conn <- ch
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (t *HttpTunnelServer) startAfter() error {
	srv.AddTunnel(t)
	t.Server.AddHandler(t)
	addr := net.JoinHostPort("", strconv.Itoa(t.Cfg.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHttpProxy(t.getRequestRoute),
		ReadHeaderTimeout: 60 * time.Second,
	}
	log.Info("Start http server:%v", t.Cfg.Port)
	t.listener = NewTcpListener()
	go func() {
		err := server.Serve(t.listener)
		if err != nil {
			log.Error("HttpTunnel server stop")
			return
		}
	}()
	return nil
}

func (t *HttpTunnelServer) getRequestRoute(req *http.Request) (net.Conn, error) {
	info := GetRouteInfo(req.URL.Path)
	if info == nil {
		return nil, errors.New("route info not found")
	}
	bindId := t.proxyToConn[info.proxyId]
	if bindId == "" {
		return nil, errors.New("proxy ID not found in proxyToConn")
	}
	ch := t.BaseTunnelServer.Managers[bindId]
	if ch == nil {
		return nil, errors.New("connection handler not found")
	}
	return ch, nil
}

func (t *HttpTunnelServer) RegisterConn(ch trp.Channel, request exchange.RegisterReqAndRsp) {
	if request.ProxyId == "" {
		log.Warn("Register http tunnel, but It' proxyId is nil")
		return
	}
	t.BaseTunnelServer.RegisterConn(ch, request)
	log.Info("Register http tunnel, proxyId: %s", request.ProxyId)
	t.proxyToConn[request.ProxyId] = request.BindId
}

func NewTcpListener() *TcpListener {
	return &TcpListener{
		conn: make(chan trp.Channel, 1),
	}
}

type TcpListener struct {
	conn chan trp.Channel
}

func (t *TcpListener) Accept() (net.Conn, error) {
	return <-t.conn, nil
}

func (t *TcpListener) Close() error {
	return nil
}

func (t *TcpListener) Addr() net.Addr {
	return (*net.TCPAddr)(nil)
}
