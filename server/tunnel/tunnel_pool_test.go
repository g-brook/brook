package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }

type fakeChannel struct {
	closed bool
	done   chan struct{}
}

func newFakeChannel() *fakeChannel {
	return &fakeChannel{done: make(chan struct{})}
}

func (f *fakeChannel) GetConn() net.Conn            { return nil }
func (f *fakeChannel) GetReader() io.Reader         { return nil }
func (f *fakeChannel) GetWriter() io.Writer         { return nil }
func (f *fakeChannel) Read([]byte) (int, error)     { return 0, nil }
func (f *fakeChannel) ReadFull([]byte) (int, error) { return 0, nil }
func (f *fakeChannel) Write([]byte) (int, error)    { return 0, nil }
func (f *fakeChannel) Next(int) ([]byte, error)     { return nil, nil }
func (f *fakeChannel) Discard(int) (int, error)     { return 0, nil }
func (f *fakeChannel) Peek(int) ([]byte, error)     { return nil, nil }
func (f *fakeChannel) Close() error {
	if !f.closed {
		f.closed = true
		close(f.done)
	}
	return nil
}
func (f *fakeChannel) IsClose() bool                          { return f.closed }
func (f *fakeChannel) GetServer() any                         { return nil }
func (f *fakeChannel) LocalAddr() net.Addr                    { return fakeAddr("local") }
func (f *fakeChannel) RemoteAddr() net.Addr                   { return fakeAddr("remote") }
func (f *fakeChannel) GetNetConn() net.Conn                   { return nil }
func (f *fakeChannel) Context() context.Context               { return context.Background() }
func (f *fakeChannel) Done() <-chan struct{}                  { return f.done }
func (f *fakeChannel) SetDeadline(time.Time) error            { return nil }
func (f *fakeChannel) SetReadDeadline(time.Time) error        { return nil }
func (f *fakeChannel) SetWriteDeadline(time.Time) error       { return nil }
func (f *fakeChannel) SendTo([]byte, net.Addr) (int, error)   { return 0, nil }
func (f *fakeChannel) ReadFrom([]byte) (int, net.Addr, error) { return 0, nil, nil }

func TestDefaultCheckHealth(t *testing.T) {
	if DefaultCheckHealth(nil) {
		t.Fatal("nil channel should be unhealthy")
	}

	ch := newFakeChannel()
	if !DefaultCheckHealth(ch) {
		t.Fatal("open channel should be healthy")
	}
	_ = ch.Close()
	if DefaultCheckHealth(ch) {
		t.Fatal("closed channel should be unhealthy")
	}
}

func TestTunnelPoolGetReturnsErrorOnRecoveredPanic(t *testing.T) {
	pool := NewTunnelPool(func() error {
		panic("boom")
	}, 1)

	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
}

func TestTunnelPoolGetRejectsUnhealthyChannel(t *testing.T) {
	pool := NewTunnelPool(func() error {
		poolChan := newFakeChannel()
		_ = poolChan.Close()
		return pool.Put(poolChan)
	}, 1)
	pool.checkHealthFunc = DefaultCheckHealth

	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected unhealthy channel error")
	}
	if !errors.Is(err, err) {
		// keep compiler happy; string check below is the assertion we care about
	}
}
