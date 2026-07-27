package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mhmdnurf/github-stats/internal/config"
	githubclient "github.com/mhmdnurf/github-stats/internal/github"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/refresh"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	githubRequestTimeout = 15 * time.Second
	refreshJobTimeout    = 2 * time.Minute
	snapshotFreshness    = 30 * time.Minute
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	jobContext, cancel := context.WithTimeout(
		signalContext,
		refreshJobTimeout,
	)
	defer cancel()

	if err := run(jobContext, logger); err != nil {
		logger.Error(
			"snapshot refresh failed",
			"error",
			err,
		)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	logger *slog.Logger,
) error {
	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	projectID := strings.TrimSpace(
		configuration.GoogleCloudProjectID,
	)
	if projectID == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT is required")
	}

	githubClient, err := githubclient.NewClient(
		configuration.GitHubToken,
		&http.Client{
			Timeout: githubRequestTimeout,
		},
	)
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	firestoreClient, err := firestore.NewClient(
		ctx,
		projectID,
	)
	if err != nil {
		return fmt.Errorf("create Firestore client: %w", err)
	}
	defer func() {
		if err := firestoreClient.Close(); err != nil {
			logger.Error(
				"close Firestore client",
				"error",
				err,
			)
		}
	}()

	statsStore, err := snapshot.NewFirestore[stats.UserStats](
		firestoreClient,
		configuration.FirestoreCollection,
	)
	if err != nil {
		return fmt.Errorf("create stats snapshot store: %w", err)
	}

	languagesStore, err :=
		snapshot.NewFirestore[languages.UserLanguages](
			firestoreClient,
			configuration.FirestoreCollection,
		)
	if err != nil {
		return fmt.Errorf(
			"create languages snapshot store: %w",
			err,
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
		return fmt.Errorf("create refresh service: %w", err)
	}

	logger.Info(
		"snapshot refresh started",
		"username",
		configuration.GitHubUsername,
	)

	if err := refreshService.Run(ctx); err != nil {
		return fmt.Errorf("refresh snapshots: %w", err)
	}

	logger.Info(
		"snapshot refresh completed",
		"username",
		configuration.GitHubUsername,
	)

	return nil
}
