package refresh

import (
	"time"

	"github.com/mhmdnurf/github-stats/internal/snapshot"
)

func newSnapshot[T any](
	value T,
	refreshedAt time.Time,
	freshness time.Duration,
) snapshot.Snapshot[T] {
	return snapshot.Snapshot[T]{
		Value:         value,
		UpdatedAt:     refreshedAt,
		RefreshAfter:  refreshedAt.Add(freshness),
		SchemaVersion: snapshot.CurrentSchemaVersion,
	}
}
