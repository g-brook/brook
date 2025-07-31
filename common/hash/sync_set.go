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
	return len(s.data)
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
