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

package metrics

import (
	"github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/srv"
)

type Channel struct {
	*srv.GChannel
	traffic *TunnelTraffic
}

func NewMetricsChannel(src transport.Channel, traffic *TunnelTraffic) *Channel {
	gch, _ := src.(*srv.GChannel)
	return &Channel{GChannel: gch, traffic: traffic}
}

func (c *Channel) Write(p []byte) (n int, err error) {
	if c == nil || c.GChannel == nil {
		return 0, nil
	}
	write, err := c.GChannel.Write(p)
	if c.traffic != nil && err == nil {
		c.traffic.AddOutBytes(write)
	}
	return write, err
}

func (c *Channel) Read(p []byte) (n int, err error) {
	if c == nil || c.GChannel == nil {
		return 0, nil
	}
	read, err := c.GChannel.Read(p)
	if c.traffic != nil && err == nil {
		c.traffic.AddInBytes(read)
	}
	return read, err
}

func (c *Channel) Next(pos int) ([]byte, error) {
	bytes, err := c.GChannel.Next(pos)
	if err != nil {
		return []byte{}, err
	}
	if c.traffic != nil {
		c.traffic.AddInBytes(len(bytes))
	}
	return bytes, err
}

func (c *Channel) Peek(n int) (buf []byte, err error) {
	bytes, err := c.GChannel.Peek(n)
	if err != nil {
		return []byte{}, err
	}
	if c.traffic != nil {
		c.traffic.AddInBytes(len(bytes))
	}
	return bytes, err
}
