# Samplings feature coverage — what kexas surface is exercised — 2026-05-03

User asked: do the 4 samplings projects exercise every kexas package's features? If not, show how to use them all.

## Pre-batch coverage

Quick survey: only `kassert`, `ktest`, `kwait` were imported across all `samplings/*/tests/*.go`. The rest were either unused or exercised only via the root `kexas` package (which re-exports kcore types).

| Package | Imported in samplings before this batch? | Surface exercised |
|---|---|---|
| `kexas` (root re-exports) | Yes | Page, Element, Cookie, Recorder (via ktest auto-record) |
| `kassert` | Yes | Most assertions used |
| `ktest` | Yes | Suite, Run |
| `kwait` | Yes | WaitUntil constants only |
| `kcore` | Indirectly (via root) | Yes — most APIs exercised across batches |
| `kapi` | **NO** | Never used |
| `klauncher` | Indirectly (via ktest auto-launch) | No direct usage |
| `errors` | Indirectly (returned from APIs) | No direct usage |

The standout was `kapi` — a useful HTTP client never touched by any test.

## What got added

### `samplings/wikitests/tests/kapi_test.go` (3 tests)

Exercises kapi against Wikipedia's REST API (`/api/rest_v1/page/summary/...`):

- `TestPAPI_GetGoArticleSummaryReturnsJSON` — basic GET + IsOK + JSONMap, asserts title contains "Go" and extract is substantial.
- `TestQAPI_NonExistentPageReturns404` — verifies `IsClientError()` returns true for 404s.
- `TestRAPI_BrowserAndAPITitleAgree` — cross-verification: fetch summary via REST API, navigate to the article in the browser, assert both report the same title prefix. Exercises the API+UI halves of kexas in one test.

All three pass in ~3.6s combined.

## How users would use the rest

Patterns for the still-uncovered surface (recorded for future samplings batches):

### `kwait` direct usage (beyond just constants)

The `kwait.For`, `kwait.UntilReady`, and `kwait.ForPageLoad` primitives are public — usable directly by consumers, not just internally:

```go
import "github.com/jankylewis/kexas/kwait"

err := kwait.For(ctx, func() (bool, error) {
    return checkSomeUserState(), nil
}, kwait.DefaultOptions())
```

A future samplings test could exercise `kwait.UntilReady` with a custom polling function for site-specific state changes (e.g., wait for a cart total to update via JS without a hard sleep).

### `klauncher` direct usage

Most users won't touch klauncher because `ktest` auto-launches. But for opt-in scenarios (custom Chrome flags, non-test scripts), direct use looks like:

```go
import "github.com/jankylewis/kexas/klauncher"

opts := klauncher.DefaultOptions()
opts.Headless = false
opts.Args = append(opts.Args, "--lang=fr-FR")
browser, _ := klauncher.Launch(ctx, opts)
defer browser.Close()
```

A future samplings test could exercise this for a "headed inspect" mode test, or for testing site behavior under a non-default locale.

### `errors` for typed checks

The `errors` package exports sentinels and structured types that consumer code can use with `errors.Is` / `errors.As`:

```go
import (
    "errors"
    kerrors "github.com/jankylewis/kexas/errors"
)

_, err := page.Find(".missing")
if errors.Is(err, kerrors.ErrElementNotFound) {
    // handle gracefully
}
```

Future samplings test could exercise the typed-error path: deliberately Find on a no-match selector and assert `errors.Is(err, kerrors.ErrElementNotFound)`.

## Recommendations

1. **Add a samplings test per session that exercises one previously-unused API.** Prevents the kapi-style "shipped feature, never sample-tested" situation.
2. **Consider an `apitests/` samplings project** if kapi grows — for now, embedding the kapi tests inside wikitests is fine because they pair naturally with Wikipedia's site.
3. **Document `kwait` and `klauncher` direct usage in the README** so users discover the public-API entry points beyond the auto-launch / auto-wait defaults.

## Final state

| Package | Imports in samplings after this batch |
|---|---|
| `kexas` (root) | sltests, wikitests, hntests, ghtests |
| `kassert` | sltests, wikitests, hntests, ghtests |
| `ktest` | sltests, wikitests, hntests, ghtests |
| `kwait` | sltests, wikitests, hntests, ghtests (constants) |
| `kapi` | **wikitests** (NEW) |
| `klauncher` | (none — opt-in only, suggest future addition) |
| `errors` | (none — opt-in only, suggest future addition) |
