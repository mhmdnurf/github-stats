package github

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/languages"
	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const userLanguagesQuery = `
query UserLanguages($username: String!, $cursor: String, $privacy: RepositoryPrivacy) {
	user(login: $username) {
		login
		repositories(
			first: 100
			after: $cursor
			ownerAffiliations: OWNER
			privacy: $privacy
		) {
			nodes {
				isFork
				isArchived
				languages(
					first: 100
					orderBy: {field: SIZE, direction: DESC}
				) {
					edges {
						size
						node {
							name
							color
						}
					}
				}
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}
`

func (client *Client) FetchLanguages(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (languages.UserLanguages, error) {
	result := languages.UserLanguages{
		Scope: scope,
	}
	totals := make(map[string]languages.LanguageUsage)

	var cursor *string
	firstPage := true

	for {
		user, err := client.fetchLanguagesPage(
			ctx,
			username,
			cursor,
			scope,
		)
		if err != nil {
			return languages.UserLanguages{}, err
		}

		if firstPage {
			result.Username = user.Login
			firstPage = false
		}

		for _, repository := range user.Repositories.Nodes {
			if repository.IsFork || repository.IsArchived {
				continue
			}

			for _, edge := range repository.Languages.Edges {
				name := strings.TrimSpace(edge.Node.Name)
				if name == "" || edge.Size <= 0 {
					continue
				}

				usage := totals[name]
				usage.Name = name
				usage.Bytes += edge.Size

				if usage.Color == "" &&
					edge.Node.Color != nil {
					usage.Color = *edge.Node.Color
				}

				totals[name] = usage
			}
		}

		if !user.Repositories.PageInfo.HasNextPage {
			break
		}

		if user.Repositories.PageInfo.EndCursor == nil {
			return languages.UserLanguages{}, errors.New(
				"github returned a repository page without an end cursor",
			)
		}

		cursor = user.Repositories.PageInfo.EndCursor
	}

	result.Languages = make(
		[]languages.LanguageUsage,
		0,
		len(totals),
	)

	for _, usage := range totals {
		result.Languages = append(
			result.Languages,
			usage,
		)
	}

	sort.Slice(
		result.Languages,
		func(left, right int) bool {
			if result.Languages[left].Bytes ==
				result.Languages[right].Bytes {
				return result.Languages[left].Name <
					result.Languages[right].Name
			}

			return result.Languages[left].Bytes >
				result.Languages[right].Bytes
		},
	)

	return result, nil
}

func (client *Client) fetchLanguagesPage(
	ctx context.Context,
	username string,
	cursor *string,
	scope repositoryScope.Scope,
) (*graphqlUser, error) {
	return client.fetchUser(
		ctx,
		userLanguagesQuery,
		map[string]any{
			"username": username,
			"cursor":   cursor,
			"privacy":  repositoryPrivacy(scope),
		},
		languages.ErrUserNotFound,
	)
}
