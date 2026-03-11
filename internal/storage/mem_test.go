package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Empty(t *testing.T) {
	s := NewMemoryStore()

	id, err := s.GetIDByURL(context.Background(), "http://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, "", id.ShortCode)

	u, err := s.GetURLByID(context.Background(), "abc")
	assert.ErrorIs(t, err, ErrURLNotFound)
	assert.Equal(t, "", u.OriginalURL)

	ok, err := s.HasID(context.Background(), "abc")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryStore_SaveAndGet(t *testing.T) {
	s := NewMemoryStore()

	err := s.Save(context.Background(), "id1", "http://example.com/", uuid.New())
	assert.NoError(t, err)

	u, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/", u.OriginalURL)

	id, err := s.GetIDByURL(context.Background(), "http://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", id.ShortCode)

	ok1, err := s.HasID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.True(t, ok1)

	ok2, err := s.HasID(context.Background(), "id2")
	assert.NoError(t, err)
	assert.False(t, ok2)
}

func TestMemoryStore_SaveOverwriteSameID(t *testing.T) {
	s := NewMemoryStore()

	err := s.Save(context.Background(), "id1", "http://a.com/", uuid.New())
	assert.NoError(t, err)
	err = s.Save(context.Background(), "id1", "http://b.com/", uuid.New())
	assert.ErrorIs(t, err, domainerr.ErrShortCodeCollision)

	u, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://a.com/", u.OriginalURL)

	id, err := s.GetIDByURL(context.Background(), "http://b.com/")
	assert.NoError(t, err)
	assert.Equal(t, "", id.ShortCode)

	oldID, err := s.GetIDByURL(context.Background(), "http://a.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", oldID.ShortCode)
}

func TestMemoryStore_SaveOverwriteSameURL(t *testing.T) {
	s := NewMemoryStore()

	err := s.Save(context.Background(), "id1", "http://a.com/", uuid.New())
	assert.NoError(t, err)
	err = s.Save(context.Background(), "id2", "http://a.com/", uuid.New())

	var existsErr *domainerr.URLAlreadyExistsError
	assert.True(t, errors.As(err, &existsErr))
	assert.Equal(t, "id1", existsErr.ShortCode)

	id, err := s.GetIDByURL(context.Background(), "http://a.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", id.ShortCode)

	u, err := s.GetURLByID(context.Background(), "id2")
	assert.ErrorIs(t, err, ErrURLNotFound)
	assert.Equal(t, "", u.OriginalURL)

	u1, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://a.com/", u1.OriginalURL)

	ok1, _ := s.HasID(context.Background(), "id1")
	ok2, _ := s.HasID(context.Background(), "id1")

	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestMemoryStore_GetURLIDByUser(t *testing.T) {
	s := NewMemoryStore()
	vid1 := uuid.New()
	vid2 := uuid.New()

	require.NoError(t, s.Save(context.Background(), "id1", "http://a.com/", vid1))
	require.NoError(t, s.Save(context.Background(), "id2", "http://b.com/", vid2))
	require.NoError(t, s.Save(context.Background(), "id3", "http://c.com/", vid1))

	urls, err := s.GetURLIDByUser(context.Background(), vid1)
	require.NoError(t, err)
	require.Len(t, urls, 2)

	got := make(map[string]string, len(urls))
	for _, u := range urls {
		got[u.ShortCode] = u.OriginalURL
	}

	assert.Equal(t, map[string]string{
		"id1": "http://a.com/",
		"id3": "http://c.com/",
	}, got)
}

func TestMemoryStore_GetAllURLs(t *testing.T) {
	s := NewMemoryStore()
	vid := uuid.New()

	require.NoError(t, s.Save(context.Background(), "id1", "http://a.com/", vid))

	urls, err := s.GetAllURLs(context.Background())
	require.NoError(t, err)
	require.Len(t, urls, 1)

	assert.Equal(t, "id1", urls[0].ShortCode)
	assert.Equal(t, "http://a.com/", urls[0].OriginalURL)
	assert.Equal(t, vid, urls[0].UserID)
}

func TestMemoryStore_BatchDelete(t *testing.T) {
	s := NewMemoryStore()
	vid := uuid.New()

	require.NoError(t, s.Save(context.Background(), "id1", "http://a.com/", vid))
	require.NoError(t, s.Save(context.Background(), "id2", "http://b.com/", vid))

	require.NoError(t, s.BatchDelete(context.Background(), "id1"))

	u1, err := s.GetURLByID(context.Background(), "id1")
	require.NoError(t, err)
	assert.True(t, u1.IsDeleted)

	u2, err := s.GetURLByID(context.Background(), "id2")
	require.NoError(t, err)
	assert.False(t, u2.IsDeleted)
}
