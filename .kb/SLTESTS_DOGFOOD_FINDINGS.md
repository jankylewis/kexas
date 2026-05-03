# Kexas dogfood findings — Sauce Demo POM suite (2026-05-02)

Built `sltests/` — a 30-test POM-style suite against https://www.saucedemo.com to dogfood kexas as a real consumer. Surface area exercised: navigate / find / type / click / evaluate / multi-tab cookie / sub-suite hooks / parallel runner / HTML report. This document captures every bug, API gap, and gotcha hit during the build — written for the next person who tries this.

**Result:** 30/30 tests pass when run with `go test -p 1 ./tests/...`. Without `-p 1`, the parallel-package Chrome flake kicks in (already-known; documented in `tests/README.md`).

## Bugs fixed in kexas during this session

### 1. CDP connection-close didn't unblock pending senders — **2-minute hang on every Chrome drop**

**Symptom:** When Chrome's WebSocket dropped (e.g., "connection reset by peer" mid-test), every subsequent `Page.Send/Evaluate/Find` call blocked forever on `respChan`. Tests took the full 2-minute test timeout to die. Stack at time of hang:

```
goroutine 50 [select, 2 minutes]:
  internal/cdp/cdp_send.go:59  // SendToSession select
  kcore/page.go:33             // Page.sendCommand
  kcore/page_content.go:131    // Page.Evaluate
  pages/react_input.go:28      // setReactInputValue
  ...
```

**Root cause:** `internal/cdp/cdp_read.go:readLoop` used to be:

```go
func (c *Client) readLoop() {
    for {
        msgType, data, err := c.conn.Read(c.ctx)
        if err != nil {
            if !c.closed.Load() {
                c.log.Error("read error", "err", err)
            }
            return    // <-- silent exit; pending senders never notified
        }
        ...
    }
}
```

The `SendToSession` select is:
```go
select {
case resp = <-respChan:
case <-ctx.Done():           // per-call timeout
case <-c.ctx.Done():         // client context
}
```

When `readLoop` returned without calling `c.cancel()`, `c.ctx.Done()` never fired and senders blocked indefinitely.

**Fix:** added `defer c.shutdown()` at the top of `readLoop` and a small `shutdown()` helper that flips `c.closed` and calls `c.cancel()`. Idempotent so it's safe whether the close path comes from `readLoop` or from `Client.Close`.

```go
func (c *Client) shutdown() {
    if !c.closed.CompareAndSwap(false, true) {
        return
    }
    c.cancel()
}
```

**Impact:** every connection failure now fast-fails with `ErrConnectionClosed` instead of hanging 2 minutes. **High priority for v0.1.0.** This single fix changed sltests dev-loop time from "do laundry while waiting" to "see error in 5s, fix, retry".

### 2. `kassert.ThatError(...).Named(...)` didn't exist

**Symptom:** `Assertion` (value flavor) had `.Named(name string)` for tagging error messages + structured logs, but `ErrorAssertion` (error flavor) didn't. Mixing the two in tests was inconsistent.

**Fix:** added `Named()` to `ErrorAssertion` (mirrors `Assertion.Named`). Updated all of `ErrorAssertion`'s methods to use `e.name` instead of the hardcoded `"error"` string in log/error output. Existing call sites unaffected — `ThatError` initializes `name = "error"` so behavior is identical when `.Named()` isn't called.

## Open kexas issues found but NOT yet fixed (workarounds in sltests)

### 3. `Element.Type()` silently no-ops against React-controlled inputs after navigation

**Symptom:** Sequence reproduces reliably:

1. Navigate to saucedemo `/`
2. Login as standard_user (works — typing succeeds, redirects to `/inventory.html`)
3. Navigate back to `/` (login form is re-rendered fresh)
4. Type into `#user-name` via `Element.Type("anything")` — **silently does nothing**

The CDP commands all succeed (no error returned), but JS evaluate of `document.querySelector("#user-name").value` after the typing returns empty string. `Click()` on the login button works correctly — only typing is affected. Both fields (username + password) fail when this state is hit.

Hypothesis (not yet confirmed): the `focusAndClear` JS that runs `this.focus(); this.value = ''` interacts badly with React-controlled inputs after the second render — possibly a React `_valueTracker` reconciliation issue where direct value assignment + subsequent native key events get "deduped" by React.

**Workaround in sltests:** `pages/react_input.go` defines `setReactInputValue(page, selector, value)` which uses the React-aware setter pattern:

```js
const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
setter.call(el, value);
el.dispatchEvent(new Event('input', { bubbles: true }));
```

LoginPage.FillUsername / FillPassword and CheckoutInfoPage.fillField both call this helper instead of `Element.Type()`.

**Recommended kexas fix:** add `Element.Fill(text string) error` that uses the React-aware setter pattern. Keep `Type()` as the keyboard-simulation method (real keyDown/char/keyUp events for forms that DO need real typing). Document the difference: `Fill()` for forms with controlled inputs (faster, more reliable); `Type()` for keyboard-event tests / autocomplete triggers / sites that use vanilla JS handlers.

