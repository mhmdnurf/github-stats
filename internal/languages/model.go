package languages

type LanguageUsage struct {
	Name  string
	Color string
	Bytes int64
}

type UserLanguages struct {
	Username  string
	Languages []LanguageUsage
}

type RepositoryScope string

const (
	RepositoryScopePublic RepositoryScope = "public"
	RepositoryScopeAll    RepositoryScope = "all"
)
