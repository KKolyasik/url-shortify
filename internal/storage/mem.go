package storage

import "sync"

type MemoryStore struct {
	mu      sync.RWMutex
	idToURL map[string]string
	urlToID map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		idToURL: make(map[string]string),
		urlToID: make(map[string]string),
	}
}

func (m *MemoryStore) GetIDByURL(u string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.urlToID[u]
	return id, ok
}

func (m *MemoryStore) GetURLByID(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.idToURL[id]
	return u, ok
}

func (m *MemoryStore) Save(id, u string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idToURL[id] = u
	m.urlToID[u] = id
}

func (m *MemoryStore) HasID(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.idToURL[id]
	return ok
}
