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

package logger

import "sync"

type SLogger interface{}

type PrintFunc[T SLogger] func(ft T)

type SubscribeFunc func()

type subscriber[T SLogger] struct {
	id uint64
	fn PrintFunc[T]
}

type Collector[T SLogger] interface {
	Subscribe(fn PrintFunc[T]) SubscribeFunc
	Print(ft T)
}

type collector[T SLogger] struct {
	mu          sync.RWMutex
	nextID      uint64
	subscribers []subscriber[T]
}

func newCollector[T SLogger]() *collector[T] {
	return &collector[T]{
		subscribers: make([]subscriber[T], 0),
	}
}

func (c *collector[T]) Print(ft T) {
	c.mu.RLock()
	subscribers := append([]subscriber[T](nil), c.subscribers...)
	c.mu.RUnlock()
	for _, subscriber := range subscribers {
		func(fn PrintFunc[T]) {
			defer func() {
				_ = recover()
			}()
			fn(ft)
		}(subscriber.fn)
	}
}

func (c *collector[T]) Subscribe(fn PrintFunc[T]) SubscribeFunc {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.subscribers = append(c.subscribers, subscriber[T]{
		id: id,
		fn: fn,
	})
	c.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			for i, subscriber := range c.subscribers {
				if subscriber.id == id {
					c.subscribers = append(c.subscribers[:i], c.subscribers[i+1:]...)
					return
				}
			}
		})
	}
}
