package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	. "github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/brook/server/defin"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

// HttpTunnelServer is a struct that represents a HTTP tunnel server.
type HttpTunnelServer struct {
	*tunnel.BaseTunnelServer

	proxyToConn map[string]map[string]*HttpTracker

	registerLock sync.Mutex

	proxy *Proxy

	tlsConfig *tls.Config

	isHttps bool
}

// NewHttpTunnelServer  is a constructor function for HttpTunnelServer. It takes a pointer to BaseTunnelServer as input
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
	server.UpdateConfigFun = func(cfg *configs.ServerTunnelConfig) {
		formatCfg(cfg, tunnelServer)
	}
	server.AddEvent(tunnel.Unregister, tunnelServer.unRegisterConn)
	server.UpdateConfig(server.Cfg)
	return tunnelServer
}

// addRoute is a function that adds route information to the HttpTunnelServer. It
func formatCfg(cfg *configs.ServerTunnelConfig, this *HttpTunnelServer) {
	RouteClean()
	for _, httpJson := range cfg.Http {
		AddRouteInfo(httpJson.Id, httpJson.Domain, httpJson.Paths, this.getProxyConnection)
		this.proxyToConn[httpJson.Id] = make(map[string]*HttpTracker, 100)
	}

	if cfg.Type == utils.Https {
		if loadTls(cfg, this) != nil {
			panic("loadTls error.")
		}
		this.isHttps = true
	}
}

func loadTls(cfg *configs.ServerTunnelConfig, this *HttpTunnelServer) error {
	kf := cfg.KeyFile
	cf := cfg.CertFile
	if kf == "" || cf == "" {
		log.Fatal("certFile or KeyFile is nil")
		return errors.New("certFile or KeyFile is nil")
	}
	if !utils.FileExists(cf) || !utils.FileExists(kf) {
		log.Fatal("certFile or KeyFile is not exist")
		return errors.New("certFile or KeyFile is not exist")
	}
	pair, _ := tls.LoadX509KeyPair(cf, kf)
	this.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{pair},
	}
	return nil
}

// verifyCfg is a function that verifies the configuration of the HttpTunnelServer. It
// It returns an error if the configuration is invalid.
func verifyCfg(cfg *configs.ServerTunnelConfig) error {
	if cfg.Http == nil {
		log.Fatal("proxy is nil")
		return errors.New("proxy is nil")
	}
	if cfg.Type == utils.Https {
		if cfg.CertFile == "" {
			log.Fatal("certFile is nil")
			return errors.New("certFile is nil")
		}
		if cfg.KeyFile == "" {
			log.Fatal("KeyFile is nil")
			return errors.New("KeyFile is nil")
		}
	}
	for _, proxy := range cfg.Http {
		if proxy.Id == "" {
			log.Fatal("proxy.id is nil")
			return errors.New("proxy.id is nil")
		}
		if proxy.Paths == nil {
			log.Fatal("proxy.paths is nil")
			return errors.New("proxy.paths is nil")
		}
		for _, path := range proxy.Paths {
			if path == "" {
				log.Fatal("proxy.paths is empty")
				return errors.New("proxy.paths is nil")
			}
		}
	}
	return nil
}

// getProxyConnection is a function that returns a net.Conn object based on the httpId. It
// It returns an error if the httpId is not found.
func (htl *HttpTunnelServer) getProxyConnection(proxyId string, reqId int64) (workConn net.Conn, err error) {
	channelIds, ok := htl.proxyToConn[proxyId]
	if !ok {
		return nil, errors.New("proxy Id not found in proxy connection:" + proxyId)
	}
	for s := range channelIds {
		channel := htl.BaseTunnelServer.TunnelChannel[s]
		bytes := make([]byte, 0)
		_, err := channel.Write(bytes)
		if err != nil {
			log.Error("Read error:", err)
			_ = channel.Close()
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

// Reader    is a method of HttpTunnelServer, which is used to process incoming requests. It
func (htl *HttpTunnelServer) Reader(ch Channel, tb srv.TraverseBy) {
	channel := ch.(srv.GContext)
	bt, err := channel.Next(-1)
	if err != nil {
		return
	}
	conn, ok := ch.GetAttr(defin.HttpChannel)
	if ok {
		conn.(*HttpConn).OnData(bt)
	}
	//skip next loop.
	tb()
}
func (htl *HttpTunnelServer) Open(ch Channel, tb srv.TraverseBy) {
	channel := ch.(srv.GContext)
	conn := newHttpConn(ch, htl.isHttps)
	channel.GetContext().AddAttr(defin.HttpChannel, conn)
	go func() {
		var rwConn net.Conn
		if htl.isHttps {
			var tlsConn *tls.Conn
			tlsConn = tls.Server(conn, htl.tlsConfig)
			errRc := newResponseWriter(tlsConn, conn, nil)
			if err := tlsConn.Handshake(); err != nil {
				log.Debug("TLS handshake failed: %v", err)
				errRc.error(err)
				_ = conn.Close()
				return
			}
			rwConn = tlsConn
		} else {
			rwConn = conn
		}
		reader := bufio.NewReader(rwConn)
		for {
			req, err := http.ReadRequest(reader)
			rc := newResponseWriter(rwConn, conn, req)
			if err != nil {
				log.Debug("Read HTTP request error: %v", err)
				rc.error(err)
				_ = rwConn.Close()
				return
			}
			htl.proxy.ServeHTTP(rc, req)
			_, _ = io.Copy(io.Discard, req.Body)
			req.Body.Close()
			rc.finish(nil)
			if req.Close || req.Header.Get("Connection") == "close" {
				_ = rwConn.Close()
				return
			}
		}
	}()
	tb()
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (htl *HttpTunnelServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	htl.proxy = NewHttpProxy(htl.getRoute)
	log.Info("HTTP tunnel server started:%v", htl.Cfg.Port)
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
func (htl *HttpTunnelServer) RegisterConn(ch Channel, request exchange.TRegister) {
	if request.GetProxyId() == "" || request.GetHttpId() == "" {
		log.Warn("Register http tunnel, but It' httpId or httpId is nil")
		return
	}
	htl.registerLock.Lock()
	htl.BaseTunnelServer.RegisterConn(ch, request)
	log.Info("Register http tunnel, httpId: %s,%s", request.GetProxyId(), request.GetHttpId())
	proxies, ok := htl.proxyToConn[request.GetHttpId()]
	if ok {
		tracker := NewHttpTracker(ch)
		proxies[ch.GetId()] = tracker
		tracker.Run()
		go func() {
			_ = htl.createConn(ch)
		}()
	} else {
		log.Warn("Register %v:%v not exists by http tunnelServer.", request.GetProxyId(), request.GetHttpId())
	}
	htl.registerLock.Unlock()
}

func (htl *HttpTunnelServer) createConn(ch Channel) (err error) {
	req := &exchange.WorkConnReq{
		ProxyId:    htl.Cfg.Id,
		RemotePort: htl.Cfg.Port,
	}
	request, err := exchange.NewRequest(req)
	_, err = ch.Write(request.Bytes())
	return
}

// unRegisterConn is a method of HttpTunnelServer, which is used to unregister a connection.
func (htl *HttpTunnelServer) unRegisterConn(ch Channel) {
	proxyId, ok := ch.GetAttr(defin.TunnelProxyId)
	if ok {
		key := proxyId.(string)
		channels := htl.proxyToConn[key]
		if channels != nil {
			delete(channels, ch.GetId())
		}
	}

}
