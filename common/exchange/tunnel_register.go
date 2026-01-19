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
	"github.com/brook/common/lang"
)

type TRegister interface {
	Cmd() Cmd

	GetTunnelType() lang.TunnelType

	GetProxyId() string

	GetHttpId() string

	GetTunnelPort() int

	GetBindId() string

	IsOpen() bool
}

// RegisterReqAndRsp
// @Description: Register Info.
type RegisterReqAndRsp struct {
	TunnelType lang.TunnelType `json:"tunnel_type"`

	// TunnelPort is port.
	TunnelPort int `json:"tunnel_port"`

	//request id.
	BindId string `json:"bind_id"`

	//proxy id. only httpx or http.
	HttpId string `json:"http_id"`

	//proxyId.
	ProxyId string `json:"proxyId"`

	Open bool `json:"open"`
}

func (r RegisterReqAndRsp) GetTunnelPort() int {
	return r.TunnelPort
}

func (r RegisterReqAndRsp) GetBindId() string {
	return r.BindId
}

func (r RegisterReqAndRsp) GetHttpId() string {
	return r.HttpId
}

func (r RegisterReqAndRsp) GetTunnelType() lang.TunnelType {
	return r.TunnelType
}

func (r RegisterReqAndRsp) GetProxyId() string {
	return r.ProxyId
}

func (r RegisterReqAndRsp) IsOpen() bool {
	return r.Open
}

type UdpRegisterReqAndRsp struct {
	*RegisterReqAndRsp
	RemoteAddress string `json:"remote_address"`
}

func (r UdpRegisterReqAndRsp) Cmd() Cmd {
	return UdpRegister
}

func (r RegisterReqAndRsp) Cmd() Cmd {
	return Register
}
