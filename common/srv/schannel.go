package srv

import (
	"bytes"
	"github.com/xtaci/smux"
	"io"
	"net"
)

type SChannel struct {
	r      io.Reader
	w      io.Writer
	stream *smux.Stream
	buf    bytes.Buffer
}

func (s *SChannel) RemoteAddr() net.Addr {
	return s.stream.RemoteAddr()
}

func (s *SChannel) LocalAddr() net.Addr {
	return s.stream.LocalAddr()
}

func NewSChannel2(stream *smux.Stream) *SChannel {
	r, w := io.Pipe()
	return &SChannel{stream: stream, r: r, w: w, buf: bytes.Buffer{}}
}

func (s *SChannel) Read(p []byte) (n int, err error) {
	return s.buf.Read(p)
}

func (s *SChannel) Write(p []byte) (n int, err error) {
	return s.stream.Write(p)
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
