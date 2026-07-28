package server

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/handler"
)

type services struct {
	configuredStats     handler.StatsService
	configuredLanguages handler.LanguagesService
	dynamic             *dynamicServices
}

func newServices(
	ctx context.Context,
	configuration config.Config,
) (services, func() error, error) {
	firestoreClient, err := firestore.NewClient(
		ctx,
		configuration.GoogleCloudProjectID,
	)
	if err != nil {
		return services{}, nil, fmt.Errorf(
			"create Firestore client: %w",
			err,
		)
	}

	closeServices := func() error {
		if err := firestoreClient.Close(); err != nil {
			return fmt.Errorf("close Firestore client: %w", err)
		}

		return nil
	}

	configured, err := newConfiguredServices(
		ctx,
		configuration,
		firestoreClient,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			err,
			closeServices,
		)
	}

	serviceSet := services{
		configuredStats:     configured.stats,
		configuredLanguages: configured.languages,
	}

	if strings.TrimSpace(configuration.GitHubToken) == "" {
		return serviceSet, closeServices, nil
	}

	dynamic, err := newDynamicServices(
		configuration,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			err,
			closeServices,
		)
	}

	serviceSet.dynamic = &dynamic

	return serviceSet, closeServices, nil
}
