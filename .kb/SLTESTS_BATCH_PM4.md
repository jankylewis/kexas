# Kexas dogfood — PM batch 4 (+10 creative tests + 1 kexas bug fix)

Companion to [`SLTESTS_BATCH_LOG.md`](SLTESTS_BATCH_LOG.md) (PM batches 1–3) and [`SLTESTS_DOGFOOD_FINDINGS.md`](SLTESTS_DOGFOOD_FINDINGS.md). This file isolates the PM-batch-4 work because the rolling log was approaching the 200-line Rule 06 cap.

## Update — 2026-05-02 PM batch 4 (+10 creative-action tests, 1 kexas bug fixed)

### Tests added (50 total now)

10 new tests deliberately exercising less-trodden kexas APIs:

| Suite | Tests | kexas APIs exercised |
|---|---|---|
| `inventory` | TestM hover, TestN scroll-to-footer, TestO images-not-broken | `Element.Hover`, `Page.ScrollToBottom`, `Element.IsVisible`, `Page.Evaluate` |
| `cart` | TestG cart-mirrors-localStorage | `Page.Evaluate` reading `localStorage.getItem` |
| `auth` | TestG localStorage-initialized-after-login | `Page.Evaluate` reading `Object.keys(localStorage).length` |
| `menu` | TestF about-link-leaves-saucedemo | `Menu.About`, `Page.WaitForLoad`, `Page.URL` (cross-origin nav check) |
| `misc` (NEW) | TestA social-link-hrefs, TestB programmatic-screenshot, TestC multi-tab-window-open, TestD Page.GetCookies | `Element.GetAttribute`, `Page.Screenshot` (programmatic), `Browser.PageCount`, `Page.GetCookies` |

### Kexas bug fixed: `Element.GetAttribute` silently rejected CSS-found elements

**Symptom:** `Find()` for `[data-test="social-twitter"]` succeeded (returns `*Element`), then `GetAttribute("href")` failed with `errors.ErrElementInvalidNodeID`. Trace showed `nodeID <= 0` guard rejecting the element. **Misleading**: the element was perfectly valid; it just had `objectID` instead of `nodeID` (CSS-found elements use the objectID path; only XPath-found ones populate nodeID).

**Root cause:** `kcore/element_methods.go::GetAttribute` had:
```go
if e.nodeID <= 0 {
    return "", errors.ErrElementInvalidNodeID
}
```
…then unconditionally used `nodeId` in the CDP `DOM.getAttributes` call. Half-implementation: works for XPath-found elements only.

**Fix:** Split into two paths:
- `getAttributeViaNodeID(name)` — original DOM.getAttributes (kept for XPath path)
- `getAttributeViaObjectID(name)` — `Runtime.callFunctionOn` with `function(name) { return this.getAttribute(name); }`
The dispatcher in `GetAttribute` checks `e.objectID != ""` first, falls back to nodeID. The pre-call guard now accepts either identifier (`e.nodeID <= 0 && e.objectID == ""`).

**Impact:** Previously **all CSS-found elements were silently broken** for `GetAttribute`. Tests using XPath worked; tests using CSS got `ErrElementInvalidNodeID`. This is the second case (after `Element.Type` from earlier today) of "kexas API works for XPath but silently fails for CSS." Worth a project-wide audit pass: search every method that takes an `*Element` for nodeID-only assumptions.

### Other new findings (no fix yet — added to priority list)

- **Saucedemo localStorage model changed.** `localStorage["session-username"]` is no longer set after login (it was in the pre-2024 saucedemo). My initial TestG assumed the old key existed; rewrote to assert `Object.keys(localStorage).length > 0` (probes that the React app fully booted, not the specific key). Lesson: when probing third-party app internals, prefer existence/non-empty checks over specific key/value asserts.
- **`Page.Evaluate` of expressions returning Window/Document objects** errors with `"Object reference chain is too long"` from CDP. Hit it when calling `window.open(url, "_blank")` directly. Workaround: prefix with `void` (`void window.open(...)`) so the expression returns `undefined`. Same trick applies to any `Page.Evaluate` that would return a complex DOM/Window reference. **Recommended kexas docstring addition** on `Page.Evaluate`: "If your expression returns a DOM node or Window, prefix with `void` or wrap in IIFE returning a primitive — CDP cannot serialize circular object references."
- **`Browser.PageCount` does NOT auto-increment on `window.open`.** kexas only tracks pages it has explicitly attached to via `attachToPage`. Popups + window.open targets are NOT auto-attached. Consumer must call `Browser.WaitForNewPage(action, timeout)` to attach + track. TestC documents this as the current behavior; if kexas adds auto-attach for popup targets, flip the test's assertion from `Equals(1)` to `>1`. **Roadmap candidate**: auto-attach option in launcher config (default off for backwards-compat).
- **`Page.GetCookies` works correctly with empty cookie list.** Saucedemo uses localStorage (no auth cookies), so `GetCookies` returns `([], nil)` not `(nil, error)`. Important kexas API smoke check — many consumers will hit empty-cookie pages.
- **`Page.ScrollToBottom + Element.IsVisible`** combo works as expected. The footer was off-screen at viewport 1280×720, scroll brought it into view, IsVisible returned true. Standard interaction pattern with no surprises.
- **`Element.Hover` on `[data-test="shopping-cart-link"]`** smoke-passes. Saucedemo's cart icon doesn't have a noticeable hover state, so this is mostly an API smoke check. Worth knowing kexas's Hover dispatches CDP events correctly.

