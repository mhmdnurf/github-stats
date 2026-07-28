package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firestoreadapter "github.com/mhmdnurf/github-stats/internal/adapter/firestore"
	"github.com/mhmdnurf/github-stats/internal/cache"
	"github.com/mhmdnurf/github-stats/internal/config"
	githubclient "github.com/mhmdnurf/github-stats/internal/github"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	snapshotPreloadTimeout = 10 * time.Second

	dynamicGitHubRequestTimeout = 20 * time.Second
	dynamicCacheTTL             = 5 * time.Minute
)

type services struct {
	configuredStats     handler.StatsService
	configuredLanguages handler.LanguagesService
	dynamicStats        handler.StatsService
	dynamicLanguages    handler.LanguagesService
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

	statsStore, err := firestoreadapter.NewStore[stats.UserStats](
		firestoreClient,
		configuration.FirestoreCollection,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create stats snapshot store: %w", err),
			closeServices,
		)
	}

	statsReader, err := snapshot.NewReader(
		snapshot.NewMemory[stats.UserStats](),
		statsStore,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create stats snapshot reader: %w", err),
			closeServices,
		)
	}

	statsService, err := snapshot.NewProvider(
		statsReader,
		snapshot.KindStats,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create stats snapshot provider: %w", err),
			closeServices,
		)
	}

	languagesStore, err :=
		firestoreadapter.NewStore[languages.UserLanguages](
			firestoreClient,
			configuration.FirestoreCollection,
		)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf(
				"create languages snapshot store: %w",
				err,
			),
			closeServices,
		)
	}

	languagesReader, err := snapshot.NewReader(
		snapshot.NewMemory[languages.UserLanguages](),
		languagesStore,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf(
				"create languages snapshot reader: %w",
				err,
			),
			closeServices,
		)
	}

	languagesService, err := snapshot.NewProvider(
		languagesReader,
		snapshot.KindLanguages,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf(
				"create languages snapshot provider: %w",
				err,
			),
			closeServices,
		)
	}

	preloadContext, cancelPreload := context.WithTimeout(
		ctx,
		snapshotPreloadTimeout,
	)
	err = preloadSnapshots(
		preloadContext,
		configuration.GitHubUsername,
		statsService,
		languagesService,
	)
	cancelPreload()
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("preload snapshots: %w", err),
			closeServices,
		)
	}

	githubClient, err := githubclient.NewClient(
		configuration.GitHubToken,
		&http.Client{Timeout: dynamicGitHubRequestTimeout},
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create GitHub client: %w", err),
			closeServices,
		)
	}

	dynamicStatsService, err := stats.NewService(
		githubClient,
		cache.NewMemory(),
		dynamicCacheTTL,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create dynamic stats service: %w", err),
			closeServices,
		)
	}

	dynamicLanguagesService, err := languages.NewService(
		githubClient,
		cache.NewLanguageMemory(),
		dynamicCacheTTL,
	)
	if err != nil {
		return services{}, nil, cleanupAfterError(
			fmt.Errorf("create dynamic languages service: %w", err),
			closeServices,
		)
	}

	return services{
		configuredStats:     statsService,
		configuredLanguages: languagesService,
		dynamicStats:        dynamicStatsService,
		dynamicLanguages:    dynamicLanguagesService,
	}, closeServices, nil
}
