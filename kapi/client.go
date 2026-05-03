// Package kapi provides an HTTP API client for making REST requests
// outside of the browser context.
//
// This is the kexas equivalent of Playwright's APIRequestContext,
// RestSharp (C#), or RestAssured (Java). It enables API-level testing
// alongside browser-level testing within the same test suite.
//
// Usage:
//
//	client := kapi.NewClient("https://api.example.com")
//	resp, err := client.Get("/users/1")
//	if err != nil { t.Fatal(err) }
//	kassert.Equal(t, resp.StatusCode, 200)
//	kassert.Contains(t, resp.Body, `"name":"John"`)
package kapi

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the core HTTP API client for making REST requests.
// Each Client instance is isolated with its own base URL, headers,
// cookies, and HTTP client — safe for parallel use across test workers.
type Client struct {
	baseURL    string
	headers    map[string]string
	cookies    []*http.Cookie
	httpClient *http.Client
	mu         sync.Mutex
}

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

// NewClient creates a new API client with the given base URL.
// Options can be provided to customize headers, timeouts, cookies, etc.
func NewClient(baseURL string, opts ...ClientOption) *Client {
	var client *Client = &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		headers: make(map[string]string),
		cookies: make([]*http.Cookie, 0),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHeader sets a default header for all requests.
func WithHeader(key string, value string) ClientOption {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithHeaders sets multiple default headers for all requests.
func WithHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithBearerToken sets the Authorization header with a Bearer token.
func WithBearerToken(token string) ClientOption {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + token
	}
}

// WithBasicAuth sets the Authorization header with Basic auth credentials.
func WithBasicAuth(username string, password string) ClientOption {
	return func(c *Client) {
		c.headers["Authorization"] = "Basic " + basicAuth(username, password)
	}
}

// WithCookie adds a cookie to all requests.
func WithCookie(name string, value string) ClientOption {
	return func(c *Client) {
		c.cookies = append(c.cookies, &http.Cookie{Name: name, Value: value})
	}
}

// WithHTTPClient sets a custom http.Client (for TLS config, proxies, etc.)
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithNoRedirect disables automatic redirect following.
func WithNoRedirect() ClientOption {
	return func(c *Client) {
		c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}
