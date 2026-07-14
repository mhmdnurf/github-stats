package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

type StatsService interface {
	Get(
		ctx context.Context,
		username string,
	) (stats.UserStats, error)
}

type CardRenderer interface {
	Render(
		userStats stats.UserStats,
		themeName string,
	) ([]byte, error)
}

type Stats struct {
	username string
	service  StatsService
	renderer CardRenderer
	logger   *slog.Logger
}

const statsRequestTimeout = 25 * time.Second

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

func (handler *Stats) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", http.MethodGet)
		writeError(
			writer,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		request.Context(),
		statsRequestTimeout,
	)
	defer cancel()

	request = request.WithContext(ctx)

	themeName := request.URL.Query().Get("theme")
	if _, err := card.ResolveTheme(themeName); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"unknown card theme",
		)
		return
	}

	userStats, err := handler.service.Get(
		request.Context(),
		handler.username,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			writeError(
				writer,
				http.StatusGatewayTimeout,
				"GitHub request timed out",
			)
			return
		}

		if errors.Is(err, stats.ErrUserNotFound) {
			writeError(
				writer,
				http.StatusNotFound,
				"GitHub user not found",
			)
			return
		}

		handler.logger.ErrorContext(
			request.Context(),
			"get GitHub user stats",
			"username",
			handler.username,
			"error",
			err,
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"failed to load GitHub statistics",
		)
		return
	}

	document, err := handler.renderer.Render(
		userStats,
		themeName,
	)
	if err != nil {
		if errors.Is(err, card.ErrUnknownTheme) {
			writeError(
				writer,
				http.StatusBadRequest,
				"unknown card theme",
			)
			return
		}

		handler.logger.ErrorContext(
			request.Context(),
			"render GitHub statistics card",
			"username",
			handler.username,
			"error",
			err,
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"failed to render statistics card",
		)
		return
	}

	writer.Header().Set(
		"Content-Type",
		"image/svg+xml; charset=utf-8",
	)
	writer.Header().Set(
		"X-Content-Type-Options",
		"nosniff",
	)
	writer.Header().Set(
		"Cache-Control",
		"public, max-age=300",
	)

	if _, err := writer.Write(document); err != nil {
		handler.logger.ErrorContext(
			request.Context(),
			"write SVG response",
			"username",
			handler.username,
			"error",
			err,
		)
	}
}

func validGitHubUsername(username string) bool {
	if len(username) == 0 || len(username) > 39 {
		return false
	}

	for index := 0; index < len(username); index++ {
		character := username[index]

		isLetter :=
			character >= 'a' && character <= 'z' ||
				character >= 'A' && character <= 'Z'
		isNumber :=
			character >= '0' && character <= '9'

		if isLetter || isNumber {
			continue
		}

		if character != '-' {
			return false
		}

		if index == 0 || index == len(username)-1 {
			return false
		}

		if username[index-1] == '-' {
			return false
		}
	}

	return true
}

func writeError(
	writer http.ResponseWriter,
	status int,
	message string,
) {
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)
	http.Error(
		writer,
		message,
		status,
	)
}
