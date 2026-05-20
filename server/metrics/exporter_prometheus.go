package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusExporter struct {
	registry *Registry
	inBytes  *prometheus.Desc
	outBytes *prometheus.Desc
	inRate   *prometheus.Desc
	outRate  *prometheus.Desc
	alive    *prometheus.Desc
}

func NewPrometheusExporter(registry *Registry) *PrometheusExporter {
	return &PrometheusExporter{
		registry: registry,
		inBytes:  prometheus.NewDesc("brook_tunnel_in_bytes_total", "Total inbound bytes per tunnel.", []string{"id", "name", "type", "port"}, nil),
		outBytes: prometheus.NewDesc("brook_tunnel_out_bytes_total", "Total outbound bytes per tunnel.", []string{"id", "name", "type", "port"}, nil),
		inRate:   prometheus.NewDesc("brook_tunnel_in_bytes_per_second", "Inbound bytes per second per tunnel.", []string{"id", "name", "type", "port"}, nil),
		outRate:  prometheus.NewDesc("brook_tunnel_out_bytes_per_second", "Outbound bytes per second per tunnel.", []string{"id", "name", "type", "port"}, nil),
		alive:    prometheus.NewDesc("brook_tunnel_alive", "Tunnel liveness state.", []string{"id", "name", "type", "port"}, nil),
	}
}

func (e *PrometheusExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.inBytes
	ch <- e.outBytes
	ch <- e.inRate
	ch <- e.outRate
	ch <- e.alive
}

func (e *PrometheusExporter) Collect(ch chan<- prometheus.Metric) {
	if e == nil || e.registry == nil {
		return
	}
	for _, snap := range e.registry.Snapshot() {
		labels := []string{snap.ID, snap.Name, snap.Type, strconv.Itoa(snap.Port)}
		ch <- prometheus.MustNewConstMetric(e.inBytes, prometheus.CounterValue, float64(snap.InBytes), labels...)
		ch <- prometheus.MustNewConstMetric(e.outBytes, prometheus.CounterValue, float64(snap.OutBytes), labels...)
		ch <- prometheus.MustNewConstMetric(e.inRate, prometheus.GaugeValue, snap.InRateBps, labels...)
		ch <- prometheus.MustNewConstMetric(e.outRate, prometheus.GaugeValue, snap.OutRateBps, labels...)
		ch <- prometheus.MustNewConstMetric(e.alive, prometheus.GaugeValue, 1, labels...)
	}
}

func NewPrometheusHandler(registry *Registry) http.Handler {
	pr := prometheus.NewRegistry()
	pr.MustRegister(NewPrometheusExporter(registry))
	return promhttp.HandlerFor(pr, promhttp.HandlerOpts{})
}
