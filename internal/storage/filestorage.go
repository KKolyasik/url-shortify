package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/google/uuid"
)

type Storage interface {
	Save(ctx context.Context, id, u string) error
	HasID(ctx context.Context, id string) (bool, error)
	GetAllIDToURLs(ctx context.Context) (map[string]string, error)
}

type URL struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	storage Storage
}

func NewFileStorage(filename string, storage Storage) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
		storage: storage,
	}, nil
}

func (f *FileStorage) Restore(ctx context.Context) error {
	logger.Log.Sugar().Info("Восстановление началось")
	var urls []URL
	err := f.decoder.Decode(&urls)

	if err != nil && err != io.EOF {
		return err
	}

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		ok, err := f.storage.HasID(ctx, url.ShortURL)
		if err != nil {
			return err
		}
		if !ok {
			err := f.storage.Save(ctx, url.ShortURL, url.OriginalURL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FileStorage) Save(ctx context.Context) error {
	logger.Log.Sugar().Info("Сохранение началось")

	if err := f.file.Truncate(0); err != nil {
		return err
	}
	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	data, err := f.storage.GetAllIDToURLs(ctx)
	if err != nil {
		return err
	}
	urls := make([]URL, 0, len(data))

	for id, url := range data {
		urls = append(urls, URL{
			UUID:        uuid.New(),
			ShortURL:    id,
			OriginalURL: url,
		})
	}
	return f.encoder.Encode(urls)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
