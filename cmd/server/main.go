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

	"github.com/mhmdnurf/github-stats/internal/cache"
	"github.com/mhmdnurf/github-stats/internal/card"
	"github.com/mhmdnurf/github-stats/internal/config"
	githubclient "github.com/mhmdnurf/github-stats/internal/github"
	"github.com/mhmdnurf/github-stats/internal/handler"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const (
	githubRequestTimeout = 15 * time.Second
	statsCacheTTL        = 10 * time.Minute
	shutdownTimeout      = 10 * time.Second
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

	githubHTTPClient := &http.Client{
		Timeout: githubRequestTimeout,
	}

	githubClient, err := githubclient.NewClient(
		configuration.GitHubToken,
		githubHTTPClient,
	)
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	memoryCache := cache.NewMemory()

	statsService, err := stats.NewService(
		githubClient,
		memoryCache,
		statsCacheTTL,
	)
	if err != nil {
		return fmt.Errorf("create stats service: %w", err)
	}

	cardRenderer, err := card.NewRenderer()
	if err != nil {
		return fmt.Errorf("create card renderer: %w", err)
	}

	statsHandler, err := handler.NewStats(
		statsService,
		cardRenderer,
		logger,
	)
	if err != nil {
		return fmt.Errorf("create stats handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/stats", statsHandler)
	mux.HandleFunc("/healthz", healthHandler)

	server := &http.Server{
		Addr:              configuration.HTTPAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

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
