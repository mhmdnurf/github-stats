package server

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firestoreadapter "github.com/mhmdnurf/github-stats/internal/adapter/firestore"
	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const snapshotPreloadTimeout = 10 * time.Second

type configuredServices struct {
	stats     handler.StatsService
	languages handler.LanguagesService
}

func newConfiguredServices(
	ctx context.Context,
	configuration config.Config,
	firestoreClient *firestore.Client,
) (configuredServices, error) {
	statsService, err := newStatsSnapshotService(
		configuration,
		firestoreClient,
	)
	if err != nil {
		return configuredServices{}, err
	}

	languagesService, err := newLanguagesSnapshotService(
		configuration,
		firestoreClient,
	)
	if err != nil {
		return configuredServices{}, err
	}

	preloadContext, cancelPreload := context.WithTimeout(
		ctx,
		snapshotPreloadTimeout,
	)
	defer cancelPreload()

	if err := preloadSnapshots(
		preloadContext,
		configuration.GitHubUsername,
		statsService,
		languagesService,
	); err != nil {
		return configuredServices{}, fmt.Errorf(
			"preload snapshots: %w",
			err,
		)
	}

	return configuredServices{
		stats:     statsService,
		languages: languagesService,
	}, nil
}

func newStatsSnapshotService(
	configuration config.Config,
	firestoreClient *firestore.Client,
) (handler.StatsService, error) {
	store, err := firestoreadapter.NewStore[stats.UserStats](
		firestoreClient,
		configuration.FirestoreCollection,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create stats snapshot store: %w",
			err,
		)
	}

	reader, err := snapshot.NewReader(
		snapshot.NewMemory[stats.UserStats](),
		store,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create stats snapshot reader: %w",
			err,
		)
	}

	service, err := snapshot.NewProvider(
		reader,
		snapshot.KindStats,
		stats.ErrUnavailable,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create stats snapshot provider: %w",
			err,
		)
	}

	return service, nil
}

func newLanguagesSnapshotService(
	configuration config.Config,
	firestoreClient *firestore.Client,
) (handler.LanguagesService, error) {
	store, err :=
		firestoreadapter.NewStore[languages.UserLanguages](
			firestoreClient,
			configuration.FirestoreCollection,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"create languages snapshot store: %w",
			err,
		)
	}

	reader, err := snapshot.NewReader(
		snapshot.NewMemory[languages.UserLanguages](),
		store,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create languages snapshot reader: %w",
			err,
		)
	}

	service, err := snapshot.NewProvider(
		reader,
		snapshot.KindLanguages,
		languages.ErrUnavailable,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create languages snapshot provider: %w",
			err,
		)
	}

	return service, nil
}
