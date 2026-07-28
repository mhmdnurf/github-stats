package cacheaside

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type cacheStub[T any] struct {
	get func(context.Context, string) (T, bool, error)
	set func(context.Context, string, T, time.Duration) error
}

func (stub cacheStub[T]) Get(
	ctx context.Context,
	key string,
) (T, bool, error) {
	return stub.get(ctx, key)
}

func (stub cacheStub[T]) Set(
	ctx context.Context,
	key string,
	value T,
	ttl time.Duration,
) error {
	return stub.set(ctx, key, value, ttl)
}

func TestLoaderReturnsCachedValue(t *testing.T) {
	const want = "cached"

	loader, err := New[string](
		cacheStub[string]{
			get: func(
				context.Context,
				string,
			) (string, bool, error) {
				return want, true, nil
			},
			set: func(
				context.Context,
				string,
				string,
				time.Duration,
			) error {
				t.Fatal("cache Set should not be called")
				return nil
			},
		},
		time.Minute,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	got, err := loader.Load(
		context.Background(),
		"key",
		func(context.Context) (string, error) {
			t.Fatal("fetch should not be called")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("load value: %v", err)
	}

	if got != want {
		t.Fatalf("value: got %q, want %q", got, want)
	}
}

func TestLoaderFetchesAndCachesMiss(t *testing.T) {
	const (
		key  = "value:key"
		want = "fetched"
		ttl  = 15 * time.Minute
	)

	loader, err := New[string](
		cacheStub[string]{
			get: func(
				_ context.Context,
				gotKey string,
			) (string, bool, error) {
				if gotKey != key {
					t.Fatalf("cache key: got %q, want %q", gotKey, key)
				}

				return "", false, nil
			},
			set: func(
				_ context.Context,
				gotKey string,
				value string,
				gotTTL time.Duration,
			) error {
				if gotKey != key {
					t.Fatalf("cache key: got %q, want %q", gotKey, key)
				}
				if value != want {
					t.Fatalf("value: got %q, want %q", value, want)
				}
				if gotTTL != ttl {
					t.Fatalf("TTL: got %s, want %s", gotTTL, ttl)
				}

				return nil
			},
		},
		ttl,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	got, err := loader.Load(
		context.Background(),
		key,
		func(context.Context) (string, error) {
			return want, nil
		},
	)
	if err != nil {
		t.Fatalf("load value: %v", err)
	}

	if got != want {
		t.Fatalf("value: got %q, want %q", got, want)
	}
}

func TestLoaderPropagatesStageErrors(t *testing.T) {
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
			name:      "cache get",
			getError:  cacheGetError,
			wantError: cacheGetError,
		},
		{
			name:       "fetch",
			fetchError: fetchError,
			wantError:  fetchError,
		},
		{
			name:      "cache set",
			setError:  cacheSetError,
			wantError: cacheSetError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader, err := New[string](
				cacheStub[string]{
					get: func(
						context.Context,
						string,
					) (string, bool, error) {
						return "", false, test.getError
					},
					set: func(
						context.Context,
						string,
						string,
						time.Duration,
					) error {
						return test.setError
					},
				},
				time.Minute,
				"value",
			)
			if err != nil {
				t.Fatalf("create loader: %v", err)
			}

			_, err = loader.Load(
				context.Background(),
				"key",
				func(context.Context) (string, error) {
					return "fetched", test.fetchError
				},
			)
			if !errors.Is(err, test.wantError) {
				t.Fatalf(
					"error: got %v, want %v",
					err,
					test.wantError,
				)
			}
		})
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	validCache := cacheStub[string]{
		get: func(
			context.Context,
			string,
		) (string, bool, error) {
			return "", false, nil
		},
		set: func(
			context.Context,
			string,
			string,
			time.Duration,
		) error {
			return nil
		},
	}

	tests := []struct {
		name      string
		cache     Cache[string]
		ttl       time.Duration
		resource  string
		wantError error
	}{
		{
			name:      "missing cache",
			ttl:       time.Minute,
			resource:  "value",
			wantError: ErrCacheRequired,
		},
		{
			name:      "invalid TTL",
			cache:     validCache,
			resource:  "value",
			wantError: ErrTTLInvalid,
		},
		{
			name:      "missing resource",
			cache:     validCache,
			ttl:       time.Minute,
			wantError: ErrResourceRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader, err := New(
				test.cache,
				test.ttl,
				test.resource,
			)
			if !errors.Is(err, test.wantError) {
				t.Fatalf(
					"error: got %v, want %v",
					err,
					test.wantError,
				)
			}
			if loader != nil {
				t.Fatalf("loader: got %+v, want nil", loader)
			}
		})
	}
}

