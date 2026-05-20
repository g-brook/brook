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

	"github.com/g-brook/brook/common/transport"
)

type TunnelMetrics interface {
	Id() string
	Name() string
	Port() int
	Type() string
	Connections() int
	Clients() int
	Runtime() time.Time
	ClientsInfo() []transport.Channel
}

type TunnelTrafficSnapshot struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Port        int       `json:"port"`
	Connections int       `json:"connections"`
	Clients     int       `json:"clients"`
	Runtime     time.Time `json:"runtime"`
	LastSeen    time.Time `json:"last_seen"`
	AgeSeconds  float64   `json:"age_seconds"`
	InBytes     uint64    `json:"in_bytes"`
	OutBytes    uint64    `json:"out_bytes"`
	InRateBps   float64   `json:"in_rate_bps"`
	OutRateBps  float64   `json:"out_rate_bps"`
	LatencyMs   float64   `json:"latency_ms"`
}
