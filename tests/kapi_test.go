
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kexas-project/kexas/kapi"
)

// ============================================================
// Test HTTP Server Helper
// ============================================================

func newTestServer() *httptest.Server {
	var mux *http.ServeMux = http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/users/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"John","email":"john@example.com"}`))
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			var resp []byte
			resp, _ = json.Marshal(map[string]interface{}{
				"id":    2,
				"name":  body["name"],
				"email": body["email"],
			})
			w.Write(resp)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`))
	})

	mux.HandleFunc("/echo-headers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var headers map[string]string = make(map[string]string)
		for key := range r.Header {
			headers[key] = r.Header.Get(key)
		}
		var resp []byte
		resp, _ = json.Marshal(headers)
		w.Write(resp)
	})

	mux.HandleFunc("/echo-cookies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cookies map[string]string = make(map[string]string)
		for _, c := range r.Cookies() {
			cookies[c.Name] = c.Value
		}
		var resp []byte
		resp, _ = json.Marshal(cookies)
		w.Write(resp)
	})

	mux.HandleFunc("/set-cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/"})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"cookie set"}`))
	})

	mux.HandleFunc("/echo-method", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"method":"%s"}`, r.Method)))
	})

	mux.HandleFunc("/echo-query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var params map[string]string = make(map[string]string)
		for key := range r.URL.Query() {
			params[key] = r.URL.Query().Get(key)
		}
		var resp []byte
		resp, _ = json.Marshal(params)
		w.Write(resp)
	})

	mux.HandleFunc("/echo-body", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		var buf []byte = make([]byte, 4096)
		var n int
		n, _ = r.Body.Read(buf)
		w.Write(buf[:n])
	})

	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		var parts []string = strings.Split(r.URL.Path, "/")
		var code int = 200
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &code)
		}
		w.WriteHeader(code)
		w.Write([]byte(fmt.Sprintf(`{"code":%d}`, code)))
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"slow":true}`))
	})

	mux.HandleFunc("/delete-me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return httptest.NewServer(mux)
}

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

// ============================================================
// Runtime Mutation Tests
// ============================================================

func TestClient_SetHeader_Runtime(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	client.SetHeader("X-Runtime", "dynamic")

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("dynamic") {
		t.Errorf("expected runtime header, body: %s", resp.Body)
	}
}

func TestClient_SetBearerToken_Runtime(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	client.SetBearerToken("refreshed-token")

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-headers")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("Bearer refreshed-token") {
		t.Errorf("expected refreshed bearer token, body: %s", resp.Body)
	}
}

func TestClient_AddCookie_Runtime(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	client.AddCookie("runtime-cookie", "value123")

	var resp *kapi.Response
	var err error
	resp, err = client.Get("/echo-cookies")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !resp.BodyContains("value123") {
		t.Errorf("expected runtime cookie, body: %s", resp.Body)
	}
}

// ============================================================
// Error Handling Tests
// ============================================================

func TestClient_Timeout(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithTimeout(500*time.Millisecond),
	)

	var _, err error
	_, err = client.Get("/slow")
	if err == nil {
		t.Error("expected timeout error for slow endpoint")
	}
}

func TestClient_InvalidURL(t *testing.T) {
	var client *kapi.Client = kapi.NewClient("http://127.0.0.1:1")

	var _, err error
	_, err = client.Get("/health")
	if err == nil {
		t.Error("expected connection error for invalid URL")
	}
}

// ============================================================
// Parallel Safety Tests
// ============================================================

func TestClient_ParallelRequests(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var wg sync.WaitGroup
	var errCount int
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var resp *kapi.Response
			var err error
			resp, err = client.Get("/health")
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}
			if resp.StatusCode != 200 {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if errCount > 0 {
		t.Errorf("parallel requests had %d errors", errCount)
	}
}

func TestClient_ParallelIsolation(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	// Two separate clients — fully isolated
	var client1 *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithHeader("X-Client", "one"),
	)
	var client2 *kapi.Client = kapi.NewClient(server.URL,
		kapi.WithHeader("X-Client", "two"),
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		var resp *kapi.Response
		var err error
		resp, err = client1.Get("/echo-headers")
		if err != nil {
			t.Errorf("client1 error: %v", err)
			return
		}
		if !resp.BodyContains("one") {
			t.Errorf("client1 should see 'one', got: %s", resp.Body)
		}
	}()

	go func() {
		defer wg.Done()
		var resp *kapi.Response
		var err error
		resp, err = client2.Get("/echo-headers")
		if err != nil {
			t.Errorf("client2 error: %v", err)
			return
		}
		if !resp.BodyContains("two") {
			t.Errorf("client2 should see 'two', got: %s", resp.Body)
		}
	}()

	wg.Wait()
}

func TestClient_SetHeader_ParallelSafe(t *testing.T) {
	var server *httptest.Server = newTestServer()
	defer server.Close()

	var client *kapi.Client = kapi.NewClient(server.URL)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client.SetHeader("X-Worker", fmt.Sprintf("worker-%d", idx))
			client.Get("/health")
		}(i)
	}
	wg.Wait()
	// If this doesn't panic or race-detect, the mutex is working
}

// ============================================================
// NoRedirect Tests
// ============================================================

func TestClient_WithNoRedirect(t *testing.T) {
	var redirectServer *httptest.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/destination", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("destination"))
	}))
	defer redirectServer.Close()

	var client *kapi.Client = kapi.NewClient(redirectServer.URL, kapi.WithNoRedirect())
	var resp *kapi.Response
	var err error
	resp, err = client.Get("/redirect")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 302 {
		t.Errorf("expected 302 (no redirect follow), got %d", resp.StatusCode)
	}
}
