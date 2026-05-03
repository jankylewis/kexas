# Kexas dogfood — test design + how-to-reproduce

Companion to [`SLTESTS_DOGFOOD_FINDINGS.md`](SLTESTS_DOGFOOD_FINDINGS.md). Captures the test-design patterns + 30-test surface coverage + commands to re-run this session's work.

## Test-design lessons (not kexas issues — applies to anyone writing kexas tests)

### A. One browser per suite, NOT per test

`ktest.Suite` reuses one browser+page across every test method in a suite (per `ktest_main.go::setupTestEnvironment`). State leaks between tests unless `BeforeEach` explicitly resets. **Symptoms when forgotten:**
- Cart persists items from previous tests (saucedemo uses `localStorage["cart-contents"]`)
- Already-logged-in user navigating to `/` doesn't re-render the login form

### B. Idempotent BeforeEach pattern that worked

```go
func (s *MySuite) BeforeEach() {
    _ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
    if loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded(); !loaded {
        login, _ := pages.NewLoginPage(s.Page).Open()
        _, _ = login.LoginAs(data.StandardUser())
    }
    // Reset cart between tests — saucedemo persists in localStorage
    _, _ = s.Page.Evaluate(`(function() { localStorage.removeItem("cart-contents"); return true; })()`)
    _ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
}
```

Fast-path: if already logged in (saucedemo localStorage persists), navigate straight to /inventory.html. Slow-path: full login flow when session is gone. Always: clear cart + re-navigate so the page is in a known state.

### C. `kexas.config.json` defaults need `parallelSet: 1` for sltests

