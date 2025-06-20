package srv

import (
	"io"
	"net"
)

type Channel interface {
	io.Reader
	io.Writer

	GetReader() io.Reader

	GetWriter() io.Writer

	RemoteAddr() net.Addr

	LocalAddr() net.Addr
}
