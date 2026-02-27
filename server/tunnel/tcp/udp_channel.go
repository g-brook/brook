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

package tcp

import (
	"encoding/json"
	"net"

	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/transport"
)

type UdpSChannel struct {
	*transport.SChannel
	bucket     *exchange.TunnelBucket
	udpConnMap hash.SyncMap[string, transport.Channel]
}

func NewUdpChannel(src *transport.SChannel) *UdpSChannel {
	bucket := exchange.NewTunnelBucket(src, src.Ctx()).Run()
	channel := &UdpSChannel{
		SChannel: src,
		bucket:   bucket,
	}
	bucket.DefaultRead(channel.read)
	return channel
}

func (r *UdpSChannel) read(p *exchange.TunnelProtocol) {
	var udpPackage exchange.UdpPackage
	err := json.Unmarshal(p.Data, &udpPackage)
	if err != nil {
		return
	}
	s := udpPackage.RemoteAddress.String()
	ct, ok := r.udpConnMap.Load(s)
	if ok {
		_, _ = ct.Write(udpPackage.Data)
	}
}

func (r *UdpSChannel) AsyncWriter(data []byte, ct transport.Channel) {
	remoteAddress, ok := ct.RemoteAddr().(*net.UDPAddr)
	if !ok {
		log.Warn("It not is udp addr %s", remoteAddress.String())
		return
	}
	udpPackage := exchange.NewUdpPackage(data, nil, remoteAddress)
	jsonData, _ := json.Marshal(udpPackage)
	s := remoteAddress.String()
	_, b := r.udpConnMap.Load(s)
	if !b {
		r.udpConnMap.Store(s, ct)
	}
	_ = r.bucket.Push(jsonData, nil)
}
