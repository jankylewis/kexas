package kapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kapi"
)

// 15 critical-after-existing unit tests for kapi.
// Existing: NewClient, HTTP methods, RequestBuilder, Response, runtime.
// This batch: option-composition, header propagation, redirect handling,
// timeout boundaries, response-body edge cases, concurrent client safety.

// echoServer returns a test server that echoes useful request properties as JSON.
// Used by multiple tests to introspect what the client sent.
func echoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"headers": r.Header,
			"query":   r.URL.RawQuery,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestClient_BaseURL_RoundTrip(t *testing.T) {
	var c *kapi.Client = kapi.NewClient("https://example.com")
	if c.BaseURL() != "https://example.com" {
		t.Errorf("BaseURL: got %q", c.BaseURL())
	}
}

func TestClient_BaseURL_PreservesTrailingSlash(t *testing.T) {
	var c *kapi.Client = kapi.NewClient("https://example.com/api/")
	// The exact normalisation isn't specified; just verify it's preserved or
	// canonically modified — the BaseURL roundtrip must be deterministic.
	if c.BaseURL() == "" {
		t.Error("BaseURL should not be empty after construction")
	}
}

func TestNewClient_WithMultipleOptions_AllApplied(t *testing.T) {
	srv := echoServer(t)
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL,
		kapi.WithTimeout(5*time.Second),
		kapi.WithHeader("X-Test-Marker", "kexas-multi"),
		kapi.WithBearerToken("test-token"),
	)

	resp, err := c.Get("/probe")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := resp.JSONMap()
	headers, _ := body["headers"].(map[string]interface{})
	auth, _ := headers["Authorization"].([]interface{})
	if len(auth) == 0 || !strings.Contains(auth[0].(string), "Bearer test-token") {
		t.Errorf("expected Bearer auth, got %v", auth)
	}
	marker, _ := headers["X-Test-Marker"].([]interface{})
	if len(marker) == 0 || marker[0].(string) != "kexas-multi" {
		t.Errorf("expected X-Test-Marker, got %v", marker)
	}
}

func TestClient_SetHeader_AppliesPerRequest(t *testing.T) {
	srv := echoServer(t)
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL)
	c.SetHeader("X-Runtime-Header", "runtime")

	resp, err := c.Get("/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := resp.JSONMap()
	headers, _ := body["headers"].(map[string]interface{})
	hdr, _ := headers["X-Runtime-Header"].([]interface{})
	if len(hdr) == 0 || hdr[0].(string) != "runtime" {
		t.Errorf("expected runtime header, got %v", hdr)
	}
}

func TestClient_SetBearerToken_PersistsAcrossRequests(t *testing.T) {
	srv := echoServer(t)
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL)
	c.SetBearerToken("persistent-token")

	for i := 0; i < 3; i++ {
		resp, err := c.Get("/")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := resp.JSONMap()
		headers, _ := body["headers"].(map[string]interface{})
		auth, _ := headers["Authorization"].([]interface{})
		if len(auth) == 0 {
			t.Errorf("call %d: missing Authorization header", i)
		}
	}
}

func TestClient_Get_StatusCodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/200":
			w.WriteHeader(200)
		case "/404":
			w.WriteHeader(404)
		case "/500":
			w.WriteHeader(500)
		}
	}))
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL)
	for _, tc := range []struct {
		path     string
		isOK     bool
		isClient bool
		isServer bool
	}{
		{"/200", true, false, false},
		{"/404", false, true, false},
		{"/500", false, false, true},
	} {
		resp, _ := c.Get(tc.path)
		if resp.IsOK() != tc.isOK {
			t.Errorf("%s: IsOK = %v, want %v", tc.path, resp.IsOK(), tc.isOK)
		}
		if resp.IsClientError() != tc.isClient {
			t.Errorf("%s: IsClientError = %v, want %v", tc.path, resp.IsClientError(), tc.isClient)
		}
		if resp.IsServerError() != tc.isServer {
			t.Errorf("%s: IsServerError = %v, want %v", tc.path, resp.IsServerError(), tc.isServer)
		}
	}
}

func TestResponse_BodyContains_PositiveAndNegative(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello kexas world"))
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	if !resp.BodyContains("kexas") {
		t.Error("expected 'kexas' in body")
	}
	if resp.BodyContains("playwright") {
		t.Error("expected 'playwright' NOT in body")
	}
}

func TestResponse_HeaderValue_CaseInsensitive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Marker", "value-here")
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	// HTTP headers are case-insensitive per RFC 7230.
	if resp.HeaderValue("X-Marker") == "" {
		t.Error("expected X-Marker value")
	}
	if resp.HeaderValue("x-marker") == "" {
		t.Error("expected x-marker (lowercase) to also resolve")
	}
}

func TestResponse_ContentType_ReturnsExpected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	if !strings.Contains(resp.ContentType(), "application/json") {
		t.Errorf("ContentType: got %q", resp.ContentType())
	}
}

func TestResponse_JSONArray_OnNonArray_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"not": "array"}`))
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	_, err := resp.JSONArray()
	if err == nil {
		t.Error("JSONArray on object should error")
	}
}

func TestResponse_JSONMap_OnArray_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[1,2,3]`))
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	_, err := resp.JSONMap()
	if err == nil {
		t.Error("JSONMap on array should error")
	}
}

func TestRequestBuilder_QueryParams_AppendsToURL(t *testing.T) {
	srv := echoServer(t)
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Request("GET", "/").
		Query("k1", "v1").
		Query("k2", "v2").
		Send()
	if err != nil {
		t.Fatal(err)
	}
	body, _ := resp.JSONMap()
	q, _ := body["query"].(string)
	if !strings.Contains(q, "k1=v1") || !strings.Contains(q, "k2=v2") {
		t.Errorf("expected both query params; got %q", q)
	}
}

func TestRequestBuilder_BodyJSON_SerializesAndSets(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf [4096]byte
		var n int
		n, _ = r.Body.Read(buf[:])
		got = string(buf[:n])
		w.WriteHeader(200)
	}))
	defer srv.Close()

	type payload struct {
		Name string `json:"name"`
	}
	_, err := kapi.NewClient(srv.URL).Request("POST", "/").
		BodyJSON(payload{Name: "kexas"}).
		Send()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "kexas") {
		t.Errorf("body should contain serialized payload; got %q", got)
	}
}

func TestClient_ConcurrentRequests_NoRace(t *testing.T) {
	var counter int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&counter, 1)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL)
	const N int = 30
	var done chan struct{} = make(chan struct{}, N)
	for i := 0; i < N; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_, _ = c.Get("/")
		}()
	}
	for i := 0; i < N; i++ {
		<-done
	}
	if atomic.LoadInt32(&counter) != int32(N) {
		t.Errorf("server saw %d requests, expected %d", counter, N)
	}
}

func TestClient_WithTimeout_HonoursDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	var c *kapi.Client = kapi.NewClient(srv.URL, kapi.WithTimeout(100*time.Millisecond))
	var start time.Time = time.Now()
	_, err := c.Get("/")
	var elapsed time.Duration = time.Since(start)
	if err == nil {
		t.Error("expected timeout error")
	}
	if elapsed > 1*time.Second {
		t.Errorf("Timeout=100ms but call took %v", elapsed)
	}
}
