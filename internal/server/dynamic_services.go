package server

import (
	"fmt"
	"net/http"
	"time"

	githubclient "github.com/mhmdnurf/github-stats/internal/adapter/github"
	"github.com/mhmdnurf/github-stats/internal/cache"
	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	dynamicGitHubRequestTimeout = 20 * time.Second
	dynamicCacheTTL             = 5 * time.Minute
)

type dynamicServices struct {
	stats     handler.StatsService
	languages handler.LanguagesService
}

func newDynamicServices(
	configuration config.Config,
) (dynamicServices, error) {
	githubClient, err := githubclient.NewClient(
		configuration.GitHubToken,
		&http.Client{
			Timeout: dynamicGitHubRequestTimeout,
		},
	)
	if err != nil {
		return dynamicServices{}, fmt.Errorf(
			"create GitHub client: %w",
			err,
		)
	}

	statsService, err := stats.NewService(
		githubClient,
		cache.NewMemory[stats.UserStats](),
		dynamicCacheTTL,
	)
	if err != nil {
		return dynamicServices{}, fmt.Errorf(
			"create dynamic stats service: %w",
			err,
		)
	}

	languagesService, err := languages.NewService(
		githubClient,
		cache.NewMemory[languages.UserLanguages](),
		dynamicCacheTTL,
	)
	if err != nil {
		return dynamicServices{}, fmt.Errorf(
			"create dynamic languages service: %w",
			err,
		)
	}

	return dynamicServices{
		stats:     statsService,
		languages: languagesService,
	}, nil
}
