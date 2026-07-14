package stats

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fetcherStub struct {
	fetch func(context.Context, string) (UserStats, error)
}

func (stub fetcherStub) Fetch(
	ctx context.Context,
	username string,
) (UserStats, error) {
	return stub.fetch(ctx, username)
}

type cacheStub struct {
	get func(context.Context, string) (UserStats, bool, error)
	set func(context.Context, string, UserStats, time.Duration) error
}

func (stub cacheStub) Get(
	ctx context.Context,
	key string,
) (UserStats, bool, error) {
	return stub.get(ctx, key)
}

func (stub cacheStub) Set(
	ctx context.Context,
	key string,
	value UserStats,
	ttl time.Duration,
) error {
	return stub.set(ctx, key, value, ttl)
}

func TestServiceGetReturnsCachedStats(t *testing.T) {
	want := UserStats{
		Name:     "Muhammad Nurfatkhur Rahman",
		Username: "mhmdnurf",
		Stars:    128,
	}

	fetchCalled := false
	setCalled := false

	service, err := NewService(
		fetcherStub{
			fetch: func(
				context.Context,
				string,
			) (UserStats, error) {
				fetchCalled = true
				return UserStats{}, nil
			},
		},
		cacheStub{
			get: func(
				_ context.Context,
				key string,
			) (UserStats, bool, error) {
				if key != "stats:v1:mhmdnurf" {
					t.Fatalf("unexpected cache key: %q", key)
				}

				return want, true, nil
			},
			set: func(
				context.Context,
				string,
				UserStats,
				time.Duration,
			) error {
				setCalled = true
				return nil
			},
		},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	got, err := service.Get(context.Background(), "  MHMDNURF  ")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected stats: got %+v, want %+v", got, want)
	}

	if fetchCalled {
		t.Fatal("expected fetcher not to be called on cache hit")
	}

	if setCalled {
		t.Fatal("expected cache Set not to be called on cache hit")
	}
}

func TestServiceGetFetchesAndCachesOnMiss(t *testing.T) {
	const expectedTTL = 15 * time.Minute

	want := UserStats{
		Name:         "Muhammad Nurfatkhur Rahman",
		Username:     "mhmdnurf",
		Repositories: 20,
		Stars:        128,
	}

	fetchCalls := 0
	setCalls := 0

	service, err := NewService(
		fetcherStub{
			fetch: func(
				_ context.Context,
				username string,
			) (UserStats, error) {
				fetchCalls++

				if username != "mhmdnurf" {
					t.Fatalf("unexpected username: %q", username)
				}

				return want, nil
			},
		},
		cacheStub{
			get: func(
				_ context.Context,
				key string,
			) (UserStats, bool, error) {
				if key != "stats:v1:mhmdnurf" {
					t.Fatalf("unexpected cache key: %q", key)
				}

				return UserStats{}, false, nil
			},
			set: func(
				_ context.Context,
				key string,
				value UserStats,
				ttl time.Duration,
			) error {
				setCalls++

				if key != "stats:v1:mhmdnurf" {
					t.Fatalf("unexpected cache key: %q", key)
				}

				if !reflect.DeepEqual(value, want) {
					t.Fatalf(
						"unexpected cached value: got %+v, want %+v",
						value,
						want,
					)
				}

				if ttl != expectedTTL {
					t.Fatalf(
						"unexpected TTL: got %s, want %s",
						ttl,
						expectedTTL,
					)
				}

				return nil
			},
		},
		expectedTTL,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	got, err := service.Get(context.Background(), "  MHMDNURF  ")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected stats: got %+v, want %+v", got, want)
	}

	if fetchCalls != 1 {
		t.Fatalf("expected one fetch call, got %d", fetchCalls)
	}

	if setCalls != 1 {
		t.Fatalf("expected one cache Set call, got %d", setCalls)
	}
}

