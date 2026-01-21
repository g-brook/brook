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
	"fmt"
	"io"
	"net"
	"time"

	"github.com/brook/common/iox"
	"github.com/brook/common/log"
	"github.com/brook/common/ringbuffer"
	"github.com/brook/common/threading"
	"github.com/panjf2000/gnet/v2"
)

type TChannel struct {
	pipeWriter io.Writer
	rawConn    *GChannel

	writerBuffer *ringbuffer.RingBuffer
	comRw        io.ReadWriteCloser
	wSign        chan int
	loop         gnet.EventLoop
	ctx          context.Context
}

func NewSmuxAdapterConn(rawConn *GChannel, ctx context.Context, loop gnet.EventLoop) *TChannel {
	reader, writer := io.Pipe()
	rw := iox.NewCompressionRw(reader, rawConn)
	return &TChannel{
		rawConn:    rawConn,
		pipeWriter: writer,
		comRw:      rw,
		loop:       loop,
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
	c.wSign <- n
	return n, nil
}

func (c *TChannel) writeWRLoop() {
	_ = c.loop.Execute(c.ctx, c)
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

func (c *TChannel) Run(ctx context.Context) error {
	threading.GoSafe(func() {
		for {
			select {
			case l := <-c.wSign:
				fmt.Println(l)
				bytes := make([]byte, c.writerBuffer.Buffered())
				_, err := c.writerBuffer.Read(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
				_, err = c.comRw.Write(bytes)
				if err != nil {
					log.Error("write error: %v", err)
				}
			case <-c.rawConn.Done():
				_ = c.Close()
				return
			}
		}
	})
	return nil
}

func (c *TChannel) Copy(p []byte) (n int, err error) {
	return c.pipeWriter.Write(p)
}

func (c *TChannel) Close() error {
	if c.writerBuffer != nil {
		c.writerBuffer.Reset()
		ringbuffer.Put(c.writerBuffer)
	}
	if c.comRw != nil {
		_ = c.comRw.Close()
	}
	return c.rawConn.Close()
}

func (c *TChannel) LocalAddr() net.Addr                { return c.rawConn.LocalAddr() }
func (c *TChannel) RemoteAddr() net.Addr               { return c.rawConn.RemoteAddr() }
func (c *TChannel) SetDeadline(t time.Time) error      { return c.rawConn.SetDeadline(t) }
func (c *TChannel) SetReadDeadline(t time.Time) error  { return c.rawConn.SetReadDeadline(t) }
func (c *TChannel) SetWriteDeadline(t time.Time) error { return c.rawConn.SetWriteDeadline(t) }
