package languages

import repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"

type LanguageUsage struct {
	Name  string
	Color string
	Bytes int64
}

type UserLanguages struct {
	Username  string
	Scope     repositoryScope.Scope
	Languages []LanguageUsage
}
