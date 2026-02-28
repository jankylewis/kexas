package kapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Response holds the result of an HTTP API request.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Cookies    []*http.Cookie
	Body       string
	BodyBytes  []byte
	Duration   time.Duration
	Request    RequestInfo
}

// RequestInfo holds metadata about the request that produced this response.
type RequestInfo struct {
	Method string
	URL    string
}

// JSON unmarshals the response body into the given target struct.
func (r *Response) JSON(target interface{}) error {
	var err error = json.Unmarshal(r.BodyBytes, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body as JSON: %w", err)
	}
	return nil
}

// JSONMap parses the response body as a JSON object (map[string]interface{}).
func (r *Response) JSONMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	var err error = json.Unmarshal(r.BodyBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body as JSON map: %w", err)
	}
	return result, nil
}

// JSONArray parses the response body as a JSON array ([]interface{}).
func (r *Response) JSONArray() ([]interface{}, error) {
	var result []interface{}
	var err error = json.Unmarshal(r.BodyBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body as JSON array: %w", err)
	}
	return result, nil
}

// IsOK returns true if the status code is 2xx.
func (r *Response) IsOK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true if the status code is 4xx.
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the status code is 5xx.
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// HeaderValue returns the value of the given response header.
func (r *Response) HeaderValue(key string) string {
	return r.Headers.Get(key)
}

// ContentType returns the Content-Type header value.
func (r *Response) ContentType() string {
	return r.Headers.Get("Content-Type")
}

// CookieValue returns the value of a response cookie by name.
// Returns empty string if cookie not found.
func (r *Response) CookieValue(name string) string {
	for _, cookie := range r.Cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// HasCookie returns true if the response contains a cookie with the given name.
func (r *Response) HasCookie(name string) bool {
	for _, cookie := range r.Cookies {
		if cookie.Name == name {
			return true
		}
	}
	return false
}

// BodyContains returns true if the response body contains the given substring.
func (r *Response) BodyContains(substr string) bool {
	return strings.Contains(r.Body, substr)
}

// String returns a human-readable summary of the response.
func (r *Response) String() string {
	return fmt.Sprintf("%s %s → %d %s (%v, %d bytes)",
		r.Request.Method, r.Request.URL,
		r.StatusCode, http.StatusText(r.StatusCode),
		r.Duration.Round(time.Millisecond), len(r.BodyBytes))
}
