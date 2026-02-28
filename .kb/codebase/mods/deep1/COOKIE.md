# Cookie Management — Deep Technical Guide (Student Level)

**File:** `kexas/cookie.go`

---

## What Are Cookies?

Cookies are small key-value pairs that websites store in your browser. When you log into a website, the server sends back a cookie like `session_id=abc123`. Your browser stores it and sends it back with every subsequent request to that domain. This is how websites "remember" you're logged in.

### Cookie Fields Explained

| Field | What It Means |
|-------|---------------|
| **Name** | The cookie's identifier (e.g., `session_id`) |
| **Value** | The data stored (e.g., `abc123`) |
| **Domain** | Which website this cookie belongs to (e.g., `.example.com`) |
| **Path** | Which URL paths can access this cookie (e.g., `/` = all paths) |
| **Expires** | When the cookie expires (Unix timestamp as float64; 0 = session cookie, deleted when browser closes) |
| **HttpOnly** | If true, JavaScript on the page CANNOT read this cookie — only the server can. This prevents XSS attacks from stealing session tokens. |
| **Secure** | If true, cookie is only sent over HTTPS connections, never plain HTTP. |
| **SameSite** | Controls when cookie is sent with cross-site requests. `Strict` = never cross-site. `Lax` = sent on top-level navigations. `None` = always sent (requires Secure=true). |
| **Size** | Total byte size of name+value. Read-only from CDP. |

### FAQ: Why would a test need to set cookies?

**To skip login flows.** Instead of filling in username/password and clicking submit on every test, you can set the authentication cookie directly and navigate to the protected page. This saves 2-5 seconds per test.

---

## How It Works Under the Hood

### CDP Network Domain

Cookies are managed through Chrome's **Network** CDP domain. Before sending any cookie command, the kexas Agent Manager automatically enables the Network domain via `ensureAgentsForCommand()`.

**What is a CDP domain?** Chrome DevTools Protocol groups related commands into "domains". The Network domain handles everything about HTTP requests, responses, and cookies. Some domains must be explicitly "enabled" before use — the Agent Manager handles this transparently.

### The parseCookie Function

CDP returns cookies as raw `map[string]interface{}` (Go's way of representing arbitrary JSON). The `parseCookie` function converts this messy map into a clean, typed `Cookie` struct:

```go
// CDP returns: {"name": "session", "value": "abc", "httpOnly": true, "expires": 1893456000.0}
// parseCookie converts to: Cookie{Name: "session", Value: "abc", HttpOnly: true, Expires: 1893456000}
```

**Why float64?** JSON numbers are always decoded as `float64` in Go's `encoding/json` and in CDP responses. That's why `Expires` is `float64` and `Size` is converted from `float64` to `int`.

### Error Handling Strategy

- **Empty name → immediate error** — Returns `ErrCookieNameEmpty` sentinel error before sending any CDP command
- **CDP failure → wrapped error** — `fmt.Errorf("set cookie failed for '%s': %w", name, err)` preserves the chain
- **Chrome rejection → ErrCookieSetFailed** — If CDP returns `{success: false}`, the cookie wasn't set (usually domain mismatch)
- **Cookie not found → ErrCookieNotFound** — `GetCookieValue` wraps this for clear error messages

### FAQ: What is a sentinel error?

A **sentinel error** is a pre-defined error variable like `var ErrCookieNotFound = fmt.Errorf("cookie not found")`. You compare against it using `errors.Is(err, ErrCookieNotFound)`. This pattern avoids fragile string matching and lets callers handle specific error types programmatically.

---

## Thread Safety

Cookie operations go through `page.sendCommand()`, which sends CDP commands over the WebSocket. The WebSocket client handles serialization internally. However, if two goroutines set cookies on the **same page** simultaneously, the results are non-deterministic (last write wins). In practice, each parallel test worker has its own browser and page, so this isn't an issue.

### FAQ: Can two tests share cookies?

Cookies are browser-level, not page-level. If you have two tabs in the same browser, they share cookies for the same domain. This matches real browser behavior. In kexas parallel testing, each worker gets its own browser, so cookies are isolated between workers.

---

## Pitfalls to Avoid

1. **Domain mismatch** — `Network.setCookie` may silently fail if the page hasn't navigated to a URL in the cookie's domain. Always navigate first, then set cookies.
2. **Session vs persistent cookies** — If `Expires` is 0 or omitted, the cookie is a "session cookie" that disappears when the browser closes. Set an Expires value for persistent cookies.
3. **SameSite=None requires Secure=true** — Chrome rejects `SameSite=None` cookies without `Secure=true`.
4. **HttpOnly cookies can't be read by JS** — `document.cookie` won't show HttpOnly cookies. Use `page.GetCookies()` (which uses CDP, not JS) to read them.
