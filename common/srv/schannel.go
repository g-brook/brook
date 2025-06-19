package srv

import (
	"github.com/xtaci/smux"
	"io"
)

type SChannel struct {
	stream *smux.Stream
}

func NewSChannel(stream *smux.Stream) *SChannel {
	return &SChannel{stream: stream}
}

func (s *SChannel) Read(p []byte) (n int, err error) {
	return s.stream.Read(p)
}

func (s *SChannel) Write(p []byte) (n int, err error) {
	return s.stream.Write(p)
}

func (s *SChannel) GetReader() io.Reader {
	return s.stream
}

func (s *SChannel) GetWriter() io.Writer {
	return s.stream
}
