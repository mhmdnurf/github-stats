package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

func newDynamicLanguagesRequest(
	t *testing.T,
	target string,
	pathUsername string,
) *http.Request {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.SetPathValue("username", pathUsername)

	return request
}

func TestDynamicLanguagesHandlerUsesUsernameFromPath(t *testing.T) {
	wantUsername := "octocat"
	wantLanguages := languages.UserLanguages{Username: wantUsername}
	wantDocument := []byte("<svg></svg>")

	service := languagesServiceStub{
		get: func(
			_ context.Context,
			username string,
			scope repositoryScope.Scope,
		) (languages.UserLanguages, error) {
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

			return wantLanguages, nil
		},
	}

	renderer := languageCardRendererStub{
		render: func(
			languages.UserLanguages,
			string,
		) ([]byte, error) {
			return wantDocument, nil
		},
	}

	handler, err := NewDynamicLanguages(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicLanguagesRequest(
		t,
		"/octocat/languages",
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

func TestDynamicLanguagesHandlerRejectsInvalidUsername(t *testing.T) {
	service := languagesServiceStub{
		get: func(
			context.Context,
			string,
			repositoryScope.Scope,
		) (languages.UserLanguages, error) {
			t.Fatal("service should not be called for an invalid username")
			return languages.UserLanguages{}, nil
		},
	}

	renderer := languageCardRendererStub{
		render: func(
			languages.UserLanguages,
			string,
		) ([]byte, error) {
			t.Fatal("renderer should not be called for an invalid username")
			return nil, nil
		},
	}

	handler, err := NewDynamicLanguages(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicLanguagesRequest(
		t,
		"/-bad-/languages",
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

func TestDynamicLanguagesHandlerRejectsAllRepositoriesScope(t *testing.T) {
	service := languagesServiceStub{
		get: func(
			context.Context,
			string,
			repositoryScope.Scope,
		) (languages.UserLanguages, error) {
			t.Fatal("service should not be called when scope is rejected")
			return languages.UserLanguages{}, nil
		},
	}

	renderer := languageCardRendererStub{
		render: func(
			languages.UserLanguages,
			string,
		) ([]byte, error) {
			t.Fatal("renderer should not be called when scope is rejected")
			return nil, nil
		},
	}

	handler, err := NewDynamicLanguages(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := newDynamicLanguagesRequest(
		t,
		"/octocat/languages?repositories=all",
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
