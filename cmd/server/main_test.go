package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)
	response := httptest.NewRecorder()

	healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	if response.Body.String() != "ok\n" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}

	if got := response.Header().Get("Content-Type"); got !=
		"text/plain; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", got)
	}

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

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

func TestRunRequiresGoogleCloudProject(t *testing.T) {
	t.Setenv("GITHUB_USERNAME", "test-user")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	err := run(logger)
	if err == nil {
		t.Fatal("expected an error")
	}

	const want = "load configuration: GOOGLE_CLOUD_PROJECT is required"
	if err.Error() != want {
		t.Fatalf("unexpected error: %v", err)
	}
}
