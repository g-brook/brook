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
	"time"

	"github.com/brook/common/lang"
	"github.com/brook/common/transport"
	"github.com/panjf2000/gnet/v2"
)

// GChannel
// @Description:
type GChannel struct {
	conn gnet.Conn

	id string

	Context *ConnContext

	Server *Server

	PipeConn *TChannel

	bgCtx context.Context

	cancel context.CancelFunc

	closeEvents []transport.CloseEvent

	protocol lang.Network

	isDatagram bool

	once sync.Once
}

func (c *GChannel) SendTo(by []byte, addr net.Addr) (int, error) {
	if c.protocol != lang.NetworkUdp {
		return 0, nil
	}
	return c.conn.SendTo(by, addr)
}

// SetDeadline is a wrapper for gnet.conn.SetDeadline.
func (c *GChannel) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *GChannel) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *GChannel) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *GChannel) GetConn() net.Conn {
	return c.conn
}

func (c *GChannel) GetReader() io.Reader {
	return c.conn
}

func (c *GChannel) GetWriter() io.Writer {
	return c.conn
}

func (c *GChannel) GetContext() *ConnContext {
	// Return the context of the GChannel
	return c.Context
}

func (c *GChannel) OnClose(event transport.CloseEvent) {
	c.closeEvents = append(c.closeEvents, event)
}

func (c *GChannel) Read(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	//ErrShortBuffer
	n, err := c.conn.Read(out)
	if errors.Is(err, io.ErrShortBuffer) {
		//try read.
		if len(out) <= 4 {
			return 0, nil
		}
		return 0, err
	}
	return n, err
}

func (c *GChannel) ReadFull(out []byte) (int, error) {
	return io.ReadFull(c.GetReader(), out)
}

func (c *GChannel) Write(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	return c.conn.Write(out)
}

func (c *GChannel) Next(pos int) ([]byte, error) {
	return c.conn.Next(pos)
}

func (c *GChannel) GetServer() *Server {
	return c.Server
}

func (c *GChannel) Close() error {
	c.once.Do(func() {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.cancel()
		for _, event := range c.closeEvents {
			event(c)
		}
		clear(c.closeEvents)
	})
	return nil
}

func (c *GChannel) IsClose() bool {
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

func (c *GChannel) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *GChannel) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *GChannel) GetNetConn() net.Conn {
	return c.conn
}

func (c *GChannel) isConnection() bool {
	return !c.IsClose()
}

func (c *GChannel) GetAttr(key lang.KeyType) (interface{}, bool) {
	i, ok := c.Context.attr[key]
	return i, ok
}

func (receiver *ConnContext) AddAttr(key lang.KeyType, value interface{}) {
	receiver.attr[key] = value
}

func (receiver *ConnContext) GetAttr(key lang.KeyType) (interface{}, bool) {
	i, ok := receiver.attr[key]
	return i, ok
}

func (c *GChannel) LastTime() time.Time {
	return c.Context.lastActive
}

func (c *GChannel) ActiveTime() time.Time {
	return c.Context.active
}

func (receiver *ConnContext) LastActive() {
	receiver.lastActive = time.Now()
}

func (receiver *ConnContext) GetLastActive() time.Time {
	return receiver.lastActive
}

func (c *GChannel) GetId() string {
	return c.id
}

func (c *GChannel) Done() <-chan struct{} {
	return c.bgCtx.Done()
}
