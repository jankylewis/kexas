# Samplings API coverage — 2026-05-02

Round 2 of samplings expansion: each project (wikitests, hntests, ghtests) grew from 3 → 10 tests. Tests D-J target kexas APIs the original A-C tests didn't touch, picked specifically to exercise lesser-used code paths and weird real-world DOM.

## What got covered

APIs that had **zero samplings exercise** before this batch, now all green:

| API | Where exercised |
|---|---|
| `Page.ScrollBy` / `ScrollPosition` | wiki D, hn F |
| `Page.ScrollToTop` | wiki E |
| `Element.ScrollIntoView` | wiki G, hn G, gh E |
| `Page.FindByXPath` | wiki F, hn E, gh D, gh F (SVG), hn D (recovery) |
| `Page.URL()` | hn H, gh H |
| `Page.SetContent` | wiki J, hn J, gh J |
| `Page.SetCookies` / `GetCookies` / `ClearCookies` | wiki H, hn I, gh I |
| `Page.LocalStorage().Set/Get/Has/Remove` (typed API, not Evaluate) | wiki I |
| `Page.WaitForElementVisible` | wiki J, hn J |
| `Element.Hover` on micro-target (10x10 px) | hn D |
| `Element.GetAttribute` on SVG | gh F |
| `Element.Fill` against `SetContent`-rendered form | gh J |

Negative path: `Page.Find` on a guaranteed-miss selector — gh G — confirms error-not-panic behavior.

## Findings

### 1. `Page.Find` and `Page.FindByXPath` are strict-single-match

The single most surface-level finding from this batch. **Three of the new tests failed initially because of this:**

- `s.Page.Find(".votearrow")` on HN → 29 matches → errors.
- `s.Page.Find(".octicon")` on GitHub → 149 matches → errors.
- `s.Page.FindByXPath("//div[@id='mw-content-text']//p[normalize-space()][1]")` on Wikipedia → 8 matches → errors. (XPath `[1]` filters per-parent, not document-wide.)

**Behavior:** when the selector matches multiple elements, kexas returns `selector X matched N elements, expected exactly 1` after waiting for the timeout. This is consistent with Playwright's strict-mode locators, and probably correct as the default. But:

- Worth documenting front-and-centre in the README so users don't bump into it the way these tests did.
- A `Page.FindAll(selector) ([]*Element, error)` companion API would close the most painful gap. Today, anything multi-element forces a fallback to `Page.Evaluate("document.querySelectorAll(...)")` with raw JS — which is what `samplings/*/pages/*.go` had to do for counting.
- For "first match in document order," XPath workaround `(...)[1]` is reliable but verbose: `(//div[@class='votearrow'])[1]`, `(//svg[contains(@class,'octicon')])[1]`.

### 2. `Find` on no-match waits the full timeout (~7s)

**Behavior:** `Page.Find(".does-not-exist")` polls for the configured timeout (default 7s) before returning the no-match error. Confirmed by the `gh G` (`TestGNonExistentSelectorReturnsErrorNotPanic`) test — it succeeded but took 7.86s.

This is correct for "wait for element to appear" semantics, but it means **intentional negative-path tests are slow**. Consider:
- A `FindNow(selector)` fast-fail variant for negative assertions.
- Or document the existing escape hatch: `WaitForElement(selector, 100*time.Millisecond)` to short-circuit the wait.

### 3. APIs that worked first try with no caveats

These were unknown-quality going in; all green on first run across all 3 sites:

- **`SetContent` + subsequent `Find`/`Fill`/`WaitForElementVisible`** — the synthetic-DOM path is solid. `SetContent` returns only after the new document is parsed and queryable; no race conditions observed.
- **`Element.Fill` on `SetContent`-rendered input** — proves Fill's React-aware setter (added during sltests dogfood) also handles plain HTML inputs without React getting in the way.
- **Cookies round-trip across navigation** — `SetCookies` → `Navigate` → `GetCookies` finds the planted cookie on all 3 domains (.wikipedia.org, .news.ycombinator.com, .github.com). `ClearCookies` clears.
- **`LocalStorage` typed API** — `Set/Has/Get/Remove` round-trips cleanly. Worth promoting in docs over the Evaluate-based pattern that sltests used initially.
- **`Element.Hover` on a 10x10 px target** — CDP `Input.dispatchMouseEvent` placement math is solid. No flakes across multiple runs.
- **`Element.GetAttribute` on SVG** — `<svg class="octicon">` returns its `class` attribute correctly. Confirms the objectID dispatch path (added during sltests dogfood) handles non-HTML namespaces.

### 4. XPath positional predicates are subtle

`//p[normalize-space()][1]` does **not** mean "first non-empty `<p>` in the document." It means "every `<p>` that is non-empty AND is the first such child within its parent." On the Go article, that matches 8 different paragraphs (one per section).

Document-order-first requires the outer-paren form: `(//p[normalize-space()])[1]`.

Worth a one-paragraph mention in `.kb/` (or the kexas README's XPath section) since this footgun is unrelated to kexas itself but combines badly with kexas's strict-single-match enforcement to produce a confusing failure.

## What didn't surface

I expected to find at least one kexas-side bug along the lines of the sltests dogfood (Element.Fill, GetAttribute-via-objectID, cookie filename slashes). Instead, every failure was a test-design issue with the selector — kexas's behavior was correct in all cases. Either:

- The earlier dogfood already swept the high-traffic paths.
- Or the XPath/Cookies/Storage/SetContent code paths really are this clean.

Probably both.

## Recommendations for next iteration

1. **Add `Page.FindAll(selector) ([]*Element, error)`.** The biggest UX gap.
2. **Document strict-single-match prominently** in the README's "Common gotchas" section. Add the XPath positional workaround as a recipe.
3. **Add fast-fail variant** for `Find` (e.g., `FindOpt{NoWait: true}` or a separate `FindNow`). Matters for negative-path tests.
4. **Promote `LocalStorage()` typed API in docs** — sltests reached for `Evaluate("localStorage.setItem(...)")` because the typed alternative wasn't visible enough.

## Test counts after this batch

| Project | Tests | Runtime (cached ffmpeg) |
|---|---|---|
| wikitests | 10 | ~10s |
| hntests | 10 | ~9s |
| ghtests | 10 | ~17s |
| sltests (carryover) | 50 | ~38s |
| **total** | **80** | **~74s** |

All green, parallel + headless.
