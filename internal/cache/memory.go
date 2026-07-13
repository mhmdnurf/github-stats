package cache

import (
	"context"
	"sync"
	"time"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

type memoryEntry struct {
	value     stats.UserStats
	expiresAt time.Time
}

type Memory struct {
	mu      sync.RWMutex
	entries map[string]memoryEntry
	now     func() time.Time
}

var _ stats.Cache = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{
		entries: make(map[string]memoryEntry),
		now:     time.Now,
	}
}

func (m *Memory) Get(
	ctx context.Context,
	key string,
) (stats.UserStats, bool, error) {
	if err := ctx.Err(); err != nil {
		return stats.UserStats{}, false, err
	}

	m.mu.RLock()
	entry, found := m.entries[key]

	if !found {
		m.mu.RUnlock()
		return stats.UserStats{}, false, nil
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
		return stats.UserStats{}, false, nil
	}

	if !m.now().Before(entry.expiresAt) {
		delete(m.entries, key)
		return stats.UserStats{}, false, nil
	}

	return entry.value, true, nil
}

func (m *Memory) Set(
	ctx context.Context,
	key string,
	value stats.UserStats,
	ttl time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if ttl <= 0 {
		return ErrInvalidTTL
	}

	entry := memoryEntry{
		value:     value,
		expiresAt: m.now().Add(ttl),
	}

	m.mu.Lock()
	m.entries[key] = entry
	m.mu.Unlock()

	return nil
}
