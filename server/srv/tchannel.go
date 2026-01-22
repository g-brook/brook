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

package srv

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/brook/common/iox"
	"github.com/brook/common/log"
	"github.com/brook/common/ringbuffer"
	"github.com/brook/common/threading"
)

type TChannel struct {
	pipeWriter io.Writer
	rawConn    *GChannel

	writerBuffer *ringbuffer.RingBuffer
	comRw        io.ReadWriteCloser
	wSign        chan int
	ctx          context.Context
	once         sync.Once
}

func NewSmuxAdapterConn(rawConn *GChannel, ctx context.Context) *TChannel {
	reader, writer := io.Pipe()
	rw := iox.NewCompressionRw(reader, rawConn)
	return &TChannel{
		rawConn:    rawConn,
		pipeWriter: writer,
		comRw:      rw,
		ctx:        ctx,
	}
}

func (c *TChannel) StartWR() {
	c.writerBuffer = ringbuffer.Get()
	c.wSign = make(chan int, 2048)
	c.writeWRLoop()
}

func (c *TChannel) WriteWR(out []byte) (int, error) {
	n, err := c.writerBuffer.Write(out)
	if err != nil {
		return 0, err
	}
	c.wSign <- 1
	return n, nil
}

func (c *TChannel) writeWRLoop() {
	threading.GoSafe(func() {
		for {
			select {
			case _ = <-c.wSign:
				bytes := make([]byte, c.writerBuffer.Buffered())
				_, err := c.writerBuffer.Read(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
				_, err = c.comRw.Write(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
			case <-c.ctx.Done():
				_ = c.Close()
				return
			}
		}
	})
}

func (c *TChannel) Read(p []byte) (int, error) {
	n, err := c.comRw.Read(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (c *TChannel) Write(p []byte) (int, error) {
	return c.WriteWR(p)
}

func (c *TChannel) Copy(p []byte) (n int, err error) {
	return c.pipeWriter.Write(p)
}

func (c *TChannel) Close() error {
	c.once.Do(func() {
		if c.writerBuffer != nil {
			c.writerBuffer.Reset()
			ringbuffer.Put(c.writerBuffer)
		}
		if c.comRw != nil {
			_ = c.comRw.Close()
		}
		_ = c.rawConn.Close()
	})
	return nil
}

func (c *TChannel) LocalAddr() net.Addr                { return c.rawConn.LocalAddr() }
func (c *TChannel) RemoteAddr() net.Addr               { return c.rawConn.RemoteAddr() }
func (c *TChannel) SetDeadline(t time.Time) error      { return c.rawConn.SetDeadline(t) }
func (c *TChannel) SetReadDeadline(t time.Time) error  { return c.rawConn.SetReadDeadline(t) }
func (c *TChannel) SetWriteDeadline(t time.Time) error { return c.rawConn.SetWriteDeadline(t) }
