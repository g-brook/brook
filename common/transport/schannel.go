package transport

import (
	"bytes"
	"github.com/xtaci/smux"
	"io"
	"net"
)

type SChannel struct {
	r      io.Reader
	w      io.Writer
	Stream *smux.Stream
	buf    bytes.Buffer
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
