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
// It implements the Conn interface from net package
package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/brook/common/lang"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
)

// SChannel struct holds the secure channel information
// r and w are the reader and writer for the channel
// Stream is the underlying smux stream
// buf is a buffer for storing data temporarily
// isBindTunnel indicates if the channel is bound to a tunnel
type SChannel struct {
	Stream       *smux.Stream
	IsOpenTunnel bool
	buf          bytes.Buffer
	id           string
	ctx          context.Context
	cancel       context.CancelFunc
	attr         map[lang.KeyType]interface{}
	closeEvents  []CloseEvent
	lastTime     time.Time
}

// NewSChannel creates a new SChannel with the given smux stream
// It initializes a pipe for reading and writing
func NewSChannel(stream *smux.Stream, parent context.Context, isOpenTunnel bool) *SChannel {
	ctx, cancelFunc := context.WithCancel(parent)
	ch := &SChannel{Stream: stream,
		ctx:          ctx,
		id:           uuid.NewString(),
		cancel:       cancelFunc,
		attr:         map[lang.KeyType]interface{}{},
		IsOpenTunnel: isOpenTunnel,
		closeEvents:  make([]CloseEvent, 0),
		lastTime:     time.Now(),
		buf:          bytes.Buffer{}} // Initialize as pointer
	return ch
}

func (c *SChannel) SendTo([]byte, net.Addr) (int, error) {
	return 0, nil
}

// Close closes the SChannel by closing the underlying stream
func (c *SChannel) Close() error {
	err := c.Stream.Close()
	c.cancel()
	for _, event := range c.closeEvents {
		if event != nil {
			event(c)
		}
	}
	clear(c.closeEvents)
	return err
}

// SetDeadline sets the deadline for both read and write operations
func (c *SChannel) SetDeadline(t time.Time) error {
	return c.Stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for read operations
func (c *SChannel) SetReadDeadline(t time.Time) error {
	return c.Stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for write operations
func (c *SChannel) SetWriteDeadline(t time.Time) error {
	return c.Stream.SetWriteDeadline(t)
}

// GetConn returns the underlying connection
func (c *SChannel) GetConn() net.Conn {
	return c.Stream
}

// RemoteAddr returns the remote network address
func (c *SChannel) RemoteAddr() net.Addr {
	return c.Stream.RemoteAddr()
}

// LocalAddr returns the local network address
func (c *SChannel) LocalAddr() net.Addr {
	return c.Stream.LocalAddr()
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
	case <-c.Stream.GetDieCh():
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
	if c.IsOpenTunnel {
		return c.Stream.Read(p)
	} else {
		n, err = c.buf.Read(p)
	}
	return
}

// Write writes data from p
func (c *SChannel) Write(p []byte) (n int, err error) {
	select {
	case <-c.Done():
		return 0, io.EOF
	case <-c.Stream.GetDieCh():
		return 0, io.EOF
	default:
		n, err = c.Stream.Write(p)
		c.lastTime = time.Now()
		if errors.Is(err, io.EOF) {
			_ = c.Close()
			return 0, err
		}
	}
	return
}

// Copy copies data into the buffer
func (c *SChannel) Copy(p []byte) (n int, err error) {
	return c.buf.Write(p) // Using pointer access
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
