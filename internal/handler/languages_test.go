package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

type languagesServiceStub struct {
	get func(
		context.Context,
		string,
		repositoryScope.Scope,
	) (languages.UserLanguages, error)
}

func (stub languagesServiceStub) Get(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (languages.UserLanguages, error) {
	return stub.get(ctx, username, scope)
}

type languageCardRendererStub struct {
	render func(languages.UserLanguages, string) ([]byte, error)
}

func (stub languageCardRendererStub) Render(
	userLanguages languages.UserLanguages,
	themeName string,
) ([]byte, error) {
	return stub.render(userLanguages, themeName)
}

func TestNewLanguagesRejectsInvalidConfiguration(t *testing.T) {
	validService := languagesServiceStub{}
	validRenderer := languageCardRendererStub{}

	tests := []struct {
		name     string
		username string
		service  LanguagesService
		renderer LanguageCardRenderer
		logger   bool
	}{
		{
			name:     "invalid username",
			username: "invalid_user",
			service:  validService,
			renderer: validRenderer,
			logger:   true,
		},
		{
			name:     "missing service",
			username: configuredTestUsername,
			service:  nil,
			renderer: validRenderer,
			logger:   true,
		},
		{
			name:     "missing renderer",
			username: configuredTestUsername,
			service:  validService,
			renderer: nil,
			logger:   true,
		},
		{
			name:     "missing logger",
			username: configuredTestUsername,
			service:  validService,
			renderer: validRenderer,
			logger:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logger = testLogger()
			if !test.logger {
				logger = nil
			}

			handler, err := NewLanguages(
				test.username,
				test.service,
				test.renderer,
				logger,
			)

			if err == nil {
				t.Fatal("expected an error")
			}

			if handler != nil {
				t.Fatalf("expected nil handler, got %+v", handler)
			}
		})
	}
}

func TestNewLanguagesNormalizesUsername(t *testing.T) {
	handler, err := NewLanguages(
		"  mhmdnurf  ",
		languagesServiceStub{},
		languageCardRendererStub{},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create languages handler: %v", err)
	}

	if handler.username != "mhmdnurf" {
		t.Fatalf(
			"unexpected username: got %q, want %q",
			handler.username,
			"mhmdnurf",
		)
	}
}

