package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/KKolyasik/url-shortify/internal/model"
	"github.com/google/uuid"
)

type Storage interface {
	Save(ctx context.Context, id, u string, vid uuid.UUID) error
	HasID(ctx context.Context, id string) (bool, error)
	GetAllURLs(ctx context.Context) ([]model.URL, error)
	BatchDelete(ctx context.Context, shortCodes ...string) error
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
	var urls []model.URL
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
		ok, err := f.storage.HasID(ctx, url.ShortCode)
		if err != nil {
			return err
		}
		if !ok {
			vid := url.UserID
			if vid == uuid.Nil {
				vid = uuid.New()
			}
			err := f.storage.Save(ctx, url.ShortCode, url.OriginalURL, vid)
			if err != nil {
				return err
			}
		}

		if url.IsDeleted {
			if err := f.storage.BatchDelete(ctx, url.ShortCode); err != nil {
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

	data, err := f.storage.GetAllURLs(ctx)
	if err != nil {
		return err
	}
	return f.encoder.Encode(data)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