### Sltests test-pattern lessons

- **`s.T().Name()` includes the suite path with slashes** (`"TestMisc/TestBProgrammatic..."`). Using it directly as a filename creates a missing-subdir error. Sanitize via `strings.ReplaceAll(name, "/", "_")`. Worth adding a `kexas/ktest.SanitizedTestName(t)` helper if the use case becomes common.
- **Probing app internals with `Page.Evaluate("Object.keys(localStorage)")`** is a useful black-box health check. When the SUT changes its storage model (saucedemo did), tests that target specific keys break — generic "is something there" checks survive.
- **Cross-origin navigation tests** (TestF: About link → saucelabs.com) require `Page.WaitForLoad(timeout)` because the kexas Page navigation is async; `URL()` returns the previous URL until the new origin's HTML responds. ~8s timeout safer than default for CDN-fronted sites with TLS handshake overhead.

### Updated v0.1.0 priority list (after PM batch 4)

1. ~~CDP connection-close hang~~ ✅
2. ~~Element.Type silent no-op on React~~ ✅ (Element.Fill)
3. ~~Config not found from sub-package CWDs~~ ✅
4. ~~No HTML report for legacy ktest.Run path~~ ✅
5. ~~Per-test screenshot was failure-only~~ ✅ (always-on)
6. ~~Per-test video missing entirely~~ ✅ (always-on, frames-fallback)
7. ~~Donut legend column misalignment~~ ✅
8. ~~`Element.GetAttribute` rejected CSS-found elements~~ ✅ (PM batch 4)
9. **Per-package `report.html` overwrite** — headline open issue
10. **Auto-step capture** — passed-test details need step-by-step
11. **`Page.FindAll` / `Locator`** — multi-element flows
12. **Project-wide nodeID-vs-objectID audit** — new from PM batch 4. Search every Element method for `nodeID <= 0` guards that should accept either identifier
13. **`Page.Evaluate` Object-reference-chain docstring** — document the `void`-prefix workaround for expressions that would return DOM/Window
14. **`Browser` auto-attach popups** — opt-in launcher config to track window.open targets without `WaitForNewPage`
15. `Page.Evaluate` typed-result helpers
16. `kassert.That(slice).Contains`
17. `Element.SelectOption`
18. `WaitForElementInteractive`
19. ffmpeg dependency for video → CDP screencast event recorder

### Files touched in PM batch 4 (kexas)

| Path | Change |
|---|---|
| `kcore/element_methods.go` | `GetAttribute` now accepts both nodeID-backed and objectID-backed elements; new `getAttributeViaObjectID` helper using `Runtime.callFunctionOn` |

### Files touched in PM batch 4 (sltests)

| Path | Change |
|---|---|
| `pages/item_detail_page.go` | New `ItemDetailPage` POM (PM batch 3, listed for completeness) |
| `pages/inventory_page.go` | New `OpenItem(product)` helper |
| `tests/inventory/inventory_test.go` | +TestI–O (item detail + creative tests M/N/O) |
| `tests/cart/cart_test.go` | +TestF (all-six), +TestG (localStorage mirror) |
| `tests/auth/login_test.go` | +TestG (localStorage initialized after login) |
| `tests/checkout/checkout_test.go` | +TestI–K (PM batch 3) |
| `tests/menu/menu_test.go` | +TestD–E (PM batch 3), +TestF (About leaves saucedemo) |
| `tests/misc/misc_test.go` | NEW — TestA hrefs / TestB screenshot / TestC multi-tab / TestD cookies |