func TestLanguagesHandlerReturnsSVG(t *testing.T) {
	type contextKey string
	const requestKey contextKey = "request-id"

	wantLanguages := languages.UserLanguages{
		Username: configuredTestUsername,
		Languages: []languages.LanguageUsage{
			{Name: "Go", Bytes: 100},
		},
	}
	wantDocument := []byte("<svg></svg>")

	service := languagesServiceStub{
		get: func(
			ctx context.Context,
			username string,
			scope repositoryScope.Scope,
		) (languages.UserLanguages, error) {
			if username != configuredTestUsername {
				t.Fatalf(
					"unexpected username: got %q, want %q",
					username,
					configuredTestUsername,
				)
			}

			if scope != repositoryScope.ScopePublic {
				t.Fatalf("unexpected scope: %q", scope)
			}

			if ctx.Value(requestKey) != "request-123" {
				t.Fatal("request context was not forwarded")
			}

			return wantLanguages, nil
		},
	}

	renderer := languageCardRendererStub{
		render: func(
			userLanguages languages.UserLanguages,
			themeName string,
		) ([]byte, error) {
			if !reflect.DeepEqual(userLanguages, wantLanguages) {
				t.Fatalf(
					"unexpected languages: got %+v, want %+v",
					userLanguages,
					wantLanguages,
				)
			}

			if themeName != card.LightTheme {
				t.Fatalf(
					"unexpected theme: got %q, want %q",
					themeName,
					card.LightTheme,
				)
			}

			return wantDocument, nil
		},
	}

	handler, err := NewLanguages(
		configuredTestUsername,
		service,
		renderer,
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create languages handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/languages?theme=light",
		nil,
	)
	request = request.WithContext(
		context.WithValue(
			request.Context(),
			requestKey,
			"request-123",
		),
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

	if !bytes.Equal(response.Body.Bytes(), wantDocument) {
		t.Fatalf(
			"unexpected response body: got %q, want %q",
			response.Body.Bytes(),
			wantDocument,
		)
	}

	if got := response.Header().Get("Content-Type"); got !=
		"image/svg+xml; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", got)
	}

	if got := response.Header().Get("X-Content-Type-Options"); got !=
		"nosniff" {
		t.Fatalf("unexpected X-Content-Type-Options: %q", got)
	}

	if got := response.Header().Get("Cache-Control"); got !=
		"public, max-age=300" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestLanguagesHandlerRejectsUnsupportedMethod(t *testing.T) {
	handler, err := NewLanguages(
		configuredTestUsername,
		languagesServiceStub{},
		languageCardRendererStub{},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create languages handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/languages",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusMethodNotAllowed,
		)
	}

	if got := response.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected Allow header: %q", got)
	}

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestLanguagesHandlerRejectsUnknownThemeBeforeService(t *testing.T) {
	service := languagesServiceStub{
		get: func(
			context.Context,
			string,
			repositoryScope.Scope,
		) (languages.UserLanguages, error) {
			t.Fatal("service should not be called")
			return languages.UserLanguages{}, nil
		},
	}

	renderer := languageCardRendererStub{
		render: func(
			languages.UserLanguages,
			string,
		) ([]byte, error) {
			t.Fatal("renderer should not be called")
			return nil, nil
		},
	}

	handler, err := NewLanguages(
		configuredTestUsername,
		service,
		renderer,
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create languages handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/languages?theme=unknown",
		nil,
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

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
}

func TestLanguagesHandlerMapsErrors(t *testing.T) {
	unexpectedServiceError := errors.New("service failed")
	rendererError := errors.New("render failed")

	tests := []struct {
		name               string
		serviceError       error
		rendererError      error
		wantStatus         int
		wantRendererCalled bool
	}{
		{
			name:         "user not found",
			serviceError: languages.ErrUserNotFound,
			wantStatus:   http.StatusNotFound,
		},
		{
			name:         "deadline exceeded",
			serviceError: context.DeadlineExceeded,
			wantStatus:   http.StatusGatewayTimeout,
		},
		{
			name:         "unexpected service error",
			serviceError: unexpectedServiceError,
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name:               "renderer unknown theme",
			rendererError:      card.ErrUnknownTheme,
			wantStatus:         http.StatusBadRequest,
			wantRendererCalled: true,
		},
		{
			name:               "renderer error",
			rendererError:      rendererError,
			wantStatus:         http.StatusInternalServerError,
			wantRendererCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendererCalled := false

			service := languagesServiceStub{
				get: func(
					context.Context,
					string,
					repositoryScope.Scope,
				) (languages.UserLanguages, error) {
					return languages.UserLanguages{
						Username: configuredTestUsername,
					}, test.serviceError
				},
			}

			renderer := languageCardRendererStub{
				render: func(
					languages.UserLanguages,
					string,
				) ([]byte, error) {
					rendererCalled = true
					return nil, test.rendererError
				},
			}

			handler, err := NewLanguages(
				configuredTestUsername,
				service,
				renderer,
				testLogger(),
			)
			if err != nil {
				t.Fatalf("create languages handler: %v", err)
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/languages",
				nil,
			)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf(
					"unexpected status: got %d, want %d",
					response.Code,
					test.wantStatus,
				)
			}

			if rendererCalled != test.wantRendererCalled {
				t.Fatalf(
					"renderer called: got %t, want %t",
					rendererCalled,
					test.wantRendererCalled,
				)
			}

			if got := response.Header().Get("Cache-Control"); got !=
				"no-store" {
				t.Fatalf("unexpected Cache-Control: %q", got)
			}
		})
	}
}
