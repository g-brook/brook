package aio

import (
	"bytes"
	"io"
	"net/http"
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

func responseToBytes(resp *http.Response) ([]byte, error) {
	// 🛡️ 为防止 resp.Body 被提前消费，我们先读出来再重置
	var bodyCopy []byte
	var err error

	if resp.Body != nil {
		bodyCopy, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// 重置 Body，让后续 Write 能读取它
		resp.Body = io.NopCloser(bytes.NewReader(bodyCopy))
	}

	// 📦 将整个 Response 写入 bytes.Buffer 中
	var buf bytes.Buffer
	err = resp.Write(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
