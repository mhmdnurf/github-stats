package refreshjob

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/config"
)

func TestNewValidatesRequiredDependencies(t *testing.T) {
	t.Parallel()

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	tests := []struct {
		name   string
		ctx    context.Context
		logger *slog.Logger
		want   string
	}{
		{
			name:   "context",
			ctx:    nil,
			logger: logger,
			want:   "context is required",
		},
		{
			name:   "logger",
			ctx:    context.Background(),
			logger: nil,
			want:   "logger is required",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(
				test.ctx,
				config.Config{},
				test.logger,
			)
			if err == nil {
				t.Fatal("New() error = nil, want error")
			}
			if err.Error() != test.want {
				t.Fatalf(
					"New() error = %q, want %q",
					err,
					test.want,
				)
			}
		})
	}
}

func TestJobRun(t *testing.T) {
	t.Parallel()

	runErr := errors.New("refresh failed")
	service := &fakeRunner{err: runErr}
	job := &Job{
		service:  service,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		username: "your-username",
	}

	err := job.Run(context.Background())
	if !errors.Is(err, runErr) {
		t.Fatalf(
			"Run() error = %v, want wrapped %v",
			err,
			runErr,
		)
	}
	if service.calls != 1 {
		t.Fatalf(
			"Run() calls = %d, want 1",
			service.calls,
		)
	}
}

func TestJobRunRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := &fakeRunner{}
	job := &Job{
		service: service,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	err := job.Run(nil)
	if err == nil || err.Error() != "context is required" {
		t.Fatalf(
			"Run() error = %v, want context validation error",
			err,
		)
	}
	if service.calls != 0 {
		t.Fatalf(
			"Run() calls = %d, want 0",
			service.calls,
		)
	}
}

func TestJobCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	closeCalls := 0
	job := &Job{
		close: func() error {
			closeCalls++
			return closeErr
		},
	}

	firstErr := job.Close()
	secondErr := job.Close()

	if !errors.Is(firstErr, closeErr) {
		t.Fatalf(
			"first Close() error = %v, want %v",
			firstErr,
			closeErr,
		)
	}
	if !errors.Is(secondErr, closeErr) {
		t.Fatalf(
			"second Close() error = %v, want %v",
			secondErr,
			closeErr,
		)
	}
	if closeCalls != 1 {
		t.Fatalf(
			"Close() calls = %d, want 1",
			closeCalls,
		)
	}
}

type fakeRunner struct {
	err   error
	calls int
}

func (runner *fakeRunner) Run(context.Context) error {
	runner.calls++
	return runner.err
}
