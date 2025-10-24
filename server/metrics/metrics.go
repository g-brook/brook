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
	"time"

	"github.com/brook/common/hash"
)

type Metrics struct {
	servers  *hash.SyncSet[TunnelMetrics]
	traffics *hash.SyncMap[string, *TunnelTraffic]
}

var M = newMetrics()

func newMetrics() *Metrics {
	return &Metrics{
		servers:  hash.NewSyncSet[TunnelMetrics](),
		traffics: hash.NewSyncMap[string, *TunnelTraffic](),
	}
}

func (receiver *Metrics) PutServer(server TunnelMetrics) *TunnelTraffic {
	receiver.servers.Add(server)
	if server != nil {
		traffic := NewTunnelTraffic(server.Id(), server.Port(), server.Name(), 1*time.Hour, 5*time.Second)
		receiver.PutTraffics(traffic)
		return traffic
	}
	return nil
}

func (receiver *Metrics) RemoveServer(server TunnelMetrics) {
	receiver.servers.Remove(server)
	receiver.traffics.Delete(server.Id())
}

func (receiver *Metrics) GetServers() []TunnelMetrics {
	return receiver.servers.List()
}

func (receiver *Metrics) PutTraffics(traffic *TunnelTraffic) {
	receiver.traffics.Store(traffic.Id, traffic)
}

func (receiver *Metrics) GetTraffics(id string) {
	receiver.traffics.Load(id)
}
