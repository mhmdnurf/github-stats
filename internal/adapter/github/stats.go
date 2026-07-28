package github

import (
	"context"
	"errors"
	"strings"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const userStatsQuery = `
query UserStats($username: String!, $cursor: String, $privacy: RepositoryPrivacy) {
	user(login: $username) {
		name
		login
		repositories(
			first: 100
			after: $cursor
			ownerAffiliations: OWNER
			privacy: $privacy
		) {
			totalCount
			nodes {
				stargazerCount
			}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
		contributionsCollection {
			totalCommitContributions
			contributionCalendar {
				weeks {
					contributionDays {
						contributionCount
					}
				}
			}
		}
		pullRequests(first: 1) {
			totalCount
		}
		followers(first: 1) {
			totalCount
		}
	}
}
`

var ErrUserNotFound = stats.ErrUserNotFound

func (client *Client) Fetch(
	ctx context.Context,
	username string,
	scope repositoryScope.Scope,
) (stats.UserStats, error) {
	var result stats.UserStats
	var cursor *string
	firstPage := true

	for {
		user, err := client.fetchStatsPage(
			ctx,
			username,
			cursor,
			scope,
		)
		if err != nil {
			return stats.UserStats{}, err
		}

		if firstPage {
			result.Name = user.Login
			if user.Name != nil &&
				strings.TrimSpace(*user.Name) != "" {
				result.Name = *user.Name
			}

			result.Username = user.Login
			result.Repositories = user.Repositories.TotalCount
			result.Commits =
				user.ContributionsCollection.
					TotalCommitContributions
			result.PullRequests = user.PullRequests.TotalCount
			result.Followers = user.Followers.TotalCount
			result.WeeklyActivity =
				user.ContributionsCollection.
					ContributionCalendar.weeklyTotals()

			firstPage = false
		}

		for _, repository := range user.Repositories.Nodes {
			result.Stars += repository.StargazerCount
		}

		if !user.Repositories.PageInfo.HasNextPage {
			break
		}

		if user.Repositories.PageInfo.EndCursor == nil {
			return stats.UserStats{}, errors.New(
				"github returned a repository page without an end cursor",
			)
		}

		cursor = user.Repositories.PageInfo.EndCursor
	}

	return result, nil
}

func (client *Client) fetchStatsPage(
	ctx context.Context,
	username string,
	cursor *string,
	scope repositoryScope.Scope,
) (*graphqlUser, error) {
	return client.fetchUser(
		ctx,
		userStatsQuery,
		map[string]any{
			"username": username,
			"cursor":   cursor,
			"privacy":  repositoryPrivacy(scope),
		},
		ErrUserNotFound,
	)
}
