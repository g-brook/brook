package http

import (
	"errors"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/common/utils"
	defin "github.com/brook/server/define"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HttpTunnelServer is a struct that represents a HTTP tunnel server.
type HttpTunnelServer struct {
	*tunnel.BaseTunnelServer

	listener *TcpListener

	// proxyId->bindId(chId)
	proxyToConn map[string]map[string]*HttpTracker

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
		proxyToConn:      make(map[string]map[string]*HttpTracker),
	}
	server.DoStart = tunnelServer.startAfter
	server.AddEvent(tunnel.Unregister, tunnelServer.unRegisterConn)
	addRoute(server.Cfg, tunnelServer)
	return tunnelServer
}

// addRoute is a function that adds route information to the HttpTunnelServer. It
func addRoute(cfg *configs.ServerTunnelConfig, this *HttpTunnelServer) {
	for _, proxy := range cfg.Proxy {
		AddRouteInfo(proxy.Id, proxy.Domain, proxy.Paths, this.getProxyConnection)
		this.proxyToConn[proxy.Id] = make(map[string]*HttpTracker, 100)
	}
}

// verifyCfg is a function that verifies the configuration of the HttpTunnelServer. It
// It returns an error if the configuration is invalid.
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

// getProxyConnection is a function that returns a net.Conn object based on the proxyId. It
// It returns an error if the proxyId is not found.
func (htl *HttpTunnelServer) getProxyConnection(proxyId string, reqId string) (workConn *ProxyConnection, err error) {
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
		tracker := channelIds[s]
		_ = tracker.AddRequest(reqId)
		workConn = NewProxyConnection(channel, reqId, tracker)
		break
	}
	if workConn == nil {
		return nil, errors.New("proxy Id not found in proxy connection:" + proxyId)
	}
	return
}

// Open  is a method of HttpTunnelServer, which is used to process incoming requests. It
func (htl *HttpTunnelServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	htl.listener.conn <- ch
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (htl *HttpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(htl.Cfg.Port))
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHttpProxy(htl.getRoute),
		ReadHeaderTimeout: 60 * time.Second,
	}
	log.Info("HTTP tunnel server started:%v", htl.Cfg.Port)
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

// getRoute is a method of HttpTunnelServer, which is used to get the route information based on the request path.
func (htl *HttpTunnelServer) getRoute(req *http.Request) (*RouteInfo, error) {
	host := req.Host
	hosts := strings.Split(host, ":")
	info := GetRouteInfo(hosts[0], req.URL.Path)
	if info == nil {
		return nil, errors.New("route info not found")
	}
	return info, nil
}

// RegisterConn is a method of HttpTunnelServer, which is used to register a connection.
func (htl *HttpTunnelServer) RegisterConn(ch trp.Channel, request exchange.RegisterReqAndRsp) {
	if request.ProxyId == "" {
		log.Warn("Register http tunnel, but It' proxyId is nil")
		return
	}
	htl.registerLock.Lock()
	htl.BaseTunnelServer.RegisterConn(ch, request)
	log.Info("Register http tunnel, proxyId: %s", request.ProxyId)
	proxies, ok := htl.proxyToConn[request.ProxyId]
	if ok {
		tracker := NewHttpTracker(ch)
		proxies[ch.GetId()] = tracker
		tracker.Run()
		go func() {
			err := htl.createConn(ch)
			if err != nil {

			}
		}()
	} else {
		log.Warn("Register %V not exists by http tunnelServer.", request.ProxyId)
	}
	htl.registerLock.Unlock()
}

func (htl *HttpTunnelServer) createConn(ch trp.Channel) (err error) {
	req := &exchange.ReqWorkConn{
		ProxyId:      "proxy3",
		Port:         htl.Port(),
		TunnelType:   htl.Cfg.Type,
		LocalAddress: "127.0.0.1",
		UnId:         "1234333",
		Network:      utils.NetworkTcp,
	}
	request, err := exchange.NewRequest(req)
	_, err = ch.Write(request.Bytes())
	return
}

// unRegisterConn is a method of HttpTunnelServer, which is used to unregister a connection.
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

// NewTcpListener is a function that returns a pointer to TcpListener.
func NewTcpListener() *TcpListener {
	return &TcpListener{
		conn: make(chan trp.Channel, 1),
	}
}

// TcpListener is a struct that implements the net.Listener interface.
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
