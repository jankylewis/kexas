
package kapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// ============================================================
// Test HTTP Server Helper
// ============================================================

func newTestServer() *httptest.Server {
	var mux *http.ServeMux = http.NewServeMux()
	registerCRUDRoutes(mux)
	registerEchoRoutes(mux)
	registerControlRoutes(mux)
	return httptest.NewServer(mux)
}

// registerCRUDRoutes registers fixed-payload "resource" endpoints used by the
// kapi tests: /health, /users/1, /users (GET+POST), /set-cookie, /delete-me.
func registerCRUDRoutes(mux *http.ServeMux) {
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

	mux.HandleFunc("/set-cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/"})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"cookie set"}`))
	})

	mux.HandleFunc("/delete-me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// registerEchoRoutes registers introspection endpoints that echo back parts of
// the request: headers, cookies, method, query string, and body.
func registerEchoRoutes(mux *http.ServeMux) {
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
}

// registerControlRoutes registers behavior-control endpoints: /status/<code>
// returns the requested HTTP status, and /slow sleeps before responding.
func registerControlRoutes(mux *http.ServeMux) {
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
}
