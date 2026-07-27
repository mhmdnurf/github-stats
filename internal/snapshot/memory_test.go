package snapshot

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMemoryGetSet(t *testing.T) {
	memory := NewMemory[string]()
	ctx := context.Background()
	value := Snapshot[string]{
		Value:         "cached",
		UpdatedAt:     time.Now(),
		RefreshAfter:  time.Now().Add(time.Hour),
		SchemaVersion: CurrentSchemaVersion,
	}

	if err := memory.Set(ctx, " stats:all ", value); err != nil {
		t.Fatalf("set snapshot: %v", err)
	}

	got, err := memory.Get(ctx, "stats:all")
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	if got.Value != value.Value {
		t.Fatalf(
			"unexpected value: got %q, want %q",
			got.Value,
			value.Value,
		)
	}
}

func TestMemoryReturnsNotFound(t *testing.T) {
	memory := NewMemory[string]()

	_, err := memory.Get(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryValidatesKey(t *testing.T) {
	memory := NewMemory[string]()
	ctx := context.Background()

	if err := memory.Set(ctx, " ", Snapshot[string]{}); !errors.Is(
		err,
		ErrKeyRequired,
	) {
		t.Fatalf("unexpected set error: %v", err)
	}

	_, err := memory.Get(ctx, "")
	if !errors.Is(err, ErrKeyRequired) {
		t.Fatalf("unexpected get error: %v", err)
	}
}

func TestMemoryRespectsCanceledContext(t *testing.T) {
	memory := NewMemory[string]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := memory.Set(ctx, "key", Snapshot[string]{}); !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf("unexpected set error: %v", err)
	}

	_, err := memory.Get(ctx, "key")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected get error: %v", err)
	}
}

func TestMemoryConcurrentAccess(t *testing.T) {
	memory := NewMemory[int]()
	ctx := context.Background()

	var group sync.WaitGroup
	for index := range 100 {
		group.Add(1)

		go func() {
			defer group.Done()

			key := "key:" + strconv.Itoa(index)
			value := Snapshot[int]{Value: index}

			if err := memory.Set(ctx, key, value); err != nil {
				t.Errorf("set %s: %v", key, err)
				return
			}

			got, err := memory.Get(ctx, key)
			if err != nil {
				t.Errorf("get %s: %v", key, err)
				return
			}

			if got.Value != index {
				t.Errorf(
					"unexpected %s value: got %d, want %d",
					key,
					got.Value,
					index,
				)
			}
		}()
	}

	group.Wait()
}
