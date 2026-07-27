package snapshot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Firestore[T any] struct {
	client     *firestore.Client
	collection string
}

func NewFirestore[T any](
	client *firestore.Client,
	collection string,
) (*Firestore[T], error) {
	if client == nil {
		return nil, errors.New("firestore client is required")
	}

	normalizedCollection := strings.TrimSpace(collection)
	if normalizedCollection == "" {
		return nil, ErrCollectionRequired
	}

	return &Firestore[T]{
		client:     client,
		collection: normalizedCollection,
	}, nil
}

func (store *Firestore[T]) Get(
	ctx context.Context,
	key string,
) (Snapshot[T], error) {
	if err := ctx.Err(); err != nil {
		return Snapshot[T]{}, err
	}

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return Snapshot[T]{}, ErrKeyRequired
	}

	document, err := store.client.
		Collection(store.collection).
		Doc(normalizedKey).
		Get(ctx)
	if status.Code(err) == codes.NotFound {
		return Snapshot[T]{}, ErrNotFound
	}
	if err != nil {
		return Snapshot[T]{}, fmt.Errorf(
			"get snapshot %q: %w",
			normalizedKey,
			err,
		)
	}

	var value Snapshot[T]
	if err := document.DataTo(&value); err != nil {
		return Snapshot[T]{}, fmt.Errorf(
			"decode snapshot %q: %w",
			normalizedKey,
			err,
		)
	}

	return value, nil
}

func (store *Firestore[T]) Set(
	ctx context.Context,
	key string,
	value Snapshot[T],
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return ErrKeyRequired
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

var _ Store[any] = (*Firestore[any])(nil)
