package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

func newDynamicStatsRequest(
	t *testing.T,
	target string,
	pathUsername string,
) *http.Request {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.SetPathValue("username", pathUsername)

	return request
}

func TestDynamicStatsHandlerUsesUsernameFromPath(t *testing.T) {
	wantUsername := "octocat"
	wantStats := stats.UserStats{
		Name:     "The Octocat",
		Username: wantUsername,
		Stars:    42,
	}
	wantDocument := []byte("<svg></svg>")

	service := statsServiceStub{
		get: func(
			_ context.Context,
			username string,
			scope repositoryScope.Scope,
		) (stats.UserStats, error) {
			if username != wantUsername {
				t.Fatalf(
					"unexpected username: got %q, want %q",
					username,
					wantUsername,
				)
			}

			if scope != repositoryScope.ScopePublic {
				t.Fatalf("unexpected scope: got %q", scope)
			}

			return wantStats, nil
		},
	}

	renderer := cardRendererStub{
		render: func(
			stats.UserStats,
			string,
		) ([]byte, error) {
			return wantDocument, nil
		},
	}

	handler, err := NewDynamicStats(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicStatsRequest(
		t,
		"/octocat/stats",
		wantUsername,
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}

	if response.Body.String() != string(wantDocument) {
		t.Fatalf(
			"unexpected body: got %q, want %q",
			response.Body.String(),
			wantDocument,
		)
	}
}

func TestDynamicStatsHandlerRejectsInvalidUsername(t *testing.T) {
	service := statsServiceStub{
		get: func(
			context.Context,
			string,
			repositoryScope.Scope,
		) (stats.UserStats, error) {
			t.Fatal("service should not be called for an invalid username")
			return stats.UserStats{}, nil
		},
	}

	renderer := cardRendererStub{
		render: func(
			stats.UserStats,
			string,
		) ([]byte, error) {
			t.Fatal("renderer should not be called for an invalid username")
			return nil, nil
		},
	}

	handler, err := NewDynamicStats(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicStatsRequest(
		t,
		"/-bad-/stats",
		"-bad-",
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}
}

func TestDynamicStatsHandlerRejectsAllRepositoriesScope(t *testing.T) {
	service := statsServiceStub{
		get: func(
			context.Context,
			string,
			repositoryScope.Scope,
		) (stats.UserStats, error) {
			t.Fatal("service should not be called when scope is rejected")
			return stats.UserStats{}, nil
		},
	}

	renderer := cardRendererStub{
		render: func(
			stats.UserStats,
			string,
		) ([]byte, error) {
			t.Fatal("renderer should not be called when scope is rejected")
			return nil, nil
		},
	}

	handler, err := NewDynamicStats(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicStatsRequest(
		t,
		"/octocat/stats?repositories=all",
		"octocat",
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}
}

func TestDynamicStatsHandlerAllowsPublicRepositoriesScope(t *testing.T) {
	wantUsername := "octocat"
	wantStats := stats.UserStats{Username: wantUsername}
	wantDocument := []byte("<svg></svg>")

	service := statsServiceStub{
		get: func(
			_ context.Context,
			username string,
			scope repositoryScope.Scope,
		) (stats.UserStats, error) {
			if scope != repositoryScope.ScopePublic {
				t.Fatalf("unexpected scope: got %q", scope)
			}

			return wantStats, nil
		},
	}

	renderer := cardRendererStub{
		render: func(
			stats.UserStats,
			string,
		) ([]byte, error) {
			return wantDocument, nil
		},
	}

	handler, err := NewDynamicStats(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicStatsRequest(
		t,
		"/octocat/stats?repositories=public",
		wantUsername,
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusOK,
		)
	}
}
