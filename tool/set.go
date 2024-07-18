package tool

import (
	"github.com/goslacker/slacker/extend/mapx"
	"sync"
)

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
