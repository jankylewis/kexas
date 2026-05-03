# KAPI — HTTP API Client Overview

**Files:** `kexas/kapi/client.go`, `kexas/kapi/response.go`
**Package:** `github.com/jankylewis/kexas/kapi`
**Status:** Implemented

---

## What It Does

KAPI is an HTTP API client for making REST requests outside of the browser context. It is the kexas equivalent of:
- **Playwright** `APIRequestContext`
- **RestSharp** (C#)
- **RestAssured** (Java)

It enables API-level testing alongside browser-level testing within the same test suite.

## API Surface

### Client Construction

| Function | Description |
|----------|-------------|
| `kapi.NewClient(baseURL, ...opts)` | Create a new API client with base URL and functional options |

### Functional Options

| Option | Description |
|--------|-------------|
| `WithTimeout(d)` | Set HTTP client timeout |
| `WithHeader(k, v)` | Set a default header |
| `WithHeaders(map)` | Set multiple default headers |
| `WithBearerToken(token)` | Set Authorization: Bearer header |
| `WithBasicAuth(user, pass)` | Set Authorization: Basic header |
| `WithCookie(name, value)` | Add a default cookie |
| `WithHTTPClient(client)` | Use a custom http.Client |
| `WithNoRedirect()` | Disable automatic redirect following |

### Convenience Methods (one-liner requests)

| Method | Description |
|--------|-------------|
| `client.Get(path)` | GET request |
| `client.Post(path, body)` | POST with JSON body |
| `client.Put(path, body)` | PUT with JSON body |
| `client.Patch(path, body)` | PATCH with JSON body |
| `client.Delete(path)` | DELETE request |
| `client.Head(path)` | HEAD request |
| `client.Options(path)` | OPTIONS request |

### RequestBuilder (advanced requests)

| Method | Description |
|--------|-------------|
| `client.Request(method, path)` | Create a RequestBuilder |
| `.Header(k, v)` | Set request-level header |
| `.Headers(map)` | Set multiple request-level headers |
| `.Query(k, v)` | Add query parameter |
| `.QueryParams(map)` | Add multiple query parameters |
| `.Body([]byte)` | Set raw body |
| `.BodyString(s)` | Set string body |
| `.BodyJSON(obj)` | Set JSON body from struct/map |
| `.BodyForm(map)` | Set URL-encoded form body |
| `.Cookie(name, value)` | Add request-level cookie |
| `.Send()` | Execute and return `*Response` |

### Response

| Method | Description |
|--------|-------------|
| `resp.JSON(&target)` | Unmarshal body into struct |
| `resp.JSONMap()` | Parse as `map[string]interface{}` |
| `resp.JSONArray()` | Parse as `[]interface{}` |
| `resp.IsOK()` | Status 2xx? |
| `resp.IsClientError()` | Status 4xx? |
| `resp.IsServerError()` | Status 5xx? |
| `resp.HeaderValue(key)` | Get response header |
| `resp.ContentType()` | Get Content-Type header |
| `resp.CookieValue(name)` | Get response cookie value |
| `resp.HasCookie(name)` | Check response cookie exists |
| `resp.BodyContains(substr)` | Check body contains string |
| `resp.String()` | Human-readable summary |

### Runtime Mutation (thread-safe)

| Method | Description |
|--------|-------------|
| `client.SetHeader(k, v)` | Update client header at runtime |
| `client.SetBearerToken(token)` | Update bearer token at runtime |
| `client.AddCookie(name, value)` | Add cookie at runtime |

## Key Design Decisions

1. **Functional options pattern** — Clean, extensible construction without bloated constructors
2. **Request-level overrides** — Per-request headers/cookies override client defaults
3. **Mutex-protected shared state** — Client headers and cookies are safe for concurrent access
4. **Each client is isolated** — Separate clients have separate state; parallel test workers each get their own client
5. **No external dependencies** — Built on Go's `net/http` stdlib only
6. **Response reads body eagerly** — Body is fully read and stored as `string` + `[]byte` before returning
