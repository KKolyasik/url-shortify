package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	idToURL map[string]string
	urlToID map[string]string

	lastURLSaved string
}

func (f *fakeStorage) GetIDByURL(ctx context.Context, u string) (string, error) {
	id, ok := f.urlToID[u]
	if !ok {
		return "", nil
	}
	return id, nil
}

func (f *fakeStorage) GetURLByID(ctx context.Context, id string) (string, error) {
	u, ok := f.idToURL[id]
	if !ok {
		return "", ErrNotFound
	}
	return u, nil
}

func (f *fakeStorage) Save(ctx context.Context, id, u string) error {
	f.idToURL[id] = u
	f.urlToID[u] = id
	f.lastURLSaved = u
	return nil
}

func (f *fakeStorage) HasID(ctx context.Context, id string) (bool, error) {
	_, ok := f.idToURL[id]
	return ok, nil
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
			u, err := svc.Resolve(context.Background(), tt.id)
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
			id, err := svc.Shorten(context.Background(), tt.url)
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
