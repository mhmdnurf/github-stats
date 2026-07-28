package server

import (
	"context"
	"errors"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type preloadStatsStub struct {
	scopes []repositoryScope.Scope
	err    error
}

func (stub *preloadStatsStub) Get(
	_ context.Context,
	_ string,
	scope repositoryScope.Scope,
) (stats.UserStats, error) {
	stub.scopes = append(stub.scopes, scope)
	return stats.UserStats{}, stub.err
}

type preloadLanguagesStub struct {
	scopes []repositoryScope.Scope
	err    error
}

func (stub *preloadLanguagesStub) Get(
	_ context.Context,
	_ string,
	scope repositoryScope.Scope,
) (languages.UserLanguages, error) {
	stub.scopes = append(stub.scopes, scope)
	return languages.UserLanguages{}, stub.err
}

func TestPreloadSnapshotsLoadsEveryScope(t *testing.T) {
	statsService := &preloadStatsStub{}
	languagesService := &preloadLanguagesStub{}

	if err := preloadSnapshots(
		context.Background(),
		"mhmdnurf",
		statsService,
		languagesService,
	); err != nil {
		t.Fatalf("preloadSnapshots() error = %v", err)
	}

	wantScopes := []repositoryScope.Scope{
		repositoryScope.ScopePublic,
		repositoryScope.ScopeAll,
	}
	assertScopes(t, statsService.scopes, wantScopes)
	assertScopes(t, languagesService.scopes, wantScopes)
}

func TestPreloadSnapshotsStopsOnFailure(t *testing.T) {
	preloadFailure := errors.New("snapshot missing")
	statsService := &preloadStatsStub{
		err: preloadFailure,
	}
	languagesService := &preloadLanguagesStub{}

	err := preloadSnapshots(
		context.Background(),
		"mhmdnurf",
		statsService,
		languagesService,
	)
	if !errors.Is(err, preloadFailure) {
		t.Fatalf(
			"preloadSnapshots() error = %v, want %v",
			err,
			preloadFailure,
		)
	}

	if len(languagesService.scopes) != 0 {
		t.Fatalf(
			"languages calls = %d, want 0",
			len(languagesService.scopes),
		)
	}
}

func assertScopes(
	t *testing.T,
	got []repositoryScope.Scope,
	want []repositoryScope.Scope,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf(
			"scope count = %d, want %d",
			len(got),
			len(want),
		)
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf(
				"scope[%d] = %q, want %q",
				index,
				got[index],
				want[index],
			)
		}
	}
}
