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
	"net/http"

	"github.com/g-brook/brook/common/hash"
)

type Metrics struct {
	servers  *hash.SyncSet[TunnelMetrics]
	traffics *hash.SyncMap[string, *TunnelTraffic]
	registry *Registry
}

var M = newMetrics()

func newMetrics() *Metrics {
	registry := Default.registry
	return &Metrics{
		servers:  hash.NewSyncSet[TunnelMetrics](),
		traffics: hash.NewSyncMap[string, *TunnelTraffic](),
		registry: registry,
	}
}

func (m *Metrics) PutServer(server TunnelMetrics) *TunnelTraffic {
	if server == nil {
		return nil
	}
	m.servers.Add(server)
	traffic := m.registry.Register(server)
	m.PutTraffics(traffic)
	return traffic
}

func (m *Metrics) RemoveServer(server TunnelMetrics) {
	if server == nil {
		return
	}
	m.servers.Remove(server)
	m.traffics.Delete(server.Id())
	m.registry.Unregister(server.Id())
}

func (m *Metrics) GetServers() []TunnelMetrics {
	return m.servers.List()
}

func (m *Metrics) PutTraffics(traffic *TunnelTraffic) {
	if traffic == nil {
		return
	}
	m.traffics.Store(traffic.Id, traffic)
}

func (m *Metrics) GetTraffics(id string) (*TunnelTraffic, bool) {
	if traffic, ok := m.registry.GetTraffic(id); ok {
		return traffic, true
	}
	return m.traffics.Load(id)
}

func (m *Metrics) Snapshot() []TunnelTrafficSnapshot {
	return m.registry.Snapshot()
}

func (m *Metrics) PrometheusHandler() http.Handler {
	return NewPrometheusHandler(m.registry)
}
