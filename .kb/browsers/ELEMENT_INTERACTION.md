# Element Interaction on WebUI

> How Kexas interacts with DOM elements: clicking, typing, hovering, reading text, and checking visibility.

**Last Updated:** February 27, 2026

---

## Overview

Element interaction is split across two files:

- **`element_interaction.go`** — Action methods: `Click`, `Type`, `Hover`, and their `WaitAnd*` variants
- **`element_methods.go`** — Query methods: `GetText`, `GetAttribute`, `Validate`
- **`element.go`** — Core struct, `IsVisible`, `resolveObjectID`

All interaction methods follow the **objectId-first pattern**: resolve the element's `objectId` (cached or via `DOM.resolveNode`), then use `Runtime.callFunctionOn` to execute JavaScript on the element.

---

## Core Pattern: resolveObjectID

Every interaction method starts by resolving the element to an `objectId`:

```go
var objectID string
objectID, err = e.resolveObjectID()
```

`resolveObjectID()` logic:
1. If `e.objectID != ""` → return cached value (fast path)
2. Otherwise → call `DOM.resolveNode({nodeId})` → extract `objectId` from response → cache it
3. This means elements found via `findByCSS`/`findByID` (which already have `objectId`) skip the extra CDP call

---

## Click Methods

### `Click()` — Immediate click, no waiting

**CDP:** `Runtime.callFunctionOn` with `this.click()`

```go
var clickFunction string = `
    function() {
        this.click();
        return true;
    }
`
```

- Uses JavaScript `this.click()` which triggers the element's click handler
- Works for buttons, links, checkboxes, etc.
- Does NOT simulate mouse coordinates — use for simple navigation clicks

### `WaitAndClick()` — Wait for clickable, then click

1. `WaitForElementClickable(selector, timeout)` — polls until element is visible and enabled
2. `resolveObjectID()` — get fresh objectId
3. `Runtime.callFunctionOn` with `this.click()`

### `WaitAndClickFor(timeout)` — Same as WaitAndClick with custom timeout

Validates `timeout >= 1s` before proceeding.

---

## Type Methods

### `Type(text)` — Immediate typing using CDP keyboard events

**This is the critical method that required the most debugging.**

**CDP Commands:** `Runtime.callFunctionOn` (focus) + `Input.dispatchKeyEvent` (per character)

**Two-phase approach:**

#### Phase 1: Focus and clear via JavaScript
```go
var focusFunction string = `
    function() {
        this.focus();
        this.value = '';
        return true;
    }
`
// Runtime.callFunctionOn with objectId
```

#### Phase 2: Type each character via CDP keyboard events
```go
for _, ch := range text {
    var charStr string = string(ch)

    // keyDown event
    Input.dispatchKeyEvent({
        type:           "keyDown",
        text:           charStr,
        unmodifiedText: charStr,
        key:            charStr,
    })

    // keyUp event
    Input.dispatchKeyEvent({
        type:           "keyUp",
        text:           charStr,
        unmodifiedText: charStr,
        key:            charStr,
    })
}
```

### Why Input.dispatchKeyEvent Instead of JavaScript Value Assignment

The original implementation used `this.value = text` via `Runtime.callFunctionOn`. This failed on Amazon's SPA because:

1. **SPA frameworks (React, Angular) listen to actual keyboard events**, not value property changes
2. Setting `this.value` bypasses the framework's event handlers
3. The form field appears filled but the framework's internal state doesn't update
4. Clicking "Continue" after JS-only typing causes the form to submit with empty/stale data

**`Input.dispatchKeyEvent` is the go-rod/Playwright pattern:**
- Simulates real keyboard input at the browser level
- Triggers `keydown`, `keypress`, `input`, `keyup` events naturally
- SPA frameworks detect and process each keystroke
- Form validation and state management work correctly

### Input Domain Auto-Enablement

The `Input` CDP domain does NOT require explicit `.enable()` — it is auto-enabled. In `ensureAgentsForCommand`:

```go
case "Input.dispatchKeyEvent", "Input.dispatchMouseEvent", "Input.dispatchTouchEvent":
    // Input domain is auto-enabled, no agent needed
    return nil
```

### `WaitAndType(text)` — Wait for visible, then type

**Note:** This method still uses the old JS `this.value = text` approach. For SPA forms, use `Type()` directly after finding the element with `Find()`.

```go
// WaitAndType uses JS value assignment (works for simple forms)
var typeFunction string = `
    function(text) {
        this.focus();
        this.value = '';
        this.value = text;
        return true;
    }
`
```

### `WaitAndTypeFor(text, timeout)` — Same with custom timeout and event dispatch

