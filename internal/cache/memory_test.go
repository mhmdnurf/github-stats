package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

func TestMemoryGetSetAndExpiration(t *testing.T) {
	now := time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC)

	memory := NewMemory()
	memory.now = func() time.Time {
		return now
	}

	ctx := context.Background()
	key := "stats:mhmdnurf"
	value := stats.UserStats{
		Name:     "Muhammad Nurfatkhur Rahman",
		Username: "mhmdnurf",
		Stars:    128,
	}

	_, found, err := memory.Get(ctx, key)
	if err != nil {
		t.Fatalf("get missing value: %v", err)
	}
	if found {
		t.Fatal("expected cache miss")
	}

	if err := memory.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("set value: %v", err)
	}

	got, found, err := memory.Get(ctx, key)
	if err != nil {
		t.Fatalf("get cached value: %v", err)
	}
	if !found {
		t.Fatal("expected cache hit")
	}
	if got != value {
		t.Fatalf("unexpected value: got %+v, want %+v", got, value)
	}

	now = now.Add(time.Minute)

	_, found, err = memory.Get(ctx, key)
	if err != nil {
		t.Fatalf("get expired value: %v", err)
	}
	if found {
		t.Fatal("expected expired value to be a cache miss")
	}

	if len(memory.entries) != 0 {
		t.Fatal("expected expired entry to be removed")
	}
}

func TestMemoryRejectsInvalidTTL(t *testing.T) {
	memory := NewMemory()
	value := stats.UserStats{Username: "mhmdnurf"}

	tests := []time.Duration{
		0,
		-time.Second,
	}

	for _, ttl := range tests {
		err := memory.Set(
			context.Background(),
			"stats:mhmdnurf",
			value,
			ttl,
		)

		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatalf(
				"TTL %s: expected ErrInvalidTTL, got %v",
				ttl,
				err,
			)
		}
	}
}

func TestMemoryRespectsCanceledContext(t *testing.T) {
	memory := NewMemory()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := memory.Set(
		ctx,
		"stats:mhmdnurf",
		stats.UserStats{},
		time.Minute,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation from Set, got %v", err)
	}

	_, _, err = memory.Get(ctx, "stats:mhmdnurf")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation from Get, got %v", err)
	}
}

func TestMemoryConcurrentAccess(t *testing.T) {
	const workers = 100

	memory := NewMemory()
	ctx := context.Background()

	var waitGroup sync.WaitGroup
	errorChannel := make(chan error, workers)

	for index := range workers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			key := fmt.Sprintf("stats:user-%d", index)
			value := stats.UserStats{
				Username: fmt.Sprintf("user-%d", index),
				Stars:    index,
			}

			if err := memory.Set(ctx, key, value, time.Minute); err != nil {
				errorChannel <- fmt.Errorf("set %s: %w", key, err)
				return
			}

			got, found, err := memory.Get(ctx, key)
			if err != nil {
				errorChannel <- fmt.Errorf("get %s: %w", key, err)
				return
			}

			if !found {
				errorChannel <- fmt.Errorf("expected cache hit for %s", key)
				return
			}

			if got != value {
				errorChannel <- fmt.Errorf(
					"unexpected value for %s: got %+v, want %+v",
					key,
					got,
					value,
				)
			}
		}()
	}

	waitGroup.Wait()
	close(errorChannel)

	for err := range errorChannel {
		t.Error(err)
	}
}
