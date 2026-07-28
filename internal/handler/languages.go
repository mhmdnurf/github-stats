package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

type LanguagesService interface {
	Get(
		ctx context.Context,
		username string,
		scope repositoryScope.Scope,
	) (languages.UserLanguages, error)
}

type LanguageCardRenderer interface {
	ThemeValidator

	Render(
		userLanguages languages.UserLanguages,
		themeName string,
	) ([]byte, error)
}

type Languages struct {
	username        string
	dynamicUsername bool
	service         LanguagesService
	renderer        LanguageCardRenderer
	logger          *slog.Logger
}

func NewLanguages(
	username string,
	service LanguagesService,
	renderer LanguageCardRenderer,
	logger *slog.Logger,
) (*Languages, error) {
	normalizedUsername := strings.TrimSpace(username)
	if !validGitHubUsername(normalizedUsername) {
		return nil, errors.New("valid GitHub username is required")
	}

	if service == nil {
		return nil, errors.New("languages service is required")
	}

	if renderer == nil {
		return nil, errors.New("language card renderer is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &Languages{
		username: normalizedUsername,
		service:  service,
		renderer: renderer,
		logger:   logger,
	}, nil
}

func NewDynamicLanguages(
	service LanguagesService,
	renderer LanguageCardRenderer,
	logger *slog.Logger,
) (*Languages, error) {
	if service == nil {
		return nil, errors.New("languages service is required")
	}

	if renderer == nil {
		return nil, errors.New("language card renderer is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &Languages{
		dynamicUsername: true,
		service:         service,
		renderer:        renderer,
		logger:          logger,
	}, nil
}
