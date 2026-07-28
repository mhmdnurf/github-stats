package snapshot

import (
	"context"
	"errors"
	"fmt"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

type SnapshotReader[T any] interface {
	Get(
		ctx context.Context,
		key string,
	) (Snapshot[T], error)
}

type Provider[T any] struct {
	reader         SnapshotReader[T]
	kind           Kind
	unavailableErr error
}

func NewProvider[T any](
	reader SnapshotReader[T],
	kind Kind,
	unavailableErr error,
) (*Provider[T], error) {
	if reader == nil {
		return nil, ErrReaderRequired
	}

	if !validKind(kind) {
		return nil, ErrKindInvalid
	}

	if unavailableErr == nil {
		return nil, errors.New("snapshot unavailable error is required")
	}

	return &Provider[T]{
		reader:         reader,
		kind:           kind,
		unavailableErr: unavailableErr,
	}, nil
}

func (provider *Provider[T]) Get(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (T, error) {
	var zero T

	key, err := Key(username, provider.kind, scope)
	if err != nil {
		return zero, fmt.Errorf(
			"create %s snapshot key: %w",
			provider.kind,
			err,
		)
	}

	value, err := provider.reader.Get(ctx, key)
	if err != nil {
		return zero, fmt.Errorf(
			"%w: get %s snapshot: %w",
			provider.unavailableErr,
			provider.kind,
			err,
		)
	}

	return value.Value, nil
}
