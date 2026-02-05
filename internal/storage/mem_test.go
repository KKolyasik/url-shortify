package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_Empty(t *testing.T) {
	s := NewMemoryStore()

	id, err := s.GetIDByURL(context.Background(), "http://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, "", id)

	u, err := s.GetURLByID(context.Background(), "abc")
	assert.ErrorIs(t, err, URLNotFound)
	assert.Equal(t, "", u)

	ok, err := s.HasID(context.Background(), "abc")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryStore_SaveAndGet(t *testing.T) {
	s := NewMemoryStore()

	s.Save(context.Background(), "id1", "http://example.com/")

	u, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/", u)

	id, err := s.GetIDByURL(context.Background(), "http://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", id)

	ok1, err := s.HasID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.True(t, ok1)

	ok2, err := s.HasID(context.Background(), "id2")
	assert.NoError(t, err)
	assert.False(t, ok2)
}

func TestMemoryStore_SaveOverwriteSameID(t *testing.T) {
	s := NewMemoryStore()

	s.Save(context.Background(), "id1", "http://a.com/")
	s.Save(context.Background(), "id1", "http://b.com/")

	u, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://b.com/", u)

	id, err := s.GetIDByURL(context.Background(), "http://b.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", id)

	oldID, err := s.GetIDByURL(context.Background(), "http://a.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id1", oldID)
}

func TestMemoryStore_SaveOverwriteSameURL(t *testing.T) {
	s := NewMemoryStore()

	s.Save(context.Background(), "id1", "http://a.com/")
	s.Save(context.Background(), "id2", "http://a.com/")

	id, err := s.GetIDByURL(context.Background(), "http://a.com/")
	assert.NoError(t, err)
	assert.Equal(t, "id2", id)

	u, err := s.GetURLByID(context.Background(), "id2")
	assert.NoError(t, err)
	assert.Equal(t, "http://a.com/", u)

	u1, err := s.GetURLByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "http://a.com/", u1)

	ok1, _ := s.HasID(context.Background(), "id1")
	ok2, _ := s.HasID(context.Background(), "id1")

	assert.True(t, ok1)
	assert.True(t, ok2)
}