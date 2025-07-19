package aio

import (
	"io"
)

// Pipe establishes a bidirectional data stream between two ReadWriteClosers, enabling data transfer in both directions.
// ... existing code ...
func Pipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) (errors []error) {
	errCh := make(chan error, 2)
	// copyData transfers data from src to dst in a goroutine.
	copyData := func(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
		defer func() {
			src.Close()
			dst.Close()
		}()
		err := WithBuffer(func(buf []byte) error {
			_, err := io.CopyBuffer(dst, src, buf)
			return err
		}, GetBuffPool16k())
		errCh <- err
	}
	// Start bidirectional data transfer
	go copyData(src, dst)
	go copyData(dst, src)
	errors = make([]error, 2)
	errors[0] = <-errCh
	errors[1] = <-errCh
	return errors
}

func SignPipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) error {
	errCh := make(chan error, 1)
	// copyData transfers data from src to dst in a goroutine.
	copyData := func(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
		defer func() {
			src.Close()
			dst.Close()
		}()
		err := WithBuffer(func(buf []byte) error {
			_, err := io.CopyBuffer(dst, src, buf)
			return err
		}, GetBuffPool16k())
		errCh <- err
	}
	// Start bidirectional data transfer
	go copyData(src, dst)
	return <-errCh
}

func Copy(src io.ReadWriteCloser, dst io.ReadWriteCloser) error {
	written := int64(0)
	err := WithBuffer(func(buf []byte) (err error) {
		for {
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[0:nr])
				if nw < 0 || nr < nw {
					nw = 0
				}
				written += int64(nw)
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					break
				}
			}
			if er != nil {
				if er == io.EOF {
					err = er
				}
				break
			}
		}
		return err
	}, GetBuffPool4k())
	return err
}
