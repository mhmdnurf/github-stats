package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type statsServiceStub struct {
	get func(context.Context, string) (stats.UserStats, error)
}

func (stub statsServiceStub) Get(
	ctx context.Context,
	username string,
) (stats.UserStats, error) {
	return stub.get(ctx, username)
}

type cardRendererStub struct {
	render func(stats.UserStats, string) ([]byte, error)
}

func (stub cardRendererStub) Render(
	userStats stats.UserStats,
	themeName string,
) ([]byte, error) {
	return stub.render(userStats, themeName)
}

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}

func TestValidGitHubUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{
			name:     "single character",
			username: "a",
			want:     true,
		},
		{
			name:     "letters numbers and dash",
			username: "octocat-123",
			want:     true,
		},
		{
			name:     "uppercase letters",
			username: "OctoCat",
			want:     true,
		},
		{
			name:     "maximum length",
			username: strings.Repeat("a", 39),
			want:     true,
		},
		{
			name:     "empty",
			username: "",
			want:     false,
		},
		{
			name:     "too long",
			username: strings.Repeat("a", 40),
			want:     false,
		},
		{
			name:     "starts with dash",
			username: "-octocat",
			want:     false,
		},
		{
			name:     "ends with dash",
			username: "octocat-",
			want:     false,
		},
		{
			name:     "consecutive dashes",
			username: "octo--cat",
			want:     false,
		},
		{
			name:     "underscore",
			username: "octo_cat",
			want:     false,
		},
		{
			name:     "unicode",
			username: "öctocat",
			want:     false,
		},
		{
			name:     "whitespace",
			username: "octo cat",
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validGitHubUsername(test.username)
			if got != test.want {
				t.Fatalf(
					"validGitHubUsername(%q): got %t, want %t",
					test.username,
					got,
					test.want,
				)
			}
		})
	}
}

func TestStatsHandlerReturnsSVG(t *testing.T) {
	type contextKey string
	const requestKey contextKey = "request-id"

	wantStats := stats.UserStats{
		Name:     "Muhammad Nurfatkhur Rahman",
		Username: "mhmdnurf",
		Stars:    128,
	}
	wantDocument := []byte("<svg></svg>")

	service := statsServiceStub{
		get: func(
			ctx context.Context,
			username string,
		) (stats.UserStats, error) {
			if username != "MHMDNURF" {
				t.Fatalf("unexpected username: %q", username)
			}

			if ctx.Value(requestKey) != "request-123" {
				t.Fatal("request context was not forwarded")
			}

			return wantStats, nil
		},
	}

	renderer := cardRendererStub{
		render: func(
			userStats stats.UserStats,
			themeName string,
		) ([]byte, error) {
			if userStats != wantStats {
				t.Fatalf(
					"unexpected stats: got %+v, want %+v",
					userStats,
					wantStats,
				)
			}

			if themeName != card.LightTheme {
				t.Fatalf("unexpected theme: %q", themeName)
			}

			return wantDocument, nil
		},
	}

	handler, err := NewStats(service, renderer, testLogger())
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?username=%20MHMDNURF%20&theme=light",
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

	if response.Body.String() != string(wantDocument) {
		t.Fatalf(
			"unexpected body: got %q, want %q",
			response.Body.String(),
			wantDocument,
		)
	}

	expectedHeaders := map[string]string{
		"Content-Type":           "image/svg+xml; charset=utf-8",
		"X-Content-Type-Options": "nosniff",
		"Cache-Control":          "public, max-age=300",
	}

	for name, want := range expectedHeaders {
		if got := response.Header().Get(name); got != want {
			t.Errorf(
				"header %s: got %q, want %q",
				name,
				got,
				want,
			)
		}
	}
}

func TestStatsHandlerRejectsUnsupportedMethod(t *testing.T) {
	handler, err := NewStats(
		statsServiceStub{
			get: func(
				context.Context,
				string,
			) (stats.UserStats, error) {
				t.Fatal("service should not be called")
				return stats.UserStats{}, nil
			},
		},
		cardRendererStub{
			render: func(
				stats.UserStats,
				string,
			) ([]byte, error) {
				t.Fatal("renderer should not be called")
				return nil, nil
			},
		},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/stats?username=mhmdnurf",
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
		t.Fatalf("unexpected Cache-Control header: %q", got)
	}

	if response.Body.String() != "method not allowed\n" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}
}

func TestStatsHandlerRejectsInvalidUsername(t *testing.T) {
	handler, err := NewStats(
		statsServiceStub{
			get: func(
				context.Context,
				string,
			) (stats.UserStats, error) {
				t.Fatal("service should not be called")
				return stats.UserStats{}, nil
			},
		},
		cardRendererStub{
			render: func(
				stats.UserStats,
				string,
			) ([]byte, error) {
				t.Fatal("renderer should not be called")
				return nil, nil
			},
		},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?username=invalid_name",
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

	if response.Body.String() != "invalid GitHub username\n" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}
}

