package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	urlToID map[string]string
	idToURL map[string]string

	hasIDFn func(id string) bool

	saveCalls int
	lastSaveID string
	lastSaveU  string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		urlToID: make(map[string]string),
		idToURL: make(map[string]string),
	}
}

func (f *fakeStorage) GetIDByURL(u string) (string, bool) {
	id, ok := f.urlToID[u]
	return id, ok
}

func (f *fakeStorage) GetURLByID(id string) (string, bool) {
	u, ok := f.idToURL[id]
	return u, ok
}

func (f *fakeStorage) Save(id, u string) {
	f.saveCalls++
	f.lastSaveID = id
	f.lastSaveU = u
	f.idToURL[id] = u
	f.urlToID[u] = id
}

func (f *fakeStorage) HasID(id string) bool {
	if f.hasIDFn != nil {
		return f.hasIDFn(id)
	}
	_, ok := f.idToURL[id]
	return ok
}

func TestService_Shorten_InvalidURL(t *testing.T) {
	s := New(newFakeStorage())

	cases := []string{
		"",
		"   ",
		"not a url",
		"ftp://example.com",
		"http://",
		"https://",
		"http://exa mple",
		"//example.com",
	}

	for _, in := range cases {
		_, err := s.Shorten(in)
		assert.ErrorIs(t, err, ErrInvalidURL, "input=%q", in)
	}
}

func TestService_Shorten_NormalizesAndSaves(t *testing.T) {
	st := newFakeStorage()
	s := New(st)

	raw := "HTTP://EXAMPLE.COM/"
	id, err := s.Shorten(raw)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	assert.Equal(t, 1, st.saveCalls)
	assert.Equal(t, id, st.lastSaveID)
	assert.Equal(t, "http://example.com", st.lastSaveU)

	assert.False(t, strings.Contains(id, "="))
}

func TestService_Shorten_ReturnsExistingID_NoSave(t *testing.T) {
	st := newFakeStorage()
	s := New(st)

	normalized := "http://example.com"
	st.urlToID[normalized] = "fixed-id"

	id, err := s.Shorten("http://EXAMPLE.com/")
	assert.NoError(t, err)
	assert.Equal(t, "fixed-id", id)
	assert.Equal(t, 0, st.saveCalls)
}

func TestService_Shorten_Collision_RetriesUntilFree(t *testing.T) {
	st := newFakeStorage()

	calls := 0
	st.hasIDFn = func(id string) bool {
		calls++
		return calls <= 2
	}

	s := New(st)

	id, err := s.Shorten("http://example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	assert.Equal(t, 1, st.saveCalls)
	assert.Equal(t, "http://example.com", st.lastSaveU)
}

func TestService_Resolve_TrimAndNotFound(t *testing.T) {
	st := newFakeStorage()
	s := New(st)

	_, err := s.Resolve("")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = s.Resolve("   ")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = s.Resolve("missing")
	assert.ErrorIs(t, err, ErrNotFound)

	// Сохранённый id
	st.idToURL["abc"] = "http://example.com"

	u, err := s.Resolve("  abc  ")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", u)
}
