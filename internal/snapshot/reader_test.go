package snapshot

import (
	"context"
	"errors"
	"testing"
	"time"
)

type storeStub[T any] struct {
	getValue Snapshot[T]
	getError error
	getCalls int
	setCalls int
}

func (stub *storeStub[T]) Get(
	_ context.Context,
	_ string,
) (Snapshot[T], error) {
	stub.getCalls++
	return stub.getValue, stub.getError
}

func (stub *storeStub[T]) Set(
	_ context.Context,
	_ string,
	_ Snapshot[T],
) error {
	stub.setCalls++
	return nil
}

func TestNewReaderValidatesDependencies(t *testing.T) {
	store := &storeStub[string]{}

	if _, err := NewReader[string](nil, store); err == nil {
		t.Fatal("expected nil memory to be rejected")
	}

	if _, err := NewReader[string](
		NewMemory[string](),
		nil,
	); !errors.Is(err, ErrStoreRequired) {
		t.Fatalf("unexpected store error: %v", err)
	}
}

func TestReaderReturnsMemorySnapshot(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory[string]()
	now := time.Now()
	cached := Snapshot[string]{
		Value:        "memory",
		RefreshAfter: now.Add(time.Hour),
	}

	if err := memory.Set(ctx, "stats:all", cached); err != nil {
		t.Fatalf("seed memory: %v", err)
	}

	store := &storeStub[string]{
		getValue: Snapshot[string]{
			Value:        "persistent",
			RefreshAfter: now.Add(time.Hour),
		},
	}
	reader, err := NewReader(memory, store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}

	got, err := reader.Get(ctx, "stats:all")
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	if got.Value != "memory" {
		t.Fatalf("unexpected value: %q", got.Value)
	}

	if store.getCalls != 0 {
		t.Fatalf(
			"persistent store called %d times",
			store.getCalls,
		)
	}
}

func TestReaderPromotesPersistentSnapshot(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory[string]()
	now := time.Now()
	store := &storeStub[string]{
		getValue: Snapshot[string]{
			Value:        "persistent",
			RefreshAfter: now.Add(time.Hour),
		},
	}
	reader, err := NewReader(memory, store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}

	for range 2 {
		got, getErr := reader.Get(ctx, "stats:all")
		if getErr != nil {
			t.Fatalf("get snapshot: %v", getErr)
		}
		if got.Value != "persistent" {
			t.Fatalf("unexpected value: %q", got.Value)
		}
	}

	if store.getCalls != 1 {
		t.Fatalf(
			"persistent store called %d times, want 1",
			store.getCalls,
		)
	}
}

func TestReaderRefreshesStaleMemorySnapshot(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory[string]()
	now := time.Date(
		2026,
		time.July,
		27,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	if err := memory.Set(
		ctx,
		"stats:all",
		Snapshot[string]{
			Value:        "stale",
			RefreshAfter: now,
		},
	); err != nil {
		t.Fatalf("seed memory: %v", err)
	}

	store := &storeStub[string]{
		getValue: Snapshot[string]{
			Value:        "fresh",
			RefreshAfter: now.Add(time.Hour),
		},
	}
	reader, err := NewReader(memory, store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}
	reader.now = func() time.Time {
		return now
	}

	got, err := reader.Get(ctx, "stats:all")
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}

	if got.Value != "fresh" {
		t.Fatalf("unexpected value: %q", got.Value)
	}

	if store.getCalls != 1 {
		t.Fatalf(
			"persistent store called %d times, want 1",
			store.getCalls,
		)
	}
}

func TestReaderReturnsStaleMemoryWhenStoreFails(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory[string]()
	now := time.Date(
		2026,
		time.July,
		27,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	if err := memory.Set(
		ctx,
		"stats:all",
		Snapshot[string]{
			Value:        "stale",
			RefreshAfter: now,
		},
	); err != nil {
		t.Fatalf("seed memory: %v", err)
	}

	store := &storeStub[string]{
		getError: errors.New("firestore unavailable"),
	}
	reader, err := NewReader(memory, store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}
	reader.now = func() time.Time {
		return now
	}

	got, err := reader.Get(ctx, "stats:all")
	if err != nil {
		t.Fatalf("get stale snapshot: %v", err)
	}

	if got.Value != "stale" {
		t.Fatalf("unexpected value: %q", got.Value)
	}
}

func TestReaderPropagatesPersistentError(t *testing.T) {
	wantError := errors.New("firestore unavailable")
	store := &storeStub[string]{
		getError: wantError,
	}
	reader, err := NewReader(NewMemory[string](), store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}

	_, err = reader.Get(context.Background(), "stats:all")
	if !errors.Is(err, wantError) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReaderRespectsCanceledContext(t *testing.T) {
	store := &storeStub[string]{
		getValue: Snapshot[string]{Value: "persistent"},
	}
	reader, err := NewReader(NewMemory[string](), store)
	if err != nil {
		t.Fatalf("create reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = reader.Get(ctx, "stats:all")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.getCalls != 0 {
		t.Fatalf(
			"persistent store called %d times",
			store.getCalls,
		)
	}
}
