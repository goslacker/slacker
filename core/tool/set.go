package tool

import (
	"sync"

	"github.com/goslacker/slacker/core/mapx"
)

func NewSet[K comparable, S []K](s S) *Set[K] {
	m := make(map[K]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return &Set[K]{
		inner: m,
	}
}

type Set[K comparable] struct {
	inner map[K]struct{}
	lock  sync.RWMutex
}

func (s *Set[K]) Add(item K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.inner[item] = struct{}{}
}

func (s *Set[K]) Slice() (ret []K) {
	return mapx.Keys(s.inner)
}
