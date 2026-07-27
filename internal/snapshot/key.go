package snapshot

import (
	"errors"
	"fmt"
	"strings"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

type Kind string

const (
	KindStats     Kind = "stats"
	KindLanguages Kind = "languages"
)

var (
	ErrUsernameRequired = errors.New("snapshot username is required")
	ErrKindInvalid      = errors.New("snapshot kind is invalid")
	ErrScopeInvalid     = errors.New("snapshot scope is invalid")
)

func Key(
	username string,
	kind Kind,
	scope repositoryScope.Scope,
) (string, error) {
	normalizedUsername := strings.ToLower(
		strings.TrimSpace(username),
	)
	if normalizedUsername == "" {
		return "", ErrUsernameRequired
	}

	if !validKind(kind) {
		return "", ErrKindInvalid
	}

	if scope != repositoryScope.ScopePublic &&
		scope != repositoryScope.ScopeAll {
		return "", ErrScopeInvalid
	}

	return fmt.Sprintf(
		"v%d:%s:%s:%s",
		CurrentSchemaVersion,
		normalizedUsername,
		kind,
		scope,
	), nil
}

func validKind(kind Kind) bool {
	return kind == KindStats || kind == KindLanguages
}
