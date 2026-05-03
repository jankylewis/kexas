package kapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
// Marshal failures store a nil body; the resulting Send will fail with a clear error.
func (rb *RequestBuilder) BodyJSON(data interface{}) *RequestBuilder {
	var jsonBytes []byte
	var err error
	jsonBytes, err = json.Marshal(data)
	if err != nil {
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

	rb.applyHeaders(req)
	rb.applyCookies(req)

	var httpResp *http.Response
	var bodyBytes []byte
	var duration time.Duration
	httpResp, bodyBytes, duration, err = rb.executeAndReadBody(req)
	if err != nil {
		return nil, err
	}

	return &Response{
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
	}, nil
}

// applyHeaders sets client-level headers first, then request-level headers (which
// override the client defaults), then Content-Type if specified.
func (rb *RequestBuilder) applyHeaders(req *http.Request) {
	rb.client.mu.Lock()
	for k, v := range rb.client.headers {
		req.Header.Set(k, v)
	}
	rb.client.mu.Unlock()

	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	if rb.contentType != "" {
		req.Header.Set("Content-Type", rb.contentType)
	}
}

// applyCookies attaches all client-level cookies, then all request-level cookies.
func (rb *RequestBuilder) applyCookies(req *http.Request) {
	rb.client.mu.Lock()
	for _, cookie := range rb.client.cookies {
		req.AddCookie(cookie)
	}
	rb.client.mu.Unlock()

	for _, cookie := range rb.cookies {
		req.AddCookie(cookie)
	}
}

// executeAndReadBody runs the HTTP request, drains the body into a byte slice,
// and returns it together with the elapsed wall-clock duration. The response body
// is closed before this helper returns; callers must not read from httpResp.Body.
func (rb *RequestBuilder) executeAndReadBody(req *http.Request) (*http.Response, []byte, time.Duration, error) {
	var start time.Time = time.Now()
	var httpResp *http.Response
	var err error
	httpResp, err = rb.client.httpClient.Do(req)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("%s %s: request failed: %w", rb.method, rb.path, err)
	}
	var duration time.Duration = time.Since(start)
	defer httpResp.Body.Close()

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("%s %s: failed to read response body: %w", rb.method, rb.path, err)
	}
	return httpResp, bodyBytes, duration, nil
}
