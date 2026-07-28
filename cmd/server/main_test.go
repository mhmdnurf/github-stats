package main

import (
	"io"
	"log/slog"
	"testing"
)

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