This matches the Playwright API split: `locator.fill(value)` vs `locator.type(text)`.

### 4. `Page.Find()` is strict-mode only — no `FindAll` / `Locator`

**Symptom:** `findByCSS` deliberately errors on `>1 match`:

```go
if matchCount > 1 {
    return nil, fmt.Errorf("selector %s matched %d elements, expected exactly 1", selector, matchCount)
}
```

Hard fail when you want to count, iterate, or operate on a list (cart items, inventory tiles, sort options).

**Workaround in sltests:** `pages/eval_helpers.go` has `numericResult / boolResult / stringSliceResult` helpers around `Page.Evaluate("document.querySelectorAll(...).length")` etc. Used by `InventoryPage.ItemCount`, `InventoryPage.ItemNames`, `CartPage.ItemCount`, etc.

**Recommended kexas additions:**
- `Page.FindAll(selector string) ([]*Element, error)` — Playwright/Selenium-style multi-match return.
- Or richer: `Page.Locator(selector string) *Locator` with `.Count()`, `.Nth(i)`, `.All()`, `.First()`, `.Last()` methods. This is closer to Playwright's design and avoids the failure mode of "I wanted a list, got an error".

### 5. `Page.Evaluate` returns `interface{}` with CDP-wrapped form — every call site needs unwrapping

**Symptom:** Page.Evaluate's return value is the raw CDP `result` field, which can be any of:
- A bare value (`float64`, `string`, `bool`, `[]interface{}`, etc.) for primitives
- A wrapped object `{type: "...", value: ...}` in some cases
- `nil` for `void` expressions

Every consumer has to write the same unwrapping boilerplate. See `sltests/pages/eval_helpers.go` for the full song and dance.

**Recommended kexas additions:** generic typed evaluators using Go 1.18+ generics:
- `EvaluateInt(page, expr) (int, error)`
- `EvaluateString(page, expr) (string, error)`
- `EvaluateBool(page, expr) (bool, error)`
- `EvaluateStrings(page, expr) ([]string, error)`

Or: a single `Evaluate[T any](page, expr) (T, error)` if the implementation can dispatch on the type parameter.

### 6. `kassert.That(slice).Contains(item)` doesn't exist (string-only)

**Symptom:** `Assertion.Contains(substring string)` only handles strings. Trying `kassert.That(t, []string{"a","b"}).Contains("a")` fails with "Expected value to be a string, but got type []string".

**Workaround in sltests:** manual loop in `tests/cart/cart_test.go::TestACartShowsAddedItems`:
```go
var hasBackpack bool
for _, n := range names {
    if n == "Sauce Labs Backpack" {
        hasBackpack = true
        break
    }
}
kassert.That(s.T(), hasBackpack).Named("cart contains Backpack").IsTrue()
```

**Recommended kexas fix:** generic slice-contains via Go generics:
- `kassert.That(t, slice).ContainsItem(item)` — type switch on actual to detect string vs slice
- OR: separate `kassert.ThatSlice(t, slice).Contains(item)` for type safety

### 7. No native `<select>` support in `Element` — must use `Page.Evaluate`

**Symptom:** Saucedemo's product-sort dropdown is a native `<select>` element. There's no kexas equivalent of Playwright's `locator.selectOption("az")`. `Element.Click()` doesn't simulate the native select behavior; `Element.Type()` types into selects (some browsers allow this for keyboard navigation, but it's flaky).

**Workaround in sltests:** `InventoryPage.SortBy()` uses Page.Evaluate:
```go
sel.value = 'az';
sel.dispatchEvent(new Event('change', { bubbles: true }));
```

**Recommended kexas fix:** `Element.SelectOption(value string) error` — sets the select's value and dispatches the appropriate native events. Per-browser behavior varies; the JS-evaluate fallback used here is safe.

### 8. `Page.WaitForElementVisible` retry timing window seems short for fully-React forms

**Symptom:** During investigation it became clear that immediately after `Navigate()` returned, the form sometimes wasn't fully interactive — focus calls were applied but key events landed nowhere because React was still hydrating. `Find()`'s 7-second retry helps, but only because Find checks for the existence of the selector — it doesn't verify the element is *interactive*. Bug #3 is partially a manifestation of this.

**Recommended kexas addition:** a stricter `WaitForElementInteractive` that verifies focusable + not disabled + page in `interactive` or `complete` readyState. Or an option on `Find`: `Find(selector, kwait.UntilInteractive)`.


---

## See also

- [`SLTESTS_TEST_DESIGN.md`](SLTESTS_TEST_DESIGN.md) — test-design lessons + coverage + reproduce-this-session command
- [`SLTESTS_BATCH_LOG.md`](SLTESTS_BATCH_LOG.md) — chronological log of all PM-batch fixes + updated v0.1.0 priority list
- [`V010_DOGFOOD_FIXES.md`](V010_DOGFOOD_FIXES.md) — concise (≤20-line) summary of shipped kexas fixes
