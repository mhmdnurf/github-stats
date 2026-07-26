package cache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/mhmdnurf/github-stats/internal/languages"
)

func TestLanguageMemoryGetSetAndExpiration(t *testing.T) {
	now := time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC)

	memory := NewLanguageMemory()
	memory.now = func() time.Time {
		return now
	}

	ctx := context.Background()
	key := "languages:mhmdnurf"
	value := languages.UserLanguages{
		Username: "mhmdnurf",
		Languages: []languages.LanguageUsage{
			{
				Name:  "Go",
				Color: "#00ADD8",
				Bytes: 125_000,
			},
		},
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
	if !reflect.DeepEqual(value, got) {
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

func TestLanguageMemoryRejectsInvalidTTL(t *testing.T) {
	memory := NewLanguageMemory()
	value := languages.UserLanguages{
		Username: "mhmdnurf",
	}

	tests := []time.Duration{
		0,
		-time.Second,
	}

	for _, ttl := range tests {
		err := memory.Set(
			context.Background(),
			"languages:mhmdnurf",
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

func TestLanguageMemoryRespectsCanceledContext(t *testing.T) {
	memory := NewLanguageMemory()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := memory.Set(
		ctx,
		"languages:mhmdnurf",
		languages.UserLanguages{},
		time.Minute,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation from Set, got %v", err)
	}

	_, _, err = memory.Get(ctx, "languages:mhmdnurf")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation from Get, got %v", err)
	}
}

func TestLanguageMemoryConcurrentAccess(t *testing.T) {
	const workers = 100

	memory := NewLanguageMemory()
	ctx := context.Background()

	var waitGroup sync.WaitGroup
	errorChannel := make(chan error, workers)

	for index := range workers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			key := fmt.Sprintf("languages:user-%d", index)
			value := languages.UserLanguages{
				Username: fmt.Sprintf("user-%d", index),
				Languages: []languages.LanguageUsage{
					{
						Name:  "Go",
						Color: "#00ADD8",
						Bytes: int64(index),
					},
				},
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

			if !reflect.DeepEqual(got, value) {
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
