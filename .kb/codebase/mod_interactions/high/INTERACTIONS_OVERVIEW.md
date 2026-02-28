# Element Interactions — High-Level Overview

Source: `kexas/element_interaction.go`, `element_hover.go`, `element_scroll.go`, `element_methods.go`

This document gives a concise overview of how kexas interacts with DOM elements. For deep protocol-level details, see `mod_interactions/detailed/INTERACTIONS.md`.

---

## Core Philosophy

Kexas uses **DOM-level JavaScript execution** (via `Runtime.callFunctionOn`) for all interactions rather than pixel-level input simulation (`Input.dispatchMouseEvent`). This means:

- **No coordinate math** — Elements are targeted by their CDP object reference, not screen position.
- **No viewport dependency** — Clicking works even if the element is scrolled off-screen.
- **Correct event dispatch** — JavaScript `this.click()`, `this.focus()`, `dispatchEvent()` fire the same DOM events that real user actions would.
- **Tradeoff** — CSS `:hover` and coordinate-dependent behaviors aren't perfectly replicated.

---

## The Three-Tier API

Every interaction exists in three variants:

| Tier | Pattern | When to Use |
|------|---------|-------------|
| **Immediate** | `Click()`, `Type(text)`, `Hover()` | Element is guaranteed to be present and ready |
| **WaitAnd** | `WaitAndClick()`, `WaitAndType(text)`, `WaitAndHover()` | General-purpose — polls for visibility first (uses element's default timeout) |
| **WaitAndFor** | `WaitAndClickFor(timeout)`, `WaitAndTypeFor(text, timeout)`, `WaitAndHoverFor(timeout)` | Need custom timeout for slow/fast elements |

### Recommended Default

Use **`WaitAnd*`** variants for all test code. They handle timing automatically and produce clear timeout error messages when elements don't appear.

Use **immediate** variants only in performance-critical loops or when you've already verified the element's state.

---

## Interaction Methods

### Click

```go
elem.Click()              // immediate
elem.WaitAndClick()       // waits for visible + enabled
elem.WaitAndClickFor(15 * time.Second)  // custom timeout
```

- Executes `this.click()` on the element via JavaScript.
- Fires: `focus`, `mousedown`, `mouseup`, `click` DOM events.
- Triggers: link navigation, form submission, checkbox toggle, button handlers.

### Type

```go
elem.Type("hello@example.com")     // immediate
elem.WaitAndType("hello@example.com")  // waits first
elem.WaitAndTypeFor("hello@example.com", 20 * time.Second)
```

- Focuses the element and clears existing text.
- Types each character via three CDP key events: `keyDown` → `char` → `keyUp`.
- Fires: `keydown`, `keypress`, `keyup`, `input`, `change` DOM events per character.
- Correctly triggers React's `onChange`, Angular's `ngModel`, and vanilla `oninput` handlers.

### Hover

```go
elem.Hover()              // immediate
elem.WaitAndHover()       // waits for visible
elem.WaitAndHoverFor(10 * time.Second)
```

- Dispatches a `mouseover` MouseEvent on the element.
- Fires: `mouseover` (with bubbling to ancestors).
- Triggers JavaScript hover handlers (tooltips, dropdown menus).
- Note: CSS `:hover` pseudo-class may not activate (browser limitation with synthetic events).

### Scroll

```go
elem.ScrollIntoView()     // always immediate (no wait variant)
```

- Calls `this.scrollIntoView({behavior: 'instant', block: 'center', inline: 'center'})`.
- Instantly scrolls the nearest scrollable container to center the element.
- Fires: `scroll` event on the container.
- Useful before screenshots or before pixel-level interaction.

---

## Data Retrieval Methods

### GetAttribute

```go
value, err := elem.GetAttribute("href")
value, err := elem.GetAttribute("class")
value, err := elem.GetAttribute("data-testid")
```

- Uses `DOM.getAttributes` (CDP DOM domain command).
- Returns the attribute's string value, or `""` if the attribute doesn't exist.
- Does NOT trigger any DOM events.

### GetText

```go
text, err := elem.GetText()
```

- Executes `this.innerText || this.textContent || ''` via JavaScript.
- `innerText` returns only visible text (respects CSS `display:none`).
- Falls back to `textContent` which returns all text including hidden elements.

### Validate

```go
err := elem.Validate()
```

- Sends `DOM.describeNode` to check if the element's `nodeID` is still valid.
- Returns `nil` if the element is still in the DOM.
- Returns `ErrElementInvalidNodeID` if the element was removed or the page navigated.

### IsVisible

```go
visible, err := elem.IsVisible()
```

- Evaluates JavaScript to check `getComputedStyle(this).display !== 'none'` and similar visibility checks.
- Returns `true` if the element has non-zero dimensions and isn't hidden.

---

## Error Handling

All methods return `error`. Common errors:

| Error | Meaning | Action |
|-------|---------|--------|
| `nil` | Success | Continue |
| `ErrElementNil` | Called method on nil element | Fix test code — `Find` probably failed |
| `ErrElementNoPage` | Element not attached to a page | Fix test code — invalid construction |
| `ErrTextEmpty` | Passed empty string to Type | Check test data |
| `ErrClickOperationFailed` | Chrome reported click failure | Element may be detached or in a weird state |
| Timeout error (wrapped) | Element didn't become visible | Increase timeout or check if element exists |
| CDP error (wrapped) | Chrome communication failure | Browser may have crashed — check logs |

### Recommended Pattern

```go
elem, err := page.Find("#submit-btn")
if err != nil { t.Fatalf("find failed: %v", err) }

err = elem.WaitAndClick()
if err != nil { t.Errorf("click failed: %v", err) }
```

Use `t.Fatalf` for element discovery failures (can't continue without the element). Use `t.Errorf` for interaction failures (may want to continue and check other things).

---

## Performance Guide

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| `Click()` | 0.5–2 ms | 1–2 CDP round-trips |
| `WaitAndClick()` | 100 ms–30 s | Depends on element load time |
| `Type("short")` | 5–15 ms | 3 CDP calls per character |
| `Type("long text...")` | 50–200 ms | Scales linearly with text length |
| `Hover()` | 0.5–2 ms | 1–2 CDP round-trips |
| `ScrollIntoView()` | 0.5–2 ms | 1–2 CDP round-trips |
| `GetAttribute()` | 0.3–1 ms | 1 CDP call |
| `GetText()` | 0.3–1 ms | 1–2 CDP calls |

### Optimization Tips

- Reuse elements instead of re-finding them.
- Use `Click()` instead of `WaitAndClick()` if you just did a `WaitAndFind()`.
- For long text, consider if you really need per-character event simulation.
- `ScrollIntoView()` before `Click()` is unnecessary — `Click()` uses DOM-level clicks that work off-screen.

---

## Interaction Flow Diagram

```
User Test Code                    Kexas Element                  Chrome Renderer
     │                               │                               │
     ├── page.Find("#btn") ─────────►│                               │
     │                               ├── DOM.querySelector ─────────►│
     │◄── *Element ──────────────────│◄── nodeId: 42 ────────────────│
     │                               │                               │
     ├── elem.WaitAndClick() ───────►│                               │
     │                               ├── poll: isVisible? ──────────►│
     │                               │   (repeat until true)         │
     │                               ├── resolveObjectID ───────────►│
     │                               │◄── objectId: {...} ───────────│
     │                               ├── this.click() ──────────────►│
     │                               │                               ├── fire: focus
     │                               │                               ├── fire: mousedown
     │                               │                               ├── fire: mouseup
     │                               │                               ├── fire: click
     │                               │◄── true ─────────────────────│
     │◄── nil (success) ────────────│                               │
```
