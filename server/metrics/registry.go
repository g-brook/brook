package metrics

import (
	"sync"
	"time"
)

type Registry struct {
	mu       sync.RWMutex
	servers  map[string]TunnelMetrics
	traffics map[string]*TunnelTraffic
	lastSeen map[string]time.Time
	window   time.Duration
	interval time.Duration
	ttl      time.Duration
	onRemove func(string)
}

func NewRegistry(window, interval, ttl time.Duration) *Registry {
	if interval <= 0 {
		interval = time.Second
	}
	if window < interval {
		window = interval
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Registry{
		servers:  make(map[string]TunnelMetrics),
		traffics: make(map[string]*TunnelTraffic),
		lastSeen: make(map[string]time.Time),
		window:   window,
		interval: interval,
		ttl:      ttl,
	}
}

func (r *Registry) Register(server TunnelMetrics) *TunnelTraffic {
	if server == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	traffic, ok := r.traffics[server.Id()]
	if !ok {
		traffic = NewTunnelTraffic(server.Id(), server.Port(), server.Name(), r.window, r.interval)
		r.traffics[server.Id()] = traffic
	}
	r.servers[server.Id()] = server
	r.lastSeen[server.Id()] = time.Now()
	return traffic
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.servers, id)
	delete(r.traffics, id)
	delete(r.lastSeen, id)
	if r.onRemove != nil {
		r.onRemove(id)
	}
}

func (r *Registry) Touch(id string) {
	r.mu.Lock()
	r.lastSeen[id] = time.Now()
	r.mu.Unlock()
}

func (r *Registry) GetTraffic(id string) (*TunnelTraffic, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	traffic, ok := r.traffics[id]
	return traffic, ok
}

func (r *Registry) Snapshot() []TunnelTrafficSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := time.Now()
	snapshots := make([]TunnelTrafficSnapshot, 0, len(r.servers))
	for id, server := range r.servers {
		traffic := r.traffics[id]
		snap := TunnelTrafficSnapshot{ID: id, Name: server.Name(), Type: server.Type(), Port: server.Port(), Connections: server.Connections(), Clients: server.Clients(), Runtime: server.Runtime()}
		if traffic != nil {
			inBytes, outBytes := traffic.Sum()
			inRate, outRate := traffic.Rate()
			snap.InBytes = inBytes
			snap.OutBytes = outBytes
			snap.InRateBps = inRate
			snap.OutRateBps = outRate
			snap.LatencyMs = traffic.LatencyMs()
		}
		if last, ok := r.lastSeen[id]; ok {
			snap.LastSeen = last
			snap.AgeSeconds = now.Sub(last).Seconds()
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots
}

func (r *Registry) SweepExpired() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	var expired []string
	for id, last := range r.lastSeen {
		if now.Sub(last) > r.ttl {
			expired = append(expired, id)
			delete(r.servers, id)
			delete(r.traffics, id)
			delete(r.lastSeen, id)
		}
	}
	return expired
}
