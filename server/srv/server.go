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

package srv

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/panjf2000/gnet/v2"
)

var connLock sync.Mutex

type TraverseBy func()

type InitConnHandler func(conn *GChannel)

func NewChannel(conn gnet.Conn, t *Server) *GChannel {
	ctx := conn.Context()
	if ctx == nil && t.isDatagram() {
		ctx = NewConnContext(t.isDatagram(), conn.RemoteAddr().String())
	}
	connContext := ctx.(*ConnContext)
	value, ok := t.connections.Load(connContext.Id)
	if ok {
		return value
	}
	connLock.Lock()
	defer connLock.Unlock()
	value, ok = t.connections.Load(connContext.Id)
	if ok {
		return value
	}
	bgCtx, cancelFunc := context.WithCancel(context.Background())
	gn := &GChannel{
		conn:        conn,
		id:          connContext.Id,
		Context:     connContext,
		Server:      t,
		bgCtx:       bgCtx,
		cancel:      cancelFunc,
		protocol:    t.opts.network,
		closeEvents: make([]trp.CloseEvent, 0),
		isDatagram:  t.isDatagram(),
	}
	if !t.isDatagram() {
		t.connections.Store(connContext.Id, gn)
	}
	if t.InitConnHandler != nil {
		t.InitConnHandler(gn)
	}
	return gn
}

// Server /*
type Server struct {
	*gnet.BuiltinEventEngine

	engine gnet.Engine

	// port.
	port int

	opts *sOptions

	connections *hash.SyncMap[string, *GChannel]

	handlers []ServerHandler

	InitConnHandler InitConnHandler
}

func NewServer(port int) *Server {
	return &Server{
		port:        port,
		connections: hash.NewSyncMap[string, *GChannel](),
		handlers:    make([]ServerHandler, 0),
	}
}

func (sever *Server) AddHandler(handler ...ServerHandler) {
	sever.handlers = append(sever.handlers, handler...)
}

func (sever *Server) AddInitConnHandler(init InitConnHandler) {
	sever.InitConnHandler = init
}

func (sever *Server) Connections() map[string]*GChannel {
	tb := make(map[string]*GChannel)
	f := func(key string, value *GChannel) bool {
		tb[key] = value
		return true
	}
	sever.connections.Range(f)
	return tb
}

func (sever *Server) GetConnection(id string) (*GChannel, bool) {
	if sever.isDatagram() {
		log.Warn("server protocol is udp, can not get connection by id: %s", id)
		return nil, false
	}
	v2, ok := sever.connections.Load(id)
	return v2, ok
}
func (sever *Server) OnBoot(engine gnet.Engine) (action gnet.Action) {
	sever.engine = engine
	log.Info("Server started %d", sever.port)
	sever.next(func(s ServerHandler, conn trp.Channel) bool {
		b := true
		_ = s.Boot(sever, func() {
			b = false
		})
		return b
	}, nil)
	return gnet.None
}

func (sever *Server) OnClose(c gnet.Conn, _ error) gnet.Action {
	log.Debug("Close an Connection: %s", c.RemoteAddr().String())
	conn := NewChannel(c, sever)
	defer sever.removeIfConnection(conn)
	_ = conn.Close()
	sever.next(func(s ServerHandler, newCh trp.Channel) bool {
		b := true
		_ = s.Close(newCh, func() {
			b = false
		})
		return b
	}, conn)

	return gnet.None
}

func (sever *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Debug("Open an Connection: %s", c.RemoteAddr().String())
	c.SetContext(NewConnContext(false, ""))
	conn := NewChannel(c, sever)
	defer sever.removeIfConnection(conn)
	sever.next(func(s ServerHandler, newCh trp.Channel) bool {
		b := true
		_ = s.Open(newCh, func() {
			b = false
		})
		return b
	}, conn)
	return nil, gnet.None
}

func (sever *Server) OnTraffic(c gnet.Conn) gnet.Action {
	conn := NewChannel(c, sever)
	if conn == nil {
		log.Warn("connection is nil, can not open connection")
		return gnet.Close
	}
	defer sever.removeIfConnection(conn)
	conn.GetContext().LastActive()
	sever.next(func(s ServerHandler, newCh trp.Channel) bool {
		b := true
		_ = s.Reader(newCh, func() {
			b = false
		})
		return b
	}, conn)
	return gnet.None
}

func (sever *Server) next(fun func(s ServerHandler, conn trp.Channel) bool, conn *GChannel) {
	for i := 0; i < len(sever.handlers); i++ {
		var newCh trp.Channel
		channelFunc := sever.opts.newChannelFunc
		if channelFunc != nil && conn != nil {
			newCh = channelFunc(conn)
		} else {
			newCh = conn
		}
		b := fun(sever.handlers[i], newCh)
		if b {
			break
		}
	}
}

// isDatagram checks if the server is using UDP protocol and initializes the event engine accordingly
func (sever *Server) isDatagram() bool {
	// Check if the server is configured to use UDP network
	return sever.opts.network == lang.NetworkUdp
}

func (sever *Server) GetPort() int {
	return sever.port
}

func (sever *Server) removeIfConnection(v2 *GChannel) {
	if v2.IsClose() {
		//This use v2.id removing map element.
		// v2.id eq context.id, so yet use v2.id.
		//Because v2.context possible is nil.
		_, ok := sever.connections.Load(v2.GetId())
		if ok {
			sever.connections.Delete(v2.GetId())
		}
		_ = v2.Close()
	}
}

// Start is function start tcp server.
func (sever *Server) Start(opt ...ServerOption) error {
	//load sOptions configs.
	sever.opts = serverOptions(opt...)
	network := sever.opts.network
	if network == "" {
		network = lang.NetworkTcp
		sever.opts.network = network
	}
	if sever.opts.withSmux != nil {
		return errors.New("server is disabled, please use dup_server")
	}
	err := gnet.Run(sever, fmt.Sprintf("%s://:%d", network, sever.port),
		gnet.WithMulticore(true),
		gnet.WithLogger(&log.GnetLogger{}),
		gnet.WithReadBufferCap(65535),
		gnet.WithWriteBufferCap(65535),
		gnet.WithReusePort(true),
		gnet.WithReuseAddr(true),
	)
	if err != nil {
		log.Error("Error %v", err)
		return err
	}
	return nil
}

func (sever *Server) Shutdown(ctx context.Context) {
	log.Info("Server shutdown: %d.", sever.GetPort())
	_ = sever.engine.Stop(ctx)
}
