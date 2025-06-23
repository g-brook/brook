package transport

import (
	"io"
	"net"
)

type Channel interface {
	io.Reader

	io.Writer

	net.Conn

	GetReader() io.Reader

	GetWriter() io.Writer

	RemoteAddr() net.Addr

	LocalAddr() net.Addr

	GetConn() net.Conn
}
