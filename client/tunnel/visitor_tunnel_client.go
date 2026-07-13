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

package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/xtaci/smux"
)

type IVisitorClient interface {
	clis.TunnelClient
	OpenVisitor(session *smux.Session, localConn net.Conn) error
}

type VisitorTunnelClient struct {
	config    *configs.ClientTunnelConfig
	listener  net.Listener
	session   *smux.Session
	mu        sync.RWMutex
	closeOnce sync.Once
	closeCtx  context.Context
	closeFun  context.CancelFunc
}

func NewVisitorTunnelClient(config *configs.ClientTunnelConfig) (clis.TunnelClient, error) {
	client := VisitorTunnelClient{
		config: config,
	}
	client.closeCtx, client.closeFun = context.WithCancel(context.Background())
	return &client, nil
}

func (v *VisitorTunnelClient) GetName() string {
	return "VisitorTunnelClient"
}

func (v *VisitorTunnelClient) Open(session *smux.Session) error {
	if v.config == nil {
		return errors.New("visitor tunnel config is nil")
	}
	if v.config.Visitor == nil {
		return errors.New("visitor config is nil")
	}
	if !v.config.IsVisitorConsumer() {
		return errors.New("provider is not allowed in visitor tunnel client")
	}
	if v.config.Visitor.LocalPort <= 0 {
		return errors.New("local port is invalid")
	}
	if session == nil || session.IsClosed() {
		return errors.New("visitor session is closed")
	}
	select {
	case <-v.Done():
		return errors.New("visitor client is closed")
	default:
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.session = session
	if v.listener != nil {
		log.Info("VTCP visitor client update session proxyId:%s", v.config.ProxyId)
		return nil
	}
	listener, err := visitorListener(v.config.Visitor.LocalPort)
	if err != nil {
		return err
	}
	v.listener = listener
	log.Info("VTCP visitor client listen local:%s proxyId:%s", listener.Addr().String(), v.config.ProxyId)
	threading.GoSafe(v.acceptLoop)
	return nil
}

func (v *VisitorTunnelClient) acceptLoop() {
	for {
		ln := v.getListener()
		if ln == nil {
			return
		}
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-v.Done():
				return
			default:
			}
			log.Error("VTCP visitor accept error:%v", err)
			return
		}
		threading.GoSafe(func() {
			_ = v.openVisitorStream(conn)
		})
	}
}

func visitorListener(localPort int) (net.Listener, error) {
	addr := fmt.Sprintf(":%d", localPort)
	return net.Listen("tcp", addr)
}

func (v *VisitorTunnelClient) Done() <-chan struct{} {
	return v.closeCtx.Done()
}

func (v *VisitorTunnelClient) Close() {
	v.closeOnce.Do(func() {
		v.closeFun()
		v.mu.Lock()
		ln := v.listener
		v.listener = nil
		v.session = nil
		v.mu.Unlock()
		if ln != nil {
			_ = ln.Close()
		}
	})
}

func (v *VisitorTunnelClient) getListener() net.Listener {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.listener
}

func (v *VisitorTunnelClient) getSession() *smux.Session {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.session
}

func (v *VisitorTunnelClient) openVisitorStream(localConn net.Conn) error {
	session := v.getSession()
	if session == nil || session.IsClosed() {
		_ = localConn.Close()
		return errors.New("visitor session is closed")
	}
	sourceTunnel, err := NewVTcpTunnelClient(v.config)
	if err != nil {
		_ = localConn.Close()
		return errors.New("new vtcp tunnel client error:" + err.Error())
	}
	if err = sourceTunnel.OpenVisitor(session, localConn); err != nil {
		sourceTunnel.Close()
		_ = localConn.Close()
		return errors.New("open visitor stream error:" + err.Error())
	}
	return nil
}
