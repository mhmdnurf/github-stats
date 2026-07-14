package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

const graphqlEndpoint = "https://api.github.com/graphql"

type Client struct {
	httpClient *http.Client
	token      string
}

var _ stats.Fetcher = (*Client)(nil)

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
query UserStats($username: String!, $cursor: String) {
	user(login: $username) {
		name
		login
		repositories(
			first: 100
			after: $cursor
			ownerAffiliations: OWNER
			privacy: PUBLIC
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
	StargazerCount int `json:"stargazerCount"`
}

type pageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

type contributionStats struct {
	TotalCommitContributions int `json:"totalCommitContributions"`
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
