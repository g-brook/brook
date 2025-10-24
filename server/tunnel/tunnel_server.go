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

package tunnel

import (
	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
)

// Save all tunnels channel. port: server.
var tunnels map[int]TunnelServer

func init() {
	tunnels = make(map[int]TunnelServer)
}

// AddTunnel
//
//	@Description: Add tunnel server.
//	@param tunnel
func AddTunnel(tunnel TunnelServer) {
	tunnels[tunnel.Port()] = tunnel
}

// GetTunnel
//
//	@Description: Get TunnelServer server.
//	@param port
//	@return TunnelServer
func GetTunnel(port int) TunnelServer {
	tunnel, ok := tunnels[port]
	if ok {
		return tunnel
	}
	return nil
}

// TunnelServer
// @Description: Define TunnelServer interface.
type TunnelServer interface {
	Id() string

	// Start is start tunnel server.
	Start(protocol utils.Network) error

	//
	// Port
	//  @Description:  Get tunnel port.
	//  @return int32
	//
	Port() int

	//
	// RegisterConn
	//  @Description: Register connection to TunnelServer.
	//  @param v2 connection.
	//  @param request request.
	//
	RegisterConn(ch transport.Channel, request exchange.TRegister)

	// PutManager put tunnel manager.
	PutManager(ch transport.Channel)

	// Shutdown shutdown.
	Shutdown()
}