func TestStatsHandlerMapsErrors(t *testing.T) {
	serviceFailure := errors.New("service failed")
	renderFailure := errors.New("render failed")

	tests := []struct {
		name              string
		serviceError      error
		renderError       error
		wantStatus        int
		wantBody          string
		wantRendererCalls int
	}{
		{
			name:              "user not found",
			serviceError:      stats.ErrUserNotFound,
			wantStatus:        http.StatusNotFound,
			wantBody:          "GitHub user not found\n",
			wantRendererCalls: 0,
		},
		{
			name:              "service failure",
			serviceError:      serviceFailure,
			wantStatus:        http.StatusInternalServerError,
			wantBody:          "failed to load GitHub statistics\n",
			wantRendererCalls: 0,
		},
		{
			name:              "renderer reports unknown theme",
			renderError:       card.ErrUnknownTheme,
			wantStatus:        http.StatusBadRequest,
			wantBody:          "unknown card theme\n",
			wantRendererCalls: 1,
		},
		{
			name:              "renderer failure",
			renderError:       renderFailure,
			wantStatus:        http.StatusInternalServerError,
			wantBody:          "failed to render statistics card\n",
			wantRendererCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendererCalls := 0

			handler, err := NewStats(
				statsServiceStub{
					get: func(
						context.Context,
						string,
					) (stats.UserStats, error) {
						return stats.UserStats{
							Username: "mhmdnurf",
						}, test.serviceError
					},
				},
				cardRendererStub{
					render: func(
						stats.UserStats,
						string,
					) ([]byte, error) {
						rendererCalls++
						return []byte("<svg></svg>"), test.renderError
					},
				},
				testLogger(),
			)
			if err != nil {
				t.Fatalf("create handler: %v", err)
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/stats?username=mhmdnurf",
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

			if response.Body.String() != test.wantBody {
				t.Fatalf(
					"unexpected body: got %q, want %q",
					response.Body.String(),
					test.wantBody,
				)
			}

			if rendererCalls != test.wantRendererCalls {
				t.Fatalf(
					"renderer calls: got %d, want %d",
					rendererCalls,
					test.wantRendererCalls,
				)
			}

			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Fatalf(
					"unexpected Cache-Control header: %q",
					got,
				)
			}
		})
	}
}

func TestStatsHandlerRejectsUnknownThemeBeforeService(t *testing.T) {
	handler, err := NewStats(
		statsServiceStub{
			get: func(
				context.Context,
				string,
			) (stats.UserStats, error) {
				t.Fatal("service should not be called")
				return stats.UserStats{}, nil
			},
		},
		cardRendererStub{
			render: func(
				stats.UserStats,
				string,
			) ([]byte, error) {
				t.Fatal("renderer should not be called")
				return nil, nil
			},
		},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?username=mhmdnurf&theme=unknown",
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

	if response.Body.String() != "unknown card theme\n" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unexpected Cache-Control header: %q", got)
	}
}

func TestStatsHandlerMapsDeadlineExceeded(t *testing.T) {
	handler, err := NewStats(
		statsServiceStub{
			get: func(
				ctx context.Context,
				_ string,
			) (stats.UserStats, error) {
				if _, found := ctx.Deadline(); !found {
					t.Fatal("expected request deadline")
				}

				return stats.UserStats{}, context.DeadlineExceeded
			},
		},
		cardRendererStub{
			render: func(
				stats.UserStats,
				string,
			) ([]byte, error) {
				t.Fatal("renderer should not be called")
				return nil, nil
			},
		},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?username=mhmdnurf",
		nil,
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusGatewayTimeout,
		)
	}

	if response.Body.String() != "GitHub request timed out\n" {
		t.Fatalf("unexpected body: %q", response.Body.String())
	}

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("unexpected Cache-Control header: %q", got)
	}
}

func TestStatsHandlerStopsWhenRequestIsCanceled(t *testing.T) {
	handler, err := NewStats(
		statsServiceStub{
			get: func(
				ctx context.Context,
				_ string,
			) (stats.UserStats, error) {
				if !errors.Is(ctx.Err(), context.Canceled) {
					t.Fatalf(
						"expected canceled context, got %v",
						ctx.Err(),
					)
				}

				return stats.UserStats{}, ctx.Err()
			},
		},
		cardRendererStub{
			render: func(
				stats.UserStats,
				string,
			) ([]byte, error) {
				t.Fatal("renderer should not be called")
				return nil, nil
			},
		},
		testLogger(),
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats?username=mhmdnurf",
		nil,
	)

	ctx, cancel := context.WithCancel(request.Context())
	cancel()

	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Body.Len() != 0 {
		t.Fatalf(
			"expected no response body, got %q",
			response.Body.String(),
		)
	}
}
