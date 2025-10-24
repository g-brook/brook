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

package queue

type MemoryQueue[T any] struct {
	data chan T
}

func NewMemoryQueue[T any](capacity int) *MemoryQueue[T] {
	return &MemoryQueue[T]{
		data: make(chan T, capacity),
	}
}

func (q *MemoryQueue[T]) Push(item T) {
	q.data <- item
}

func (q *MemoryQueue[T]) TryPush(item T) bool {
	select {
	case q.data <- item:
		return true
	default:
		return false
	}
}

func (q *MemoryQueue[T]) Pop() T {
	return <-q.data
}

func (q *MemoryQueue[T]) TryPop() (T, bool) {
	select {
	case item := <-q.data:
		return item, true
	default:
		var zero T
		return zero, false
	}
}

func (q *MemoryQueue[T]) Len() int {
	return len(q.data)
}

func (q *MemoryQueue[T]) Cap() int {
	return cap(q.data)
}
