package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type StatsService interface {
	Get(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (stats.UserStats, error)
}

type ThemeValidator interface {
	SupportsTheme(themeName string) bool
}

type CardRenderer interface {
	ThemeValidator

	Render(
		userStats stats.UserStats,
		themeName string,
	) ([]byte, error)
}

type Stats struct {
	username        string
	dynamicUsername bool
	service         StatsService
	renderer        CardRenderer
	logger          *slog.Logger
}

func NewStats(
	username string,
	service StatsService,
	renderer CardRenderer,
	logger *slog.Logger,
) (*Stats, error) {
	normalizedUsername := strings.TrimSpace(username)
	if !validGitHubUsername(normalizedUsername) {
		return nil, errors.New("valid GitHub username is required")
	}

	if service == nil {
		return nil, errors.New("stats service is required")
	}

	if renderer == nil {
		return nil, errors.New("card renderer is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &Stats{
		username: normalizedUsername,
		service:  service,
		renderer: renderer,
		logger:   logger,
	}, nil
}

func NewDynamicStats(
	service StatsService,
	renderer CardRenderer,
	logger *slog.Logger,
) (*Stats, error) {
	if service == nil {
		return nil, errors.New("stats service is required")
	}

	if renderer == nil {
		return nil, errors.New("card renderer is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &Stats{
		dynamicUsername: true,
		service:         service,
		renderer:        renderer,
		logger:          logger,
	}, nil
}
