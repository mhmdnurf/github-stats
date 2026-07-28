package refreshjob

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firestoreadapter "github.com/mhmdnurf/github-stats/internal/adapter/firestore"
	"github.com/mhmdnurf/github-stats/internal/config"
	githubclient "github.com/mhmdnurf/github-stats/internal/github"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/refresh"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	githubRequestTimeout = 15 * time.Second
	snapshotFreshness    = 30 * time.Minute
)

// New wires the refresh use case to its external adapters.
func New(
	ctx context.Context,
	configuration config.Config,
	logger *slog.Logger,
) (*Job, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	projectID := strings.TrimSpace(
		configuration.GoogleCloudProjectID,
	)
	if projectID == "" {
		return nil, errors.New(
			"GOOGLE_CLOUD_PROJECT is required",
		)
	}

	githubClient, err := githubclient.NewClient(
		configuration.GitHubToken,
		&http.Client{
			Timeout: githubRequestTimeout,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create GitHub client: %w",
			err,
		)
	}

	firestoreClient, err := firestore.NewClient(
		ctx,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create Firestore client: %w",
			err,
		)
	}

	closeFirestore := func() error {
		if err := firestoreClient.Close(); err != nil {
			return fmt.Errorf(
				"close Firestore client: %w",
				err,
			)
		}
		return nil
	}

	statsStore, err := firestoreadapter.NewStore[stats.UserStats](
		firestoreClient,
		configuration.FirestoreCollection,
	)
	if err != nil {
		return nil, cleanupAfterError(
			fmt.Errorf(
				"create stats snapshot store: %w",
				err,
			),
			closeFirestore,
		)
	}

	languagesStore, err :=
		firestoreadapter.NewStore[languages.UserLanguages](
			firestoreClient,
			configuration.FirestoreCollection,
		)
	if err != nil {
		return nil, cleanupAfterError(
			fmt.Errorf(
				"create languages snapshot store: %w",
				err,
			),
			closeFirestore,
		)
	}

	refreshService, err := refresh.NewService(
		configuration.GitHubUsername,
		githubClient,
		githubClient,
		statsStore,
		languagesStore,
		snapshotFreshness,
	)
	if err != nil {
		return nil, cleanupAfterError(
			fmt.Errorf(
				"create refresh service: %w",
				err,
			),
			closeFirestore,
		)
	}

	return &Job{
		service:  refreshService,
		close:    closeFirestore,
		logger:   logger,
		username: configuration.GitHubUsername,
	}, nil
}

func cleanupAfterError(
	setupErr error,
	closeResource func() error,
) error {
	if closeErr := closeResource(); closeErr != nil {
		return errors.Join(setupErr, closeErr)
	}

	return setupErr
}
