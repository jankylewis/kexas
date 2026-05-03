
package kapi_test

import (
	"net/http/httptest"
	"testing"

	"github.com/jankylewis/kexas/kapi"
)

// ============================================================
// HTTP Method Tests
// ============================================================

func TestClient_Get(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !resp.BodyContains(`"status":"ok"`) {
		t.Errorf("unexpected body: %s", resp.Body)
	}
}

func TestClient_Post_JSON(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var body map[string]string = map[string]string{
		"name":  "Jane",
		"email": "jane@example.com",
	}
	var resp *kapi.Response
	var err error
	resp, err = client.Post("/users", body)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if !resp.BodyContains("Jane") {
		t.Errorf("expected name in response, body: %s", resp.Body)
	}
}

func TestClient_Put(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Put("/users", map[string]string{"name": "Updated"})
	if err != nil {
		t.Fatalf("PUT failed: %v", err)
	}

	// /users handler returns 200 for non-POST methods
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Delete(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Delete("/delete-me")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestClient_Head(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Head("/health")
	if err != nil {
		t.Fatalf("HEAD failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Options(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var resp *kapi.Response
	var err error
	resp, err = client.Options("/health")
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
