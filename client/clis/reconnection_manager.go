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

package clis

import (
	"sync"
	"time"

	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
)

type ReconnectFunction func() bool
type ReconnectManager struct {
	timer             *time.Timer
	reconnectInterval time.Duration
	retries           int
	isStart           bool
	lock              sync.Mutex
}

func NewReconnectionManager(t time.Duration) *ReconnectManager {
	return &ReconnectManager{
		timer:             time.NewTimer(t),
		reconnectInterval: t,
	}
}

func (r *ReconnectManager) TryReconnect(rf ReconnectFunction) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.isStart {
		return
	}
	r.isStart = true
	r.timer.Reset(r.reconnectInterval)
	threading.GoSafe(func() {
		for {
			select {
			case <-r.timer.C:
				r.retries++
				log.Info("Try reconnect %v count, now.", r.retries)
				b := rf()
				if b {
					r.isStart = false
					r.retries = 0
					return
				}
				r.timer.Reset(r.reconnectInterval)
			}

		}
	})
}
