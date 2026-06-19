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
	"sync"
	"sync/atomic"
	"time"

	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/transport"
)

var visitorAddrIndex atomic.Uint32

type VisitorContextChannel struct {
	transport.Channel
	ctx        *ConnContext
	buf        []byte
	mu         sync.Mutex
	remoteAddr net.Addr
}

func NewVisitorContextChannel(ch transport.Channel) *VisitorContextChannel {
	ctx := NewConnContext(false, "")
	ctx.Id = ch.GetId()
	port := int(visitorAddrIndex.Add(1)%60000) + 1024
	return &VisitorContextChannel{
		Channel:    ch,
		ctx:        ctx,
		remoteAddr: &net.UDPAddr{IP: net.IPv4zero, Port: port},
	}
}

func (c *VisitorContextChannel) GetContext() *ConnContext {
	return c.ctx
}

func (c *VisitorContextChannel) GetAttr(key lang.KeyType) (interface{}, bool) {
	return c.ctx.GetAttr(key)
}

func (c *VisitorContextChannel) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *VisitorContextChannel) Peek(n int) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.fill(n); err != nil {
		return nil, err
	}
	if n < 0 || n > len(c.buf) {
		n = len(c.buf)
	}
	out := make([]byte, n)
	copy(out, c.buf[:n])
	return out, nil
}

func (c *VisitorContextChannel) Next(n int) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.fill(n); err != nil {
		return nil, err
	}
	if n < 0 || n > len(c.buf) {
		n = len(c.buf)
	}
	out := make([]byte, n)
	copy(out, c.buf[:n])
	c.buf = c.buf[n:]
	return out, nil
}

func (c *VisitorContextChannel) Discard(n int) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 0 || n > len(c.buf) {
		n = len(c.buf)
	}
	c.buf = c.buf[n:]
	return n, nil
}

func (c *VisitorContextChannel) fill(n int) error {
	for len(c.buf) == 0 || (n > 0 && len(c.buf) < n) {
		tmp := make([]byte, 65535)
		read, err := c.Channel.Read(tmp)
		if read > 0 {
			c.buf = append(c.buf, tmp[:read]...)
		}
		if err != nil {
			if len(c.buf) > 0 {
				return nil
			}
			if err == io.EOF {
				return io.EOF
			}
			return err
		}
		if n <= 0 {
			return nil
		}
	}
	return nil
}

func (c *VisitorContextChannel) LastTime() time.Time {
	return c.Channel.LastTime()
}

func (c *VisitorContextChannel) ActiveTime() time.Time {
	return c.Channel.ActiveTime()
}
