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

package remote

import (
	"fmt"

	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/transport"
)

var Inserver *InServer

type TunnelCfg struct {
	RemotePort  int
	Destination string
}

func NewTunnelCfg(remotePort int, destination string) *TunnelCfg {
	return &TunnelCfg{
		RemotePort:  remotePort,
		Destination: destination,
	}
}

var OpenTunnelServerFun func(req *exchange.OpenTunnelReq, ch transport.Channel) (*TunnelCfg, error)

var ManagerTunnelServerFun func(proxyId string, ch transport.Channel) error

func OpenTunnelServer(req *exchange.OpenTunnelReq, ch transport.Channel) (*TunnelCfg, error) {
	if OpenTunnelServerFun == nil {
		log.Error("not found open tunnel function")
		return nil, fmt.Errorf("not found open tunnel function")
	}
	return OpenTunnelServerFun(req, ch)
}

func ManagerTunnelServer(req *exchange.Heartbeat, ch transport.Channel) error {
	if ManagerTunnelServerFun == nil {
		log.Error("not found manager tunnel function")
		return fmt.Errorf("not found manager tunnel function")
	}
	proxyIds := req.ProxyId
	if proxyIds == nil || len(proxyIds) == 0 {
		log.Error("proxy id is empty")
	}
	for _, id := range proxyIds {
		_ = ManagerTunnelServerFun(id, ch)
	}
	return nil
}
