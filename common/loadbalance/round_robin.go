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

package loadbalance

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brook/common/hash"
)

type WeightedRoundRobin struct {
	current    atomic.Int64
	weight     int
	lastUpdate time.Time
}

func newWrr(weight int, now time.Time) *WeightedRoundRobin {
	return &WeightedRoundRobin{
		weight:     weight,
		lastUpdate: now,
	}
}

func (w *WeightedRoundRobin) sel(total int64) {
	w.current.Add(-1 * total)
}

func (w *WeightedRoundRobin) add() int64 {
	return w.current.Add(int64(w.weight))
}
func (w *WeightedRoundRobin) setWeight(weight int) {
	w.weight = weight
	w.current.Store(0)
}

type RoundRobin struct {
	weighted *hash.SyncMap[string, *WeightedRoundRobin]
	lock     sync.Mutex
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		weighted: hash.NewSyncMap[string, *WeightedRoundRobin](),
	}
}

func (r *RoundRobin) Select(channels []string) string {
	if len(channels) == 0 {
		return ""
	}
	totalWeight := 0
	var maxCurrent int64
	maxCurrent = math.MinInt
	now := time.Now()
	var selectedInvoker string
	var selRr *WeightedRoundRobin
	for _, c := range channels {
		id := c
		weight := getWeight(id)
		load, b := r.weighted.Load(id)
		if !b {
			r.lock.Lock()
			load, b = r.weighted.Load(id)
			if !b {
				load = newWrr(weight, now)
				r.weighted.Store(id, load)
			}
			r.lock.Unlock()
		}
		if weight != load.weight {
			load.setWeight(weight)
		}
		cur := load.add()
		if cur > maxCurrent {
			maxCurrent = cur
			selectedInvoker = c
			selRr = load
		}
		totalWeight += weight
	}
	if len(channels) != r.weighted.Len() {
		r.weighted.Range(func(key string, value *WeightedRoundRobin) (shouldContinue bool) {
			if value.lastUpdate.Before(now.Add(-time.Minute)) {
				r.weighted.Delete(key)
			}
			return true
		})
	}
	if selectedInvoker != "" && selRr != nil {
		selRr.sel(int64(totalWeight))
		return selectedInvoker
	}
	return channels[0]

}
