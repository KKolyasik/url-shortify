package storage

import (
	"encoding/json"
	"io"
	"os"

	"github.com/KKolyasik/url-shortify/internal/logger"
)

type Storage interface {
	Save(id, u string)
	HasID(id string) bool
	GetAllIDToURLs() map[string]string
}

type URL struct {
	UUID        uint   `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func (f *FileStorage) Restore() error {
	logger.Log.Sugar().Info("Восстановление началось")
	var urls []URL
	err := f.decoder.Decode(&urls)

	if err != nil && err != io.EOF {
		return err
	}

	for _, url := range urls {
		if !f.storage.HasID(url.ShortURL) {
			f.storage.Save(url.ShortURL, url.OriginalURL)
		}
	}
	return nil
}

func (f *FileStorage) Save() error {
	logger.Log.Sugar().Info("Сохранение началось")

	if err := f.file.Truncate(0); err != nil {
		return err
	}
	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	var uuid uint
	data := f.storage.GetAllIDToURLs()
	urls := make([]URL, 0, len(data))
	
	for id, url := range data {
		uuid++
		urls = append(urls, URL{
			UUID: uuid,
			ShortURL: id,
			OriginalURL: url,
		})
	}
	return f.encoder.Encode(urls)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}