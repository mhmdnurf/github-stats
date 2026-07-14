package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestEnv(
	t *testing.T,
	content string,
) string {
	t.Helper()

	filename := filepath.Join(
		t.TempDir(),
		".env",
	)

	if err := os.WriteFile(
		filename,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf("write test env: %v", err)
	}

	return filename
}

func unsetEnvironment(
	t *testing.T,
	name string,
) {
	t.Helper()

	previous, found := os.LookupEnv(name)

	if err := os.Unsetenv(name); err != nil {
		t.Fatalf("unset %s: %v", name, err)
	}

	t.Cleanup(func() {
		if found {
			_ = os.Setenv(name, previous)
			return
		}

		_ = os.Unsetenv(name)
	})
}

func TestLoadFromFile(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n"+
			"HTTP_ADDRESS=:7000\n",
	)

	got, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := Config{
		GitHubToken:    "file-token",
		GitHubUsername: "file-user",
		HTTPAddress:    ":7000",
	}

	if got != want {
		t.Fatalf(
			"unexpected config: got %+v, want %+v",
			got,
			want,
		)
	}
}

func TestLoadEnvironmentTakesPrecedence(t *testing.T) {
	t.Setenv("GITHUB_USERNAME", "environment-user")
	t.Setenv("GITHUB_TOKEN", "environment-token")
	t.Setenv("HTTP_ADDRESS", ":7000")

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n"+
			"HTTP_ADDRESS=:7000\n",
	)

	got, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := Config{
		GitHubToken:    "environment-token",
		GitHubUsername: "environment-user",
		HTTPAddress:    ":7000",
	}

	if got != want {
		t.Fatalf(
			"unexpected config: got %+v, want %+v",
			got,
			want,
		)
	}
}

func TestLoadUsesDefaultAddress(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n",
	)

	got, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got.HTTPAddress != defaultHTTPAddress {
		t.Fatalf(
			"unexpected address: got %q, want %q",
			got.HTTPAddress,
			defaultHTTPAddress,
		)
	}
}

func TestLoadRequiresGitHubToken(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")

	filename := writeTestEnv(t, "GITHUB_USERNAME=file-user\n"+"")

	config, err := load(filename)
	if err == nil {
		t.Fatal("expected an error")
	}

	if config != (Config{}) {
		t.Fatalf("expected zero config, got %+v", config)
	}

	if !strings.Contains(
		err.Error(),
		"GITHUB_TOKEN is required",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsMalformedFile(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")

	filename := writeTestEnv(
		t,
		"GITHUB_TOKEN=\"unterminated\n",
	)

	config, err := load(filename)
	if err == nil {
		t.Fatal("expected an error")
	}

	if config != (Config{}) {
		t.Fatalf("expected zero config, got %+v", config)
	}

	if !strings.Contains(err.Error(), "load") {
		t.Fatalf("unexpected error: %v", err)
	}
}