Default would otherwise dispatch test methods within a suite to multiple workers, each with its own browser. For ~30 tests this triggers the parallel-Chrome flake (already documented in `tests/README.md`). `parallelSet: 1` keeps everything serial within a suite. Cross-suite parallelism (Go's `go test ./...` running packages in parallel) needs `-p 1`.

### D. Test-name prefixes A/B/C/... force ordering

`ktest.SortTestsByPriority` orders by Priority then by registration order. With no explicit `Priority`, methods run in reflection order — which Go does NOT guarantee. Naming tests `TestA…`, `TestB…`, `TestC…` gives a stable, readable order. Kexas's house style (per `.kb/KEXAS_NAMING_PHILOSOPHY.md`) explicitly endorses this. The sltests follow the convention.

### E. Per-test screenshots land in `<test-package-dir>/test-results/screenshots/`

Not in the project root. Each test package's screenshots directory is relative to the package being tested, which means each `tests/<feature>/test-results/` ends up gitignored separately. The `kexas.config.json::screenshotDir` field is interpreted relative to the test's working directory, which is its package dir.

## Test surface coverage

Final 30-test breakdown:

| Suite | Count | Coverage |
|---|---|---|
| `auth/login_test.go` | 6 | login form display, standard/locked/invalid users, empty username/password validation |
| `inventory/inventory_test.go` | 8 | product count, page title, sort A→Z / Z→A / price low→hi / price hi→lo, add-to-cart badge update, social-link footer presence |
| `cart/cart_test.go` | 5 | items shown, remove item, continue shopping, persist across nav, empty-cart no badge |
| `checkout/checkout_test.go` | 8 | first-name / last-name / postal-code required, valid-info advances, overview totals, finish completes order, cancel returns to cart, back-to-products |
| `menu/menu_test.go` | 3 | logout to login, reset-app-state clears cart, burger menu opens + closes |
| **Total** | **30** | All passing with `go test -p 1 ./tests/...` |

## Files written during this session

| Path | Purpose |
|---|---|
| `sltests/go.mod` | Module: `github.com/jankylewis/sltests` with `replace github.com/jankylewis/kexas => ../` |
| `sltests/kexas.config.json` | Headless, parallelSet=1, screenshot-on-fail, viewport 1280×720 |
| `sltests/.gitignore` | `test-results/`, `*.log`, `.claudoholic/` |
| `sltests/data/users.go` | 7 user fixtures (standard, locked, problem, performance_glitch, error, visual, invalid) |
| `sltests/data/products.go` | 6 product fixtures with kebab-case SelectorIDs (incl. AllTheThings parens special case) |
| `sltests/components/header.go` | Header component — burger trigger, cart badge count, cart link |
| `sltests/components/menu.go` | Side-menu actions (Logout, ResetAppState, AllItems, About, Close) |
| `sltests/pages/login_page.go` | LoginPage with Open / FillUsername / FillPassword / ClickLogin / LoginAs / ErrorMessage |
| `sltests/pages/inventory_page.go` | InventoryPage with IsLoaded / Title / ItemCount / ItemNames / ItemPrices / SortBy / ActiveSortLabel / AddItemToCart / RemoveItemFromCart |
| `sltests/pages/cart_page.go` | CartPage with IsLoaded / ItemCount / ItemNames / RemoveItem / Checkout / ContinueShopping |
| `sltests/pages/checkout_info_page.go` | Step 1: FillInfo / Continue / Cancel / ErrorMessage |
| `sltests/pages/checkout_overview_page.go` | Step 2: SubtotalLabel / TaxLabel / TotalLabel / Finish / Cancel |
| `sltests/pages/checkout_complete_page.go` | Step 3: HeaderText / BodyText / BackToProducts |
| `sltests/pages/eval_helpers.go` | numericResult / boolResult / stringSliceResult — CDP-evaluate result coercion |
| `sltests/pages/react_input.go` | setReactInputValue — workaround for kexas Element.Type bug #3 |
| `sltests/tests/<feature>/<feature>_test.go` × 5 | 30 tests across auth/inventory/cart/checkout/menu |

## Files modified in kexas during this session

| Path | Change |
|---|---|
| `internal/cdp/cdp_read.go` | Added `defer c.shutdown()` to readLoop + `shutdown()` helper. Fix #1. |
| `kassert/kassert_errors.go` | Added `Named(name string)` to ErrorAssertion + replaced hardcoded `"error"` with `e.name` in log/error output. Fix #2. |

## Priority for kexas v0.1.0

1. **Bug #1 (CDP connection-close hang)** — already fixed this session. Verify with `go test ./...` in kexas repo + a Chrome-kill stress test.
2. **Bug #3 (Type silently no-ops on React forms)** — add `Element.Fill()`. Without this, anyone using kexas against a React/Vue/Angular SPA will hit silent failures. **Worst category of bug for a test framework.**
3. **Gap #4 (no FindAll)** — add `Page.FindAll()` or `Page.Locator()`. Required for any non-trivial page (lists, tables, multi-element assertions).
4. **Gap #5 (Evaluate result unwrapping)** — add typed evaluators. Quality-of-life; not blocking but every consumer will write the same boilerplate without it.
5. **Gap #6 (kassert.That.Contains for slices)** — small change, big QoL win.
6. **Gap #7 (Element.SelectOption)** — common case, easy fix.
7. **Gap #8 (WaitForElementInteractive)** — nice-to-have; partially mitigated by Bug #3 fix.

`HtmlOnly` rename (separate session) is done. Module path alignment is done. LICENSE is done.

## How to reproduce this session

```bash
cd ~/Documents/se/kexas/sltests
go test -count=1 -timeout=120s -p 1 ./tests/...
```

Expected: `ok` × 5 packages, ~80s total wall-clock. `-p 1` is required until the parallel-package Chrome flake is fixed (`tests/README.md` documents this for the kexas internal tests too).

To re-trigger the bug #3 reproduction without the workaround, edit `sltests/pages/login_page.go::FillUsername/FillPassword` to call `field.Type(value)` (the kexas method) instead of `setReactInputValue(...)` and re-run the auth suite — TestC, TestD, TestF will all fail with "Username is required" because typing silently no-ops.
