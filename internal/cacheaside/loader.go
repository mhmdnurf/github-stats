package cacheaside

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"
)

var (
	ErrCacheRequired    = errors.New("cache is required")
	ErrFetchRequired    = errors.New("fetch function is required")
	ErrResourceRequired = errors.New("cache resource is required")
	ErrTTLInvalid       = errors.New("cache TTL must be greater than zero")
)

var errFlightResultInvalid = errors.New(
	"singleflight returned an invalid result",
)

type Cache[T any] interface {
	Get(
		ctx context.Context,
		key string,
	) (value T, found bool, err error)

	Set(
		ctx context.Context,
		key string,
		value T,
		ttl time.Duration,
	) error
}

type Loader[T any] struct {
	cache    Cache[T]
	flights  singleflight.Group
	ttl      time.Duration
	resource string
}

type flightResult[T any] struct {
	value T
}

func New[T any](
	cache Cache[T],
	ttl time.Duration,
	resource string,
) (*Loader[T], error) {
	if cache == nil {
		return nil, ErrCacheRequired
	}

	if ttl <= 0 {
		return nil, ErrTTLInvalid
	}

	normalizedResource := strings.TrimSpace(resource)
	if normalizedResource == "" {
		return nil, ErrResourceRequired
	}

	return &Loader[T]{
		cache:    cache,
		ttl:      ttl,
		resource: normalizedResource,
	}, nil
}

func (loader *Loader[T]) Load(
	ctx context.Context,
	key string,
	fetch func(context.Context) (T, error),
) (T, error) {
	var zero T

	if fetch == nil {
		return zero, ErrFetchRequired
	}

	cached, found, err := loader.getCached(ctx, key)
	if err != nil || found {
		return cached, err
	}

	result, err, _ := loader.flights.Do(
		key,
		func() (any, error) {
			cached, found, err := loader.getCached(ctx, key)
			if err != nil {
				return nil, err
			}

			if found {
				return flightResult[T]{value: cached}, nil
			}

			fetched, err := loader.fetchAndCache(
				ctx,
				key,
				fetch,
			)
			if err != nil {
				return nil, err
			}

			return flightResult[T]{value: fetched}, nil
		},
	)
	if err != nil {
		return zero, err
	}

	typedResult, ok := result.(flightResult[T])
	if !ok {
		return zero, errFlightResultInvalid
	}

	return typedResult.value, nil
}

func (loader *Loader[T]) getCached(
	ctx context.Context,
	key string,
) (T, bool, error) {
	var zero T

	cached, found, err := loader.cache.Get(ctx, key)
	if err != nil {
		return zero, false, fmt.Errorf(
			"get cached %s: %w",
			loader.resource,
			err,
		)
	}

	if found {
		return cached, true, nil
	}

	return zero, false, nil
}

func (loader *Loader[T]) fetchAndCache(
	ctx context.Context,
	key string,
	fetch func(context.Context) (T, error),
) (T, error) {
	var zero T

	fetched, err := fetch(ctx)
	if err != nil {
		return zero, fmt.Errorf(
			"fetch %s: %w",
			loader.resource,
			err,
		)
	}

	if err := loader.cache.Set(
		ctx,
		key,
		fetched,
		loader.ttl,
	); err != nil {
		return zero, fmt.Errorf(
			"cache fetched %s: %w",
			loader.resource,
			err,
		)
	}

	return fetched, nil
}
