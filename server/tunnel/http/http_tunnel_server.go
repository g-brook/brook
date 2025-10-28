/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	"github.com/brook/common/filex"
	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	. "github.com/brook/common/transport"
	"github.com/brook/server/defin"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

// HttpTunnelServer is a struct that represents a HTTP tunnel server.
type HttpTunnelServer struct {
	*tunnel.BaseTunnelServer

	proxyToConn *hash.SyncMap[string, *hash.SyncMap[string, *HttpTracker]]

	registerLock sync.Mutex

	proxy *Proxy

	tlsConfig *tls.Config

	isHttps bool
}

// NewHttpTunnelServer  is a constructor function for HttpTunnelServer. It takes a pointer to BaseTunnelServer as input
// and returns a pointer to HttpTunnelServer. The constructor sets the DoStart field of BaseTunnelServer to the startAfter
// method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter the server
// processes the request. The constructor also returns a pointer to HttpTunnelServer.
func NewHttpTunnelServer(server *tunnel.BaseTunnelServer) (*HttpTunnelServer, error) {
	if server.Cfg == nil {
		log.Error("start http tunnel server error, cfg is nil")
		return nil, errors.New("cfg is nil")
	}
	if err := verifyCfg(server.Cfg); err != nil {
		log.Error("http tunnel server cfg verify is false")
		return nil, err
	}
	tunnelServer := &HttpTunnelServer{
		BaseTunnelServer: server,
		proxyToConn:      hash.NewSyncMap[string, *hash.SyncMap[string, *HttpTracker]](),
	}
	server.DoStart = tunnelServer.startAfter
	server.UpdateConfigFun = func(cfg *configs.ServerTunnelConfig) {
		formatCfg(cfg, tunnelServer)
	}
	server.AddEvent(tunnel.Unregister, tunnelServer.unRegisterConn)
	server.UpdateConfig(server.Cfg)
	return tunnelServer, nil
}

// addRoute is a function that adds route information to the HttpTunnelServer. It
func formatCfg(cfg *configs.ServerTunnelConfig, this *HttpTunnelServer) {
	RouteClean()
	for _, httpJson := range cfg.Http {
		AddRouteInfo(httpJson.Id, httpJson.Domain, httpJson.Paths, this.getProxyConnection)
		this.proxyToConn.Store(httpJson.Id, hash.NewSyncMap[string, *HttpTracker]())
	}

	if cfg.Type == lang.Https {
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
		log.Error("certFile or KeyFile is nil")
		return errors.New("certFile or KeyFile is nil")
	}
	if !filex.FileExists(cf) || !filex.FileExists(kf) {
		log.Error("certFile or KeyFile is not exist")
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
		log.Error("proxy is nil")
		return errors.New("proxy is nil")
	}
	if cfg.Type == lang.Https {
		if cfg.CertFile == "" {
			log.Error("certFile is nil")
			return errors.New("certFile is nil")
		}
		if cfg.KeyFile == "" {
			log.Error("KeyFile is nil")
			return errors.New("KeyFile is nil")
		}
	}
	for _, proxy := range cfg.Http {
		if proxy.Id == "" {
			log.Error("proxy.id is nil")
			return errors.New("proxy.id is nil")
		}
		if proxy.Paths == nil {
			log.Error("proxy.paths is nil")
			return errors.New("proxy.paths is nil")
		}
		for _, path := range proxy.Paths {
			if path == "" {
				log.Error("proxy.paths is empty")
				return errors.New("proxy.paths is nil")
			}
		}
	}
	return nil
}

// getProxyConnection is a function that returns a net.Conn object based on the httpId. It
// It returns an error if the httpId is not found.
func (htl *HttpTunnelServer) getProxyConnection(proxyId string, reqId int64) (workConn net.Conn, err error) {
	err = errors.New("proxy Id not found in proxy connection:" + proxyId)
	channelIds, ok := htl.proxyToConn.Load(proxyId)
	if !ok {
		return nil, err
	}
	var channel Channel
	var tracker *HttpTracker
	channelIds.Range(func(key string, value *HttpTracker) (shouldContinue bool) {
		channel = htl.BaseTunnelServer.TunnelChannel[key]
		tracker = value
		if channel != nil {
			return false
		}
		return true
	})
	if channel == nil || tracker == nil {
		return nil, err
	}
	bytes := make([]byte, 0)
	_, err = channel.Write(bytes)
	if err != nil {
		log.Error("Read error:", err)
		_ = channel.Close()
		return nil, errors.New("write error:" + err.Error())
	}
	_ = tracker.AddRequest(reqId)
	workConn = NewProxyConnection(channel, reqId, tracker)
	if workConn == nil {
		return nil, err
	}
	err = nil
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
	threading.GoSafe(func() {
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
	})
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
	proxies, ok := htl.proxyToConn.Load(request.GetHttpId())
	if ok {
		tracker := NewHttpTracker(ch)
		proxies.Store(ch.GetId(), tracker)
		tracker.Run()
		threading.GoSafe(func() {
			_ = htl.createConn(ch)
		})
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
	proxyId, ok := ch.GetAttr(defin.HttpIdKey)
	if ok {
		key := proxyId.(string)
		channels, ok := htl.proxyToConn.Load(key)
		if ok {
			channels.Delete(ch.GetId())
		}
	}

}
