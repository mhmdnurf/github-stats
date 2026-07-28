package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type routerStatsService struct{}

func (routerStatsService) Get(
	_ context.Context,
	username string,
	_ repositoryScope.Scope,
) (stats.UserStats, error) {
	return stats.UserStats{
		Username: username,
	}, nil
}

type routerLanguagesService struct{}

func (routerLanguagesService) Get(
	_ context.Context,
	username string,
	scope repositoryScope.Scope,
) (languages.UserLanguages, error) {
	return languages.UserLanguages{
		Username: username,
		Scope:    scope,
	}, nil
}

func newTestRouter(
	t *testing.T,
	dynamicEnabled bool,
) http.Handler {
	t.Helper()

	serviceSet := services{
		configuredStats:     routerStatsService{},
		configuredLanguages: routerLanguagesService{},
	}
	if dynamicEnabled {
		serviceSet.dynamic = &dynamicServices{
			stats:     routerStatsService{},
			languages: routerLanguagesService{},
		}
	}

	router, err := newRouter(
		config.Config{
			GitHubUsername: "mhmdnurf",
		},
		serviceSet,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	return router
}

func TestRouterWithoutDynamicServices(t *testing.T) {
	router := newTestRouter(t, false)

	tests := []struct {
		path       string
		wantStatus int
	}{
		{
			path:       "/stats",
			wantStatus: http.StatusOK,
		},
		{
			path:       "/languages",
			wantStatus: http.StatusOK,
		},
		{
			path:       "/octocat/stats",
			wantStatus: http.StatusNotFound,
		},
		{
			path:       "/octocat/languages",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				test.path,
				nil,
			)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf(
					"status: got %d, want %d",
					response.Code,
					test.wantStatus,
				)
			}
		})
	}
}

func TestRouterWithDynamicServices(t *testing.T) {
	router := newTestRouter(t, true)

	for _, path := range []string{
		"/octocat/stats",
		"/octocat/languages",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf(
					"status: got %d, want %d",
					response.Code,
					http.StatusOK,
				)
			}
		})
	}
}

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
