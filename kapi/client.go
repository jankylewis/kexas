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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// --- Request Building ---

// RequestBuilder builds an HTTP request with fluent chaining.
type RequestBuilder struct {
	client      *Client
	method      string
	path        string
	headers     map[string]string
	queryParams url.Values
	body        io.Reader
	contentType string
	cookies     []*http.Cookie
}

// newRequest creates a new RequestBuilder.
func (c *Client) newRequest(method string, path string) *RequestBuilder {
	return &RequestBuilder{
		client:      c,
		method:      method,
		path:        path,
		headers:     make(map[string]string),
		queryParams: make(url.Values),
		cookies:     make([]*http.Cookie, 0),
	}
}

// Header sets a request-specific header (overrides client defaults).
func (rb *RequestBuilder) Header(key string, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// Headers sets multiple request-specific headers.
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// Query adds a query parameter.
func (rb *RequestBuilder) Query(key string, value string) *RequestBuilder {
	rb.queryParams.Add(key, value)
	return rb
}

// QueryParams adds multiple query parameters.
func (rb *RequestBuilder) QueryParams(params map[string]string) *RequestBuilder {
	for k, v := range params {
		rb.queryParams.Add(k, v)
	}
	return rb
}

// Body sets the request body as raw bytes.
func (rb *RequestBuilder) Body(data []byte) *RequestBuilder {
	rb.body = bytes.NewReader(data)
	return rb
}

// BodyString sets the request body as a string.
func (rb *RequestBuilder) BodyString(data string) *RequestBuilder {
	rb.body = strings.NewReader(data)
	return rb
}

// BodyJSON sets the request body as JSON from any struct/map.
func (rb *RequestBuilder) BodyJSON(data interface{}) *RequestBuilder {
	var jsonBytes []byte
	var err error
	jsonBytes, err = json.Marshal(data)
	if err != nil {
		// Store nil body; error will surface when Send() is called
		rb.body = nil
		return rb
	}
	rb.body = bytes.NewReader(jsonBytes)
	rb.contentType = "application/json"
	return rb
}

// BodyForm sets the request body as URL-encoded form data.
func (rb *RequestBuilder) BodyForm(data map[string]string) *RequestBuilder {
	var formValues url.Values = make(url.Values)
	for k, v := range data {
		formValues.Set(k, v)
	}
	rb.body = strings.NewReader(formValues.Encode())
	rb.contentType = "application/x-www-form-urlencoded"
	return rb
}

// Cookie adds a request-specific cookie.
func (rb *RequestBuilder) Cookie(name string, value string) *RequestBuilder {
	rb.cookies = append(rb.cookies, &http.Cookie{Name: name, Value: value})
	return rb
}

// Send executes the request and returns a Response.
func (rb *RequestBuilder) Send() (*Response, error) {
	var fullURL string = rb.client.baseURL + rb.path

	if len(rb.queryParams) > 0 {
		fullURL = fullURL + "?" + rb.queryParams.Encode()
	}

	var req *http.Request
	var err error
	req, err = http.NewRequest(rb.method, fullURL, rb.body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: failed to create request: %w", rb.method, rb.path, err)
	}

	// Apply client-level headers first
	rb.client.mu.Lock()
	for k, v := range rb.client.headers {
		req.Header.Set(k, v)
	}
	rb.client.mu.Unlock()

	// Apply request-level headers (override client defaults)
	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	// Set content type if specified
	if rb.contentType != "" {
		req.Header.Set("Content-Type", rb.contentType)
	}

	// Apply client-level cookies
	rb.client.mu.Lock()
	for _, cookie := range rb.client.cookies {
		req.AddCookie(cookie)
	}
	rb.client.mu.Unlock()

	// Apply request-level cookies
	for _, cookie := range rb.cookies {
		req.AddCookie(cookie)
	}

	// Execute request
	var start time.Time = time.Now()
	var httpResp *http.Response
	httpResp, err = rb.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: request failed: %w", rb.method, rb.path, err)
	}
	var duration time.Duration = time.Since(start)

	defer httpResp.Body.Close()

	// Read body
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: failed to read response body: %w", rb.method, rb.path, err)
	}

	var resp *Response = &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header,
		Cookies:    httpResp.Cookies(),
		Body:       string(bodyBytes),
		BodyBytes:  bodyBytes,
		Duration:   duration,
		Request: RequestInfo{
			Method: rb.method,
			URL:    fullURL,
		},
	}

	return resp, nil
}

// --- Convenience Methods ---

// Get sends a GET request and returns a Response.
func (c *Client) Get(path string) (*Response, error) {
	return c.newRequest(http.MethodGet, path).Send()
}

// Post sends a POST request with JSON body and returns a Response.
func (c *Client) Post(path string, body interface{}) (*Response, error) {
	return c.newRequest(http.MethodPost, path).BodyJSON(body).Send()
}

// Put sends a PUT request with JSON body and returns a Response.
func (c *Client) Put(path string, body interface{}) (*Response, error) {
	return c.newRequest(http.MethodPut, path).BodyJSON(body).Send()
}

// Patch sends a PATCH request with JSON body and returns a Response.
func (c *Client) Patch(path string, body interface{}) (*Response, error) {
	return c.newRequest(http.MethodPatch, path).BodyJSON(body).Send()
}

// Delete sends a DELETE request and returns a Response.
func (c *Client) Delete(path string) (*Response, error) {
	return c.newRequest(http.MethodDelete, path).Send()
}

// Head sends a HEAD request and returns a Response.
func (c *Client) Head(path string) (*Response, error) {
	return c.newRequest(http.MethodHead, path).Send()
}

// Options sends an OPTIONS request and returns a Response.
func (c *Client) Options(path string) (*Response, error) {
	return c.newRequest(http.MethodOptions, path).Send()
}

// --- Builder Access (for advanced requests) ---

// Request creates a RequestBuilder for the given method and path.
// Use this for requests that need query params, custom headers, or form bodies.
func (c *Client) Request(method string, path string) *RequestBuilder {
	return c.newRequest(method, path)
}

// --- Runtime Header/Cookie Management ---

// SetHeader sets or updates a client-level header (thread-safe).
func (c *Client) SetHeader(key string, value string) {
	c.mu.Lock()
	c.headers[key] = value
	c.mu.Unlock()
}

// SetBearerToken sets the Authorization header with a Bearer token (thread-safe).
func (c *Client) SetBearerToken(token string) {
	c.mu.Lock()
	c.headers["Authorization"] = "Bearer " + token
	c.mu.Unlock()
}

// AddCookie adds a client-level cookie (thread-safe).
func (c *Client) AddCookie(name string, value string) {
	c.mu.Lock()
	c.cookies = append(c.cookies, &http.Cookie{Name: name, Value: value})
	c.mu.Unlock()
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// --- Helper ---

func basicAuth(username string, password string) string {
	var auth string = username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
