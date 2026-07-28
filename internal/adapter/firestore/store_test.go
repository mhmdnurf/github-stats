package firestore

import (
	"errors"
	"testing"

	cloudfirestore "cloud.google.com/go/firestore"
)

func TestNewStoreValidatesConfiguration(t *testing.T) {
	_, err := NewStore[string](nil, "snapshots")
	if err == nil {
		t.Fatal("expected nil client to be rejected")
	}

	_, err = NewStore[string](&cloudfirestore.Client{}, " ")
	if !errors.Is(err, ErrCollectionRequired) {
		t.Fatalf("unexpected collection error: %v", err)
	}
}

func TestNewStoreNormalizesCollection(t *testing.T) {
	store, err := NewStore[string](
		&cloudfirestore.Client{},
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
