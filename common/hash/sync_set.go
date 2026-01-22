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

package hash

import "sync"

type SyncSet[T comparable] struct {
	*Set[T]
	lock sync.Mutex
}

func NewSyncSet[T comparable](value ...T) *SyncSet[T] {
	s := &SyncSet[T]{
		Set: NewSet[T](value...),
	}
	for _, v := range value {
		s.Add(v)
	}
	return s

}

func (s *SyncSet[T]) Contains(v T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.data[v]
	return ok
}

func (s *SyncSet[T]) Add(v T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[v] = s.obj
}

func (s *SyncSet[T]) Remove(v T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, v)
}

func (s *SyncSet[T]) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.data)
}

func (s *SyncSet[T]) ForEach(f func(v T) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for k := range s.data {
		b := f(k)
		if !b {
			break
		}
	}
}

func (s *SyncSet[T]) List() []T {
	s.lock.Lock()
	defer s.lock.Unlock()
	keys := make([]T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
