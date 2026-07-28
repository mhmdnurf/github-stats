package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/middleware"
)

const (
	dynamicRateLimitPerSecond = 1
	dynamicRateLimitBurst     = 10
)

func newRouter(
	configuration config.Config,
	serviceSet services,
	logger *slog.Logger,
) (http.Handler, error) {
	cardRenderer, err := card.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("create card renderer: %w", err)
	}

	languageRenderer, err := card.NewLanguageRenderer()
	if err != nil {
		return nil, fmt.Errorf("create language renderer: %w", err)
	}

	statsHandler, err := handler.NewStats(
		configuration.GitHubUsername,
		serviceSet.configuredStats,
		cardRenderer,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("create stats handler: %w", err)
	}

	languagesHandler, err := handler.NewLanguages(
		configuration.GitHubUsername,
		serviceSet.configuredLanguages,
		languageRenderer,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("create languages handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/stats", statsHandler)
	mux.Handle("/languages", languagesHandler)
	mux.HandleFunc("/health", healthHandler)

	if serviceSet.dynamic != nil {
		dynamicStatsHandler, err := handler.NewDynamicStats(
			serviceSet.dynamic.stats,
			cardRenderer,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create dynamic stats handler: %w",
				err,
			)
		}

		dynamicLanguagesHandler, err := handler.NewDynamicLanguages(
			serviceSet.dynamic.languages,
			languageRenderer,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create dynamic languages handler: %w",
				err,
			)
		}

		rateLimiter := middleware.NewRateLimiter(
			dynamicRateLimitPerSecond,
			dynamicRateLimitBurst,
		)

		mux.Handle(
			"/{username}/stats",
			rateLimiter.Middleware(dynamicStatsHandler),
		)
		mux.Handle(
			"/{username}/languages",
			rateLimiter.Middleware(dynamicLanguagesHandler),
		)
	}

	return mux, nil
}

func healthHandler(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writer.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)

	writer.WriteHeader(http.StatusOK)

	_, _ = writer.Write([]byte("ok\n"))
}
