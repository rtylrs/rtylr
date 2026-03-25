package cacher

import (
	"sync"

	"github.com/puzpuzpuz/xsync/v3"
)

type swapMap struct {
	m       *sync.RWMutex
	current *xsync.Map
	next    *xsync.Map
}

func newSwapMap() *swapMap {
	return &swapMap{
		m:       &sync.RWMutex{},
		current: xsync.NewMap(),
		next:    xsync.NewMap(),
	}
}

func (s *swapMap) swap() {
	// Swapping blocks all reads
	s.m.Lock()
	s.current = s.next
	s.next = xsync.NewMap()
	s.m.Unlock()
}

func (s *swapMap) load(key string) (any, bool) {
	// Protect the current/next maps from being swapped while loading
	s.m.RLock()
	defer s.m.RUnlock()

	return s.current.Load(key)
}

func (s *swapMap) store(key string, value any) {
	// Protect the current/next maps from being swapped while storing
	s.m.RLock()
	defer s.m.RUnlock()

	s.current.Store(key, value)
	s.next.Store(key, value)
}
