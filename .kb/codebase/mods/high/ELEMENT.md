# Element (High-Level)

Source: `kexas/element.go`, `element_interaction.go`, `element_hover.go`, `element_methods.go`, `element_scroll.go`

## Mission

Represent a single DOM node with a fluent API for clicking, typing, hovering, scrolling, and reading attributes/text. `kexas.Element` is the unit test authors interact with after finding an element on a page.

## What It Holds

| Field | Purpose |
|-------|---------|
| `page` | Back-pointer to the owning `Page` (for sending CDP commands) |
| `selector` | The CSS selector used to find this element (for re-querying in wait methods) |
| `nodeID` | Chrome's internal DOM node identifier (integer, volatile — changes on navigation) |
| `objectID` | CDP remote object handle for `Runtime.callFunctionOn` (stable within a page session) |
| `timeout` | Default timeout for `WaitAnd*` methods (typically 10–30 seconds) |

## Three-Tier API

Every interaction exists in three variants:

| Tier | Example | Behavior |
|------|---------|----------|
| Immediate | `Click()`, `Type(text)`, `Hover()` | Execute now — no waiting |
| WaitAnd | `WaitAndClick()`, `WaitAndType(text)` | Poll for visibility first, then act (default timeout) |
| WaitAndFor | `WaitAndClickFor(timeout)` | Same as WaitAnd but with custom timeout |

**Recommendation**: Use `WaitAnd*` for all test code. Use immediate variants only when the element is guaranteed ready.

## Key Functions

### Actions
- **`Click()` / `WaitAndClick()`** — Executes `this.click()` via JavaScript. Fires focus, mousedown, mouseup, click events.
- **`Type(text)` / `WaitAndType(text)`** — Focuses the element, clears it, types each character via keyDown/char/keyUp CDP events.
- **`Hover()` / `WaitAndHover()`** — Dispatches a `mouseover` MouseEvent on the element.
- **`ScrollIntoView()`** — Scrolls the nearest container to center the element in the viewport.

### Data Retrieval
- **`GetAttribute(name)`** — Reads a DOM attribute (e.g., `href`, `class`) via `DOM.getAttributes`.
- **`GetText()`** — Returns visible text via `innerText` (falls back to `textContent`).
- **`IsVisible()`** — Checks computed style and dimensions.
- **`Validate()`** — Confirms the element's `nodeID` is still valid in Chrome's DOM.

## Lifecycle

```
page.Find("#login") → *Element{nodeID:42, selector:"#login"}
  │
  ├── First interaction: resolveObjectID() → DOM.resolveNode → objectID cached
  ├── Subsequent interactions: reuse cached objectID (no CDP call)
  │
  ├── elem.WaitAndClick()
  │     ├── Poll: is element visible + enabled?
  │     └── Runtime.callFunctionOn("this.click()")
  │
  └── Page navigates → objectID becomes stale → must re-Find
```

## Error Handling

| Error | Condition |
|-------|-----------|
| `ErrElementNil` | Called method on nil element |
| `ErrElementNoPage` | Element has no page reference |
| `ErrElementInvalidNodeID` | Stale element after DOM change |
| `ErrTextEmpty` | Empty string passed to `Type` |
| `ErrClickOperationFailed` | Chrome reported click failure |
| Timeout (wrapped) | Element didn't become visible in time |

## Why It Matters

Elements are the atomic unit of test interaction. Every assertion ("this button should say 'Submit'"), every action ("click the login link"), and every wait ("wait for the spinner to disappear") operates on an Element. The three-tier API pattern and lazy objectID caching make interactions both reliable and fast.
