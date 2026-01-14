package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL = errors.New("invalid url")
	ErrNotFound   = errors.New("not found")
)

type Storage interface {
	GetIDByURL(u string) (string, bool)
	GetURLByID(id string) (string, bool)
	Save(id, u string)
	HasID(id string) bool
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Shorten(raw string) (string, error) {
	u, err := normalizeURL(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	if id, ok := s.storage.GetIDByURL(u); ok {
		return id, nil
	}

	for {
		id, err := generateID(16)
		if err != nil {
			return "", err
		}
		if s.storage.HasID(id) {
			continue
		}
		s.storage.Save(id, u)
		return id, nil
	}

}

func (s *Service) Resolve(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrNotFound
	}

	u, ok := s.storage.GetURLByID(id)
	if !ok {
		return "", ErrNotFound
	}
	return u, nil
}

func generateID(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func normalizeURL(in string) (string, error) {
	in = strings.TrimSpace(in)
	u, err := url.ParseRequestURI(in)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if u.Path == "/" {
		u.Path = ""
	}

	u.Host = strings.ToLower(u.Host)

	return u.String(), nil
}
