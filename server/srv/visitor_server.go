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
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/modules"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/common/transport"
)

var VisitorServerPlugin = modules.ModuleID("visitor_server_plugin")

var errVisitorListenerClosed = errors.New("visitor listener closed")

type VisitorServer struct {
	port     int
	handlers []ServerHandler
	ln       *VisitorListener
	ctx      context.Context
	cancel   context.CancelFunc
	active   atomic.Int64
}

type VisitorListener struct {
	port          int
	sourceChannel chan net.Conn
	once          sync.Once
	closeCh       chan struct{}
}

func (v *VisitorListener) Accept() (net.Conn, error) {
	select {
	case <-v.closeCh:
		return nil, errVisitorListenerClosed
	case ch, ok := <-v.sourceChannel:
		if !ok || ch == nil {
			return nil, errVisitorListenerClosed
		}
		return ch, nil
	}
}

func (v *VisitorListener) Close() error {
	v.once.Do(func() {
		close(v.closeCh)
		close(v.sourceChannel)
	})
	return nil
}

func (v *VisitorListener) Addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4zero,
		Port: v.port,
	}
}

func (t *VisitorServer) doOpen(conn transport.Channel) error {
	return t.next(func(s ServerHandler, newCh transport.Channel) (bool, error) {
		b := true
		err := s.Open(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (t *VisitorServer) doReader(conn transport.Channel) error {
	return t.next(func(s ServerHandler, newCh transport.Channel) (bool, error) {
		b := true
		err := s.Reader(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (t *VisitorServer) doClose(conn transport.Channel) error {
	return t.next(func(s ServerHandler, newCh transport.Channel) (bool, error) {
		b := true
		err := s.Close(newCh, func() {
			b = false
		})
		return b, err
	}, conn)
}

func (t *VisitorServer) doError(conn transport.Channel, err error) {
	_ = t.next(func(s ServerHandler, newCh transport.Channel) (bool, error) {
		b := true
		s.Error(newCh, err, func() {
			b = false
		})
		return b, nil
	}, conn)
}

func (t *VisitorServer) doBoot() error {
	return t.next(func(s ServerHandler, newCh transport.Channel) (bool, error) {
		b := true
		err := s.Boot(t, func() {
			b = false
		})
		return b, err
	}, nil)
}

func (t *VisitorServer) Start(_ ...ServerOption) error {
	if t.ln == nil {
		return errors.New("visitor listener is nil")
	}
	if err := t.doBoot(); err != nil {
		return err
	}
	threading.GoSafe(func() {
		for {
			channel, err := t.ln.Accept()
			if err != nil {
				if !errors.Is(err, errVisitorListenerClosed) {
					log.Error("Visitor server accept error: %v", err)
					t.doError(nil, err)
				}
				return
			}
			trChannel, ok := channel.(transport.Channel)
			if !ok {
				_ = channel.Close()
				err = errors.New("visitor channel does not implement transport.Channel")
				log.Error("Visitor server accept error: %v", err)
				t.doError(nil, err)
				continue
			}
			t.active.Add(1)
			if err = t.doOpen(trChannel); err != nil {
				t.active.Add(-1)
				_ = trChannel.Close()
				log.Error("Visitor server open error: %v", err)
				t.doError(trChannel, err)
				continue
			}
			threading.GoSafe(func() {
				defer t.active.Add(-1)
				defer func() {
					if err := t.doClose(trChannel); err != nil {
						log.Error("Visitor server close error: %v", err)
					}
					_ = trChannel.Close()
				}()
				for {
					select {
					case <-t.ctx.Done():
						return
					default:
					}
					err := t.doReader(trChannel)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							log.Error("Visitor server reader error: %v", err)
							t.doError(trChannel, err)
						}
						return
					}
					if trChannel.IsClose() {
						return
					}
				}
			})
		}
	})
	return nil
}

func (t *VisitorServer) AddLastChannel(channel transport.Channel) error {
	if t.ln == nil {
		return errors.New("visitor listener is nil")
	}
	select {
	case <-t.ctx.Done():
		return errVisitorListenerClosed
	case <-t.ln.closeCh:
		return errVisitorListenerClosed
	case t.ln.sourceChannel <- channel:
		return nil
	}
}

func (t *VisitorServer) AddHandler(handler ...ServerHandler) {
	t.handlers = append(t.handlers, handler...)
}

func (t *VisitorServer) Shutdown(ctx context.Context) {
	_ = ctx
	if t.cancel != nil {
		t.cancel()
	}
	if t.ln != nil {
		_ = t.ln.Close()
	}
}

func (t *VisitorServer) Connections() int {
	return int(t.active.Load())
}

func (t *VisitorServer) next(fun func(s ServerHandler, conn transport.Channel) (bool, error), conn transport.Channel) error {
	for i := 0; i < len(t.handlers); i++ {
		b, err := fun(t.handlers[i], conn)
		if err != nil {
			return err
		}
		if b {
			break
		}
	}
	return nil
}

func init() {
	modules.RegisterModule(&VisitorServer{})
}

func NewVisitorServer() *VisitorServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &VisitorServer{ctx: ctx, cancel: cancel}
}

func (t *VisitorServer) Module() modules.ModuleInfo {
	return modules.ModuleInfo{
		ID:         VisitorServerPlugin,
		ModuleType: modules.TServerModule,
		New:        func() modules.Module { return NewVisitorServer() },
	}
}

func (t *VisitorServer) Open(port int) BootServer {
	t.port = port
	t.ln = &VisitorListener{
		port:          port,
		sourceChannel: make(chan net.Conn, 1000),
		closeCh:       make(chan struct{}),
	}
	if t.ctx == nil || t.cancel == nil {
		t.ctx, t.cancel = context.WithCancel(context.Background())
	}
	return t
}
