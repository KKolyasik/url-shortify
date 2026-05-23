package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/KKolyasik/url-shortify/internal/domainerr"
	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/KKolyasik/url-shortify/internal/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidURL         = errors.New("invalid url")
	ErrNotFound           = errors.New("not found")
	ErrURLNotBelongToUser = errors.New("url not belong to user")
)

type Storage interface {
	GetURLByID(ctx context.Context, id string) (model.URL, error)
	Save(ctx context.Context, id, u string, vid uuid.UUID) error
	HasID(ctx context.Context, id string) (bool, error)
	GetURLIDByUser(ctx context.Context, vid uuid.UUID) ([]model.URL, error)
	BatchDelete(ctx context.Context, shortCodes ...string) error
}

type Service struct {
	storage Storage

	doneCh            chan struct{}
	shortCodeDeleteCh chan string
}

func New(ctx context.Context, storage Storage) *Service {
	svc := &Service{
		storage:           storage,
		doneCh: make(chan struct{}),
		shortCodeDeleteCh: make(chan string, 1024),
	}

	go svc.deleteShortCodes(ctx)

	return svc
}

func (s *Service) Shorten(ctx context.Context, raw string, vid uuid.UUID) (string, error) {
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
		err = s.storage.Save(ctx, id, u, vid)
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
	if u.IsDeleted {
		return "", domainerr.ErrURLDeleted
	}
	return u.OriginalURL, nil
}

func (s *Service) UserResolve(ctx context.Context, vid uuid.UUID) ([]model.URL, error) {
	urls, err := s.storage.GetURLIDByUser(ctx, vid)
	if err != nil {
		return nil, err
	}

	return urls, nil
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

func (s *Service) Delete(ctx context.Context, vid uuid.UUID, shortCodes ...string) error {
	logger.Log.Sugar().Infow("Начало удаления")
	urls, err := s.storage.GetURLIDByUser(ctx, vid)
	if err != nil {
		return err
	}

	urlsSet := make(map[string]struct{}, len(shortCodes))

	for _, url := range urls {
		urlsSet[url.ShortCode] = struct{}{}
	}

	for _, code := range shortCodes {
		if _, ok := urlsSet[code]; !ok {
			logger.Log.Sugar().Infow("Ссылка не принадлежит пользователю")
			return ErrURLNotBelongToUser
		}

		s.shortCodeDeleteCh <- code
	}

	return nil
}

func (s *Service) Stop() {
	<-s.doneCh
}

func (s *Service) deleteShortCodes(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	codes := make([]string, 0)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			for {
				select {
				case code := <-s.shortCodeDeleteCh:
					codes = append(codes, code)
				default:
					if len(codes) > 0 {
						flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						s.storage.BatchDelete(flushCtx, codes...)
						cancel()
					}
					close(s.doneCh)
					return
				}
			}
		case code := <-s.shortCodeDeleteCh:
			logger.Log.Sugar().Infow("Добавили ссылку в Batch", "code", code)
			codes = append(codes, code)
		case <-ticker.C:
			if len(codes) == 0 {
				continue
			}

			err := s.storage.BatchDelete(ctx, codes...)
			if err != nil {
				continue
			}

			codes = nil
		}
	}
}
