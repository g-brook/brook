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
	"github.com/brook/common/loadbalance"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	. "github.com/brook/common/transport"
	"github.com/brook/server/defin"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

// TunnelHttpServer is a struct that represents a HTTP tunnel server.
type TunnelHttpServer struct {
	*tunnel.BaseTunnelServer

	proxyToConn *hash.SyncMap[string, *hash.SyncMap[string, *Tracker]]

	registerLock sync.Mutex

	httpProxy *Proxy

	websocketProxy *WebsocketProxy

	tlsConfig *tls.Config

	isHttps bool
}

// NewHttpTunnelServer  is a constructor function for HttpTunnelServer. It takes a pointer to BaseTunnelServer as input
// and returns a pointer to HttpTunnelServer. The constructor sets the DoStart field of BaseTunnelServer to the startAfter
// method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter the server
// processes the request. The constructor also returns a pointer to HttpTunnelServer.
func NewHttpTunnelServer(server *tunnel.BaseTunnelServer) (*TunnelHttpServer, error) {
	if server.Cfg == nil {
		log.Error("start http tunnel server error, cfg is nil")
		return nil, errors.New("cfg is nil")
	}
	if err := verifyCfg(server.Cfg); err != nil {
		log.Error("http tunnel server cfg verify is false")
		return nil, err
	}
	tunnelServer := &TunnelHttpServer{
		BaseTunnelServer: server,
		proxyToConn:      hash.NewSyncMap[string, *hash.SyncMap[string, *Tracker]](),
	}
	server.DoStart = tunnelServer.startAfter
	server.UpdateConfigFun = func(cfg *configs.ServerTunnelConfig) {
		log.Info("http tunnel server config updated")
		formatCfg(cfg, tunnelServer)
	}
	server.AddEvent(tunnel.Unregister, tunnelServer.unRegisterConn)
	server.UpdateConfig(server.Cfg)
	return tunnelServer, nil
}

// addRoute is a function that adds route information to the HttpTunnelServer. It
func formatCfg(cfg *configs.ServerTunnelConfig, this *TunnelHttpServer) {
	RouteClean()
	for _, httpJson := range cfg.Http {
		AddRouteInfo(httpJson.Id, httpJson.Domain, httpJson.Paths, this.getProxyConnection)
		if _, ok := this.proxyToConn.Load(httpJson.Id); !ok {
			this.proxyToConn.Store(httpJson.Id, hash.NewSyncMap[string, *Tracker]())
		}
	}

	if cfg.Type == lang.Https {
		if loadTls(cfg, this) != nil {
			panic("loadTls error.")
		}
		this.isHttps = true
	}
}

