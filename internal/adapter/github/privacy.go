package github

import repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"

func repositoryPrivacy(
	scope repositoryScope.Scope,
) *string {
	if scope == repositoryScope.ScopeAll {
		return nil
	}

	privacy := "PUBLIC"
	return &privacy
}
