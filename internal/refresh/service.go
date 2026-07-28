package refresh

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

var (
	ErrUsernameRequired         = errors.New("refresh username is required")
	ErrStatsFetcherRequired     = errors.New("stats fetcher is required")
	ErrLanguagesFetcherRequired = errors.New("languages fetcher is required")
	ErrFreshnessInvalid         = errors.New("snapshot freshness must be positive")
)

type StatsFetcher interface {
	Fetch(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (stats.UserStats, error)
}

type LanguagesFetcher interface {
	FetchLanguages(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (languages.UserLanguages, error)
}

type Service struct {
	username         string
	statsFetcher     StatsFetcher
	languagesFetcher LanguagesFetcher
	statsStore       snapshot.Store[stats.UserStats]
	languagesStore   snapshot.Store[languages.UserLanguages]
	freshness        time.Duration
	now              func() time.Time
}

func NewService(
	username string,
	statsFetcher StatsFetcher,
	languagesFetcher LanguagesFetcher,
	statsStore snapshot.Store[stats.UserStats],
	languagesStore snapshot.Store[languages.UserLanguages],
	freshness time.Duration,
) (*Service, error) {
	normalizedUsername := strings.ToLower(
		strings.TrimSpace(username),
	)
	if normalizedUsername == "" {
		return nil, ErrUsernameRequired
	}
	if statsFetcher == nil {
		return nil, ErrStatsFetcherRequired
	}
	if languagesFetcher == nil {
		return nil, ErrLanguagesFetcherRequired
	}
	if statsStore == nil || languagesStore == nil {
		return nil, snapshot.ErrStoreRequired
	}
	if freshness <= 0 {
		return nil, ErrFreshnessInvalid
	}

	return &Service{
		username:         normalizedUsername,
		statsFetcher:     statsFetcher,
		languagesFetcher: languagesFetcher,
		statsStore:       statsStore,
		languagesStore:   languagesStore,
		freshness:        freshness,
		now:              time.Now,
	}, nil
}
