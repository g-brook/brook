package aio

import (
	"io"
	"sync"
)

// Pipe establishes a bidirectional data stream between two ReadWriteClosers, enabling data transfer in both directions.
// ... existing code ...
func Pipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) (errors []error) {
	var wait sync.WaitGroup
	errors2 := make([]error, 2)
	// copyData transfers data from src to dst in a goroutine.
	copyData := func(index int, src io.ReadWriteCloser, dst io.ReadWriteCloser) {
		defer func() {
			wait.Done()
			src.Close()
			dst.Close()
		}()
		errors2[index] = WithBuffer(func(buf []byte) error {
			_, err := io.CopyBuffer(dst, src, buf)
			return err
		}, GetBytePool16k())
	}
	wait.Add(2)
	// Start bidirectional data transfer
	go copyData(0, src, dst)
	go copyData(1, dst, src)
	wait.Wait()
	for _, e := range errors2 {
		if e != nil {
			errors = append(errors, e)
		}
	}
	return
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
		}, GetBytePool16k())
		errCh <- err
	}
	// Start bidirectional data transfer
	go copyData(src, dst)
	return <-errCh
}

func Copy(src io.ReadWriteCloser, dst io.ReadWriteCloser) error {
	written := int64(0)
	return WithBuffer(func(buf []byte) (err error) {
		for {
			nr, er := src.Read(buf)
			if nr > 0 {
				bytes := buf[0:nr]
				nw, ew := dst.Write(bytes)
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
			if er != nil && er != io.EOF {
				if er == io.EOF {
					err = er
				}
				break
			}
		}
		return err
	}, GetBytePool16k())
}
