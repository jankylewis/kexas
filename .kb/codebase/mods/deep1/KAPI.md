# KAPI — HTTP API Client Deep Technical Guide

**Files:** `kexas/kapi/client.go`, `kexas/kapi/response.go`

---

## Architecture

KAPI is a standalone HTTP API client with no dependency on the kexas browser primitives. It uses only Go's `net/http` stdlib.

### Core Types

```
Client → owns: baseURL, headers map, cookies slice, *http.Client, sync.Mutex
RequestBuilder → owns: method, path, headers, queryParams, body, contentType, cookies
Response → owns: StatusCode, Headers, Cookies, Body, BodyBytes, Duration, RequestInfo
```

### Request Flow

```
client.Get("/users/1")
  → newRequest("GET", "/users/1")  → creates RequestBuilder
  → Send()
    → build full URL: baseURL + path + queryParams
    → http.NewRequest(method, url, body)
    → apply client headers (under mutex)
    → apply request headers (override)
    → apply content type
    → apply client cookies (under mutex)
    → apply request cookies
    → httpClient.Do(req)  → actual HTTP call
    → io.ReadAll(resp.Body)
    → build Response struct
```

### Thread Safety Model

The `Client` struct has shared mutable state (headers map, cookies slice) that can be modified at runtime via `SetHeader()`, `SetBearerToken()`, `AddCookie()`. All reads and writes to these fields are protected by `sync.Mutex`.

**Critical:** The mutex is locked only for the duration of copying headers/cookies to the request — NOT for the entire HTTP round-trip. This prevents one slow request from blocking all other goroutines.

```go
// Lock only to read shared state
c.mu.Lock()
for k, v := range c.headers { req.Header.Set(k, v) }
c.mu.Unlock()
// Unlock before the actual HTTP call
httpResp, err = c.httpClient.Do(req)  // no lock held
```

### Functional Options Pattern

Instead of a config struct with many fields, we use functions:

```go
var client = kapi.NewClient(url,
    kapi.WithTimeout(5*time.Second),
    kapi.WithBearerToken("token"),
)
```

Each option is a `func(*Client)` that mutates the client during construction. This is Go's idiomatic builder pattern — extensible without breaking API compatibility.

### Response Eager-Read

The response body is fully read into memory (`io.ReadAll`) before returning. This means:
- The `http.Response.Body` is closed immediately (no resource leak)
- `resp.Body` (string) and `resp.BodyBytes` ([]byte) are always available
- JSON parsing can be called multiple times without re-reading

Trade-off: Large response bodies (>10MB) consume proportional memory. For API testing, response bodies are typically small (<1MB).

## Pitfalls

1. **Base URL trailing slash** — `NewClient` trims trailing slashes. Path must start with `/`.
2. **BodyJSON nil body** — If `json.Marshal` fails, body is nil and Send() will send an empty body.
3. **Header case** — `http.Header` canonicalizes keys (e.g., `x-custom` → `X-Custom`). The echo server returns canonical keys.
4. **Cookie jar** — This client does NOT use `http.CookieJar`. Cookies are managed manually. Response cookies are NOT auto-sent on subsequent requests.
