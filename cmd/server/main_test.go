package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/healthz",
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

func TestRunRequiresGitHubToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	err := run(logger)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "GITHUB_TOKEN is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
