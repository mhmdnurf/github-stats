package stats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const cacheKeyPrefix = "stats:v1:"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUserNotFound     = errors.New("user not found")
)

type Fetcher interface {
	Fetch(ctx context.Context, username string) (UserStats, error)
}

type Cache interface {
	Get(
		ctx context.Context,
		key string,
	) (value UserStats, found bool, err error)

	Set(
		ctx context.Context,
		key string,
		value UserStats,
		ttl time.Duration,
	) error
}

type Service struct {
	fetcher Fetcher
	cache   Cache
	ttl     time.Duration
}

func NewService(
	fetcher Fetcher,
	cache Cache,
	ttl time.Duration,
) (*Service, error) {
	if fetcher == nil {
		return nil, errors.New("stats fetcher is required")
	}

	if cache == nil {
		return nil, errors.New("stats cache is required")
	}

	if ttl <= 0 {
		return nil, errors.New("stats cache TTL must be greater than zero")
	}

	return &Service{
		fetcher: fetcher,
		cache:   cache,
		ttl:     ttl,
	}, nil
}

func (s *Service) Get(
	ctx context.Context,
	username string,
) (UserStats, error) {
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if normalizedUsername == "" {
		return UserStats{}, ErrUsernameRequired
	}

	key := cacheKeyPrefix + normalizedUsername

	cached, found, err := s.cache.Get(ctx, key)
	if err != nil {
		return UserStats{}, fmt.Errorf("get cached stats: %w", err)
	}

	if found {
		return cached, nil
	}

	fetched, err := s.fetcher.Fetch(ctx, normalizedUsername)
	if err != nil {
		return UserStats{}, fmt.Errorf("fetch stats: %w", err)
	}

	if err := s.cache.Set(ctx, key, fetched, s.ttl); err != nil {
		return UserStats{}, fmt.Errorf("cache fetched stats: %w", err)
	}

	return fetched, nil
}
