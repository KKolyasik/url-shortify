package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
)

var (
	ErrIDNotFound  = errors.New("id not found")
	ErrURLNotFound = errors.New("url not found")
)

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

func (m *MemoryStore) GetIDByURL(ctx context.Context, u string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	id, ok := m.urlToID[u]
	if !ok {
		return "", nil
	}
	return id, nil
}

func (m *MemoryStore) GetURLByID(ctx context.Context, id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	u, ok := m.idToURL[id]
	if !ok {
		return "", ErrURLNotFound
	}
	return u, nil
}

func (m *MemoryStore) Save(ctx context.Context, id, u string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if existingID, ok := m.urlToID[u]; ok {
		return &domainerr.URLAlreadyExistsError{ShortCode: existingID}
	}
	if _, ok := m.idToURL[id]; ok {
		return domainerr.ErrShortCodeCollision
	}

	m.idToURL[id] = u
	m.urlToID[u] = id
	return nil
}

func (m *MemoryStore) HasID(ctx context.Context, id string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	_, ok := m.idToURL[id]
	return ok, nil
}

func (m *MemoryStore) GetAllIDToURLs(ctx context.Context) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dst := make(map[string]string, len(m.idToURL))

	for id, url := range m.idToURL {
		select {
		case <-ctx.Done():
			return dst, ctx.Err()
		default:
		}
		dst[id] = url
	}

	return dst, nil
}
