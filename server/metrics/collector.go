package metrics

import (
	"sync/atomic"
	"time"
)

type Collector struct {
	registry *Registry
	reporter *ConsoleReporter
}

func NewCollector(registry *Registry) *Collector {
	collector := &Collector{registry: registry}
	collector.reporter = NewConsoleReporter(registry, 30*time.Second)
	collector.reporter.Start()
	return collector
}

func (c *Collector) Register(server TunnelMetrics) *TunnelTraffic { return c.registry.Register(server) }
func (c *Collector) Unregister(id string)                         { c.registry.Unregister(id) }
func (c *Collector) Snapshot() []TunnelTrafficSnapshot            { return c.registry.Snapshot() }
func (c *Collector) GetTraffic(id string) (*TunnelTraffic, bool)  { return c.registry.GetTraffic(id) }

var Default = NewCollector(NewRegistry(1*time.Hour, 5*time.Second, 10*time.Minute))

var _ = atomic.Int64{}
