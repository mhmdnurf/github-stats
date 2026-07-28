package firestore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cloudfirestore "cloud.google.com/go/firestore"
	"github.com/mhmdnurf/github-stats/internal/snapshot"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrCollectionRequired = errors.New(
	"snapshot collection is required",
)

type Store[T any] struct {
	client     *cloudfirestore.Client
	collection string
}

func NewStore[T any](
	client *cloudfirestore.Client,
	collection string,
) (*Store[T], error) {
	if client == nil {
		return nil, errors.New("firestore client is required")
	}

	normalizedCollection := strings.TrimSpace(collection)
	if normalizedCollection == "" {
		return nil, ErrCollectionRequired
	}

	return &Store[T]{
		client:     client,
		collection: normalizedCollection,
	}, nil
}

func (store *Store[T]) Get(
	ctx context.Context,
	key string,
) (snapshot.Snapshot[T], error) {
	if err := ctx.Err(); err != nil {
		return snapshot.Snapshot[T]{}, err
	}

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return snapshot.Snapshot[T]{}, snapshot.ErrKeyRequired
	}

	document, err := store.client.
		Collection(store.collection).
		Doc(normalizedKey).
		Get(ctx)
	if status.Code(err) == codes.NotFound {
		return snapshot.Snapshot[T]{}, snapshot.ErrNotFound
	}
	if err != nil {
		return snapshot.Snapshot[T]{}, fmt.Errorf(
			"get snapshot %q: %w",
			normalizedKey,
			err,
		)
	}

	var value snapshot.Snapshot[T]
	if err := document.DataTo(&value); err != nil {
		return snapshot.Snapshot[T]{}, fmt.Errorf(
			"decode snapshot %q: %w",
			normalizedKey,
			err,
		)
	}

	return value, nil
}

func (store *Store[T]) Set(
	ctx context.Context,
	key string,
	value snapshot.Snapshot[T],
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return snapshot.ErrKeyRequired
	}

	_, err := store.client.
		Collection(store.collection).
		Doc(normalizedKey).
		Set(ctx, value)
	if err != nil {
		return fmt.Errorf(
			"set snapshot %q: %w",
			normalizedKey,
			err,
		)
	}

	return nil
}

var _ snapshot.Store[any] = (*Store[any])(nil)
