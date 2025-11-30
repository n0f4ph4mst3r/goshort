package redis

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/n0f4ph4mst3r/goshort/internal/config"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
	cfg    *config.CacheConfig
}

func New(connStr string, cfg *config.CacheConfig) (*Storage, error) {
	const op = "storage.redis.New"

	if !cfg.Enabled {
		return &Storage{client: nil, cfg: cfg}, nil
	}

	if connStr == "" {
		return nil, fmt.Errorf("%s: cache enabled but connection string empty", op)
	}

	u, err := url.Parse(connStr)
	if err != nil {
		return nil, err
	}

	addr := u.Host

	db := 0
	if u.Path != "" {
		dbStr := strings.TrimPrefix(u.Path, "/")
		db, _ = strconv.Atoi(dbStr)
	}

	client := redis.NewClient(&redis.Options{Addr: addr, DB: db})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: unable to connect: %w", op, err)
	}

	return &Storage{client: client, cfg: cfg}, nil
}

func (s *Storage) SetURL(ctx context.Context, u, alias string) error {
	const op = "storage.redis.SetURL"

	if err := s.client.Set(ctx, s.cfg.PrefixURL+alias, u, s.cfg.TTL).Err(); err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}
	if err := s.client.Set(ctx, s.cfg.PrefixRev+u, alias, s.cfg.ReverseIndexTTL).Err(); err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.redis.GetURL"

	u, err := s.client.Get(ctx, s.cfg.PrefixURL+alias).Result()
	if err != nil {
		return "", fmt.Errorf("%s: %s", op, err)
	}

	return u, nil
}

func (s *Storage) DelURL(ctx context.Context, u, alias string) error {
	const op = "storage.redis.DelURL"

	if err := s.client.Del(ctx, s.cfg.PrefixURL+alias).Err(); err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}
	if err := s.client.Del(ctx, s.cfg.PrefixRev+u).Err(); err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