func TestLoaderRejectsMissingFetch(t *testing.T) {
	loader, err := New[string](
		cacheStub[string]{
			get: func(
				context.Context,
				string,
			) (string, bool, error) {
				return "", false, nil
			},
			set: func(
				context.Context,
				string,
				string,
				time.Duration,
			) error {
				return nil
			},
		},
		time.Minute,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	_, err = loader.Load(context.Background(), "key", nil)
	if !errors.Is(err, ErrFetchRequired) {
		t.Fatalf("error: got %v, want %v", err, ErrFetchRequired)
	}
}

func TestLoaderDeduplicatesConcurrentFetchesForSameKey(t *testing.T) {
	const workers = 50

	loader, err := New[string](
		cacheStub[string]{
			get: func(
				context.Context,
				string,
			) (string, bool, error) {
				return "", false, nil
			},
			set: func(
				context.Context,
				string,
				string,
				time.Duration,
			) error {
				return nil
			},
		},
		time.Minute,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	var (
		fetchCalls atomic.Int32
		ready      sync.WaitGroup
		waiters    sync.WaitGroup
	)
	ready.Add(workers)
	waiters.Add(workers)

	start := make(chan struct{})
	errorsChannel := make(chan error, workers)

	for range workers {
		go func() {
			defer waiters.Done()
			ready.Done()
			<-start

			value, loadErr := loader.Load(
				context.Background(),
				"shared-key",
				func(context.Context) (string, error) {
					fetchCalls.Add(1)
					time.Sleep(50 * time.Millisecond)
					return "fetched", nil
				},
			)
			if loadErr != nil {
				errorsChannel <- loadErr
				return
			}
			if value != "fetched" {
				errorsChannel <- fmt.Errorf(
					"value: got %q, want %q",
					value,
					"fetched",
				)
			}
		}()
	}

	ready.Wait()
	close(start)
	waiters.Wait()
	close(errorsChannel)

	for loadErr := range errorsChannel {
		t.Error(loadErr)
	}

	if got := fetchCalls.Load(); got != 1 {
		t.Fatalf("fetch calls: got %d, want 1", got)
	}
}

func TestLoaderDoesNotDeduplicateDifferentKeys(t *testing.T) {
	loader, err := New[string](
		cacheStub[string]{
			get: func(
				context.Context,
				string,
			) (string, bool, error) {
				return "", false, nil
			},
			set: func(
				context.Context,
				string,
				string,
				time.Duration,
			) error {
				return nil
			},
		},
		time.Minute,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	started := make(chan string, 2)
	release := make(chan struct{})
	results := make(chan error, 2)

	for _, key := range []string{"first", "second"} {
		go func() {
			value, loadErr := loader.Load(
				context.Background(),
				key,
				func(context.Context) (string, error) {
					started <- key
					<-release
					return key, nil
				},
			)
			if loadErr != nil {
				results <- loadErr
				return
			}
			if value != key {
				results <- fmt.Errorf(
					"value: got %q, want %q",
					value,
					key,
				)
				return
			}

			results <- nil
		}()
	}

	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("different keys did not fetch concurrently")
		}
	}
	close(release)

	for range 2 {
		if loadErr := <-results; loadErr != nil {
			t.Error(loadErr)
		}
	}
}

func TestLoaderRetriesAfterSharedFetchError(t *testing.T) {
	wantError := errors.New("fetch failed")
	var fetchCalls int

	loader, err := New[string](
		cacheStub[string]{
			get: func(
				context.Context,
				string,
			) (string, bool, error) {
				return "", false, nil
			},
			set: func(
				context.Context,
				string,
				string,
				time.Duration,
			) error {
				return nil
			},
		},
		time.Minute,
		"value",
	)
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	fetch := func(context.Context) (string, error) {
		fetchCalls++
		if fetchCalls == 1 {
			return "", wantError
		}

		return "fetched", nil
	}

	if _, err := loader.Load(
		context.Background(),
		"key",
		fetch,
	); !errors.Is(err, wantError) {
		t.Fatalf("first load error: got %v, want %v", err, wantError)
	}

	value, err := loader.Load(
		context.Background(),
		"key",
		fetch,
	)
	if err != nil {
		t.Fatalf("retry load: %v", err)
	}
	if value != "fetched" {
		t.Fatalf("value: got %q, want %q", value, "fetched")
	}
	if fetchCalls != 2 {
		t.Fatalf("fetch calls: got %d, want 2", fetchCalls)
	}
}
