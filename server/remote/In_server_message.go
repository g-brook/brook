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

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
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

var OpenTunnelServerFun func(req exchange.OpenTunnelReq, ch transport.Channel) (*TunnelCfg, error)

// InWriteMessage This function takes a user ID and an InBound request as parameters and sends the request to the user's connection
func InWriteMessage(unId string, r exchange.InBound) error {
	//Get the connection associated with the user ID
	connection, ok := Inserver.GetConnection(unId)
	//If the connection exists
	if ok {
		//Create a new request from the InBound request
		request, _ := exchange.NewRequest(r)
		//Write the request to the connection
		_, err := connection.Write(request.Bytes())
		return err
	} else {
		return fmt.Errorf("not found user %s connection", unId)
	}
}

func OpenTunnelServer(req exchange.OpenTunnelReq, ch transport.Channel) (*TunnelCfg, error) {
	if OpenTunnelServerFun == nil {
		log.Error("not found open tunnel function")
		return nil, fmt.Errorf("not found open tunnel function")
	}
	return OpenTunnelServerFun(req, ch)
}
