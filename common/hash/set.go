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

type Set[T comparable] struct {
	data map[T]struct{}
	obj  struct{}
}

func NewSet[T comparable](value ...T) *Set[T] {
	s := &Set[T]{
		data: make(map[T]struct{}),
		obj:  struct{}{},
	}
	for _, v := range value {
		s.Add(v)
	}
	return s

}

func (s *Set[T]) Contains(v T) bool {
	_, ok := s.data[v]
	return ok
}

func (s *Set[T]) Add(v T) {
	s.data[v] = s.obj
}

func (s *Set[T]) Remove(v T) {
	delete(s.data, v)
}

func (s *Set[T]) Len() int {
	return len(s.data)
}

func (s *Set[T]) List() []T {
	keys := make([]T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
