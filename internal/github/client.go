package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/languages"
	"github.com/mhmdnurf/github-stats/internal/stats"
)

const graphqlEndpoint = "https://api.github.com/graphql"

type Client struct {
	httpClient *http.Client
	token      string
}

var _ stats.Fetcher = (*Client)(nil)
var _ languages.Fetcher = (*Client)(nil)

func NewClient(token string, httpClient *http.Client) (*Client, error) {
	normalizedToken := strings.TrimSpace(token)
	if normalizedToken == "" {
		return nil, errors.New("github token is required")
	}

	if httpClient == nil {
		return nil, errors.New("http client is required")
	}

	return &Client{
		httpClient: httpClient,
		token:      normalizedToken,
	}, nil
}

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

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphqlResponse struct {
	Data   graphqlData    `json:"data"`
	Errors []graphqlError `json:"errors"`
}

type graphqlData struct {
	User *graphqlUser `json:"user"`
}

type graphqlUser struct {
	Name                    *string              `json:"name"`
	Login                   string               `json:"login"`
	Repositories            repositoryConnection `json:"repositories"`
	ContributionsCollection contributionStats    `json:"contributionsCollection"`
	PullRequests            countConnection      `json:"pullRequests"`
	Followers               countConnection      `json:"followers"`
}

type repositoryConnection struct {
	TotalCount int          `json:"totalCount"`
	Nodes      []repository `json:"nodes"`
	PageInfo   pageInfo     `json:"pageInfo"`
}

type repository struct {
	StargazerCount int                `json:"stargazerCount"`
	IsFork         bool               `json:"isFork"`
	IsArchived     bool               `json:"isArchived"`
	Languages      languageConnection `json:"languages"`
}

type languageConnection struct {
	Edges []languageEdge `json:"edges"`
}

type languageEdge struct {
	Size int64        `json:"size"`
	Node languageNode `json:"node"`
}

type languageNode struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type pageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

type contributionStats struct {
	TotalCommitContributions int                  `json:"totalCommitContributions"`
	ContributionCalendar     contributionCalendar `json:"contributionCalendar"`
}

type contributionCalendar struct {
	Weeks []contributionWeek `json:"weeks"`
}

type contributionWeek struct {
	ContributionDays []contributionDay `json:"contributionDays"`
}

type contributionDay struct {
	ContributionCount int `json:"contributionCount"`
}

func (calendar contributionCalendar) weeklyTotals() []int {
	if len(calendar.Weeks) == 0 {
		return nil
	}

	totals := make([]int, 0, len(calendar.Weeks))
	for _, week := range calendar.Weeks {
		total := 0
		for _, day := range week.ContributionDays {
			total += day.ContributionCount
		}
		totals = append(totals, total)
	}

	return totals
}

type countConnection struct {
	TotalCount int `json:"totalCount"`
}

type graphqlError struct {
	Message string `json:"message"`
}

var ErrUserNotFound = stats.ErrUserNotFound

func (c *Client) Fetch(ctx context.Context, username string) (stats.UserStats, error) {
	var result stats.UserStats
	var cursor *string
	firstPage := true

	for {
		user, err := c.fetchPage(ctx, username, cursor)
		if err != nil {
			return stats.UserStats{}, err
		}

		if firstPage {
			result.Name = user.Login
			if user.Name != nil && strings.TrimSpace(*user.Name) != "" {
				result.Name = *user.Name
			}

			result.Username = user.Login
			result.Repositories = user.Repositories.TotalCount
			result.Commits = user.ContributionsCollection.TotalCommitContributions
			result.PullRequests = user.PullRequests.TotalCount
			result.Followers = user.Followers.TotalCount
			result.WeeklyActivity = user.ContributionsCollection.
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

func (c *Client) FetchLanguages(
	ctx context.Context,
	username string,
	scope languages.RepositoryScope,
) (languages.UserLanguages, error) {
	result := languages.UserLanguages{}
	totals := make(map[string]languages.LanguageUsage)

	var cursor *string
	firstPage := true

	for {
		user, err := c.fetchLanguagesPage(ctx, username, cursor, scope)
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

				if usage.Color == "" && edge.Node.Color != nil {
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
		result.Languages = append(result.Languages, usage)
	}

	sort.Slice(result.Languages, func(left, right int) bool {
		if result.Languages[left].Bytes == result.Languages[right].Bytes {
			return result.Languages[left].Name <
				result.Languages[right].Name
		}

		return result.Languages[left].Bytes >
			result.Languages[right].Bytes
	})

	return result, nil
}

func (c *Client) fetchPage(
	ctx context.Context,
	username string,
	cursor *string,
) (*graphqlUser, error) {
	payload := graphqlRequest{
		Query: userStatsQuery,
		Variables: map[string]any{
			"username": username,
			"cursor":   cursor,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode github graphql request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		graphqlEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create github graphql request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "github-stats")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute github graphql request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"github graphql returned HTTP status %d",
			response.StatusCode,
		)
	}

	var graphqlResult graphqlResponse
	if err := json.NewDecoder(response.Body).Decode(&graphqlResult); err != nil {
		return nil, fmt.Errorf("decode github graphql response: %w", err)
	}

	if len(graphqlResult.Errors) > 0 {
		return nil, fmt.Errorf(
			"github graphql error: %s",
			graphqlResult.Errors[0].Message,
		)
	}

	if graphqlResult.Data.User == nil {
		return nil, ErrUserNotFound
	}

	return graphqlResult.Data.User, nil
}

func (c *Client) fetchLanguagesPage(
	ctx context.Context,
	username string,
	cursor *string,
	scope languages.RepositoryScope,
) (*graphqlUser, error) {
	payload := graphqlRequest{
		Query: userLanguagesQuery,
		Variables: map[string]any{
			"username": username,
			"cursor":   cursor,
			"privacy":  repositoryPrivacy(scope),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode github graphql request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		graphqlEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create github graphql request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "github-stats")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute github graphql request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"github graphql returned HTTP status %d",
			response.StatusCode,
		)
	}

	var graphqlResult graphqlResponse
	if err := json.NewDecoder(response.Body).Decode(&graphqlResult); err != nil {
		return nil, fmt.Errorf("decode github graphql response: %w", err)
	}

	if len(graphqlResult.Errors) > 0 {
		return nil, fmt.Errorf(
			"github graphql error: %s",
			graphqlResult.Errors[0].Message,
		)
	}

	if graphqlResult.Data.User == nil {
		return nil, languages.ErrUserNotFound
	}

	return graphqlResult.Data.User, nil
}

func repositoryPrivacy(
	scope languages.RepositoryScope,
) *string {
	if scope == languages.RepositoryScopeAll {
		return nil
	}

	privacy := "PUBLIC"
	return &privacy
}
