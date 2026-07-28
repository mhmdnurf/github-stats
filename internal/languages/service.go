package languages

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mhmdnurf/github-stats/internal/cacheaside"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const cacheKeyPrefix = "languages:v1:"

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUserNotFound     = errors.New("user not found")
	ErrUnavailable      = errors.New("snapshot unavailable")
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
	loader  *cacheaside.Loader[UserLanguages]
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

	loader, err := cacheaside.New[UserLanguages](
		cache,
		ttl,
		"languages",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create languages cache loader: %w",
			err,
		)
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
) (UserLanguages, error) {
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	if normalizedUsername == "" {
		return UserLanguages{}, ErrUsernameRequired
	}

	key := cacheKeyPrefix + normalizedUsername + ":" + string(scope)

	return s.loader.Load(
		ctx,
		key,
		func(ctx context.Context) (UserLanguages, error) {
			return s.fetcher.FetchLanguages(
				ctx,
				normalizedUsername,
				scope,
			)
		},
	)
}
