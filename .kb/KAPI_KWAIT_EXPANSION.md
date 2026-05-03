# kapi + kwait expansion — 2026-05-03

User asked to expand kapi to RestSharp/RestAssured/Axios parity and kwait to Playwright/Rod/Selenium/Cypress parity, then exercise everything in samplings.

## kapi additions

### `Client.Trace(path)` (RFC 9110 §9.3.8)

Completes the standard HTTP method set: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, **TRACE**. CONNECT is intentionally omitted — it's a proxy-tunnel verb that doesn't fit a typed test client. Custom verbs remain available via `Client.Request(method, path).Send()` (used in samplings/apitests for PROPFIND).

### `RequestBuilder.BodyMultipart(fields, files)`

```go
type MultipartFile struct {
    FieldName, Filename, ContentType string
    Content                           []byte
}

c.Request("POST", "/upload").
    BodyMultipart(
        map[string]string{"caption": "kexas-test"},
        []kapi.MultipartFile{{FieldName: "file", Filename: "a.txt", Content: []byte("data")}},
    ).Send()
```

Mirrors RestSharp's `AddFile`, RestAssured's `multiPart()`, Axios's `FormData`. Negotiates boundary + sets Content-Type automatically. Default `application/octet-stream` for files without an explicit ContentType.

### `Response.SaveToFile(path)`

```go
resp, _ := c.Get("/image/png")
_ = resp.SaveToFile("/tmp/downloaded.png")
```

Common need (RestSharp's `DownloadData` + write step, Axios's `responseType: 'stream'` + pipe). Creates parent dirs with `os.MkdirAll`. Overwrites existing files.

### Tests added

- **12 new unit tests** in `tests/kapi/kapi_new_methods_test.go` — TRACE roundtrip, multipart serialisation (single file, multiple files, default content-type, empty), SaveToFile (basic, mkdir-p, binary, overwrite, bad-path).
- **New samplings project: `samplings/apitests/`** against `httpbin.org` — 11 tests, one per HTTP method + multipart upload + PNG download. Real-network round-trips for the full surface.

## kwait additions

Five new advanced waiters in `kcore/page_wait_advanced.go`. Each follows the same shape: 100ms timeout floor, nil-page guard, polling at 100ms interval.

| Waiter | Maps to | Use case |
|---|---|---|
| `Page.WaitForElementHidden(sel, timeout)` | Playwright `state: 'hidden'` / Selenium `invisibilityOfElementLocated` / Rod `WaitInvisible` | Element exists but display:none / visibility:hidden / opacity:0 / zero-size |
| `Page.WaitForElementDetached(sel, timeout)` | Playwright `state: 'detached'` | Element fully removed from DOM (stricter than Hidden) |
| `Page.WaitForFunction(jsExpr, timeout)` | Playwright `page.waitForFunction()` | Custom JS predicate — escape hatch when no built-in matcher fits |
| `Page.WaitForTitleContains(substr, timeout)` | Selenium `titleContains` | Title-change assertion (post-navigation, post-update) |
| `Page.WaitForElementText(sel, substr, timeout)` | Selenium `textToBePresentInElement` / Cypress `should('contain.text', ...)` | Element exists AND text contains substring |

Shared helper: `pollUntilJSTrue(jsExpr, timeout, msg)` + `jsResultIsTruthy(result)` — both used by Hidden / Detached / Function / ElementText.

### Tests added

- **15 new unit tests** in `tests/kcore/page_wait_advanced_test.go` — nil-page guards + timeout-validation for every new waiter (3 tests × 5 waiters).
- **5 new samplings tests** in `samplings/wikitests/tests/wait_strategies_test.go` — synthetic-DOM via `Page.SetContent` with JS-driven dynamic state; each waiter resolves on a real DOM mutation.

## Verification

| Test pool | Pre-batch | Post-batch | Status |
|---|---|---|---|
| `tests/kapi/` unit | 59 | **71** (+12) | green ~6.4s |
| `tests/kcore/` unit | 139 | **154** (+15) | green ~3.1s |
| `samplings/apitests/` | NEW | **11** | green ~4.4s |
| `samplings/wikitests/` | 16 | **21** (+5) | green ~25s |
| Total kexas test functions visible to default `go test` | 574 | **616** | full sweep ~67s |
| Samplings projects | 4 | **5** | all green parallel + headless |

## Findings

### `httpbin.org` echoes file body but not filename

httpbin's `/post` returns `files.<field>: <content>` and `form.<field>: <text>` but doesn't include the original filename in the response. Caught during the multipart-upload sample test; assertion adjusted accordingly. Not a kexas bug — httpbin's response shape is just narrower than I expected.

### `WaitForElementHidden`'s "hidden" definition is broader than Playwright

Kexas's check fires on display:none, visibility:hidden, opacity:0, AND zero-size bounding rect. Playwright's `state: 'hidden'` fires on the first three but treats zero-size as "still attached". The kexas variant matches user intuition better ("can the user see it?") but worth noting if porting test logic from Playwright.

### `WaitForFunction` is the escape hatch

Any condition that doesn't fit a built-in waiter — wait for cart total to update, wait for a chart to render, wait for an animation to finish — is `WaitForFunction("...", timeout)`. The wikitests sample exercises a global `window.kexasReady` flag pattern that's typical for SPA initialisation gates.

## Open follow-ups

- **`Client.Stream(path)` for chunked downloads** — for large files where loading the full body into memory is wasteful. Today's `SaveToFile` reads the whole body first.
- **Request/response interceptors** — RestSharp + Axios both expose middleware-style hooks for logging / auth refresh. Useful but lower priority than methods.
- **Retry policy** with backoff on 5xx / network errors. Common need; small additive feature.
- **`Element.WaitFor(state, timeout)` shorthand** matching Playwright's `locator.waitFor({state})` — sugar over the page-level waiters.
- **Auto-discover `WaitForElementVisible`'s twin** — it's currently the only "visible state" waiter without an explicit "becomes visible after delay" sample test.
