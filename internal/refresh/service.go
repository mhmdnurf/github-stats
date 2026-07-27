package refresh

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const maxConcurrentFetches = 2

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

func (service *Service) Run(ctx context.Context) error {
	type task func(context.Context) error

	tasks := []task{
		func(ctx context.Context) error {
			return service.refreshStats(
				ctx,
				repositoryScope.ScopePublic,
			)
		},
		func(ctx context.Context) error {
			return service.refreshStats(
				ctx,
				repositoryScope.ScopeAll,
			)
		},
		func(ctx context.Context) error {
			return service.refreshLanguages(
				ctx,
				repositoryScope.ScopePublic,
			)
		},
		func(ctx context.Context) error {
			return service.refreshLanguages(
				ctx,
				repositoryScope.ScopeAll,
			)
		},
	}

	semaphore := make(chan struct{}, maxConcurrentFetches)
	errorsChannel := make(chan error, len(tasks))
	var waitGroup sync.WaitGroup

	for _, runTask := range tasks {
		runTask := runTask
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() {
					<-semaphore
				}()
			case <-ctx.Done():
				errorsChannel <- ctx.Err()
				return
			}

			if err := runTask(ctx); err != nil {
				errorsChannel <- err
			}
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)

	var refreshErrors []error
	for err := range errorsChannel {
		refreshErrors = append(refreshErrors, err)
	}

	return errors.Join(refreshErrors...)
}

func (service *Service) refreshStats(
	ctx context.Context,
	scope repositoryScope.Scope,
) error {
	value, err := service.statsFetcher.Fetch(
		ctx,
		service.username,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch stats snapshot for %s: %w",
			scope,
			err,
		)
	}

	key, err := snapshot.Key(
		service.username,
		snapshot.KindStats,
		scope,
	)
	if err != nil {
		return fmt.Errorf("create stats snapshot key: %w", err)
	}

	now := service.now()
	if err := service.statsStore.Set(
		ctx,
		key,
		snapshot.Snapshot[stats.UserStats]{
			Value:         value,
			UpdatedAt:     now,
			RefreshAfter:  now.Add(service.freshness),
			SchemaVersion: snapshot.CurrentSchemaVersion,
		},
	); err != nil {
		return fmt.Errorf(
			"store stats snapshot for %s: %w",
			scope,
			err,
		)
	}

	return nil
}

func (service *Service) refreshLanguages(
	ctx context.Context,
	scope repositoryScope.Scope,
) error {
	value, err := service.languagesFetcher.FetchLanguages(
		ctx,
		service.username,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch languages snapshot for %s: %w",
			scope,
			err,
		)
	}

	key, err := snapshot.Key(
		service.username,
		snapshot.KindLanguages,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"create languages snapshot key: %w",
			err,
		)
	}

	now := service.now()
	if err := service.languagesStore.Set(
		ctx,
		key,
		snapshot.Snapshot[languages.UserLanguages]{
			Value:         value,
			UpdatedAt:     now,
			RefreshAfter:  now.Add(service.freshness),
			SchemaVersion: snapshot.CurrentSchemaVersion,
		},
	); err != nil {
		return fmt.Errorf(
			"store languages snapshot for %s: %w",
			scope,
			err,
		)
	}

	return nil
}
