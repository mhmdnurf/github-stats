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

func unsetSnapshotEnvironment(t *testing.T) {
	t.Helper()

	unsetEnvironment(t, "GOOGLE_CLOUD_PROJECT")
	unsetEnvironment(t, "FIRESTORE_COLLECTION")
}

func TestLoadFromFile(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")
	unsetSnapshotEnvironment(t)

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n"+
			"HTTP_ADDRESS=:7000\n"+
			"GOOGLE_CLOUD_PROJECT=file-project\n"+
			"FIRESTORE_COLLECTION=file-snapshots\n",
	)

	got, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := Config{
		GitHubToken:          "file-token",
		GitHubUsername:       "file-user",
		HTTPAddress:          ":7000",
		GoogleCloudProjectID: "file-project",
		FirestoreCollection:  "file-snapshots",
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
	t.Setenv("GOOGLE_CLOUD_PROJECT", "environment-project")
	t.Setenv("FIRESTORE_COLLECTION", "environment-snapshots")

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n"+
			"HTTP_ADDRESS=:7000\n"+
			"GOOGLE_CLOUD_PROJECT=file-project\n"+
			"FIRESTORE_COLLECTION=file-snapshots\n",
	)

	got, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := Config{
		GitHubToken:          "environment-token",
		GitHubUsername:       "environment-user",
		HTTPAddress:          ":7000",
		GoogleCloudProjectID: "environment-project",
		FirestoreCollection:  "environment-snapshots",
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
	unsetSnapshotEnvironment(t)

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GITHUB_TOKEN=file-token\n"+
			"GOOGLE_CLOUD_PROJECT=file-project\n",
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

	if got.FirestoreCollection != defaultFirestoreCollection {
		t.Fatalf(
			"unexpected Firestore collection: got %q, want %q",
			got.FirestoreCollection,
			defaultFirestoreCollection,
		)
	}
}

func TestLoadAllowsMissingGitHubToken(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")
	unsetSnapshotEnvironment(t)

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n"+
			"GOOGLE_CLOUD_PROJECT=file-project\n",
	)

	configuration, err := load(filename)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if configuration.GitHubToken != "" {
		t.Fatalf(
			"GitHubToken = %q, want empty",
			configuration.GitHubToken,
		)
	}
}

func TestLoadRequiresGoogleCloudProject(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")
	unsetSnapshotEnvironment(t)

	filename := writeTestEnv(
		t,
		"GITHUB_USERNAME=file-user\n",
	)

	configuration, err := load(filename)
	if err == nil {
		t.Fatal("expected an error")
	}

	if configuration != (Config{}) {
		t.Fatalf(
			"expected zero config, got %+v",
			configuration,
		)
	}

	if !strings.Contains(
		err.Error(),
		"GOOGLE_CLOUD_PROJECT is required",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsMalformedFile(t *testing.T) {
	unsetEnvironment(t, "GITHUB_USERNAME")
	unsetEnvironment(t, "GITHUB_TOKEN")
	unsetEnvironment(t, "HTTP_ADDRESS")
	unsetSnapshotEnvironment(t)

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
