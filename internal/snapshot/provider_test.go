package snapshot

import (
	"context"
	"errors"
	"testing"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

type readerStub[T any] struct {
	value Snapshot[T]
	err   error
	key   string
}

func (stub *readerStub[T]) Get(
	_ context.Context,
	key string,
) (Snapshot[T], error) {
	stub.key = key
	return stub.value, stub.err
}

func TestProviderReturnsSnapshotValue(t *testing.T) {
	t.Parallel()

	unavailableFailure := errors.New("snapshot unavailable")
	reader := &readerStub[string]{
		value: Snapshot[string]{
			Value: "stored value",
		},
	}
	provider, err := NewProvider[string](
		reader,
		KindStats,
		unavailableFailure,
	)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	got, err := provider.Get(
		context.Background(),
		"  MHMDNURF  ",
		repositoryScope.ScopeAll,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "stored value" {
		t.Fatalf("Get() = %q, want %q", got, "stored value")
	}

	const wantKey = "v1:mhmdnurf:stats:all"
	if reader.key != wantKey {
		t.Fatalf(
			"reader key = %q, want %q",
			reader.key,
			wantKey,
		)
	}
}

func TestProviderMapsReaderFailureToUnavailable(
	t *testing.T,
) {
	t.Parallel()

	readFailure := errors.New("Firestore unavailable")
	unavailableFailure := errors.New("snapshot unavailable")
	provider, err := NewProvider[string](
		&readerStub[string]{
			err: readFailure,
		},
		KindLanguages,
		unavailableFailure,
	)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	_, err = provider.Get(
		context.Background(),
		"mhmdnurf",
		repositoryScope.ScopePublic,
	)
	if !errors.Is(err, unavailableFailure) {
		t.Fatalf(
			"Get() error = %v, want %v",
			err,
			unavailableFailure,
		)
	}
	if !errors.Is(err, readFailure) {
		t.Fatalf(
			"Get() error = %v, want wrapped reader error",
			err,
		)
	}
}

func TestProviderPreservesContextFailure(t *testing.T) {
	t.Parallel()

	unavailableFailure := errors.New("snapshot unavailable")
	provider, err := NewProvider[string](
		&readerStub[string]{
			err: context.DeadlineExceeded,
		},
		KindStats,
		unavailableFailure,
	)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	_, err = provider.Get(
		context.Background(),
		"mhmdnurf",
		repositoryScope.ScopePublic,
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"Get() error = %v, want deadline exceeded",
			err,
		)
	}
}

func TestNewProviderValidatesDependencies(t *testing.T) {
	t.Parallel()

	if _, err := NewProvider[string](
		nil,
		KindStats,
		errors.New("snapshot unavailable"),
	); !errors.Is(err, ErrReaderRequired) {
		t.Fatalf(
			"nil reader error = %v, want %v",
			err,
			ErrReaderRequired,
		)
	}

	if _, err := NewProvider[string](
		&readerStub[string]{},
		Kind("unknown"),
		errors.New("snapshot unavailable"),
	); !errors.Is(err, ErrKindInvalid) {
		t.Fatalf(
			"invalid kind error = %v, want %v",
			err,
			ErrKindInvalid,
		)
	}

	if _, err := NewProvider[string](
		&readerStub[string]{},
		KindStats,
		nil,
	); err == nil {
		t.Fatal("expected nil unavailable error to be rejected")
	}
}
