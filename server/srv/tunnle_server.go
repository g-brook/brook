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
	"github.com/g-brook/brook/common/modules"
)

var tunnelServerPlugin = modules.ModuleID("gnet_tunnel_server_plugin")

type TServer interface {
	modules.Module

	Open(port int) BootServer
}

func init() {
	modules.RegisterModule(&TServer2GnetServer{})
}

type TServer2GnetServer struct {
	sever *GnetServer
}

func NewTServer2GnetServer() *TServer2GnetServer {
	return &TServer2GnetServer{}
}

func (T *TServer2GnetServer) Module() modules.ModuleInfo {
	return modules.ModuleInfo{
		ID:         tunnelServerPlugin,
		ModuleType: modules.TServerModule,
		New:        func() modules.Module { return NewTServer2GnetServer() },
	}
}

func (T *TServer2GnetServer) Open(port int) BootServer {
	T.sever = NewServer(port)
	return T.sever
}
