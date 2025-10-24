package ordered

import (
	"encoding/json"
	"maps"
	"slices"
)

type Set[T comparable] struct {
	keys []T
	seen map[T]struct{}
}

func NewSet[T comparable](opts ...Option) *Set[T] {
	var opt option
	for _, o := range opts {
		o(&opt)
	}
	return &Set[T]{
		seen: make(map[T]struct{}, opt.capacity),
		keys: make([]T, 0, opt.capacity),
	}
}

func (s *Set[T]) Add(key T) {
	if _, ok := s.seen[key]; ok {
		return
	}
	s.seen[key] = struct{}{}
	s.keys = append(s.keys, key)
}

func (s *Set[T]) Remove(key T) {
	idx := slices.Index(s.keys, key)
	if idx != -1 {
		s.keys = slices.Delete(s.keys, idx, idx+1)
		delete(s.seen, key)
	}
}

func (s *Set[T]) Contains(key T) bool {
	_, ok := s.seen[key]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.keys)
}

func (s *Set[T]) Iter(yield func(key T) bool) {
	for _, key := range s.keys {
		if !yield(key) {
			break
		}
	}
}

func (s *Set[T]) Clear() {
	clear(s.seen)
	s.keys = s.keys[:0]
}

func (s *Set[T]) Clone() *Set[T] {
	return &Set[T]{
		seen: maps.Clone(s.seen),
		keys: slices.Clone(s.keys),
	}
}

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.keys)
}

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &s.keys)
	if err != nil {
		return err
	}
	for _, key := range s.keys {
		s.seen[key] = struct{}{}
	}
	return nil
}
