
package kapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jankylewis/kexas/kapi"
)

// ============================================================
// RequestBuilder Tests
// ============================================================

func TestRequestBuilder_Query(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodGet, "/echo-query").
		Query("page", "2").
		Query("limit", "10").
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains(`"page":"2"`) {
		t.Errorf("expected page param, body: %s", resp.Body)
	}
	if !resp.BodyContains(`"limit":"10"`) {
		t.Errorf("expected limit param, body: %s", resp.Body)
	}
}

func TestRequestBuilder_QueryParams(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodGet, "/echo-query").
		QueryParams(map[string]string{"foo": "bar", "baz": "qux"}).
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("bar") || !resp.BodyContains("qux") {
		t.Errorf("expected query params, body: %s", resp.Body)
	}
}

func TestRequestBuilder_Header(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodGet, "/echo-headers").
		Header("X-Request-Level", "per-request").
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("per-request") {
		t.Errorf("expected request-level header, body: %s", resp.Body)
	}
}

func TestRequestBuilder_HeaderOverridesClient(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithHeader("X-Level", "client"),
	)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodGet, "/echo-headers").
		Header("X-Level", "request").
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("request") {
		t.Errorf("expected request-level to override client-level, body: %s", resp.Body)
	}
}

func TestRequestBuilder_BodyString(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodPost, "/echo-body").
		Header("Content-Type", "text/plain").
		BodyString("hello world").
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.Body != "hello world" {
		t.Errorf("expected echoed body, got: %s", resp.Body)
	}
}

func TestRequestBuilder_BodyJSON(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodPost, "/users").
		BodyJSON(map[string]string{"name": "Test", "email": "test@test.com"}).
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestRequestBuilder_BodyForm(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodPost, "/echo-body").
		BodyForm(map[string]string{"username": "admin", "password": "secret"}).
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("username=admin") {
		t.Errorf("expected form data, body: %s", resp.Body)
	}
}

func TestRequestBuilder_Cookie(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Request(http.MethodGet, "/echo-cookies").
		Cookie("token", "xyz789").
		Send()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("xyz789") {
		t.Errorf("expected cookie, body: %s", resp.Body)
	}
}
