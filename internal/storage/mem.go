package storage

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/KKolyasik/url-shortify/internal/model"
	"github.com/google/uuid"
)

var (
	ErrIDNotFound  = errors.New("id not found")
	ErrURLNotFound = errors.New("url not found")
)

type MemoryStore struct {
	mu          sync.RWMutex
	idToURL     map[string]string
	urlToID     map[string]string
	idToUser    map[string]uuid.UUID
	idToDeleted map[string]bool
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		idToURL:     make(map[string]string),
		urlToID:     make(map[string]string),
		idToUser:    make(map[string]uuid.UUID),
		idToDeleted: make(map[string]bool),
	}
}

func (m *MemoryStore) GetIDByURL(ctx context.Context, u string) (model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return model.URL{}, ctx.Err()
	default:
	}
	id, ok := m.urlToID[u]
	if !ok {
		return model.URL{}, nil
	}
	return model.URL{
		ShortCode: id,
		IsDeleted: m.idToDeleted[id],
	}, nil
}

func (m *MemoryStore) GetURLByID(ctx context.Context, id string) (model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return model.URL{}, ctx.Err()
	default:
	}
	u, ok := m.idToURL[id]
	if !ok {
		return model.URL{}, ErrURLNotFound
	}
	return model.URL{
		ShortCode:   id,
		OriginalURL: u,
		UserID:      m.idToUser[id],
		IsDeleted:   m.idToDeleted[id],
	}, nil
}

func (m *MemoryStore) Save(ctx context.Context, id, u string, vid uuid.UUID) error {
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
	m.idToUser[id] = vid
	m.idToDeleted[id] = false
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

func (m *MemoryStore) GetURLIDByUser(ctx context.Context, vid uuid.UUID) ([]model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.idToUser))
	for id := range m.idToUser {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	urls := make([]model.URL, 0)
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return urls, ctx.Err()
		default:
		}

		if m.idToUser[id] != vid {
			continue
		}

		originalURL, ok := m.idToURL[id]
		if !ok {
			continue
		}

		urls = append(urls, model.URL{
			OriginalURL: originalURL,
			ShortCode:   id,
			UserID:      m.idToUser[id],
			IsDeleted:   m.idToDeleted[id],
		})
	}

	return urls, nil
}

func (m *MemoryStore) GetAllURLs(ctx context.Context) ([]model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.idToURL))
	for id := range m.idToURL {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	urls := make([]model.URL, 0, len(ids))
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return urls, ctx.Err()
		default:
		}

		urls = append(urls, model.URL{
			UUID:        uuid.New(),
			ShortCode:   id,
			OriginalURL: m.idToURL[id],
			UserID:      m.idToUser[id],
			IsDeleted:   m.idToDeleted[id],
		})
	}

	return urls, nil
}

func (m *MemoryStore) BatchDelete(ctx context.Context, shortCodes ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, code := range shortCodes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, ok := m.idToURL[code]; !ok {
			continue
		}

		m.idToDeleted[code] = true
	}

	return nil
}
