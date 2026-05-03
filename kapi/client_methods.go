package kapi

import (
	"encoding/base64"
	"net/http"
)

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

// Trace sends a TRACE request and returns a Response. RFC 9110 §9.3.8 — used
// to echo the received request for debugging proxies / loop detection.
// Most servers either disable TRACE for security or return the request as-is.
func (c *Client) Trace(path string) (*Response, error) {
	return c.newRequest(http.MethodTrace, path).Send()
}

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

// basicAuth encodes "username:password" as base64 for the Basic auth header.
func basicAuth(username string, password string) string {
	var auth string = username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
