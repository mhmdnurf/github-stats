package server

import (
	"strings"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/config"
)

func TestNewDynamicServicesWrapsClientError(t *testing.T) {
	t.Parallel()

	_, err := newDynamicServices(config.Config{})
	if err == nil {
		t.Fatal("newDynamicServices() error = nil, want error")
	}

	const want = "create GitHub client: github token is required"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf(
			"newDynamicServices() error = %q, want containing %q",
			err,
			want,
		)
	}
}
