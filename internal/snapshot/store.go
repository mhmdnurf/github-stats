package snapshot

import (
	"context"
	"errors"
)

var (
	ErrKeyRequired        = errors.New("snapshot key is required")
	ErrNotFound           = errors.New("snapshot not found")
	ErrStoreRequired      = errors.New("snapshot store is required")
	ErrCollectionRequired = errors.New(
		"snapshot collection is required",
	)
)

type Store[T any] interface {
	Get(
		ctx context.Context,
		key string,
	) (Snapshot[T], error)

	Set(
		ctx context.Context,
		key string,
		value Snapshot[T],
	) error
}
