package handler

import (
	"context"
	"errors"
	"net/http"
)

func writeContextError(
	writer http.ResponseWriter,
	err error,
) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		writeError(
			writer,
			http.StatusGatewayTimeout,
			"GitHub request timed out",
		)
		return true
	}

	return false
}

func writeSVGResponse(
	writer http.ResponseWriter,
	document []byte,
) error {
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

	_, err := writer.Write(document)
	return err
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
