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
	"math"
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
	window   time.Duration
	latency  latencyStats
}

type latencyStats struct {
	lastMs float64
	sumMs  float64
	count  uint64
}

func NewTunnelTraffic(id string, port int, name string, window time.Duration, interval time.Duration) *TunnelTraffic {
	if interval <= 0 {
		interval = time.Second
	}
	if window < interval {
		window = interval
	}
	size := int(window / interval)
	if size < 1 {
		size = 1
	}
	return &TunnelTraffic{
		Id:       id,
		buckets:  make([]bucket, size),
		size:     size,
		interval: interval,
		window:   window,
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
	if ts == nil || bytes <= 0 {
		return
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now().UnixNano() / ts.interval.Nanoseconds()
	idx := int(now % int64(ts.size))
	if ts.buckets[idx].timestamp != now {
		ts.buckets[idx] = bucket{timestamp: now}
	}
	if isIn {
		ts.buckets[idx].inBytes += uint64(bytes)
		return
	}
	ts.buckets[idx].outBytes += uint64(bytes)
}

func (ts *TunnelTraffic) Sum() (in uint64, out uint64) {
	if ts == nil {
		return 0, 0
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now().UnixNano() / ts.interval.Nanoseconds()
	oldest := now - int64(ts.size)
	for _, b := range ts.buckets {
		if b.timestamp >= oldest {
			in += b.inBytes
			out += b.outBytes
		}
	}
	return
}

func (ts *TunnelTraffic) Rate() (float64, float64) {
	in, out := ts.Sum()
	seconds := ts.window.Seconds()
	if seconds <= 0 {
		seconds = ts.interval.Seconds()
	}
	return float64(in) / seconds, float64(out) / seconds
}

func (ts *TunnelTraffic) InRateBps() float64 {
	in, _ := ts.Rate()
	return in
}

func (ts *TunnelTraffic) OutRateBps() float64 {
	_, out := ts.Rate()
	return out
}

func (ts *TunnelTraffic) Port() int {
	if ts == nil {
		return 0
	}
	return ts.port
}

func (ts *TunnelTraffic) ObserveLatency(d time.Duration) {
	if ts == nil || d <= 0 {
		return
	}
	ms := float64(d) / float64(time.Millisecond)
	if math.IsNaN(ms) || math.IsInf(ms, 0) {
		return
	}
	ts.mu.Lock()
	ts.latency.lastMs = ms
	ts.latency.sumMs += ms
	ts.latency.count++
	ts.mu.Unlock()
}

func (ts *TunnelTraffic) LatencyMs() float64 {
	if ts == nil {
		return 0
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.latency.count == 0 {
		return 0
	}
	return ts.latency.lastMs
}

func (ts *TunnelTraffic) Name() string {
	if ts == nil {
		return ""
	}
	return ts.name
}
