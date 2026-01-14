package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_Empty(t *testing.T) {
	s := NewMemoryStore()

	id, ok := s.GetIDByURL("http://example.com/")
	assert.False(t, ok)
	assert.Equal(t, "", id)

	u, ok := s.GetURLByID("abc")
	assert.False(t, ok)
	assert.Equal(t, "", u)

	assert.False(t, s.HasID("abc"))
}

func TestMemoryStore_SaveAndGet(t *testing.T) {
	s := NewMemoryStore()

	s.Save("id1", "http://example.com/")

	u, ok := s.GetURLByID("id1")
	assert.True(t, ok)
	assert.Equal(t, "http://example.com/", u)

	id, ok := s.GetIDByURL("http://example.com/")
	assert.True(t, ok)
	assert.Equal(t, "id1", id)

	assert.True(t, s.HasID("id1"))
	assert.False(t, s.HasID("missing"))
}

func TestMemoryStore_SaveOverwriteSameID(t *testing.T) {
	s := NewMemoryStore()

	s.Save("id1", "http://a.com/")
	s.Save("id1", "http://b.com/")

	u, ok := s.GetURLByID("id1")
	assert.True(t, ok)
	assert.Equal(t, "http://b.com/", u)

	id, ok := s.GetIDByURL("http://b.com/")
	assert.True(t, ok)
	assert.Equal(t, "id1", id)

	oldID, ok := s.GetIDByURL("http://a.com/")
	assert.True(t, ok)
	assert.Equal(t, "id1", oldID)
}

func TestMemoryStore_SaveOverwriteSameURL(t *testing.T) {
	s := NewMemoryStore()

	s.Save("id1", "http://a.com/")
	s.Save("id2", "http://a.com/")

	id, ok := s.GetIDByURL("http://a.com/")
	assert.True(t, ok)
	assert.Equal(t, "id2", id)

	u, ok := s.GetURLByID("id2")
	assert.True(t, ok)
	assert.Equal(t, "http://a.com/", u)

	u1, ok := s.GetURLByID("id1")
	assert.True(t, ok)
	assert.Equal(t, "http://a.com/", u1)

	assert.True(t, s.HasID("id1"))
	assert.True(t, s.HasID("id2"))
}