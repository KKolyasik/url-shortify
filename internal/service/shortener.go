package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
)

var (
	ErrInvalidURL = errors.New("invalid url")
	ErrNotFound   = errors.New("not found")
)

type Storage interface {
	GetURLByID(ctx context.Context, id string) (string, error)
	Save(ctx context.Context, id, u string) error
	HasID(ctx context.Context, id string) (bool, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Shorten(ctx context.Context, raw string) (string, error) {
	u, err := normalizeURL(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		id, err := generateID(16)
		if err != nil {
			return "", err
		}
		ok, err := s.storage.HasID(ctx, id)
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		err = s.storage.Save(ctx, id, u)
		if err != nil {
			if errors.Is(err, domainerr.ErrShortCodeCollision) {
				continue
			}
			return "", err
		}
		return id, nil
	}
}

func (s *Service) Resolve(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrNotFound
	}

	u, err := s.storage.GetURLByID(ctx, id)
	if err != nil {
		return "", err
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
