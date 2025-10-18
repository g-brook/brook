package metrics

import (
	"runtime"
	"sync"
	"time"
)

type bucket struct {
	timestamp int64
	inBytes   uint64
	outBytes  uint64
}

type TunnelTraffic struct {
	Id       string
	buckets  []bucket
	port     int
	name     string
	mu       sync.Mutex
	interval time.Duration
	size     int
}

func NewTunnelTraffic(Id string, port int, name string, window time.Duration, interval time.Duration) *TunnelTraffic {
	if window < interval {
		window = interval
	}
	size := int(window / interval)
	return &TunnelTraffic{
		Id:       Id,
		buckets:  make([]bucket, size),
		size:     size,
		interval: interval,
		port:     port,
		name:     name,
	}
}

func (ts *TunnelTraffic) AddInBytes(bytes int) {
	ts.addBytes(bytes, true)
}

func (ts *TunnelTraffic) AddOutBytes(bytes int) {
	ts.addBytes(bytes, false)
}

func (ts *TunnelTraffic) addBytes(bytes int, isIn bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now().UnixNano() / ts.interval.Nanoseconds()
	idx := int(now % int64(ts.size))

	// 如果桶过期，清零
	if ts.buckets[idx].timestamp != now {
		ts.buckets[idx] = bucket{timestamp: now}
	}
	if isIn {
		ts.buckets[idx].inBytes += uint64(bytes)
	} else {
		ts.buckets[idx].outBytes += uint64(bytes)
	}

}

// Sum calculates the total incoming and outgoing traffic in the tunnel
// It only considers the buckets within the time window defined by the interval and size
// Returns:
//   - in: total incoming traffic (bytes)
//   - out: total outgoing traffic (bytes)
func (ts *TunnelTraffic) Sum() (in uint64, out uint64) {
	// Lock the mutex to ensure thread-safe access to the buckets
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now().UnixNano() / ts.interval.Nanoseconds()
	for _, b := range ts.buckets {
		if b.timestamp >= now-int64(ts.size) {
			in += b.inBytes   // Add incoming bytes
			out += b.outBytes // Add outgoing bytes
		}
	}
	return
}

func (ts *TunnelTraffic) Print() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			in, out := ts.Sum()
			println("TunnelTraffic", ts.Id, ts.port, ts.name, in, out, runtime.NumGoroutine())
		}
	}()
}
