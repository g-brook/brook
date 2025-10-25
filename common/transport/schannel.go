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
		buf:          bytes.Buffer{}} // Initialize as pointer
	return ch
}

func (c *SChannel) SendTo([]byte, net.Addr) (int, error) {
	return 0, nil
}

// Close closes the SChannel by closing the underlying stream
func (s *SChannel) Close() error {
	err := s.Stream.Close()
	s.cancel()
	for _, event := range s.closeEvents {
		if event != nil {
			event(s)
		}
	}
	clear(s.closeEvents)
	return err
}

// SetDeadline sets the deadline for both read and write operations
func (s *SChannel) SetDeadline(t time.Time) error {
	return s.Stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for read operations
func (s *SChannel) SetReadDeadline(t time.Time) error {
	return s.Stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for write operations
func (s *SChannel) SetWriteDeadline(t time.Time) error {
	return s.Stream.SetWriteDeadline(t)
}

// GetConn returns the underlying connection
func (s *SChannel) GetConn() net.Conn {
	return s.Stream
}

// RemoteAddr returns the remote network address
func (s *SChannel) RemoteAddr() net.Addr {
	return s.Stream.RemoteAddr()
}

// LocalAddr returns the local network address
func (s *SChannel) LocalAddr() net.Addr {
	return s.Stream.LocalAddr()
}

func (s *SChannel) AddAttr(key lang.KeyType, value interface{}) {
	s.attr[key] = value
}

func (s *SChannel) OnClose(event CloseEvent) {
	s.closeEvents = append(s.closeEvents, event)
}

func (s *SChannel) IsClose() bool {
	select {
	case <-s.Done():
		return true
	case <-s.Stream.GetDieCh():
		return true
	default:
		return false
	}
}

func (s *SChannel) GetAttr(key lang.KeyType) (interface{}, bool) {
	value, ok := s.attr[key]
	return value, ok
}

// Read reads data into p
func (s *SChannel) Read(p []byte) (n int, err error) {
	if s.IsClose() {
		return 0, io.EOF
	}
	if s.IsOpenTunnel {
		return s.Stream.Read(p)
	} else {
		n, err = s.buf.Read(p)
	}
	return
}

// Write writes data from p
func (s *SChannel) Write(p []byte) (n int, err error) {
	select {
	case <-s.Done():
		return 0, io.EOF
	case <-s.Stream.GetDieCh():
		return 0, io.EOF
	default:
		n, err = s.Stream.Write(p)
		if errors.Is(err, io.EOF) {
			_ = s.Close()
			return 0, err
		}
	}
	return
}

// Copy copies data into the buffer
func (s *SChannel) Copy(p []byte) (n int, err error) {
	return s.buf.Write(p) // Using pointer access
}

// GetReader returns the reader for this channel
func (s *SChannel) GetReader() io.Reader {
	return s
}

// GetWriter returns the writer for this channel
func (s *SChannel) GetWriter() io.Writer {
	return s
}

func (s *SChannel) GetId() string {
	return s.id
}

func (s *SChannel) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *SChannel) Ctx() context.Context {
	return s.ctx
}
