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

package wheel

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type HierarchicalWheelTimer struct {
	workerState atomic.Int32

	delayQueue *DelayQueue

	taskCount int

	timingWheel *TimingWheel

	readLock sync.RWMutex

	writeLock sync.RWMutex
}

func (h *HierarchicalWheelTimer) Add(task *TimerTask) error {
	if task == nil {
		return errors.New("task is nil")
	}
	h.readLock.Lock()
	h.start()
	defer h.readLock.Unlock()
	return nil
}

func (h *HierarchicalWheelTimer) AdvanceClock(timeoutMs int64) {
	exitC := make(chan struct{})
	defer close(exitC)
	h.delayQueue.Poll(exitC, func() int64 {
		return time.Now().UnixMilli() + timeoutMs
	})

	//select {
	//case v := <-h.delayQueue.C:
	//default:
	//}
	//for v != nil {
	//	h.writeLock.Lock()
	//	h.taskCount--
	//	h.writeLock.Unlock()
	//	v.(TaskEntity).Cancel()
	//	v = <-h.delayQueue.C
	//}
}

func (h *HierarchicalWheelTimer) Size() int {
	return h.taskCount
}

func (h *HierarchicalWheelTimer) Shutdown() {
}

func (h *HierarchicalWheelTimer) start() {
	if h.workerState.CompareAndSwap(0, 1) {
		go h.run()
	}
}

func (h *HierarchicalWheelTimer) run() {
	for {
		h.AdvanceClock(100)
	}
}
