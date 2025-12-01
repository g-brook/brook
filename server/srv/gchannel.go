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
	"time"

	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/ringbuffer"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
	"github.com/panjf2000/gnet/v2"
)

// GChannel
// @Description:
type GChannel struct {
	Conn gnet.Conn

	Id string

	Context *ConnContext

	Server *Server

	PipeConn *SmuxAdapterConn

	bgCtx context.Context

	cancel context.CancelFunc

	closeEvents []transport.CloseEvent

	protocol lang.Network

	isDatagram bool

	writerBuffer *ringbuffer.RingBuffer

	writer chan int
}

func (c *GChannel) SendTo(by []byte, addr net.Addr) (int, error) {
	if c.protocol != lang.NetworkUdp {
		return 0, nil
	}
	return c.Conn.SendTo(by, addr)
}

// SetDeadline is a wrapper for gnet.Conn.SetDeadline.
func (c *GChannel) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *GChannel) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *GChannel) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}

func (c *GChannel) GetConn() net.Conn {
	return c.Conn
}

func (c *GChannel) StartWR() {
	c.writerBuffer = ringbuffer.Get()
	c.writer = make(chan int, 2048)
	c.writeWRLoop()
}

// GetReader
//
//	@Description: Get from gnet.conn.
//	@receiver receiver
//	@return iox.Reader
func (c *GChannel) GetReader() io.Reader {
	return c.Conn
}

// GetWriter
//
//	@Description:
//	@receiver receiver
//	@return iox.Writer
func (c *GChannel) GetWriter() io.Writer {
	return c.Conn
}

// GetContext
//
// This function returns the context of the GChannel
func (c *GChannel) GetContext() *ConnContext {
	// Return the context of the GChannel
	return c.Context
}

// OnClose CloseEvent This function takes a pointer to a GChannel
// and a function as parameters. The function does not return anything.
func (c *GChannel) OnClose(event transport.CloseEvent) {
	c.closeEvents = append(c.closeEvents, event)
}

// Reader
func (c *GChannel) Read(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	//ErrShortBuffer
	n, err := c.Conn.Read(out)
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

func (c *GChannel) WriteWR(out []byte) (int, error) {
	n, err := c.writerBuffer.Write(out)
	if err != nil {
		return 0, err
	}
	c.writer <- n
	return n, nil
}

func (c *GChannel) writeWRLoop() {
	threading.GoSafe(func() {
		for {
			select {
			case l := <-c.writer:
				bytes := make([]byte, l)
				_, err := c.writerBuffer.Read(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
				_, err = c.Conn.Write(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
			case <-c.Done():
				return
			}
		}
	})
}

// Writer
//
//	@Description:
//	@receiver receiver
//	@param out
//	@return error
func (c *GChannel) Write(out []byte) (int, error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	return c.Conn.Write(out)
}

// Next
//
//	@Description: Next()
//	@receiver receiver
//	@param pos
//	@return net.Conn
func (c *GChannel) Next(pos int) ([]byte, error) {
	return c.Conn.Next(pos)
}

func (c *GChannel) GetServer() *Server {
	return c.Server
}

func (c *GChannel) Close() error {
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	c.Context.IsClosed = true
	c.cancel()
	for _, event := range c.closeEvents {
		event(c)
	}
	clear(c.closeEvents)
	if c.writerBuffer != nil {
		ringbuffer.Put(c.writerBuffer)
	}
	return nil
}

func (c *GChannel) IsClose() bool {
	if c.Context.IsClosed {
		return true
	}
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

func (c *GChannel) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *GChannel) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *GChannel) GetNetConn() net.Conn {
	return c.Conn
}

func (c *GChannel) isConnection() bool {
	return !c.Context.IsClosed
}

func (c *GChannel) GetAttr(key lang.KeyType) (interface{}, bool) {
	i, ok := c.Context.attr[key]
	return i, ok
}

// AddAttr
//
//	@Description: Add a attr info on Conn.
//	@receiver receiver
func (receiver *ConnContext) AddAttr(key lang.KeyType, value interface{}) {
	receiver.attr[key] = value
}

// GetAttr
//
//	@Description: Get conn attr value.
//	@receiver receiver
//	@param key
//	@return interface{}
//	@return bool
func (receiver *ConnContext) GetAttr(key lang.KeyType) (interface{}, bool) {
	i, ok := receiver.attr[key]
	return i, ok
}

// LastActive
//
//	@Description:
//	@receiver receiver
func (receiver *ConnContext) LastActive() {
	receiver.lastActive = time.Now()
	//conn.context = receiver
}

func (c *GChannel) LastTime() time.Time {
	return c.Context.lastActive
}

// GetLastActive
//
//	@Description:
//	@receiver receiver
//	@return time.Time
func (receiver *ConnContext) GetLastActive() time.Time {
	return receiver.lastActive
}

func (c *GChannel) GetId() string {
	return c.Id
}

func (c *GChannel) Done() <-chan struct{} {
	return c.bgCtx.Done()
}
