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
	"io"
	"net"
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
}

func NewSmuxAdapterConn(rawConn *GChannel) *TChannel {
	reader, writer := io.Pipe()
	rw := iox.NewCompressionRw(reader, rawConn)
	return &TChannel{
		rawConn:    rawConn,
		pipeWriter: writer,
		comRw:      rw,
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
	threading.GoSafe(func() {
		for {
			select {
			case l := <-c.wSign:
				bytes := make([]byte, l)
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
}

func (s *TChannel) Read(p []byte) (int, error) {
	n, err := s.comRw.Read(p)
	if err != nil {
		return 0, err
	}
	return n, err
}

func (s *TChannel) Write(p []byte) (int, error) {
	return s.WriteWR(p)
}

func (s *TChannel) Copy(p []byte) (n int, err error) {
	return s.pipeWriter.Write(p)
}

func (s *TChannel) Close() error {
	if s.writerBuffer != nil {
		ringbuffer.Put(s.writerBuffer)
	}
	if s.comRw != nil {
		_ = s.comRw.Close()
	}
	return s.rawConn.Close()
}

func (s *TChannel) LocalAddr() net.Addr                { return s.rawConn.LocalAddr() }
func (s *TChannel) RemoteAddr() net.Addr               { return s.rawConn.RemoteAddr() }
func (s *TChannel) SetDeadline(t time.Time) error      { return s.rawConn.SetDeadline(t) }
func (s *TChannel) SetReadDeadline(t time.Time) error  { return s.rawConn.SetReadDeadline(t) }
func (s *TChannel) SetWriteDeadline(t time.Time) error { return s.rawConn.SetWriteDeadline(t) }
