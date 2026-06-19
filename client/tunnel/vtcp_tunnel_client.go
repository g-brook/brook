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
	"io"
	"net"
	"sync"

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/iox"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/common/transport"
	"github.com/xtaci/smux"
)

type VTCPTunnelClient struct {
	*clis.BaseTunnelClient
	session   *smux.Session
	listener  net.Listener
	closeOnce sync.Once
}

func NewVTcpTunnelClient(config *configs.ClientTunnelConfig, _ *MultipleTunnelClient) (*VTCPTunnelClient, error) {
	tunnelClient := clis.NewBaseTunnelClient(config, false)
	client := VTCPTunnelClient{
		BaseTunnelClient: tunnelClient,
	}
	client.BaseTunnelClient.DoOpen = client.initOpen
	return &client, nil
}

func (c *VTCPTunnelClient) initOpen(ch *transport.SChannel) error {
	return c.AsyncVisitorRegister(nil, func(p *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
		if !p.IsSuccess() {
			return errors.New("vtcp tunnel open fail")
		}
		//open local server.
		return nil
	})
}

func (c *VTCPTunnelClient) GetName() string {
	return "VTcpTunnelClient"
}

func (c *VTCPTunnelClient) Open(session *smux.Session) error {
	if c.GetCfg().Visitor == nil {
		return errors.New("visitor config is nil")
	}
	if c.GetCfg().Visitor.LocalPort <= 0 {
		return errors.New("visitor local port is invalid")
	}
	c.session = session
	addr := fmt.Sprintf(":%d", c.GetCfg().Visitor.LocalPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	c.listener = ln
	log.Info("VTCP visitor client listen local:%s proxyId:%s", addr, c.GetCfg().ProxyId)
	threading.GoSafe(c.acceptLoop)
	return nil
}

func (c *VTCPTunnelClient) Close() {
	c.closeOnce.Do(func() {
		c.BaseTunnelClient.Close()
		if c.listener != nil {
			_ = c.listener.Close()
		}
	})
}

func (c *VTCPTunnelClient) acceptLoop() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			select {
			case <-c.Done():
				return
			default:
			}
			log.Error("VTCP visitor accept error:%v", err)
			return
		}
		threading.GoSafe(func() {
			c.openVisitorStream(conn)
		})
	}
}

func (c *VTCPTunnelClient) openVisitorStream(localConn net.Conn) {
	defer func() {
		_ = localConn.Close()
	}()
	if c.session == nil || c.session.IsClosed() {
		log.Error("VTCP visitor session closed proxyId:%s", c.GetCfg().ProxyId)
		return
	}
	stream, err := c.session.OpenStream()
	if err != nil {
		log.Error("VTCP visitor open stream error:%v", err)
		return
	}
	channel := transport.NewSChannel(stream, c.TcControl.Context(), true)
	bucket := exchange.NewMessageBucket(channel, channel.Ctx())
	bucket.AddHandler(exchange.RegisterVisitor, func(p *exchange.Protocol, _ io.ReadWriteCloser, _ context.Context) error {
		if !p.IsSuccess() {
			log.Error("VTCP visitor register fail:%s", p.RspMsg)
			return exchange.CloseError
		}
		errs := iox.Pipe(channel, localConn)
		if len(errs) > 0 {
			log.Debug("VTCP visitor pipe exit:%v", errs)
		}
		return exchange.CloseError
	})
	bucket.Run()
	if err = bucket.PushWitchRequest(c.GetVisitorReq()); err != nil {
		log.Error("VTCP visitor register push error:%v", err)
		bucket.Close()
		_ = channel.Close()
		return
	}
	select {
	case <-bucket.Done():
	case <-c.Done():
		bucket.Close()
		_ = channel.Close()
	}
}
