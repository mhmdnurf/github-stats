package refresh

import (
	"context"
	"fmt"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
)

func (service *Service) refreshStats(
	ctx context.Context,
	scope repositoryScope.Scope,
) error {
	value, err := service.statsFetcher.Fetch(
		ctx,
		service.username,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch stats snapshot for %s: %w",
			scope,
			err,
		)
	}

	key, err := snapshot.Key(
		service.username,
		snapshot.KindStats,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"create stats snapshot key: %w",
			err,
		)
	}

	if err := service.statsStore.Set(
		ctx,
		key,
		newSnapshot(
			value,
			service.now(),
			service.freshness,
		),
	); err != nil {
		return fmt.Errorf(
			"store stats snapshot for %s: %w",
			scope,
			err,
		)
	}

	return nil
}
