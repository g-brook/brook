package transport

import (
	"bytes"
	"github.com/xtaci/smux"
	"io"
	"net"
	"time"
)

type SChannel struct {
	r            io.Reader
	w            io.Writer
	Stream       *smux.Stream
	buf          bytes.Buffer
	isBindTunnel bool
}

func (s *SChannel) Close() error {
	return s.Stream.Close()
}

func (s *SChannel) SetDeadline(t time.Time) error {
	return s.Stream.SetDeadline(t)
}

func (s *SChannel) SetReadDeadline(t time.Time) error {
	return s.Stream.SetReadDeadline(t)
}

func (s *SChannel) SetWriteDeadline(t time.Time) error {
	return s.Stream.SetWriteDeadline(t)
}

func (s *SChannel) GetConn() net.Conn {
	return s.Stream
}

func (s *SChannel) RemoteAddr() net.Addr {
	return s.Stream.RemoteAddr()
}

func (s *SChannel) LocalAddr() net.Addr {
	return s.Stream.LocalAddr()
}

func NewSChannel2(stream *smux.Stream) *SChannel {
	r, w := io.Pipe()
	return &SChannel{Stream: stream, r: r, w: w, buf: bytes.Buffer{}}
}

func (s *SChannel) Read(p []byte) (n int, err error) {
	return s.buf.Read(p)
}

func (s *SChannel) Write(p []byte) (n int, err error) {
	return s.Stream.Write(p)
}

func (s *SChannel) Copy(p []byte) (n int, err error) {
	return s.buf.Write(p)
}

func (s *SChannel) GetReader() io.Reader {
	return s
}

func (s *SChannel) GetWriter() io.Writer {
	return s
}

func (s *SChannel) SetIsBindTunnel(f bool) {
	s.isBindTunnel = f
}

func (s *SChannel) IsBindTunnel() bool {
	return s.isBindTunnel
}
