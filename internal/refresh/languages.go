package refresh

import (
	"context"
	"fmt"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
)

func (service *Service) refreshLanguages(
	ctx context.Context,
	scope repositoryScope.Scope,
) error {
	value, err := service.languagesFetcher.FetchLanguages(
		ctx,
		service.username,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch languages snapshot for %s: %w",
			scope,
			err,
		)
	}

	key, err := snapshot.Key(
		service.username,
		snapshot.KindLanguages,
		scope,
	)
	if err != nil {
		return fmt.Errorf(
			"create languages snapshot key: %w",
			err,
		)
	}

	if err := service.languagesStore.Set(
		ctx,
		key,
		newSnapshot(
			value,
			service.now(),
			service.freshness,
		),
	); err != nil {
		return fmt.Errorf(
			"store languages snapshot for %s: %w",
			scope,
			err,
		)
	}

	return nil
}
