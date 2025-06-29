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
