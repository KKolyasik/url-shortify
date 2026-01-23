package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	idToURL map[string]string
	urlToID map[string]string

	lastURLSaved string
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
	f.idToURL[id] = u
	f.urlToID[u] = id
	f.lastURLSaved = u
}

func (f *fakeStorage) HasID(id string) bool {
	_, ok := f.idToURL[id]
	return ok
}

func TestServise_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		fs      fakeStorage
		id      string
		want    string
		wantErr bool
	}{
		{
			name: "success resolve",
			fs: fakeStorage{
				idToURL: map[string]string{
					"id1": "http://example.com/",
				},
			},
			id:      "id1",
			want:    "http://example.com/",
			wantErr: false,
		},
		{
			name:    "empty id",
			wantErr: true,
		},
		{
			name: "unkown id",
			fs: fakeStorage{
				idToURL: map[string]string{
					"id1": "url1",
				},
			},
			id:      "id2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&tt.fs)
			u, err := svc.Resolve(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, u)
		})
	}
}

func TestServise_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		fs      fakeStorage
		url     string
		length  int
		want    string
		wantErr bool
	}{
		{
			name: "success shorten",
			fs: fakeStorage{
				urlToID: make(map[string]string),
				idToURL: make(map[string]string),
			},
			url:     "http://example.com",
			length:  1,
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name: "success normalize",
			fs: fakeStorage{
				urlToID: make(map[string]string),
				idToURL: make(map[string]string),
			},
			url:     "http://EXaMpLE.COm/",
			length:  1,
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name: "wrong url",
			url: "http://",
			length: 0,
			wantErr: true,
		},
		{
			name: "wrong url2",
			url: "http",
			length: 0,
			wantErr: true,
		},
		{
			name: "wrong url3",
			url: "",
			length: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&tt.fs)
			id, err := svc.Shorten(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.fs.lastURLSaved)
			assert.Equal(t, tt.length, len(tt.fs.idToURL))
			assert.False(t, strings.Contains(id, "="))
		})
	}
}
