package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mhmdnurf/github-stats/internal/config"
	"github.com/mhmdnurf/github-stats/internal/refreshjob"
)

const refreshJobTimeout = 2 * time.Minute

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

	job, err := refreshjob.New(
		ctx,
		configuration,
		logger,
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := job.Close(); err != nil {
			logger.Error(
				"close refresh job resources",
				"error",
				err,
			)
		}
	}()

	return job.Run(ctx)
}
