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
)

type SmuxAdapterConn struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	rawConn *GChannel
}

func NewSmuxAdapterConn(rawConn *GChannel) *SmuxAdapterConn {
	pipe, writer := io.Pipe()
	return &SmuxAdapterConn{
		rawConn: rawConn,
		reader:  pipe,
		writer:  writer,
	}
}

func (s *SmuxAdapterConn) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *SmuxAdapterConn) Write(p []byte) (int, error) {
	return s.rawConn.WriteWR(p)
}

func (s *SmuxAdapterConn) Copy(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

func (s *SmuxAdapterConn) Close() error {
	_ = s.writer.Close()
	return s.rawConn.Close()
}

func (s *SmuxAdapterConn) LocalAddr() net.Addr                { return s.rawConn.LocalAddr() }
func (s *SmuxAdapterConn) RemoteAddr() net.Addr               { return s.rawConn.RemoteAddr() }
func (s *SmuxAdapterConn) SetDeadline(t time.Time) error      { return s.rawConn.SetDeadline(t) }
func (s *SmuxAdapterConn) SetReadDeadline(t time.Time) error  { return s.rawConn.SetReadDeadline(t) }
func (s *SmuxAdapterConn) SetWriteDeadline(t time.Time) error { return s.rawConn.SetWriteDeadline(t) }
