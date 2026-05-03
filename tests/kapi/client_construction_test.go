
package kapi_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kapi"
)

// ============================================================
// Client Construction Tests
// ============================================================

func TestNewClient_DefaultTimeout(t *testing.T) {
	var client *kapi.Client = kapi.NewClient("https://api.example.com")

	if client.BaseURL() != "https://api.example.com" {
		t.Errorf("expected base URL 'https://api.example.com', got '%s'", client.BaseURL())
	}
}

func TestNewClient_TrailingSlashTrimmed(t *testing.T) {
	var client *kapi.Client = kapi.NewClient("https://api.example.com/")

	if client.BaseURL() != "https://api.example.com" {
		t.Errorf("expected trailing slash trimmed, got '%s'", client.BaseURL())
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	var client *kapi.Client = kapi.NewClient("https://api.example.com",
		kapi.WithTimeout(5*time.Second),
	)

	if client.BaseURL() != "https://api.example.com" {
		t.Errorf("unexpected base URL: %s", client.BaseURL())
	}
}

func TestNewClient_WithHeader(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithHeader("X-Custom", "test-value"),
	)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.BodyContains("test-value") {
		t.Errorf("expected header to be echoed, body: %s", resp.Body)
	}
}

func TestNewClient_WithHeaders(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithHeaders(map[string]string{
			"X-One": "1",
			"X-Two": "2",
		}),
	)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.BodyContains("1") || !resp.BodyContains("2") {
		t.Errorf("expected both headers, body: %s", resp.Body)
	}
}

func TestNewClient_WithBearerToken(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithBearerToken("mytoken123"),
	)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.BodyContains("Bearer mytoken123") {
		t.Errorf("expected Bearer token in headers, body: %s", resp.Body)
	}
}

func TestNewClient_WithBasicAuth(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithBasicAuth("user", "pass"),
	)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.BodyContains("Basic") {
		t.Errorf("expected Basic auth in headers, body: %s", resp.Body)
	}
}

func TestNewClient_WithCookie(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithCookie("session", "abc123"),
	)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-cookies")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.BodyContains("abc123") {
		t.Errorf("expected cookie to be echoed, body: %s", resp.Body)
	}
}
