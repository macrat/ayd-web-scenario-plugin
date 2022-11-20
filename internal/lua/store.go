package lua

import (
	"sync"
)

type store[T any] struct {
	sync.RWMutex

	idx    int
	values map[int]T
}

func newStore[T any]() *store[T] {
	return &store[T]{
		values: make(map[int]T),
	}
}

func (s *store[T]) Push(x T) int {
	s.Lock()
	defer s.Unlock()

	s.idx++
	s.values[s.idx] = x
	return s.idx
}

func (s *store[T]) Get(h int) T {
	s.RLock()
	defer s.RUnlock()

	return s.values[h]
}

func (s *store[T]) Pop(h int) {
	s.Lock()
	defer s.Unlock()

	delete(s.values, s.idx)
}

type syncMap[K comparable, V any] struct {
	sync.RWMutex

	values map[K]V
}

func newSyncMap[K comparable, V any]() *syncMap[K, V] {
	return &syncMap[K, V]{
		values: make(map[K]V),
	}
}

func (s *syncMap[K, V]) Set(key K, value V) {
	s.Lock()
	defer s.Unlock()
	s.values[key] = value
}

func (s *syncMap[K, V]) Get(key K) (value V, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.values[key]
	return v, ok
}

func (s *syncMap[K, V]) Pop(key K) (value V, ok bool) {
	s.Lock()
	defer s.Unlock()

	if v, ok := s.values[key]; ok {
		delete(s.values, key)
		return v, true
	}
	return
}
