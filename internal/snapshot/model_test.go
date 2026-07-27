package snapshot

import (
	"testing"
	"time"
)

func TestSnapshotIsStale(t *testing.T) {
	refreshAfter := time.Date(
		2026,
		time.July,
		27,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	value := Snapshot[string]{
		Value:         "cached",
		UpdatedAt:     refreshAfter.Add(-time.Hour),
		RefreshAfter:  refreshAfter,
		SchemaVersion: CurrentSchemaVersion,
	}

	if value.IsStale(refreshAfter.Add(-time.Nanosecond)) {
		t.Fatal("expected snapshot to be fresh before refresh time")
	}

	if !value.IsStale(refreshAfter) {
		t.Fatal("expected snapshot to be stale at refresh time")
	}

	if !value.IsStale(refreshAfter.Add(time.Second)) {
		t.Fatal("expected snapshot to be stale after refresh time")
	}
}
