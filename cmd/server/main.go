package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/middleware"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	snapshotPreloadTimeout = 10 * time.Second
	shutdownTimeout        = 10 * time.Second

	dynamicRateLimitPerSecond = 1
	dynamicRateLimitBurst     = 10
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

	if err := run(logger); err != nil {
		logger.Error(
			"server stopped",
			"error",
			err,
		)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	configuration, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	firestoreClient, err := firestore.NewClient(
		signalContext,
		configuration.GoogleCloudProjectID,
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

	statsReader, err := snapshot.NewReader(
		snapshot.NewMemory[stats.UserStats](),
		statsStore,
	)
	if err != nil {
		return fmt.Errorf("create stats snapshot reader: %w", err)
	}

	statsService, err := snapshot.NewProvider(
		statsReader,
		snapshot.KindStats,
	)
	if err != nil {
		return fmt.Errorf("create stats snapshot provider: %w", err)
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

	languagesReader, err := snapshot.NewReader(
		snapshot.NewMemory[languages.UserLanguages](),
		languagesStore,
	)
	if err != nil {
		return fmt.Errorf(
			"create languages snapshot reader: %w",
			err,
		)
	}

	languagesService, err := snapshot.NewProvider(
		languagesReader,
		snapshot.KindLanguages,
	)
	if err != nil {
		return fmt.Errorf(
			"create languages snapshot provider: %w",
			err,
		)
	}

	preloadContext, cancelPreload := context.WithTimeout(
		signalContext,
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
		return fmt.Errorf("preload snapshots: %w", err)
	}

	cardRenderer, err := card.NewRenderer()
	if err != nil {
		return fmt.Errorf("create card renderer: %w", err)
	}

	languageRenderer, err := card.NewLanguageRenderer()
	if err != nil {
		return fmt.Errorf("create language renderer: %w", err)
	}

	statsHandler, err := handler.NewStats(
		configuration.GitHubUsername,
		statsService,
		cardRenderer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create stats handler: %w", err)
	}

	languagesHandler, err := handler.NewLanguages(
		configuration.GitHubUsername,
		languagesService,
		languageRenderer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create languages handler: %w", err)
	}

	dynamicStatsHandler, err := handler.NewDynamicStats(
		statsService,
		cardRenderer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create dynamic stats handler: %w", err)
	}

	dynamicLanguagesHandler, err := handler.NewDynamicLanguages(
		languagesService,
		languageRenderer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create dynamic languages handler: %w", err)
	}

	rateLimiter := middleware.NewRateLimiter(
		dynamicRateLimitPerSecond,
		dynamicRateLimitBurst,
	)

	mux := http.NewServeMux()
	mux.Handle("/stats", statsHandler)
	mux.Handle("/languages", languagesHandler)
	mux.Handle(
		"/{username}/stats",
		rateLimiter.Middleware(dynamicStatsHandler),
	)
	mux.Handle(
		"/{username}/languages",
		rateLimiter.Middleware(dynamicLanguagesHandler),
	)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              configuration.HTTPAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"server started",
			"address",
			server.Addr,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve HTTP: %w", err)

	case <-signalContext.Done():
		logger.Info("shutting down server")

		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}

		err := <-serverErrors
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP during shutdown: %w", err)
		}

		return nil
	}
}

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

func healthHandler(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writer.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)

	writer.WriteHeader(http.StatusOK)

	_, _ = writer.Write([]byte("ok\n"))
}
