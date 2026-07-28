package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinBurst(t *testing.T) {
	limiter := NewRateLimiter(1, 3)

	handler := limiter.Middleware(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	))

	for i := 0; i < 3; i++ {
		request := httptest.NewRequest(http.MethodGet, "/octocat/stats", nil)
		request.RemoteAddr = "203.0.113.1:1234"

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf(
				"request %d: unexpected status: got %d, want %d",
				i,
				response.Code,
				http.StatusOK,
			)
		}
	}
}

func TestRateLimiterRejectsBeyondBurst(t *testing.T) {
	limiter := NewRateLimiter(1, 2)

	handler := limiter.Middleware(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	))

	newRequest := func() *http.Request {
		request := httptest.NewRequest(http.MethodGet, "/octocat/stats", nil)
		request.RemoteAddr = "203.0.113.2:1234"
		return request
	}

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, newRequest())

		if response.Code != http.StatusOK {
			t.Fatalf(
				"request %d: unexpected status: got %d, want %d",
				i,
				response.Code,
				http.StatusOK,
			)
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, newRequest())

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			response.Code,
			http.StatusTooManyRequests,
		)
	}
}

func TestRateLimiterTracksClientsIndependently(t *testing.T) {
	limiter := NewRateLimiter(1, 1)

	handler := limiter.Middleware(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	))

	first := httptest.NewRequest(http.MethodGet, "/a/stats", nil)
	first.RemoteAddr = "203.0.113.3:1234"

	second := httptest.NewRequest(http.MethodGet, "/b/stats", nil)
	second.RemoteAddr = "203.0.113.4:1234"

	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	if firstResponse.Code != http.StatusOK {
		t.Fatalf("unexpected status for first client: %d", firstResponse.Code)
	}

	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)
	if secondResponse.Code != http.StatusOK {
		t.Fatalf("unexpected status for second client: %d", secondResponse.Code)
	}
}

func TestRateLimiterRefillsOverTime(t *testing.T) {
	limiter := NewRateLimiter(1, 1)

	current := time.Unix(0, 0)
	limiter.now = func() time.Time { return current }

	handler := limiter.Middleware(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	))

	newRequest := func() *http.Request {
		request := httptest.NewRequest(http.MethodGet, "/octocat/stats", nil)
		request.RemoteAddr = "203.0.113.5:1234"
		return request
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, newRequest())
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}

	blocked := httptest.NewRecorder()
	handler.ServeHTTP(blocked, newRequest())
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %d", blocked.Code)
	}

	current = current.Add(2 * time.Second)

	refilled := httptest.NewRecorder()
	handler.ServeHTTP(refilled, newRequest())
	if refilled.Code != http.StatusOK {
		t.Fatalf("unexpected status after refill: %d", refilled.Code)
	}
}
