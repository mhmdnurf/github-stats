package refreshjob

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type runner interface {
	Run(context.Context) error
}

// Job owns the refresh use case and its infrastructure resources.
type Job struct {
	service  runner
	close    func() error
	logger   *slog.Logger
	username string

	closeOnce sync.Once
	closeErr  error
}

// Run refreshes and persists every snapshot managed by the job.
func (job *Job) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	job.logger.Info(
		"snapshot refresh started",
		"username",
		job.username,
	)

	if err := job.service.Run(ctx); err != nil {
		return fmt.Errorf("refresh snapshots: %w", err)
	}

	job.logger.Info(
		"snapshot refresh completed",
		"username",
		job.username,
	)

	return nil
}

// Close releases infrastructure resources owned by the job.
func (job *Job) Close() error {
	job.closeOnce.Do(func() {
		if job.close != nil {
			job.closeErr = job.close()
		}
	})

	return job.closeErr
}
