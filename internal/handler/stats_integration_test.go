package handler_test

import (
	"context"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mhmdnurf/github-stats/internal/cache"
	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type integrationFetcher struct {
	mu        sync.Mutex
	value     stats.UserStats
	usernames []string
}

const configuredTestUsername = "mhmdnurf"

func (fetcher *integrationFetcher) Fetch(
	ctx context.Context,
	username string,
) (stats.UserStats, error) {
	if err := ctx.Err(); err != nil {
		return stats.UserStats{}, err
	}

	fetcher.mu.Lock()
	defer fetcher.mu.Unlock()

	fetcher.usernames = append(
		fetcher.usernames,
		username,
	)

	return fetcher.value, nil
}

func (fetcher *integrationFetcher) calls() []string {
	fetcher.mu.Lock()
	defer fetcher.mu.Unlock()

	return append([]string(nil), fetcher.usernames...)
}

func TestStatsHandlerEndToEnd(t *testing.T) {
	wantStats := stats.UserStats{
		Name:         "Muhammad Nurfatkhur Rahman",
		Username:     "mhmdnurf",
		Repositories: 101,
		Stars:        202,
		Commits:      303,
		PullRequests: 404,
		Followers:    505,
	}

	fetcher := &integrationFetcher{
		value: wantStats,
	}

	memoryCache := cache.NewMemory()

	service, err := stats.NewService(
		fetcher,
		memoryCache,
		time.Minute,
	)
	if err != nil {
		t.Fatalf("create stats service: %v", err)
	}

	renderer, err := card.NewRenderer()
	if err != nil {
		t.Fatalf("create card renderer: %v", err)
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	statsHandler, err := handler.NewStats(
		configuredTestUsername,
		service,
		renderer,
		logger,
	)
	if err != nil {
		t.Fatalf("create stats handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/stats", statsHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	requestURLs := []string{
		server.URL +
			"/stats?username=%20MHMDNURF%20&theme=light",
		server.URL +
			"/stats?username=mhmdnurf&theme=light",
	}

	var firstDocument []byte

	for index, requestURL := range requestURLs {
		response, err := server.Client().Get(requestURL)
		if err != nil {
			t.Fatalf(
				"request %d: %v",
				index+1,
				err,
			)
		}

		document, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()

		if readErr != nil {
			t.Fatalf(
				"read response %d: %v",
				index+1,
				readErr,
			)
		}

		if closeErr != nil {
			t.Fatalf(
				"close response %d: %v",
				index+1,
				closeErr,
			)
		}

		if response.StatusCode != http.StatusOK {
			t.Fatalf(
				"response %d status: got %d, want %d; body: %s",
				index+1,
				response.StatusCode,
				http.StatusOK,
				document,
			)
		}

		if got := response.Header.Get("Content-Type"); got !=
			"image/svg+xml; charset=utf-8" {
			t.Errorf(
				"response %d Content-Type: %q",
				index+1,
				got,
			)
		}

		if got := response.Header.Get("Cache-Control"); got !=
			"public, max-age=300" {
			t.Errorf(
				"response %d Cache-Control: %q",
				index+1,
				got,
			)
		}

		if got := response.Header.Get(
			"X-Content-Type-Options",
		); got != "nosniff" {
			t.Errorf(
				"response %d X-Content-Type-Options: %q",
				index+1,
				got,
			)
		}

		var root struct {
			XMLName xml.Name
		}

		if err := xml.Unmarshal(document, &root); err != nil {
			t.Fatalf(
				"response %d is not valid XML: %v",
				index+1,
				err,
			)
		}

		if root.XMLName.Local != "svg" {
			t.Fatalf(
				"response %d root element: got %q, want %q",
				index+1,
				root.XMLName.Local,
				"svg",
			)
		}

		output := string(document)

		expectedValues := []string{
			"Muhammad Nurfatkhur Rahman",
			"@mhmdnurf",
			"REPOSITORIES",
			"101",
			"STARS",
			"202",
			"COMMITS",
			"303",
			"PULL REQUESTS",
			"404",
			"FOLLOWERS",
			"505",
			"#ffffff",
			"#0969da",
		}

		for _, expected := range expectedValues {
			if !strings.Contains(output, expected) {
				t.Errorf(
					"response %d does not contain %q",
					index+1,
					expected,
				)
			}
		}

		if index == 0 {
			firstDocument = append(
				[]byte(nil),
				document...,
			)
		} else if string(document) != string(firstDocument) {
			t.Fatal(
				"expected cached request to produce identical SVG",
			)
		}
	}

	calls := fetcher.calls()

	if len(calls) != 1 {
		t.Fatalf(
			"expected one fetch call after two HTTP requests, got %d",
			len(calls),
		)
	}

	if calls[0] != "mhmdnurf" {
		t.Fatalf(
			"expected normalized username, got %q",
			calls[0],
		)
	}
}
