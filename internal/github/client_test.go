package github

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mhmdnurf/github-stats/internal/stats"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestClientFetch(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", request.Method)
			}

			if request.URL.String() != graphqlEndpoint {
				t.Fatalf("unexpected URL: %s", request.URL)
			}

			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("expected bearer token")
			}

			body := `{
				"data": {
					"user": {
						"name": "Muhammad Nurfatkhur Rahman",
						"login": "mhmdnurf",
						"repositories": {
							"totalCount": 2,
							"nodes": [
								{"stargazerCount": 100},
								{"stargazerCount": 28}
							],
							"pageInfo": {
								"hasNextPage": false,
								"endCursor": null
							}
						},
						"contributionsCollection": {
							"totalCommitContributions": 1245
						},
						"pullRequests": {
							"totalCount": 86
						},
						"followers": {
							"totalCount": 57
						}
					}
				}
			}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	client, err := NewClient("test-token", httpClient)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	got, err := client.Fetch(context.Background(), "mhmdnurf")
	if err != nil {
		t.Fatalf("fetch stats: %v", err)
	}

	want := stats.UserStats{
		Name:         "Muhammad Nurfatkhur Rahman",
		Username:     "mhmdnurf",
		Repositories: 2,
		Stars:        128,
		Commits:      1245,
		PullRequests: 86,
		Followers:    57,
	}

	if got != want {
		t.Fatalf("unexpected stats: got %+v, want %+v", got, want)
	}
}

func TestClientFetchPaginatesRepositories(t *testing.T) {
	requestCount := 0

	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			requestCount++

			var payload graphqlRequest
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}

			var body string

			switch requestCount {
			case 1:
				if payload.Variables["cursor"] != nil {
					t.Fatalf(
						"expected nil first cursor, got %v",
						payload.Variables["cursor"],
					)
				}

				body = `{
					"data": {
						"user": {
							"name": "Muhammad Nurfatkhur Rahman",
							"login": "mhmdnurf",
							"repositories": {
								"totalCount": 101,
								"nodes": [
									{"stargazerCount": 100}
								],
								"pageInfo": {
									"hasNextPage": true,
									"endCursor": "cursor-1"
								}
							},
							"contributionsCollection": {
								"totalCommitContributions": 1245
							},
							"pullRequests": {
								"totalCount": 86
							},
							"followers": {
								"totalCount": 57
							}
						}
					}
				}`

			case 2:
				if payload.Variables["cursor"] != "cursor-1" {
					t.Fatalf(
						"expected cursor-1, got %v",
						payload.Variables["cursor"],
					)
				}

				body = `{
					"data": {
						"user": {
							"repositories": {
								"totalCount": 101,
								"nodes": [
									{"stargazerCount": 28}
								],
								"pageInfo": {
									"hasNextPage": false,
									"endCursor": null
								}
							}
						}
					}
				}`

			default:
				t.Fatalf("unexpected request number: %d", requestCount)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}

	client, err := NewClient("test-token", httpClient)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	got, err := client.Fetch(context.Background(), "mhmdnurf")
	if err != nil {
		t.Fatalf("fetch stats: %v", err)
	}

	if requestCount != 2 {
		t.Fatalf("expected 2 requests, got %d", requestCount)
	}

	want := stats.UserStats{
		Name:         "Muhammad Nurfatkhur Rahman",
		Username:     "mhmdnurf",
		Repositories: 101,
		Stars:        128,
		Commits:      1245,
		PullRequests: 86,
		Followers:    57,
	}

	if got != want {
		t.Fatalf("unexpected stats: got %+v, want %+v", got, want)
	}
}

func TestClientFetchErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		body          string
		wantError     error
		errorContains string
	}{
		{
			name:       "user not found",
			statusCode: http.StatusOK,
			body: `{
				"data": {
					"user": null
				}
			}`,
			wantError: ErrUserNotFound,
		},
		{
			name:       "graphql error",
			statusCode: http.StatusOK,
			body: `{
				"errors": [
					{"message": "rate limit exceeded"}
				]
			}`,
			errorContains: "github graphql error: rate limit exceeded",
		},
		{
			name:          "http error",
			statusCode:    http.StatusUnauthorized,
			body:          `{}`,
			errorContains: "HTTP status 401",
		},
		{
			name:       "missing pagination cursor",
			statusCode: http.StatusOK,
			body: `{
				"data": {
					"user": {
						"name": "Muhammad Nurfatkhur Rahman",
						"login": "mhmdnurf",
						"repositories": {
							"totalCount": 101,
							"nodes": [],
							"pageInfo": {
								"hasNextPage": true,
								"endCursor": null
							}
						},
						"contributionsCollection": {
							"totalCommitContributions": 1245
						},
						"pullRequests": {
							"totalCount": 86
						},
						"followers": {
							"totalCount": 57
						}
					}
				}
			}`,
			errorContains: "without an end cursor",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{
				Transport: roundTripFunc(
					func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: test.statusCode,
							Header:     make(http.Header),
							Body: io.NopCloser(
								strings.NewReader(test.body),
							),
						}, nil
					},
				),
			}

			client, err := NewClient("test-token", httpClient)
			if err != nil {
				t.Fatalf("create client: %v", err)
			}

			_, err = client.Fetch(context.Background(), "mhmdnurf")
			if err == nil {
				t.Fatal("expected an error")
			}

			if test.wantError != nil &&
				!errors.Is(err, test.wantError) {
				t.Fatalf(
					"expected error %v, got %v",
					test.wantError,
					err,
				)
			}

			if test.errorContains != "" &&
				!strings.Contains(err.Error(), test.errorContains) {
				t.Fatalf(
					"expected error containing %q, got %q",
					test.errorContains,
					err,
				)
			}
		})
	}
}

func TestNewClientValidatesDependencies(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		httpClient *http.Client
		wantError  string
	}{
		{
			name:       "empty token",
			token:      "",
			httpClient: &http.Client{},
			wantError:  "github token is required",
		},
		{
			name:       "whitespace token",
			token:      "   ",
			httpClient: &http.Client{},
			wantError:  "github token is required",
		},
		{
			name:       "nil HTTP client",
			token:      "test-token",
			httpClient: nil,
			wantError:  "http client is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewClient(
				test.token,
				test.httpClient,
			)

			if err == nil {
				t.Fatal("expected an error")
			}

			if client != nil {
				t.Fatal("expected a nil client")
			}

			if !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf(
					"expected error containing %q, got %q",
					test.wantError,
					err,
				)
			}
		})
	}
}

func TestNewClientNormalizesToken(t *testing.T) {
	httpClient := &http.Client{}

	client, err := NewClient(
		"  test-token  ",
		httpClient,
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if client.token != "test-token" {
		t.Fatalf("unexpected token: %q", client.token)
	}

	if client.httpClient != httpClient {
		t.Fatal("expected the supplied HTTP client")
	}
}