func loadTls(cfg *configs.ServerTunnelConfig, this *TunnelHttpServer) error {
	if cfg.IsFileCert {
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
	} else {
		cert, err := tls.X509KeyPair([]byte(cfg.CertContent), []byte(cfg.KeyContent))
		if err != nil {
			log.Error("load tls error:", err)
			return err
		}
		this.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return nil

}

// verifyCfg is a function that verifies the configuration of the HttpTunnelServer. It
// It returns an error if the configuration is invalid.
func verifyCfg(cfg *configs.ServerTunnelConfig) error {
	if cfg.Http == nil {
		log.Error("http is nil")
		return errors.New("http is nil")
	}
	if cfg.Type == lang.Https {
		if cfg.IsFileCert {
			if cfg.CertFile == "" {
				log.Error("certFile is nil")
				return errors.New("certFile is nil")
			}
			if cfg.KeyFile == "" {
				log.Error("KeyFile is nil")
				return errors.New("KeyFile is nil")
			}
		} else {
			if cfg.CertContent == "" {
				log.Error("certContent is nil")
				return errors.New("certContent is nil")
			}
			if cfg.KeyContent == "" {
				log.Error("KeyContent is nil")
				return errors.New("KeyContent is nil")
			}
		}
	}
	for _, hcfg := range cfg.Http {
		if hcfg.Id == "" {
			log.Error("http.id is nil")
			return errors.New("http.id is nil")
		}
		if hcfg.Paths == nil {
			log.Error("http.paths is nil")
			return errors.New("http.paths is nil")
		}
		for _, path := range hcfg.Paths {
			if path == "" {
				log.Error("http.paths is empty")
				return errors.New("http.paths is nil")
			}
		}
	}
	return nil
}

// getProxyConnection is a function that returns a net.Conn object based on the httpId. It
// It returns an error if the httpId is not found.
func (htl *TunnelHttpServer) getProxyConnection(httpId string) (workConn net.Conn, err error) {
	err = errors.New("http Id not found in http connection:" + httpId)
	channelIds, ok := htl.proxyToConn.Load(httpId)
	var selectKeys []string
	channelIds.Range(func(key string, value *Tracker) (shouldContinue bool) {
		channel, _ := htl.TunnelChannel.Load(key)
		if channel != nil {
			selectKeys = append(selectKeys, key)
		} else {
			htl.proxyToConn.Delete(key)
		}
		return true
	})
	if !ok || len(selectKeys) == 0 {
		return nil, err
	}
	var channel Channel
	var tracker *Tracker
	key := loadbalance.Select(selectKeys)
	channel, _ = htl.TunnelChannel.Load(key)
	tracker, _ = channelIds.Load(key)
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
	workConn = NewProxyConnection(channel,
		tracker)
	if workConn == nil {
		return nil, err
	}
	err = nil
	return
}

// Reader    is a method of HttpTunnelServer, which is used to process incoming requests. It
func (htl *TunnelHttpServer) Reader(ch Channel, tb srv.TraverseBy) {
	channel := ch.(srv.GContext)
	bt, err := channel.Next(-1)
	if err != nil {
		return
	}
	conn, ok := ch.GetAttr(defin.HttpChannel)
	if ok {
		conn.(*Conn).OnData(bt)
	}
	//skip next loop.
	tb()
}
func (htl *TunnelHttpServer) Open(ch Channel, tb srv.TraverseBy) {
	channel := ch.(srv.GContext)
	httpConn := newHttpConn(ch, htl.isHttps)
	channel.GetContext().AddAttr(defin.HttpChannel, httpConn)
	threading.GoSafe(func() {
		var rwConn net.Conn
		if htl.isHttps {
			var tlsConn *tls.Conn
			tlsConn = tls.Server(httpConn, htl.tlsConfig)
			errRc := newResponseWriter(tlsConn, httpConn, nil)
			if err := tlsConn.Handshake(); err != nil {
				log.Debug("TLS handshake failed: %v", err)
				errRc.error(err)
				_ = httpConn.Close()
				return
			}
			rwConn = tlsConn
		} else {
			rwConn = httpConn
		}
		reader := bufio.NewReader(rwConn)
		for {
			req, err := http.ReadRequest(reader)
			rc := newResponseWriter(rwConn, httpConn, req)
			if err != nil {
				log.Debug("Read HTTP request error: %v", err)
				rc.error(err)
				_ = rwConn.Close()
				return
			}
			if isWebSocket(req) {
				htl.websocketProxy.ServeHTTP(rc, req)
			} else {
				htl.httpProxy.ServeHTTP(rc, req)
				_, _ = io.Copy(io.Discard, req.Body)
				_ = req.Body.Close()
				err := rc.finish(nil, req)
				if err != nil {
					_ = rwConn.Close()
					return
				}
			}
		}
	})
	tb()
}

// After is a method of HttpTunnelServer, which is used to perform cleanup or subsequent processing operations startAfter
// the server processes the request.This method currently does not perform any operation, and returns nil directly.
// This may be a reserved hook point for future additions.Parameters:
// None Return value: error, indicating the result of the execution of the operation, and always returns nil.
func (htl *TunnelHttpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	htl.httpProxy = NewHttpProxy(htl.getRoute, htl.Cfg.Id)
	htl.websocketProxy = NewWebsocketProxy(htl.getRoute)
	log.Info("Http tunnel server started:%v", htl.Cfg.Port)
	return nil
}

// getRoute is a method of HttpTunnelServer, which is used to get the route information based on the request path.
func (htl *TunnelHttpServer) getRoute(req *http.Request) (*RouteInfo, error) {
	host := req.Host
	hosts := strings.Split(host, ":")
	info := GetRouteInfo(hosts[0], req.URL.Path)
	if info == nil {
		return nil, errors.New("route info not found:" + hosts[0] + ":" + req.URL.Path)
	}
	return info, nil
}

// RegisterConn is a method of HttpTunnelServer, which is used to register a connection.
func (htl *TunnelHttpServer) RegisterConn(ch Channel, request exchange.TRegister) (serverId string, err error) {
	if request.GetProxyId() == "" || request.GetHttpId() == "" {
		log.Warn("Register http tunnel, but It' ProxyId or httpId is nil")
		return "", errors.New("ProxyId or httpId is nil")
	}
	htl.registerLock.Lock()
	defer htl.registerLock.Unlock()
	serverId, err = htl.BaseTunnelServer.RegisterConn(ch, request)
	log.Info("Register http tunnel, proxyId: %s,httpId:%s, waiting for open worker", request.GetTunnelPort(), request.GetHttpId())
	return
}
func (htl *TunnelHttpServer) OpenWorker(ch Channel, request *exchange.ClientWorkConnReq) error {
	proxies, ok := htl.proxyToConn.Load(request.HttpId)
	if ok {
		log.Info("Open Worker http tunnel, proxyId: %s,httpId:%s", request.ProxyId, request.HttpId)

	} else {
		log.Error("Open Worker %v:%v not exists by http tunnelServer.", request.ProxyId, request.HttpId)
		return errors.New("Open Worker " + request.ProxyId + ":" + request.HttpId + " not exists by http tunnelServer.")
	}
	id := request.ServerId
	userCh, b := htl.TunnelChannel.Load(id)
	if b && !ch.IsClose() {
		tracker := NewHttpTracker(userCh)
		proxies.Store(userCh.GetId(), tracker)
		tracker.Run()
		log.Info("add http tracker, proxyId: %s", request.ProxyId)
		return nil
	}
	return errors.New("channel is nil or closed")
}

// unRegisterConn is a method of HttpTunnelServer, which is used to unregister a connection.
func (htl *TunnelHttpServer) unRegisterConn(ch Channel) {
	httpId, ok := ch.GetAttr(defin.HttpIdKey)
	if ok {
		log.Debug("unRegister http tunnel, httpId: %v,channelId:%v", httpId, ch.GetId())
		key := httpId.(string)
		channels, ok := htl.proxyToConn.Load(key)
		if ok {
			channels.Delete(ch.GetId())
		}
	}

}