This variant also dispatches `input` and `change` events after setting the value:

```go
this.dispatchEvent(new Event('input', { bubbles: true }));
this.dispatchEvent(new Event('change', { bubbles: true }));
```

### Typing Method Selection Guide

| Scenario | Recommended Method |
|----------|-------------------|
| SPA forms (React, Angular, Amazon) | `Type()` — uses `Input.dispatchKeyEvent` |
| Simple HTML forms | `WaitAndType()` — uses JS value assignment |
| Forms with custom validation | `Type()` — triggers real keyboard events |
| Forms needing input/change events | `WaitAndTypeFor()` — dispatches events after assignment |

---

## Hover Methods

### `Hover()` — Immediate mouseover event

**CDP:** `Runtime.callFunctionOn` with `MouseEvent('mouseover')`

```go
var hoverFunction string = `
    function() {
        var event = new MouseEvent('mouseover', {
            'view': window,
            'bubbles': true,
            'cancelable': true
        });
        this.dispatchEvent(event);
        return true;
    }
`
```

### `WaitAndHover()` / `WaitAndHoverFor(timeout)`

Same pattern: wait for visibility, then dispatch mouseover.

---

## Query Methods

### `GetText()` — Get visible text content

**CDP:** `Runtime.callFunctionOn` with `this.innerText`

```go
function() { return this.innerText || this.textContent || ''; }
```

- Uses `innerText` first (visible text only, respects CSS `display: none`)
- Falls back to `textContent` (all text including hidden)
- Returns empty string if neither is available

**Previous bug:** Originally used `DOM.getOuterHTML` which returned the full HTML markup, not just text. Fixed to use `Runtime.callFunctionOn` with `innerText`.

### `GetAttribute(name)` — Get attribute value

**CDP:** `DOM.getAttributes` with `nodeId`

- Returns all attributes as alternating name/value pairs
- Iterates to find the matching attribute name
- Returns empty string (not error) if attribute not found

**Note:** This method still requires `nodeId > 0`. Elements found via CSS/ID (objectId-only) will fail. Consider refactoring to use `Runtime.callFunctionOn` with `this.getAttribute(name)`.

### `IsVisible()` — Check element visibility

**CDP:** `Runtime.callFunctionOn` with computed style check

```javascript
function() {
    var style = window.getComputedStyle(this);
    if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') {
        return false;
    }
    var rect = this.getBoundingClientRect();
    return rect.width > 0 && rect.height > 0;
}
```

Checks three conditions:
1. `display` is not `none`
2. `visibility` is not `hidden`
3. `opacity` is not `0`
4. Element has non-zero width and height

### `Validate()` — Check if element is still in DOM

**CDP:** `DOM.describeNode` with `nodeId`

- Returns error if node no longer exists
- Useful after navigation to check if element references are stale

---

## Error Handling Pattern

All interaction methods follow the same guard clause pattern:

```go
func (e *Element) SomeAction() error {
    // 1. Nil checks
    if e == nil { return errors.ErrElementNil }
    if e.page == nil { return errors.ErrElementNoPage }

    // 2. Dual guard: accept nodeId OR objectId
    if e.nodeID <= 0 && e.objectID == "" {
        return errors.ErrElementInvalidNodeID
    }

    // 3. Resolve objectId
    objectID, err := e.resolveObjectID()

    // 4. Execute CDP command
    result, err := e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)

    // 5. Validate result
    success, ok := result["result"].(map[string]interface{})["value"].(bool)
}
```

---

## Sentinel Errors

Defined in `errors/` package:

| Error | Meaning |
|-------|---------|
| `ErrElementNil` | Element pointer is nil |
| `ErrElementNoPage` | Element has no associated page |
| `ErrElementInvalidNodeID` | Neither nodeId nor objectId is valid |
| `ErrTextEmpty` | Empty string passed to Type() |
| `ErrClickOperationFailed` | Click JS returned false |
| `ErrTypeOperationFailed` | Type JS returned false |
| `ErrHoverOperationFailed` | Hover JS returned false |

---

## File Reference

- **`element.go`** — `Element` struct, `NewElement()`, `NewElementWithObject()`, `resolveObjectID()`, `IsVisible()`
- **`element_interaction.go`** — `Click()`, `WaitAndClick()`, `Type()`, `WaitAndType()`, `Hover()`, `WaitAndHover()`, and `*For()` variants
- **`element_methods.go`** — `GetText()`, `GetAttribute()`, `Validate()`
- **`page_wait.go`** — `WaitForElementVisible()`, `WaitForElementClickable()`