func TestServiceGetRejectsBlankUsername(t *testing.T) {
	service, err := NewService(
		fetcherStub{
			fetch: func(
				context.Context,
				string,
			) (UserStats, error) {
				t.Fatal("fetcher should not be called")
				return UserStats{}, nil
			},
		},
		cacheStub{
			get: func(
				context.Context,
				string,
			) (UserStats, bool, error) {
				t.Fatal("cache Get should not be called")
				return UserStats{}, false, nil
			},
			set: func(
				context.Context,
				string,
				UserStats,
				time.Duration,
			) error {
				t.Fatal("cache Set should not be called")
				return nil
			},
		},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.Get(context.Background(), "   ")
	if !errors.Is(err, ErrUsernameRequired) {
		t.Fatalf("expected ErrUsernameRequired, got %v", err)
	}
}

func TestServiceGetPropagatesErrors(t *testing.T) {
	cacheGetError := errors.New("cache get failed")
	fetchError := errors.New("fetch failed")
	cacheSetError := errors.New("cache set failed")

	tests := []struct {
		name       string
		getError   error
		fetchError error
		setError   error
		wantError  error
	}{
		{
			name:      "cache get error",
			getError:  cacheGetError,
			wantError: cacheGetError,
		},
		{
			name:       "fetch error",
			fetchError: fetchError,
			wantError:  fetchError,
		},
		{
			name:      "cache set error",
			setError:  cacheSetError,
			wantError: cacheSetError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewService(
				fetcherStub{
					fetch: func(
						context.Context,
						string,
					) (UserStats, error) {
						return UserStats{
							Username: "mhmdnurf",
						}, test.fetchError
					},
				},
				cacheStub{
					get: func(
						context.Context,
						string,
					) (UserStats, bool, error) {
						return UserStats{}, false, test.getError
					},
					set: func(
						context.Context,
						string,
						UserStats,
						time.Duration,
					) error {
						return test.setError
					},
				},
				time.Minute,
			)
			if err != nil {
				t.Fatalf("create service: %v", err)
			}

			_, err = service.Get(context.Background(), "mhmdnurf")
			if !errors.Is(err, test.wantError) {
				t.Fatalf(
					"expected error %v, got %v",
					test.wantError,
					err,
				)
			}
		})
	}
}

func TestNewServiceRejectsInvalidConfiguration(t *testing.T) {
	validFetcher := fetcherStub{
		fetch: func(
			context.Context,
			string,
		) (UserStats, error) {
			return UserStats{}, nil
		},
	}

	validCache := cacheStub{
		get: func(
			context.Context,
			string,
		) (UserStats, bool, error) {
			return UserStats{}, false, nil
		},
		set: func(
			context.Context,
			string,
			UserStats,
			time.Duration,
		) error {
			return nil
		},
	}

	tests := []struct {
		name    string
		fetcher Fetcher
		cache   Cache
		ttl     time.Duration
	}{
		{
			name:    "missing fetcher",
			fetcher: nil,
			cache:   validCache,
			ttl:     time.Minute,
		},
		{
			name:    "missing cache",
			fetcher: validFetcher,
			cache:   nil,
			ttl:     time.Minute,
		},
		{
			name:    "zero TTL",
			fetcher: validFetcher,
			cache:   validCache,
			ttl:     0,
		},
		{
			name:    "negative TTL",
			fetcher: validFetcher,
			cache:   validCache,
			ttl:     -time.Minute,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewService(
				test.fetcher,
				test.cache,
				test.ttl,
			)

			if err == nil {
				t.Fatal("expected an error")
			}

			if service != nil {
				t.Fatal("expected a nil service")
			}
		})
	}
}

func TestServiceGetStopsAfterCacheGetError(t *testing.T) {
	wantError := errors.New("cache unavailable")

	service, err := NewService(
		fetcherStub{
			fetch: func(
				context.Context,
				string,
			) (UserStats, error) {
				t.Fatal("fetcher should not be called")
				return UserStats{}, nil
			},
		},
		cacheStub{
			get: func(
				context.Context,
				string,
			) (UserStats, bool, error) {
				return UserStats{}, false, wantError
			},
			set: func(
				context.Context,
				string,
				UserStats,
				time.Duration,
			) error {
				t.Fatal("cache Set should not be called")
				return nil
			},
		},
		time.Minute,
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.Get(context.Background(), "mhmdnurf")
	if !errors.Is(err, wantError) {
		t.Fatalf("expected error %v, got %v", wantError, err)
	}
}
