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
	"io"
	"net"

	"github.com/brook/common/iox"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	trp "github.com/brook/common/transport"
	"github.com/xtaci/smux"
)

type DupServer struct {
	ln                net.Listener
	handlers          []ServerHandler
	opts              *sOptions
	startTunnelServer func(conn net.Conn, option *SmuxServerOption) error
	port              int
}

func NewDupServer(port int, opt ...ServerOption) *DupServer {
	options := serverOptions(opt...)
	network := options.network
	if network == "" {
		network = lang.NetworkTcp
		options.network = network
	}
	server := &DupServer{
		opts:     options,
		handlers: make([]ServerHandler, 0),
		port:     port,
	}
	return server
}

func (sever *DupServer) Start() error {
	if !sever.isTunnelServer() {
		return errors.New("server is disabled,please use server")
	}
	sever.streamAssignment()
	addr := fmt.Sprintf(":%d", sever.port)
	listener, err := net.Listen(string(sever.opts.network), addr)
	if err != nil {
		log.Error("Server Listen %s error: %v", addr, err)
		return err
	}
	log.Info("Server Listen %s success", addr)
	sever.ln = listener
	for {
		conn, err := sever.ln.Accept()
		if err != nil {
			log.Error("Server Accept error: %v", err)
			sever.OnError(nil, err)
			break
		} else {
			sever.openStream(conn)
		}
	}
	return nil
}

func (sever *DupServer) OnOpen(conn trp.Channel) error {
	return sever.next(func(s ServerHandler, newCh trp.Channel) (bool, error) {
		b := true
		err := s.Open(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (sever *DupServer) OnRead(conn trp.Channel) error {
	return sever.next(func(s ServerHandler, newCh trp.Channel) (bool, error) {
		b := true
		err := s.Reader(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (sever *DupServer) OnError(conn trp.Channel, err error) {
	_ = sever.next(func(s ServerHandler, newCh trp.Channel) (bool, error) {
		b := true
		s.Error(newCh, err, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (sever *DupServer) OnClose(conn trp.Channel) {
	log.Debug("Close an Connection: %s", conn.RemoteAddr().String())
	_ = conn.Close()
	_ = sever.next(func(s ServerHandler, newCh trp.Channel) (bool, error) {
		b := true
		err := s.Close(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}
func (sever *DupServer) openStream(ch net.Conn) {
	log.Debug("Start smux server.%s", ch.RemoteAddr().String())
	err := sever.startTunnelServer(
		ch, nil,
	)
	if err != nil {
		log.Error("Smux start error: %v", err)
		return
	}
}

func (sever *DupServer) isDatagram() bool {
	// Check if the server is configured to use UDP network
	return sever.opts.network == lang.NetworkUdp
}

func (sever *DupServer) isTunnelServer() bool {
	// Check if the server is configured to use UDP network
	return sever.opts.withSmux != nil && sever.opts.withSmux.enable
}

func (sever *DupServer) next(fun func(s ServerHandler, conn trp.Channel) (bool, error), conn trp.Channel) error {
	for i := 0; i < len(sever.handlers); i++ {
		var newCh trp.Channel
		channelFunc := sever.opts.newChannelFunc
		if channelFunc != nil && conn != nil {
			newCh = channelFunc(conn)
		} else {
			newCh = conn
		}
		b, err := fun(sever.handlers[i], newCh)
		if err != nil {
			return err
		}
		if b {
			break
		}
	}
	return nil
}

func (sever *DupServer) streamAssignment() {
	sever.startTunnelServer = func(conn net.Conn, option *SmuxServerOption) error {
		threading.GoSafe(func() {
			config := smux.DefaultConfig()
			compressConn := iox.NewCompressConn(conn)
			session, err := smux.Server(compressConn, config)
			if err != nil {
				log.Error("Start server error. %v", err)
				_ = conn.Close()
				return
			}
			log.Debug("Start server accept stream. %s:%s", conn.LocalAddr(), conn.RemoteAddr())
			for {
				if session.IsClosed() {
					return
				}
				stream, err := session.AcceptStream()
				if err != nil {
					log.Error("session is close.PORT:%v, %v", conn.LocalAddr(), err.Error())
					_ = conn.Close()
					return
				}
				log.Info("accept success stream. %s:%s", conn.LocalAddr(), stream.RemoteAddr())
				channel := trp.NewSChannel(stream, context.Background(), false)
				err = sever.OnOpen(channel)
				if err != nil {
					if err == io.EOF {
						sever.OnClose(channel)
						return
					}
					log.Error("Tunnel Server next error. %v", err)
					sever.OnError(channel, err)
				}
				threading.GoSafe(func() {
					sever.readLoopStream(channel)
				})
				addHealthyCheckStream(channel)
			}

		})
		return nil
	}
}

func (sever *DupServer) readLoopStream(ch *trp.SChannel) {
	for {
		if ch.IsOpenTunnel {
			return
		}
		if ch.IsClose() {
			break
		}
		err := sever.OnRead(ch)
		if err != nil {
			if err != io.EOF {
				log.Debug("Tunnel Server error. %v", err)
				sever.OnError(ch, err)
			} else {
				break
			}
		}
	}
	sever.OnClose(ch)
}

func (sever *DupServer) AddHandler(handler ...ServerHandler) {
	sever.handlers = append(sever.handlers, handler...)
}

func (sever *DupServer) Shutdown(ctx context.Context) {
	log.Info("Server shutdown: %d.", sever.port)
	_ = sever.ln.Close()
}
