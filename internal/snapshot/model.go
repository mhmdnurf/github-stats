package snapshot

import "time"

const CurrentSchemaVersion = 1

type Snapshot[T any] struct {
	Value         T         `firestore:"value"`
	UpdatedAt     time.Time `firestore:"updated_at"`
	RefreshAfter  time.Time `firestore:"refresh_after"`
	SchemaVersion int       `firestore:"schema_version"`
}

func (snapshot Snapshot[T]) IsStale(now time.Time) bool {
	return !now.Before(snapshot.RefreshAfter)
}
