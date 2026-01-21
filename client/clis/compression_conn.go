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

package clis

import (
	"net"

	"github.com/brook/common/iox"
)

type CompressConn struct {
	net.Conn
	rw *iox.CompressionRw
}

func NewCompressConn(conn net.Conn) *CompressConn {
	return &CompressConn{
		Conn: conn,
		rw:   iox.NewCompressionRw(conn, conn),
	}
}

func (c *CompressConn) Read(b []byte) (n int, err error) {
	read, err := c.rw.Read(b)
	return read, err
}

func (c *CompressConn) Write(b []byte) (n int, err error) {
	return c.rw.Write(b)
}

func (c *CompressConn) Close() error {
	_ = c.Conn.Close()
	return c.rw.Close()
}
