package server

import (
	"context"
	"fmt"

	"github.com/mhmdnurf/github-stats/internal/handler"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

func preloadSnapshots(
	ctx context.Context,
	username string,
	statsService handler.StatsService,
	languagesService handler.LanguagesService,
) error {
	for _, scope := range []repositoryScope.Scope{
		repositoryScope.ScopePublic,
		repositoryScope.ScopeAll,
	} {
		if _, err := statsService.Get(
			ctx,
			username,
			scope,
		); err != nil {
			return fmt.Errorf(
				"preload stats for %s: %w",
				scope,
				err,
			)
		}

		if _, err := languagesService.Get(
			ctx,
			username,
			scope,
		); err != nil {
			return fmt.Errorf(
				"preload languages for %s: %w",
				scope,
				err,
			)
		}
	}

	return nil
}
