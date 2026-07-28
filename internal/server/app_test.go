package server

import (
	"errors"
	"testing"
)

func TestAppCloseIsIdempotent(t *testing.T) {
	closeCalls := 0
	closeFailure := errors.New("close failed")

	application := &App{
		close: func() error {
			closeCalls++
			return closeFailure
		},
	}

	for range 2 {
		if err := application.Close(); !errors.Is(err, closeFailure) {
			t.Fatalf("Close() error = %v, want %v", err, closeFailure)
		}
	}

	if closeCalls != 1 {
		t.Fatalf("close calls = %d, want 1", closeCalls)
	}
}
