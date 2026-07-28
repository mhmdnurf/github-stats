package github

import (
	"errors"
	"net/http"
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

func NewClient(
	token string,
	httpClient *http.Client,
) (*Client, error) {
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
