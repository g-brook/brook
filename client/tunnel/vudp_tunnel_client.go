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
	"time"

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/common/transport"
	"github.com/xtaci/smux"
)

type VUdpTunnelClient struct {
	*clis.BaseTunnelClient
	session   *smux.Session
	localConn *net.UDPConn
	streams   *hash.SyncMap[string, *visitorUDPStream]
	bufSize   int
	closeOnce sync.Once
}

type visitorUDPStream struct {
	ch    *transport.SChannel
	ready chan struct{}
	once  sync.Once
	err   error
}

func NewVUdpTunnelClient(cfg *configs.ClientTunnelConfig, _ *MultipleTunnelClient) (*VUdpTunnelClient, error) {
	if cfg.UdpSize == 0 {
		cfg.UdpSize = 1500
	}
	return &VUdpTunnelClient{
		BaseTunnelClient: clis.NewBaseTunnelClient(cfg, false),
		streams:          hash.NewSyncMap[string, *visitorUDPStream](),
		bufSize:          cfg.UdpSize,
	}, nil
}

func (c *VUdpTunnelClient) GetName() string {
	return "VUdpTunnelClient"
}

func (c *VUdpTunnelClient) Open(session *smux.Session) error {
	if c.GetCfg().Visitor == nil {
		return errors.New("visitor config is nil")
	}
	if c.GetCfg().Visitor.LocalPort <= 0 {
		return errors.New("visitor local port is invalid")
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", c.GetCfg().Visitor.LocalPort))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	c.session = session
	c.localConn = conn
	log.Info("VUDP visitor client listen local:%s proxyId:%s", addr.String(), c.GetCfg().ProxyId)
	threading.GoSafe(c.readLoop)
	return nil
}

func (c *VUdpTunnelClient) Close() {
	c.closeOnce.Do(func() {
		c.BaseTunnelClient.Close()
		if c.localConn != nil {
			_ = c.localConn.Close()
		}
		c.streams.Range(func(_ string, stream *visitorUDPStream) bool {
			if stream != nil && stream.ch != nil {
				_ = stream.ch.Close()
			}
			return true
		})
		c.streams.Clear()
	})
}

func (c *VUdpTunnelClient) readLoop() {
	buf := make([]byte, c.bufSize)
	for {
		n, remoteAddr, err := c.localConn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-c.Done():
				return
			default:
			}
			log.Error("VUDP visitor read error:%v", err)
			return
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		threading.GoSafe(func() {
			c.writeToStream(remoteAddr, data)
		})
	}
}

func (c *VUdpTunnelClient) writeToStream(remoteAddr *net.UDPAddr, data []byte) {
	stream, err := c.getStream(remoteAddr)
	if err != nil {
		log.Error("VUDP visitor get stream error:%v", err)
		return
	}
	select {
	case <-stream.ready:
	case <-time.After(5 * time.Second):
		log.Error("VUDP visitor register timeout proxyId:%s", c.GetCfg().ProxyId)
		return
	case <-c.Done():
		return
	}
	if stream.err != nil {
		log.Error("VUDP visitor register error:%v", stream.err)
		return
	}
	if _, err = stream.ch.Write(data); err != nil {
		log.Error("VUDP visitor write stream error:%v", err)
		c.streams.Delete(remoteAddr.String())
		_ = stream.ch.Close()
	}
}

func (c *VUdpTunnelClient) getStream(remoteAddr *net.UDPAddr) (*visitorUDPStream, error) {
	key := remoteAddr.String()
	if stream, ok := c.streams.Load(key); ok && stream != nil && stream.ch != nil && !stream.ch.IsClose() {
		return stream, nil
	}
	if c.session == nil || c.session.IsClosed() {
		return nil, errors.New("visitor session closed")
	}
	smuxStream, err := c.session.OpenStream()
	if err != nil {
		return nil, err
	}
	channel := transport.NewSChannel(smuxStream, c.TcControl.Context(), true)
	stream := &visitorUDPStream{
		ch:    channel,
		ready: make(chan struct{}),
	}
	c.streams.Store(key, stream)
	c.registerStream(stream, remoteAddr)
	return stream, nil
}

func (c *VUdpTunnelClient) registerStream(stream *visitorUDPStream, remoteAddr *net.UDPAddr) {
	bucket := exchange.NewMessageBucket(stream.ch, stream.ch.Ctx())
	bucket.AddHandler(exchange.RegisterVisitor, func(p *exchange.Protocol, _ io.ReadWriteCloser, _ context.Context) error {
		if !p.IsSuccess() {
			stream.err = errors.New(p.RspMsg)
			stream.once.Do(func() {
				close(stream.ready)
			})
			return exchange.CloseError
		}
		stream.once.Do(func() {
			close(stream.ready)
		})
		threading.GoSafe(func() {
			c.readStream(stream, remoteAddr)
		})
		<-stream.ch.Done()
		return exchange.CloseError
	})
	bucket.Run()
	if err := bucket.PushWitchRequest(c.GetVisitorReq()); err != nil {
		stream.err = err
		stream.once.Do(func() {
			close(stream.ready)
		})
		bucket.Close()
		_ = stream.ch.Close()
	}
}

func (c *VUdpTunnelClient) readStream(stream *visitorUDPStream, remoteAddr *net.UDPAddr) {
	defer c.streams.Delete(remoteAddr.String())
	buf := make([]byte, c.bufSize)
	for {
		n, err := stream.ch.Read(buf)
		if err != nil {
			_ = stream.ch.Close()
			return
		}
		if n > 0 {
			_, _ = c.localConn.WriteToUDP(buf[:n], remoteAddr)
		}
	}
}
