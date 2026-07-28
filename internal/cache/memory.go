package cache

import (
	"context"
	"sync"
	"time"
)

type memoryEntry[T any] struct {
	value     T
	expiresAt time.Time
}

type Memory[T any] struct {
	mu      sync.RWMutex
	entries map[string]memoryEntry[T]
	now     func() time.Time
}

func NewMemory[T any]() *Memory[T] {
	return &Memory[T]{
		entries: make(map[string]memoryEntry[T]),
		now:     time.Now,
	}
}

func (m *Memory[T]) Get(
	ctx context.Context,
	key string,
) (T, bool, error) {
	var zero T

	if err := ctx.Err(); err != nil {
		return zero, false, err
	}

	m.mu.RLock()
	entry, found := m.entries[key]

	if !found {
		m.mu.RUnlock()
		return zero, false, nil
	}

	if m.now().Before(entry.expiresAt) {
		m.mu.RUnlock()
		return entry.value, true, nil
	}

	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, found = m.entries[key]
	if !found {
		return zero, false, nil
	}

	if !m.now().Before(entry.expiresAt) {
		delete(m.entries, key)
		return zero, false, nil
	}

	return entry.value, true, nil
}

func (m *Memory[T]) Set(
	ctx context.Context,
	key string,
	value T,
	ttl time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if ttl <= 0 {
		return ErrInvalidTTL
	}

	entry := memoryEntry[T]{
		value:     value,
		expiresAt: m.now().Add(ttl),
	}

	m.mu.Lock()
	m.entries[key] = entry
	m.mu.Unlock()

	return nil
}
