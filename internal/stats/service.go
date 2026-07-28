package stats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mhmdnurf/github-stats/internal/cacheaside"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const cacheKeyPrefix = "stats:v1:"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUserNotFound     = errors.New("user not found")
	ErrUnavailable      = errors.New("snapshot unavailable")
)

type Fetcher interface {
	Fetch(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (UserStats, error)
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
	loader  *cacheaside.Loader[UserStats]
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

	loader, err := cacheaside.New[UserStats](
		cache,
		ttl,
		"stats",
	)
	if err != nil {
		return nil, fmt.Errorf("create stats cache loader: %w", err)
	}

	return &Service{
		fetcher: fetcher,
		loader:  loader,
	}, nil
}

func (s *Service) Get(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (UserStats, error) {
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if normalizedUsername == "" {
		return UserStats{}, ErrUsernameRequired
	}

	key := cacheKeyPrefix + normalizedUsername + ":" + string(scope)

	return s.loader.Load(
		ctx,
		key,
		func(ctx context.Context) (UserStats, error) {
			return s.fetcher.Fetch(
				ctx,
				normalizedUsername,
				scope,
			)
		},
	)
}
