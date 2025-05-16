package kv

import "sync"

type Store struct {
    mu    sync.RWMutex
    store map[string]string
}

func NewStore() *Store {
    return &Store{store: make(map[string]string)}
}

func (s *Store) Get(key string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    val, ok := s.store[key]
    return val, ok
}

func (s *Store) Set(key, value string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.store[key] = value
}