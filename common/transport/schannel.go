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

	"github.com/brook/common"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
)

// SChannel struct holds the secure channel information
// r and w are the reader and writer for the channel
// Stream is the underlying smux stream
// buf is a buffer for storing data temporarily
// isBindTunnel indicates if the channel is bound to a tunnel
type SChannel struct {
	Stream      *smux.Stream
	IsTunnel    bool
	buf         bytes.Buffer
	id          string
	ctx         context.Context
	cancel      context.CancelFunc
	attr        map[common.KeyType]interface{}
	closeEvents []CloseEvent
}

// NewSChannel creates a new SChannel with the given smux stream
// It initializes a pipe for reading and writing
func NewSChannel(stream *smux.Stream, parent context.Context, isTunnel bool) *SChannel {
	ctx, cancelFunc := context.WithCancel(parent)
	ch := &SChannel{Stream: stream,
		ctx:         ctx,
		id:          uuid.NewString(),
		cancel:      cancelFunc,
		attr:        map[common.KeyType]interface{}{},
		IsTunnel:    isTunnel,
		closeEvents: make([]CloseEvent, 0),
		buf:         bytes.Buffer{}} // Initialize as pointer
	return ch
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

func (s *SChannel) AddAttr(key common.KeyType, value interface{}) {
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

func (s *SChannel) GetAttr(key common.KeyType) (interface{}, bool) {
	value, ok := s.attr[key]
	return value, ok
}

// Read reads data into p
func (s *SChannel) Read(p []byte) (n int, err error) {
	if s.IsClose() {
		return 0, io.EOF
	}
	if s.IsTunnel {
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
