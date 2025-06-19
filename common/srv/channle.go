package srv

import "io"

type Channel interface {
	io.Reader
	io.Writer

	GetReader() io.Reader

	GetWriter() io.Writer
}
