# SCROLL_ACTIONS — Kexas Scroll API

## Overview

Kexas provides two levels of scroll actions:

1. **Page-level**: `Page.ScrollToTop()`, `Page.ScrollToBottom()`, `Page.ScrollBy(x, y)`, `Page.ScrollPosition()`
2. **Element-level**: `Element.ScrollIntoView()`

## Architecture Decisions

### Why Two Levels?

- **Page-level scroll** operates on `window.scrollTo/scrollBy` — useful for navigating long pages, lazy-loading content, or reaching off-screen elements.
- **Element-level scroll** operates on `Element.scrollIntoView()` — useful for scrolling a specific element into the viewport before interacting with it (click, type, hover).

### Why Not CDP `Input.dispatchMouseEvent` for Scrolling?

Playwright and go-rod both use JavaScript-based scrolling (`window.scrollBy`, `element.scrollIntoView`) rather than synthetic mouse wheel events for several reasons:

- **Reliability**: `Input.dispatchMouseEvent` with `type: "mouseWheel"` is flaky across platforms; the delta values vary by OS and display scaling.
- **Simplicity**: JavaScript scrolling is deterministic — `window.scrollBy(0, 500)` always scrolls exactly 500px.
- **Compatibility**: Works identically in headed and headless modes.

### ScrollIntoView Strategy

We use `scrollIntoView({ behavior: 'instant', block: 'center', inline: 'center' })`:

- **`behavior: 'instant'`** — No smooth animation; immediate for test speed.
- **`block: 'center'`** — Centers vertically so the element isn't obscured by sticky headers/footers.
- **`inline: 'center'`** — Centers horizontally for wide pages.

This mirrors Playwright's `scrollIntoViewIfNeeded()` approach.

## API Reference

### Page-Level

| Method | Description | JS Equivalent |
|--------|-------------|---------------|
| `ScrollToTop()` | Scroll to (0, 0) | `window.scrollTo(0, 0)` |
| `ScrollToBottom()` | Scroll to page bottom | `window.scrollTo(0, document.body.scrollHeight)` |
| `ScrollBy(x, y)` | Scroll by relative pixels | `window.scrollBy(x, y)` |
| `ScrollPosition()` | Get current (x, y) scroll | `[window.scrollX, window.scrollY]` |

### Element-Level

| Method | Description | JS Equivalent |
|--------|-------------|---------------|
| `ScrollIntoView()` | Center element in viewport | `el.scrollIntoView({block:'center'})` |

## Guard Clauses (Error Priority)

All scroll methods follow the standard Kexas guard order:

1. **Page-level**: `p == nil` → `ErrPageNil`
2. **Element-level**: `e == nil` → `ErrElementNil` → `e.page == nil` → `ErrElementNoPage` → nodeID/objectID check → `ErrElementInvalidNodeID`

Special case: `ScrollBy(0, 0)` returns `ErrScrollInvalidPixels` (checked after nil page).

## Sentinel Errors

```go
ErrScrollOperationFailed = "scroll operation failed"
ErrScrollInvalidPixels   = "scroll pixels must not be zero"
```

## Pitfalls & Lessons Learned

1. **Scroll settle time**: After `ScrollToBottom()`, elements that were off-screen may need a brief delay (200-500ms) before they become interactable. This is because the browser needs to render the newly visible content.

2. **Lazy-loaded content**: Some SPAs load content on scroll. After scrolling, you may need to `page.Find()` again as new DOM nodes may have been injected.

3. **Sticky headers/footers**: Using `block: 'center'` in `ScrollIntoView` avoids the common bug where an element is scrolled to the top but hidden behind a sticky header.

4. **ScrollBy(0, 0) is invalid**: We explicitly reject zero-delta scrolls as a likely programming error rather than silently no-oping.

5. **ScrollPosition parsing**: CDP returns JavaScript numbers as `float64`; we use `Math.round()` in the JS expression and parse integers for a clean API.

6. **Cross-browser**: All scroll methods use standard `window.scrollTo/scrollBy` and `Element.scrollIntoView` which work consistently across Chrome, Chromium, and headless modes.

## Files

- `kexas/page_scroll.go` — Page-level scroll actions
- `kexas/element_scroll.go` — Element-level ScrollIntoView
- `kexas/errors/errors.go` — Scroll sentinel errors
- `kexas/tests/scroll_test.go` — Unit tests

---

**Created**: 2026-02-27
**Applies to**: Kexas scroll action development
