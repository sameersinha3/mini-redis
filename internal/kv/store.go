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

func (s *Store) Delete(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.store[key]; !exists {
        return false
    }
    
    delete(s.store, key)
    return true
}

func (s *Store) GetAll() map[string]string {
    s.mu.RLock()
    defer s.mu.RUnlock()

    copy := make(map[string]string, len(s.store))
    for k, v := range s.store {
        copy[k] = v
    }
    return copy
}

func (s *Store) LoadFromSnapshot(snapshot map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store = make(map[string]string, len(snapshot))
	for k, v := range snapshot {
		s.store[k] = v
	}
}