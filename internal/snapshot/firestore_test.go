package snapshot

import (
	"errors"
	"testing"

	"cloud.google.com/go/firestore"
)

func TestNewFirestoreValidatesConfiguration(t *testing.T) {
	_, err := NewFirestore[string](nil, "snapshots")
	if err == nil {
		t.Fatal("expected nil client to be rejected")
	}

	_, err = NewFirestore[string](&firestore.Client{}, " ")
	if !errors.Is(err, ErrCollectionRequired) {
		t.Fatalf("unexpected collection error: %v", err)
	}
}

func TestNewFirestoreNormalizesCollection(t *testing.T) {
	store, err := NewFirestore[string](
		&firestore.Client{},
		" snapshots ",
	)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if store.collection != "snapshots" {
		t.Fatalf(
			"unexpected collection: %q",
			store.collection,
		)
	}
}
