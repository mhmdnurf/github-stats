package cache

import (
	"context"
	"sync"
	"time"

	"github.com/mhmdnurf/github-stats/internal/languages"
)

type languageMemoryEntry struct {
	value     languages.UserLanguages
	expiresAt time.Time
}

type LanguageMemory struct {
	mu      sync.RWMutex
	entries map[string]languageMemoryEntry
	now     func() time.Time
}

var _ languages.Cache = (*LanguageMemory)(nil)

func NewLanguageMemory() *LanguageMemory {
	return &LanguageMemory{
		entries: make(map[string]languageMemoryEntry),
		now:     time.Now,
	}
}

func (m *LanguageMemory) Get(
	ctx context.Context,
	key string,
) (languages.UserLanguages, bool, error) {
	if err := ctx.Err(); err != nil {
		return languages.UserLanguages{}, false, err
	}

	m.mu.RLock()
	entry, found := m.entries[key]

	if !found {
		m.mu.RUnlock()
		return languages.UserLanguages{}, false, nil
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
		return languages.UserLanguages{}, false, nil
	}

	if !m.now().Before(entry.expiresAt) {
		delete(m.entries, key)
		return languages.UserLanguages{}, false, nil
	}

	return entry.value, true, nil
}

func (m *LanguageMemory) Set(
	ctx context.Context,
	key string,
	value languages.UserLanguages,
	ttl time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if ttl <= 0 {
		return ErrInvalidTTL
	}

	entry := languageMemoryEntry{
		value:     value,
		expiresAt: m.now().Add(ttl),
	}

	m.mu.Lock()
	m.entries[key] = entry
	m.mu.Unlock()

	return nil
}
