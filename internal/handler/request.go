package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mhmdnurf/github-stats/internal/card"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const cardRequestTimeout = 25 * time.Second

type cardRequestOptions struct {
	username  string
	themeName string
	scope     repositoryScope.Scope
}

type cardRequestFailure struct {
	status  int
	message string
}

func allowGET(
	writer http.ResponseWriter,
	request *http.Request,
) bool {
	if request.Method == http.MethodGet {
		return true
	}

	writer.Header().Set("Allow", http.MethodGet)
	writeError(
		writer,
		http.StatusMethodNotAllowed,
		"method not allowed",
	)
	return false
}

func withCardRequestTimeout(
	request *http.Request,
) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(
		request.Context(),
		cardRequestTimeout,
	)

	return request.WithContext(ctx), cancel
}

func parseCardRequestOptions(
	request *http.Request,
	configuredUsername string,
	dynamicUsername bool,
) (cardRequestOptions, *cardRequestFailure) {
	username := configuredUsername
	if dynamicUsername {
		username = strings.TrimSpace(
			request.PathValue("username"),
		)
		if !validGitHubUsername(username) {
			return cardRequestOptions{}, &cardRequestFailure{
				status:  http.StatusBadRequest,
				message: "invalid GitHub username",
			}
		}
	}

	themeName := request.URL.Query().Get("theme")
	if _, err := card.ResolveTheme(themeName); err != nil {
		return cardRequestOptions{}, &cardRequestFailure{
			status:  http.StatusBadRequest,
			message: "unknown card theme",
		}
	}

	scope := repositoryScope.Scope(
		request.URL.Query().Get("repositories"),
	)
	if scope == "" {
		scope = repositoryScope.ScopePublic
	}

	if scope != repositoryScope.ScopePublic &&
		scope != repositoryScope.ScopeAll {
		return cardRequestOptions{}, &cardRequestFailure{
			status:  http.StatusBadRequest,
			message: "invalid repositories parameter",
		}
	}

	if dynamicUsername &&
		scope != repositoryScope.ScopePublic {
		return cardRequestOptions{}, &cardRequestFailure{
			status: http.StatusBadRequest,
			message: "repositories=all is only available on " +
				"self-hosted instances using your own GitHub token",
		}
	}

	return cardRequestOptions{
		username:  username,
		themeName: themeName,
		scope:     scope,
	}, nil
}
