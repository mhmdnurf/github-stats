package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/mhmdnurf/github-stats/internal/config"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	httpServer *http.Server
	close      func() error
	closeOnce  sync.Once
	closeErr   error
	logger     *slog.Logger
}

func New(
	ctx context.Context,
	configuration config.Config,
	logger *slog.Logger,
) (*App, error) {
	if ctx == nil {
		return nil, errors.New("server context is required")
	}

	if logger == nil {
		return nil, errors.New("server logger is required")
	}

	serviceSet, closeServices, err := newServices(
		ctx,
		configuration,
	)
	if err != nil {
		return nil, err
	}

	router, err := newRouter(
		configuration,
		serviceSet,
		logger,
	)
	if err != nil {
		return nil, cleanupAfterError(err, closeServices)
	}

	return &App{
		httpServer: &http.Server{
			Addr:              configuration.HTTPAddress,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
		close:  closeServices,
		logger: logger,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("server context is required")
	}

	serverErrors := make(chan error, 1)

	go func() {
		app.logger.Info(
			"server started",
			"address",
			app.httpServer.Addr,
		)

		serverErrors <- app.httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve HTTP: %w", err)

	case <-ctx.Done():
		app.logger.Info("shutting down server")

		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := app.httpServer.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}

		err := <-serverErrors
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP during shutdown: %w", err)
		}

		return nil
	}
}

func (app *App) Close() error {
	app.closeOnce.Do(func() {
		app.closeErr = app.close()
	})

	return app.closeErr
}

func cleanupAfterError(
	primaryErr error,
	closeResources func() error,
) error {
	if closeErr := closeResources(); closeErr != nil {
		return errors.Join(
			primaryErr,
			fmt.Errorf("close server resources: %w", closeErr),
		)
	}

	return primaryErr
}
