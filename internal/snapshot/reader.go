package snapshot

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Reader[T any] struct {
	memory *Memory[T]
	store  Store[T]
	now    func() time.Time
}

func NewReader[T any](
	memory *Memory[T],
	store Store[T],
) (*Reader[T], error) {
	if memory == nil {
		return nil, errors.New(
			"snapshot memory cache is required",
		)
	}

	if store == nil {
		return nil, ErrStoreRequired
	}

	return &Reader[T]{
		memory: memory,
		store:  store,
		now:    time.Now,
	}, nil
}

func (reader *Reader[T]) Get(
	ctx context.Context,
	key string,
) (Snapshot[T], error) {
	cached, err := reader.memory.Get(ctx, key)
	if err == nil {
		if !cached.IsStale(reader.now()) {
			return cached, nil
		}

		persisted, persistentErr := reader.store.Get(ctx, key)
		if persistentErr != nil {
			return cached, nil
		}

		if setErr := reader.memory.Set(
			ctx,
			key,
			persisted,
		); setErr != nil {
			return Snapshot[T]{}, fmt.Errorf(
				"cache persistent snapshot: %w",
				setErr,
			)
		}

		return persisted, nil
	}

	if !errors.Is(err, ErrNotFound) {
		return Snapshot[T]{}, fmt.Errorf(
			"get memory snapshot: %w",
			err,
		)
	}

	persisted, err := reader.store.Get(ctx, key)
	if err != nil {
		return Snapshot[T]{}, fmt.Errorf(
			"get persistent snapshot: %w",
			err,
		)
	}

	if err := reader.memory.Set(ctx, key, persisted); err != nil {
		return Snapshot[T]{}, fmt.Errorf(
			"cache persistent snapshot: %w",
			err,
		)
	}

	return persisted, nil
}
