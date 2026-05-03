
package kapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kapi"
)

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
