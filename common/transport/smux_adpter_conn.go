package transport

import (
	"io"
	"net"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type SmuxAdapterConn struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	rawConn gnet.Conn
}

func NewSmuxAdapterConn(rawConn gnet.Conn) *SmuxAdapterConn {
	pipe, writer := io.Pipe()
	return &SmuxAdapterConn{
		rawConn: rawConn,
		reader:  pipe,
		writer:  writer,
	}
}

// 实现 io.Reader
func (s *SmuxAdapterConn) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

// 实现 io.Writer
func (s *SmuxAdapterConn) Write(p []byte) (int, error) {
	return s.rawConn.Write(p)
}

func (s *SmuxAdapterConn) Copy(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

func (s *SmuxAdapterConn) Close() error {
	_ = s.writer.Close()
	return s.rawConn.Close()
}

func (s *SmuxAdapterConn) LocalAddr() net.Addr                { return s.rawConn.LocalAddr() }
func (s *SmuxAdapterConn) RemoteAddr() net.Addr               { return s.rawConn.RemoteAddr() }
func (s *SmuxAdapterConn) SetDeadline(t time.Time) error      { return s.rawConn.SetDeadline(t) }
func (s *SmuxAdapterConn) SetReadDeadline(t time.Time) error  { return s.rawConn.SetReadDeadline(t) }
func (s *SmuxAdapterConn) SetWriteDeadline(t time.Time) error { return s.rawConn.SetWriteDeadline(t) }
