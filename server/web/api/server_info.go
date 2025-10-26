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

package api

import (
	"fmt"
	"strconv"

	"github.com/brook/server/metrics"
	"github.com/brook/server/web/sql"
)

func init() {
	RegisterRoute(NewRoute("/getServerInfo", "POST"), getServerInfo)
}

// GetServerInfo retrieves information about the server
// This function is designed to gather and return various details
// about the server's current status and configuration
func getServerInfo(req *Request[QueryServerInfo]) *Response {
	servers := metrics.M.GetServers()
	var v []*ServerInfo
	for _, item := range servers {
		newItem := sql.GetProxyConfigByProxyId(item.Id())
		if newItem == nil {
			return NewResponseSuccess(nil)
		}
		v = append(v, &ServerInfo{
			Name:        newItem.Name,
			Port:        strconv.Itoa(item.Port()),
			TunnelType:  item.Type(),
			TAG:         newItem.Tag,
			Connections: item.Connections(),
			Users:       item.Clients(),
		})
		fmt.Println(item.Name())
	}
	return NewResponseSuccess(v)
}
