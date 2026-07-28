package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

type graphqlError struct {
	Message string `json:"message"`
}

func (client *Client) fetchUser(
	ctx context.Context,
	query string,
	variables map[string]any,
	userNotFoundError error,
) (*graphqlUser, error) {
	body, err := json.Marshal(graphqlRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"encode github graphql request: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		graphqlEndpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create github graphql request: %w",
			err,
		)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+client.token,
	)
	request.Header.Set(
		"Content-Type",
		"application/json",
	)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "github-stats")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"execute github graphql request: %w",
			err,
		)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"github graphql returned HTTP status %d",
			response.StatusCode,
		)
	}

	var result graphqlResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"decode github graphql response: %w",
			err,
		)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf(
			"github graphql error: %s",
			result.Errors[0].Message,
		)
	}

	if result.Data.User == nil {
		return nil, userNotFoundError
	}

	return result.Data.User, nil
}
