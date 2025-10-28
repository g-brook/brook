/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iox

import (
	"io"
	"sync"

	"github.com/brook/common/threading"
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
	threading.GoSafe(func() {
		copyData(0, src, dst)
	})
	threading.GoSafe(func() {
		copyData(1, dst, src)
	})
	wait.Wait()
	for _, e := range errors2 {
		if e != nil {
			errors = append(errors, e)
		}
	}
	return
}

func SinglePipe(src io.ReadWriteCloser, dst io.ReadWriteCloser) error {
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
	threading.GoSafe(func() {
		copyData(src, dst)
	})
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
