package srv

import (
	"github.com/panjf2000/gnet/v2"
	"io"
	"net"
	"time"
)

type SmuxAdapterConn struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	rawConn gnet.Conn
}

// 实现 io.Reader
func (s *SmuxAdapterConn) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

// 实现 io.Writer
func (s *SmuxAdapterConn) Write(p []byte) (int, error) {
	err := s.rawConn.AsyncWrite(p, nil)
	if err != nil {
		return 0, err
	}
	return len(p), nil
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
