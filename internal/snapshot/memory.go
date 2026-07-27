package snapshot

import (
	"context"
	"strings"
	"sync"
)

type Memory[T any] struct {
	mu      sync.RWMutex
	entries map[string]Snapshot[T]
}

func NewMemory[T any]() *Memory[T] {
	return &Memory[T]{
		entries: make(map[string]Snapshot[T]),
	}
}

func (memory *Memory[T]) Get(
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

	memory.mu.RLock()
	value, found := memory.entries[normalizedKey]
	memory.mu.RUnlock()

	if !found {
		return Snapshot[T]{}, ErrNotFound
	}

	return value, nil
}

func (memory *Memory[T]) Set(
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

	memory.mu.Lock()
	memory.entries[normalizedKey] = value
	memory.mu.Unlock()

	return nil
}

var _ Store[any] = (*Memory[any])(nil)
