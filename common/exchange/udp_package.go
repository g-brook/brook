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

package exchange

import (
	"net"
)

type UdpPackage struct {
	Data []byte `json:"data"`

	LocalAddress *net.UDPAddr `json:"local_address"`

	RemoteAddress *net.UDPAddr `json:"remote_address"`
}

func NewUdpPackage(data []byte, localAddr, remoteAddr *net.UDPAddr) *UdpPackage {
	return &UdpPackage{
		Data:          data,
		LocalAddress:  localAddr,
		RemoteAddress: remoteAddr,
	}
}

func (p *UdpPackage) GetRemoteAddress() *net.UDPAddr {
	return p.RemoteAddress
}

func (p *UdpPackage) GetLocalAddress() *net.UDPAddr {
	return p.LocalAddress
}

func (p *UdpPackage) GetData() []byte {
	return p.Data
}
