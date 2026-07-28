package handler

import (
	"errors"
	"net/http"

	"github.com/mhmdnurf/github-stats/internal/languages"
)

func (handler *Languages) ServeHTTP(
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

	userLanguages, err := handler.service.Get(
		request.Context(),
		options.username,
		options.scope,
	)
	if err != nil {
		if writeContextError(writer, err) {
			return
		}

		if errors.Is(err, languages.ErrUserNotFound) {
			writeError(
				writer,
				http.StatusNotFound,
				"GitHub user not found",
			)
			return
		}

		if errors.Is(err, languages.ErrUnavailable) {
			writer.Header().Set("Retry-After", "60")
			writeError(
				writer,
				http.StatusServiceUnavailable,
				"languages snapshot unavailable",
			)
			return
		}

		handler.logger.ErrorContext(
			request.Context(),
			"get GitHub user languages",
			"username",
			options.username,
			"error",
			err,
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"failed to load GitHub languages",
		)
		return
	}

	document, err := handler.renderer.Render(
		userLanguages,
		options.themeName,
	)
	if err != nil {
		handler.logger.ErrorContext(
			request.Context(),
			"render GitHub language card",
			"username",
			options.username,
			"error",
			err,
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"failed to render language card",
		)
		return
	}

	if err := writeSVGResponse(writer, document); err != nil {
		handler.logger.ErrorContext(
			request.Context(),
			"write language SVG response",
			"username",
			options.username,
			"error",
			err,
		)
	}
}
