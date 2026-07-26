package languages

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const cacheKeyPrefix = "languages:v1:"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUserNotFound     = errors.New("user not found")
)

type Fetcher interface {
	FetchLanguages(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (UserLanguages, error)
}

type Cache interface {
	Get(
		ctx context.Context,
		key string,
	) (value UserLanguages, found bool, err error)

	Set(
		ctx context.Context,
		key string,
		value UserLanguages,
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
		return nil, errors.New("languages fetcher is required")
	}

	if cache == nil {
		return nil, errors.New("languages cache is required")
	}

	if ttl <= 0 {
		return nil, errors.New(
			"languages cache TTL must be greater than zero",
		)
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
	scope repositoryScope.Scope,
) (UserLanguages, error) {
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if normalizedUsername == "" {
		return UserLanguages{}, ErrUsernameRequired
	}

	key := cacheKeyPrefix + normalizedUsername + ":" + string(scope)
	cached, found, err := s.cache.Get(ctx, key)
	if err != nil {
		return UserLanguages{}, fmt.Errorf(
			"get cached languages: %w",
			err,
		)
	}
	if found {
		return cached, nil
	}

	fetched, err := s.fetcher.FetchLanguages(ctx, normalizedUsername, scope)
	if err != nil {
		return UserLanguages{}, fmt.Errorf(
			"fetch languages: %w",
			err,
		)
	}

	if err := s.cache.Set(ctx, key, fetched, s.ttl); err != nil {
		return UserLanguages{}, fmt.Errorf(
			"cache fetched languages: %w",
			err,
		)
	}

	return fetched, nil
}
