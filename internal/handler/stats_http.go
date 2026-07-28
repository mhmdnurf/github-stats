package handler

import (
	"errors"
	"net/http"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

func (handler *Stats) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if !allowGET(writer, request) {
		return
	}

	request, cancel := withCardRequestTimeout(request)
	defer cancel()

	options, failure := parseCardRequestOptions(
		request,
		handler.username,
		handler.dynamicUsername,
		handler.renderer,
	)
	if failure != nil {
		writeError(
			writer,
			failure.status,
			failure.message,
		)
		return
	}

	userStats, err := handler.service.Get(
		request.Context(),
		options.username,
		options.scope,
	)
	if err != nil {
		if writeContextError(writer, err) {
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

		if errors.Is(err, stats.ErrUnavailable) {
			writer.Header().Set("Retry-After", "60")
			writeError(
				writer,
				http.StatusServiceUnavailable,
				"statistics snapshot unavailable",
			)
			return
		}

		handler.logger.ErrorContext(
			request.Context(),
			"get GitHub user stats",
			"username",
			options.username,
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
		options.themeName,
	)
	if err != nil {
		handler.logger.ErrorContext(
			request.Context(),
			"render GitHub statistics card",
			"username",
			options.username,
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

	if err := writeSVGResponse(writer, document); err != nil {
		handler.logger.ErrorContext(
			request.Context(),
			"write SVG response",
			"username",
			options.username,
			"error",
			err,
		)
	}
}
