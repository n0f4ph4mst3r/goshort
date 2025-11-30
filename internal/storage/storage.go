package storage

import (
	"context"
	"errors"
	"log/slog"
)

type UrlStorage struct {
	service UrlService
	cache   CacheClient
	log     *slog.Logger
}

type UrlService interface {
	SaveURL(ctx context.Context, originalURL, alias string) error
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) (string, error)
}

type CacheClient interface {
	SetURL(ctx context.Context, u, alias string) error
	GetURL(ctx context.Context, alias string) (string, error)
	DelURL(ctx context.Context, u, alias string) error
}

func New(log *slog.Logger, service UrlService, cache CacheClient) *UrlStorage {
	return &UrlStorage{
		service: service,
		cache:   cache,
		log:     log,
	}
}

func (s *UrlStorage) SaveURL(ctx context.Context, u, alias string) error {
	if err := s.service.SaveURL(ctx, u, alias); err != nil {
		return err
	}

	if s.cache != nil {
		s.log.Info("caching URL", slog.String("alias", alias))
		err := s.cache.SetURL(ctx, u, alias)
		if err != nil {
			s.log.Warn("failed to cache URL", slog.String("alias", alias), slog.Any("err", err.Error()))
		} else {
			s.log.Info("URL cached", slog.String("alias", alias))
		}
	}

	return nil
}

func (s *UrlStorage) GetURL(ctx context.Context, alias string) (string, error) {
	if s.cache != nil {
		s.log.Info("checking cache for URL", slog.String("alias", alias))
		u, err := s.cache.GetURL(ctx, alias)
		if err == nil {
			s.log.Info("URL found in cache", slog.String("alias", alias))
			return u, nil
		} else {
			s.log.Info("URL not found in cache", slog.String("alias", alias))
		}
	}

	u, err := s.service.GetURL(ctx, alias)
	if err != nil {
		return "", err
	}

	if s.cache != nil {
		s.log.Info("caching URL", slog.String("alias", alias))
		err := s.cache.SetURL(ctx, u, alias)
		if err != nil {
			s.log.Warn("failed to cache URL", slog.String("alias", alias), slog.Any("err", err.Error()))
		} else {
			s.log.Info("URL cached", slog.String("alias", alias))
		}
	}

	return u, nil
}

func (s *UrlStorage) DeleteURL(ctx context.Context, alias string) (string, error) {
	u, err := s.service.DeleteURL(ctx, alias)
	if err != nil {
		return "", err
	}

	if s.cache != nil {
		s.log.Info("deleting URL from cache", slog.String("alias", alias))
		err := s.cache.DelURL(ctx, u, alias)
		if err != nil {
			s.log.Warn("failed to delete URL from cache", slog.String("alias", alias), slog.Any("err", err.Error()))
		} else {
			s.log.Info("URL deleted from cache", slog.String("alias", alias))
		}
	}

	return u, nil
}

var (
	ErrUrlNotFound = errors.New("URL not found")
	ErrUrlExists   = errors.New("URL already exists")
)
