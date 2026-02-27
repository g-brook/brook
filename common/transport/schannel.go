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

// Package transport SChannel is a struct that represents a secure channel for communication
// It implements the conn interface from net package
package transport

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/g-brook/brook/common/lang"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
)

// SChannel struct holds the secure channel information
// r and w are the reader and writer for the channel
// stream is the underlying smux stream
// buf is a buffer for storing data temporarily
// isBindTunnel indicates if the channel is bound to a tunnel
type SChannel struct {
	stream       *smux.Stream
	IsOpenTunnel bool
	id           string
	ctx          context.Context
	cancel       context.CancelFunc
	attr         map[lang.KeyType]interface{}
	closeEvents  []CloseEvent
	lastTime     time.Time
	active       time.Time
	once         sync.Once
}

// NewSChannel creates a new SChannel with the given smux stream
// It initializes a pipe for reading and writing
func NewSChannel(
	stream *smux.Stream,
	parent context.Context,
	isOpenTunnel bool) *SChannel {
	ctx, cancelFunc := context.WithCancel(parent)
	ch := &SChannel{
		stream:       stream,
		ctx:          ctx,
		id:           uuid.NewString(),
		cancel:       cancelFunc,
		attr:         map[lang.KeyType]interface{}{},
		IsOpenTunnel: isOpenTunnel,
		closeEvents:  make([]CloseEvent, 0),
		lastTime:     time.Now(),
		active:       time.Now(),
	} // Initialize as pointer
	return ch
}

func (c *SChannel) SendTo([]byte, net.Addr) (int, error) {
	return 0, nil
}

// Close closes the SChannel by closing the underlying stream
func (c *SChannel) Close() error {
	c.once.Do(func() {
		_ = c.stream.Close()
		c.cancel()
		for _, event := range c.closeEvents {
			if event != nil {
				event(c)
			}
		}
		clear(c.closeEvents)
	})
	return nil
}

func (c *SChannel) ActiveTime() time.Time {
	return c.active
}

// SetDeadline sets the deadline for both read and write operations
func (c *SChannel) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for read operations
func (c *SChannel) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for write operations
func (c *SChannel) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// GetConn returns the underlying connection
func (c *SChannel) GetConn() net.Conn {
	return c.stream
}

// RemoteAddr returns the remote network address
func (c *SChannel) RemoteAddr() net.Addr {
	return c.stream.RemoteAddr()
}

// LocalAddr returns the local network address
func (c *SChannel) LocalAddr() net.Addr {
	return c.stream.LocalAddr()
}

func (c *SChannel) AddAttr(key lang.KeyType, value interface{}) {
	c.attr[key] = value
}

func (c *SChannel) OnClose(event CloseEvent) {
	c.closeEvents = append(c.closeEvents, event)
}

func (c *SChannel) IsClose() bool {
	select {
	case <-c.Done():
		return true
	case <-c.stream.GetDieCh():
		return true
	default:
		return false
	}
}

func (c *SChannel) GetAttr(key lang.KeyType) (interface{}, bool) {
	value, ok := c.attr[key]
	return value, ok
}

// Read reads data into p
func (c *SChannel) Read(p []byte) (n int, err error) {
	if c.IsClose() {
		return 0, io.EOF
	}
	c.lastTime = time.Now()
	n, err = c.stream.Read(p)
	return
}

// Write writes data from p
func (c *SChannel) Write(p []byte) (n int, err error) {
	select {
	case <-c.Done():
		return 0, io.EOF
	case <-c.stream.GetDieCh():
		return 0, io.EOF
	default:
		if len(p) > 0 {
			n, err = c.stream.Write(p)
		}
	}
	return
}

// GetReader returns the reader for this channel
func (c *SChannel) GetReader() io.Reader {
	return c
}

// GetWriter returns the writer for this channel
func (c *SChannel) GetWriter() io.Writer {
	return c
}

func (c *SChannel) GetId() string {
	return c.id
}

func (c *SChannel) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *SChannel) Ctx() context.Context {
	return c.ctx
}

func (c *SChannel) IsHealthy() bool {
	now := time.Now()
	sub := now.Sub(c.lastTime)
	return sub <= 500*time.Second
}

func (c *SChannel) LastTime() time.Time {
	return c.lastTime
}
