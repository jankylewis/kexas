# Unit-test gap fill — 2026-05-03

User asked whether kcore's recently-added Web UI APIs had critical unit tests; gap survey + 16 new tests landed in the same session.

## Existing coverage (kept as-is)

`tests/kcore/` already had **108 unit tests** (no build tag) + **109 integration tests** (`-tags integration`, real browser). The unit tests exhaustively cover the Web UI primitives that existed before the dogfood batch:

- `Page.Find` — ID/CSS/XPath dispatch, empty selector, special chars, numeric IDs, child/combination selectors, XPath functions, parentheses (12 tests across `page_find_test.go` + `page_find_strict_test.go`)
- `Element.Click` — invalid timeout, zero timeout, nil element (4 tests)
- `Element.Type` — invalid timeout, empty text, special characters, nil element (5 tests)
- `Element.Hover` + `WaitAndHover` + `WaitAndHoverFor` — no-page, invalid timeout, no-objectID (7 tests)
- `Element.ScrollIntoView` — nil element, no page, invalid nodeID, with-object/no-page (4 tests)
- `Element.GetText` / `GetAttribute` / `IsVisible` / `Validate` / `Equals` / `String` — full coverage incl. nil/zero/negative permutations
- `Page.WaitForElement{,Visible,Clickable}` — timeout validation, empty selector, special chars (8 tests)
- `Page.Navigate` / `URL` / `Close` / `WaitForLoad` — happy + error paths (~10 tests)
- `Page.LocalStorage` returns Storage (1 test)
- `Page.SetCookies` / `ClearCookies` (2 tests)
- `Page.ScrollToTop` / `ScrollToBottom` / `ScrollBy` / `ScrollPosition` — nil-page guards (5 tests)

## Gap identified

Three APIs added during dogfood had **zero unit-test coverage**:

| API | Added | Why no tests |
|---|---|---|
| `Element.Fill` | sltests dogfood (Rule-07 era predates) | Shipped without — gap |
| `Page.FindAll` | today (2026-05-03) | Shipped without — gap |
| `Page.WaitForURLContains` | yesterday (2026-05-02) | Shipped without — gap |

## Tests added (16 new)

### `tests/kcore/element_fill_test.go` (5 tests)

- `TestElement_Fill_NilElement` — nil receiver returns `ErrElementNil`, no panic.
- `TestElement_Fill_NoPage` — element with nil page returns `ErrElementNoPage`.
- `TestElement_Fill_EmptyTextAllowed` — `Fill("")` is the "clear field" operation; should NOT trigger the empty-text guard that `Type` has.
- `TestElement_Fill_SpecialCharacters` — special-char text passes through validation.
- `TestElement_Fill_ZeroNodeID_NoObjectID` — element with both nodeID=0 and empty objectID is unusable; rejects without panic.

### `tests/kcore/page_find_all_test.go` (6 tests)

- `TestPage_FindAll_NilPage` — returns `"FindAll: nil page"` error.
- `TestPage_FindAll_EmptySelector` — guard order: nil-page fires before empty-selector.
- `TestPage_FindAll_CSS_Dispatch` — `.class` selector reaches the dispatcher.
- `TestPage_FindAll_XPath_Dispatch` — `//div[…]` selector reaches the dispatcher.
- `TestPage_FindAll_XPathPositional_Dispatch` — `(...)[1]` form also dispatches XPath.
- `TestPage_FindAll_IDShortcut_Dispatch` — `#x` form folds into CSS path.

### `tests/kcore/page_wait_for_url_test.go` (5 tests)

- `TestPage_WaitForURLContains_ZeroTimeout` — sub-100ms timeout rejected with timeout-validation error.
- `TestPage_WaitForURLContains_NegativeTimeout` — same path.
- `TestPage_WaitForURLContains_BelowFloor` — exactly 99ms still rejected.
- `TestPage_WaitForURLContains_NilPage` — nil page returns "URL did not contain X" error format.
- `TestPage_WaitForURLContains_ReturnsBeforeTimeoutOnNilPage` — nil-page guard short-circuits (<1s) instead of burning the requested 5s timeout.

## Bug surfaced + fixed during test-writing

`WaitForURLContains` was missing the nil-page guard that every other `WaitFor*` method has. Without it, `p.URL()` on a nil receiver would panic. Patched by adding `if p == nil { return errors.URLDidNotContainWithin(substr, timeout) }` after the timeout-validation block. Caught by `TestPage_WaitForURLContains_NilPage` — the test was written first, the impl was patched to make it pass.

This is the kind of bug that's nearly invisible until the unit test forces the nil-page case. Argument for the gap-fill effort.

## Final state

| Metric | Before | After |
|---|---|---|
| Unit tests in `tests/kcore/` | 108 | **124** |
| Integration tests (with `-tags integration`) | 109 | 109 |
| Total declared | 217 | **233** |
| Web UI APIs with no unit-test coverage | 3 (Fill, FindAll, WaitForURLContains) | **0** |

All 124 unit tests pass in 0.4s without `-tags integration`.

## Open / follow-ups

- **Lower the timeout floor for WaitForURLContains** to e.g. 50ms — current 100ms is arbitrary and the floor exists only to avoid `0` causing a tight loop. Could relax once we trust the polling cadence.
- **Survey other Wait\* methods for similar nil-page gaps.** The fix above was a missing guard pattern; other recently-added methods may have the same omission.
- **Auto-detect missing unit tests for new public APIs.** Could be a CI check: every `func (p *Page) Foo` or `func (e *Element) Foo` requires at least one `TestPage_Foo*` / `TestElement_Foo*` in `tests/kcore/`.
