package refresh

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type fakeFetcher struct {
	mu            sync.Mutex
	active        int
	maxActive     int
	statsFailures map[repositoryScope.Scope]error
}

func (fetcher *fakeFetcher) begin() {
	fetcher.mu.Lock()
	fetcher.active++
	if fetcher.active > fetcher.maxActive {
		fetcher.maxActive = fetcher.active
	}
	fetcher.mu.Unlock()
}

func (fetcher *fakeFetcher) end() {
	fetcher.mu.Lock()
	fetcher.active--
	fetcher.mu.Unlock()
}

func (fetcher *fakeFetcher) Fetch(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (stats.UserStats, error) {
	fetcher.begin()
	defer fetcher.end()

	select {
	case <-time.After(5 * time.Millisecond):
	case <-ctx.Done():
		return stats.UserStats{}, ctx.Err()
	}

	if err := fetcher.statsFailures[scope]; err != nil {
		return stats.UserStats{}, err
	}

	return stats.UserStats{
		Username:     username,
		Repositories: scopeValue(scope),
	}, nil
}

func (fetcher *fakeFetcher) FetchLanguages(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (languages.UserLanguages, error) {
	fetcher.begin()
	defer fetcher.end()

	select {
	case <-time.After(5 * time.Millisecond):
	case <-ctx.Done():
		return languages.UserLanguages{}, ctx.Err()
	}

	return languages.UserLanguages{
		Username: username,
		Scope:    scope,
	}, nil
}

func scopeValue(scope repositoryScope.Scope) int {
	if scope == repositoryScope.ScopeAll {
		return 2
	}
	return 1
}

func TestServiceRunStoresAllSnapshots(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{}
	statsStore := snapshot.NewMemory[stats.UserStats]()
	languagesStore := snapshot.NewMemory[languages.UserLanguages]()

	service, err := NewService(
		"  MHMDNURF  ",
		fetcher,
		fetcher,
		statsStore,
		languagesStore,
		30*time.Minute,
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time {
		return now
	}

	if err := service.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, scope := range []repositoryScope.Scope{
		repositoryScope.ScopePublic,
		repositoryScope.ScopeAll,
	} {
		statsKey, err := snapshot.Key(
			"mhmdnurf",
			snapshot.KindStats,
			scope,
		)
		if err != nil {
			t.Fatalf("stats Key() error = %v", err)
		}

		statsSnapshot, err := statsStore.Get(
			context.Background(),
			statsKey,
		)
		if err != nil {
			t.Fatalf("stats Get(%q) error = %v", statsKey, err)
		}
		if statsSnapshot.Value.Repositories != scopeValue(scope) {
			t.Errorf(
				"stats repositories for %s = %d",
				scope,
				statsSnapshot.Value.Repositories,
			)
		}
		assertMetadata(t, statsSnapshot.UpdatedAt, statsSnapshot.RefreshAfter, statsSnapshot.SchemaVersion, now)

		languagesKey, err := snapshot.Key(
			"mhmdnurf",
			snapshot.KindLanguages,
			scope,
		)
		if err != nil {
			t.Fatalf("languages Key() error = %v", err)
		}

		languagesSnapshot, err := languagesStore.Get(
			context.Background(),
			languagesKey,
		)
		if err != nil {
			t.Fatalf(
				"languages Get(%q) error = %v",
				languagesKey,
				err,
			)
		}
		if languagesSnapshot.Value.Scope != scope {
			t.Errorf(
				"languages scope = %q, want %q",
				languagesSnapshot.Value.Scope,
				scope,
			)
		}
		assertMetadata(t, languagesSnapshot.UpdatedAt, languagesSnapshot.RefreshAfter, languagesSnapshot.SchemaVersion, now)
	}

	fetcher.mu.Lock()
	maxActive := fetcher.maxActive
	fetcher.mu.Unlock()
	if maxActive > maxConcurrentFetches {
		t.Fatalf(
			"maximum concurrent fetches = %d, want <= %d",
			maxActive,
			maxConcurrentFetches,
		)
	}
}

func TestServiceRunKeepsSuccessfulSnapshotsOnPartialFailure(
	t *testing.T,
) {
	t.Parallel()

	fetchFailure := errors.New("GitHub unavailable")
	fetcher := &fakeFetcher{
		statsFailures: map[repositoryScope.Scope]error{
			repositoryScope.ScopeAll: fetchFailure,
		},
	}
	statsStore := snapshot.NewMemory[stats.UserStats]()
	languagesStore := snapshot.NewMemory[languages.UserLanguages]()

	service, err := NewService(
		"mhmdnurf",
		fetcher,
		fetcher,
		statsStore,
		languagesStore,
		time.Hour,
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	err = service.Run(context.Background())
	if !errors.Is(err, fetchFailure) {
		t.Fatalf("Run() error = %v, want %v", err, fetchFailure)
	}

	failedKey, err := snapshot.Key(
		"mhmdnurf",
		snapshot.KindStats,
		repositoryScope.ScopeAll,
	)
	if err != nil {
		t.Fatalf("Key() error = %v", err)
	}
	if _, err := statsStore.Get(
		context.Background(),
		failedKey,
	); !errors.Is(err, snapshot.ErrNotFound) {
		t.Fatalf(
			"failed snapshot Get() error = %v, want %v",
			err,
			snapshot.ErrNotFound,
		)
	}

	publicStatsKey, _ := snapshot.Key(
		"mhmdnurf",
		snapshot.KindStats,
		repositoryScope.ScopePublic,
	)
	if _, err := statsStore.Get(
		context.Background(),
		publicStatsKey,
	); err != nil {
		t.Fatalf("successful stats snapshot missing: %v", err)
	}

	for _, scope := range []repositoryScope.Scope{
		repositoryScope.ScopePublic,
		repositoryScope.ScopeAll,
	} {
		key, _ := snapshot.Key(
			"mhmdnurf",
			snapshot.KindLanguages,
			scope,
		)
		if _, err := languagesStore.Get(
			context.Background(),
			key,
		); err != nil {
			t.Fatalf(
				"successful languages snapshot for %s missing: %v",
				scope,
				err,
			)
		}
	}
}

func TestNewServiceValidation(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{}
	statsStore := snapshot.NewMemory[stats.UserStats]()
	languagesStore := snapshot.NewMemory[languages.UserLanguages]()

	tests := []struct {
		name             string
		username         string
		statsFetcher     StatsFetcher
		languagesFetcher LanguagesFetcher
		statsStore       snapshot.Store[stats.UserStats]
		languagesStore   snapshot.Store[languages.UserLanguages]
		freshness        time.Duration
		wantErr          error
	}{
		{
			name:             "username",
			statsFetcher:     fetcher,
			languagesFetcher: fetcher,
			statsStore:       statsStore,
			languagesStore:   languagesStore,
			freshness:        time.Hour,
			wantErr:          ErrUsernameRequired,
		},
		{
			name:             "stats fetcher",
			username:         "mhmdnurf",
			languagesFetcher: fetcher,
			statsStore:       statsStore,
			languagesStore:   languagesStore,
			freshness:        time.Hour,
			wantErr:          ErrStatsFetcherRequired,
		},
		{
			name:           "languages fetcher",
			username:       "mhmdnurf",
			statsFetcher:   fetcher,
			statsStore:     statsStore,
			languagesStore: languagesStore,
			freshness:      time.Hour,
			wantErr:        ErrLanguagesFetcherRequired,
		},
		{
			name:             "stats store",
			username:         "mhmdnurf",
			statsFetcher:     fetcher,
			languagesFetcher: fetcher,
			languagesStore:   languagesStore,
			freshness:        time.Hour,
			wantErr:          snapshot.ErrStoreRequired,
		},
		{
			name:             "languages store",
			username:         "mhmdnurf",
			statsFetcher:     fetcher,
			languagesFetcher: fetcher,
			statsStore:       statsStore,
			freshness:        time.Hour,
			wantErr:          snapshot.ErrStoreRequired,
		},
		{
			name:             "freshness",
			username:         "mhmdnurf",
			statsFetcher:     fetcher,
			languagesFetcher: fetcher,
			statsStore:       statsStore,
			languagesStore:   languagesStore,
			wantErr:          ErrFreshnessInvalid,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewService(
				test.username,
				test.statsFetcher,
				test.languagesFetcher,
				test.statsStore,
				test.languagesStore,
				test.freshness,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"NewService() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func assertMetadata(
	t *testing.T,
	updatedAt time.Time,
	refreshAfter time.Time,
	schemaVersion int,
	now time.Time,
) {
	t.Helper()

	if !updatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %s, want %s", updatedAt, now)
	}
	if !refreshAfter.Equal(now.Add(30 * time.Minute)) {
		t.Errorf(
			"RefreshAfter = %s, want %s",
			refreshAfter,
			now.Add(30*time.Minute),
		)
	}
	if schemaVersion != snapshot.CurrentSchemaVersion {
		t.Errorf(
			"SchemaVersion = %d, want %d",
			schemaVersion,
			snapshot.CurrentSchemaVersion,
		)
	}
}
