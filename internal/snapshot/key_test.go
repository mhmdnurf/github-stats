package snapshot

import (
	"errors"
	"testing"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

func TestKey(t *testing.T) {
	t.Parallel()

	got, err := Key(
		"  MHMDNURF  ",
		KindStats,
		repositoryScope.ScopeAll,
	)
	if err != nil {
		t.Fatalf("Key() error = %v", err)
	}

	const want = "v1:mhmdnurf:stats:all"
	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestKeyRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		username string
		kind     Kind
		scope    repositoryScope.Scope
		wantErr  error
	}{
		{
			name:    "missing username",
			kind:    KindStats,
			scope:   repositoryScope.ScopePublic,
			wantErr: ErrUsernameRequired,
		},
		{
			name:     "invalid kind",
			username: "mhmdnurf",
			kind:     Kind("unknown"),
			scope:    repositoryScope.ScopePublic,
			wantErr:  ErrKindInvalid,
		},
		{
			name:     "invalid scope",
			username: "mhmdnurf",
			kind:     KindStats,
			scope:    repositoryScope.Scope("unknown"),
			wantErr:  ErrScopeInvalid,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := Key(
				test.username,
				test.kind,
				test.scope,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Key() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}
