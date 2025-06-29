package http

import (
	"errors"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	defin "github.com/brook/server/define"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var ANY any

type HttpTunnelServer struct {
	*tunnel.BaseTunnelServer

	listener *TcpListener

	// proxyId->bindId(chId)
	proxyToConn map[string]map[string]any

	registerLock sync.Mutex
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
	tunnelServer := &HttpTunnelServer{
		BaseTunnelServer: server,
		proxyToConn:      make(map[string]map[string]any),
	}
	server.DoStart = tunnelServer.startAfter
	server.AddHandler(tunnel.Unregister, tunnelServer.unRegisterConn)
	addRoute(server.Cfg, tunnelServer)
	return tunnelServer
}

func addRoute(cfg *configs.ServerTunnelConfig, this *HttpTunnelServer) {
	for _, proxy := range cfg.Proxy {
		AddRouteInfo(proxy.Id, proxy.Paths, this.getProxyConnection)
		this.proxyToConn[proxy.Id] = make(map[string]any, 100)
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

func (htl *HttpTunnelServer) getProxyConnection(proxyId string) (workConn net.Conn, err error) {
	channelIds, ok := htl.proxyToConn[proxyId]
	if !ok {
		return nil, errors.New("proxy Id not found in proxy connection:" + proxyId)
	}
	for s := range channelIds {
		channel := htl.BaseTunnelServer.Managers[s]
		bytes := make([]byte, 0)
		_, err := channel.Write(bytes)
		if err != nil {
			_ = channel.Close()
			log.Error("Read error:", err)
			continue
		}
		workConn = channel
		break
	}
	if workConn == nil {
		return nil, errors.New("proxy Id not found in proxy connection:" + proxyId)
	}
	return
}

func (htl *HttpTunnelServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	htl.listener.conn <- ch
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (htl *HttpTunnelServer) startAfter() error {
	srv.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	addr := net.JoinHostPort("", strconv.Itoa(htl.Cfg.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHttpProxy(htl.getRoute),
		ReadHeaderTimeout: 60 * time.Second,
	}
	log.Info("Start http server:%v", htl.Cfg.Port)
	htl.listener = NewTcpListener()
	go func() {
		err := server.Serve(htl.listener)
		if err != nil {
			log.Error("HttpTunnel server stop")
			return
		}
	}()
	return nil
}

func (htl *HttpTunnelServer) getRoute(req *http.Request) (*RouteInfo, error) {
	info := GetRouteInfo(req.URL.Path)
	if info == nil {
		return nil, errors.New("route info not found")
	}
	return info, nil
}

func (htl *HttpTunnelServer) RegisterConn(ch trp.Channel, request exchange.RegisterReqAndRsp) {
	if request.ProxyId == "" {
		log.Warn("Register http tunnel, but It' proxyId is nil")
		return
	}
	htl.registerLock.Lock()
	htl.BaseTunnelServer.RegisterConn(NewProxyConnection(ch), request)
	log.Info("Register http tunnel, proxyId: %s", request.ProxyId)
	proxies, ok := htl.proxyToConn[request.ProxyId]
	if ok {
		proxies[ch.GetId()] = ANY
	} else {
		log.Warn("Register %V not exists by http tunnelServer.", request.ProxyId)
	}
	htl.registerLock.Unlock()
	go func() {
		newRequest, _ := exchange.NewRequest(&exchange.ReqWorkConn{})
		_, _ = ch.Write(newRequest.Bytes())
		select {
		case <-ch.Done():
			log.Info("Close http tunnel, proxyId: %s", request.ProxyId)
			return
		}
	}()
}

func (htl *HttpTunnelServer) unRegisterConn(ch trp.Channel) {
	proxyId, ok := ch.GetAttr(defin.TunnelProxyId)
	if ok {
		key := proxyId.(string)
		channels := htl.proxyToConn[key]
		if channels != nil {
			delete(channels, ch.GetId())
		}
	}

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
