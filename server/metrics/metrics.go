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
