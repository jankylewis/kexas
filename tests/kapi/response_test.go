
package kapi_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jankylewis/kexas/kapi"
)

// ============================================================
// Response Tests
// ============================================================

func TestResponse_IsOK(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !resp.IsOK() {
		t.Error("expected IsOK true for 200")
	}
}

func TestResponse_IsClientError(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/status/404")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !resp.IsClientError() {
		t.Errorf("expected IsClientError true for 404, got status %d", resp.StatusCode)
	}
}

func TestResponse_IsServerError(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/status/500")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !resp.IsServerError() {
		t.Errorf("expected IsServerError true for 500, got status %d", resp.StatusCode)
	}
}

func TestResponse_JSON_Struct(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/users/1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var user struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	err = resp.JSON(&user)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected id 1, got %d", user.ID)
	}
	if user.Name != "John" {
		t.Errorf("expected name John, got %s", user.Name)
	}
}

func TestResponse_JSONMap(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/users/1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var data map[string]interface{}
	data, err = resp.JSONMap()
	if err != nil {
		t.Fatalf("JSONMap failed: %v", err)
	}
	if data["name"] != "John" {
		t.Errorf("expected name John, got %v", data["name"])
	}
}

func TestResponse_JSONArray(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/users")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var data []interface{}
	data, err = resp.JSONArray()
	if err != nil {
		t.Fatalf("JSONArray failed: %v", err)
	}
	if len(data) != 2 {
		t.Errorf("expected 2 users, got %d", len(data))
	}
}

func TestResponse_HeaderValue(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.HeaderValue("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", resp.HeaderValue("Content-Type"))
	}
}

func TestResponse_ContentType(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.ContentType() != "application/json" {
		t.Errorf("expected application/json, got %s", resp.ContentType())
	}
}

func TestResponse_SetCookie(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/set-cookie")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.HasCookie("session") {
		t.Error("expected response to have 'session' cookie")
	}
	if resp.CookieValue("session") != "abc123" {
		t.Errorf("expected cookie value abc123, got %s", resp.CookieValue("session"))
	}
}

func TestResponse_CookieValue_NotFound(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.CookieValue("nonexistent") != "" {
		t.Error("expected empty string for missing cookie")
	}
	if resp.HasCookie("nonexistent") {
		t.Error("expected HasCookie false for missing cookie")
	}
}

func TestResponse_BodyContains(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("ok") {
		t.Error("expected body to contain 'ok'")
	}
	if resp.BodyContains("error") {
		t.Error("expected body to NOT contain 'error'")
	}
}

func TestResponse_String(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var str string = resp.String()
	if !strings.Contains(str, "GET") {
		t.Errorf("expected String to contain GET, got: %s", str)
	}
	if !strings.Contains(str, "200") {
		t.Errorf("expected String to contain 200, got: %s", str)
	}
}

func TestResponse_Duration(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.Duration <= 0 {
		t.Error("expected positive duration")
	}
}
