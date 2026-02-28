# Element Module — Deep Detailed Walkthrough

Source: `kexas/element.go`, `element_interaction.go`, `element_hover.go`, `element_methods.go`, `element_scroll.go`

This document explains every type, function, and mechanism in the Element module at the OS, process, memory, and protocol level.

---

## 1. The `Element` Struct — What Lives in Memory

```go
type Element struct {
    page     *Page         // reference to the page containing this element
    selector string        // the CSS selector used to find this element
    nodeID   cdp.NodeID    // CDP node identifier (integer)
    objectID string        // CDP remote object ID for Runtime.callFunctionOn
    timeout  time.Duration // default timeout for wait-and-* operations
}
```

### Memory Layout

An `Element` is heap-allocated (returned as `*Element` from `Page.Find`). Its fields:

- **`page`** — 8-byte pointer back to the `Page`. This is how the Element sends CDP commands. The Element does NOT own the Page — it's a borrowed reference. If the Page is closed while an Element reference still exists, any Element method will fail with a CDP error (the session is invalidated).

- **`selector`** — A Go string (16-byte header: pointer + length). The backing byte array holds the CSS selector text (e.g., `"#login-form"`, `".product-card:nth-child(3)"`). This is stored for two purposes: (1) debugging/logging, (2) re-querying in `WaitAndClick` / `WaitAndType` which call `page.WaitForElementClickable(e.selector, ...)`.

- **`nodeID`** — A `cdp.NodeID` (alias for `int`). This is Chrome's internal DOM node identifier, assigned when the DOM agent inspects the document. Node IDs are **session-scoped and volatile** — they change whenever the DOM is re-fetched (e.g., after navigation). A stale `nodeID` means CDP will return an error like `"Could not find node with given id"`.

- **`objectID`** — A string like `"{\\"injectedScriptId\\":2,\\"id\\":15}"`. This is a CDP `RemoteObjectId` — a handle to a JavaScript object in the renderer's V8 heap. Unlike `nodeID`, an `objectID` is stable as long as the page hasn't navigated. It allows calling JavaScript functions directly on the DOM element via `Runtime.callFunctionOn`. This is the preferred way to interact with elements (following go-rod's pattern).

- **`timeout`** — A `time.Duration` (8-byte int64, nanoseconds). Typically 10–30 seconds. Used as the default deadline for `WaitAndClick`, `WaitAndType`, etc.

### Total Memory Footprint

Each `Element` struct: ~64 bytes for the struct itself + ~50 bytes for the selector string backing array + ~80 bytes for the objectID string backing array ≈ **~200 bytes** per element. Creating 100 elements costs ~20 KB — negligible.

---

## 2. Constructors — `NewElement` vs `NewElementWithObject`

### `NewElement(page, selector, nodeID, timeout)`

Creates an Element with only a `nodeID`. The `objectID` is left empty (`""`). This means the first call to `Click()`, `Type()`, etc. will trigger a `DOM.resolveNode` call to obtain the `objectID`. This adds one extra CDP round-trip (~0.5–2 ms).

### `NewElementWithObject(page, selector, nodeID, objectID, timeout)`

Creates an Element with both `nodeID` and `objectID` pre-populated. This is the **preferred constructor** because it avoids the extra `DOM.resolveNode` call. `Page.Find` uses this constructor when it can obtain both IDs during the initial query.

### Why Two Constructors?

Some CDP commands return only a `nodeId` (e.g., `DOM.querySelector`). Others return both `nodeId` and the resolved `objectId`. Rather than always making a second call, the code provides two constructors to handle both cases efficiently.

---

## 3. `resolveObjectID()` — The Lazy Resolution Pattern

```go
func (e *Element) resolveObjectID() (string, error) {
    if e.objectID != "" {
        return e.objectID, nil  // cached — no I/O
    }
    result, err = e.page.sendCommand("DOM.resolveNode", map[string]interface{}{
        "nodeId": e.nodeID,
    })
    // Extract objectId from result["object"]["objectId"]
    e.objectID = objectID  // cache for future use
    return objectID, nil
}
```

### What Happens in Chrome

`DOM.resolveNode` tells Chrome: "Given this DOM `nodeId`, create a V8 JavaScript wrapper object for it and give me a handle (`objectId`)."

1. Chrome's DOM agent looks up the internal DOM node by `nodeId` in its `NodeId → Node*` map.
2. It creates (or reuses) a V8 wrapper object — this is a JavaScript `HTMLElement` (or `HTMLInputElement`, `HTMLDivElement`, etc.) backed by the C++ Blink DOM node.
3. Chrome's runtime agent assigns a `RemoteObjectId` to this wrapper and stores it in an `InjectedScript` table.
4. The `objectId` string is returned.

### Caching Strategy

Once resolved, the `objectID` is cached in `e.objectID`. All subsequent calls to `Click()`, `Type()`, `Hover()`, etc. skip the resolution step. This is critical for performance — a test that clicks, types, and checks visibility on the same element makes 3 `Runtime.callFunctionOn` calls but only 1 `DOM.resolveNode` call.

### Staleness

If the page navigates, all `objectId` handles become invalid (Chrome's V8 context is destroyed and recreated). Using a stale `objectID` will cause `Runtime.callFunctionOn` to fail with `"Cannot find context with specified id"`. The caller must re-find the element.

---

## 4. `Click()` — Immediate Click

```go
func (e *Element) Click() error {
    // 1. Nil checks
    if e == nil { return errors.ErrElementNil }
    if e.page == nil { return errors.ErrElementNoPage }
    if e.nodeID <= 0 && e.objectID == "" { return errors.ErrElementInvalidNodeID }

    // 2. Resolve objectID
    objectID, err = e.resolveObjectID()

    // 3. Execute click via JavaScript
    var clickFunction string = `function() { this.click(); return true; }`
    result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
        "functionDeclaration": clickFunction,
        "objectId":            objectID,
        "returnByValue":       true,
    })

    // 4. Verify success
    success, ok = result["result"].(map[string]interface{})["value"].(bool)
}
```

### What Happens in Chrome's Renderer

`Runtime.callFunctionOn` with `objectId` and `functionDeclaration`:

1. Chrome's V8 engine looks up the object by `objectId` in the `InjectedScript` table.
2. The function string is compiled by V8 into bytecode.
3. The function is called with `this` set to the DOM element.
4. `this.click()` triggers the following in Blink's event system:
   - **Focus** — The element receives focus (if focusable). The previously focused element loses focus.
   - **mousedown** event — Dispatched to the element and bubbles up the DOM tree.
   - **mouseup** event — Same bubbling behavior.
   - **click** event — Dispatched after mousedown+mouseup. This is what triggers `<a href>` navigation, `<button>` form submission, or any JavaScript `addEventListener("click", ...)` handlers.
5. The function returns `true`, which V8 serializes as a JSON boolean (since `returnByValue: true`).

### Why JavaScript `click()` Instead of CDP `Input.dispatchMouseEvent`?

`this.click()` is a **DOM-level click** — it fires the click event directly on the element, regardless of its position on screen. `Input.dispatchMouseEvent` is a **pixel-level click** — you specify x,y coordinates and Chrome simulates mouse movement + click at those coordinates. The DOM-level approach is:
- **Faster** — No need to compute element position via `getBoxModel`.
- **More reliable** — Works even if the element is partially obscured or outside the viewport.
- **Simpler** — No coordinate math.

The tradeoff: DOM-level clicks don't trigger hover/mousemove events, which some websites rely on. For those cases, you'd use `Hover()` first.

---

## 5. `WaitAndClick()` — Click With Visibility Wait

```go
func (e *Element) WaitAndClick() error {
    // ... nil checks ...
    _, err = e.page.WaitForElementClickable(e.selector, e.timeout)
    if err != nil { return fmt.Errorf("element not clickable: %w", err) }
    // ... resolve objectID + click (same as Click()) ...
}
```

### The Wait Loop

`page.WaitForElementClickable(selector, timeout)` polls Chrome repeatedly:

1. Query the DOM for the element by `selector`.
2. Check if it's visible (not `display:none`, not `visibility:hidden`, has non-zero dimensions).
3. Check if it's enabled (not `disabled` attribute on `<input>`/`<button>`).
4. If both conditions are true, return the element.
5. If not, sleep for `pollInterval` (typically 100ms) and retry.
6. If `timeout` expires, return an error.

### OS-Level Sleep Behavior

`time.Sleep(100 * time.Millisecond)` parks the goroutine in Go's scheduler. The goroutine is removed from the run queue and placed in a timer heap. After 100ms, the runtime's `sysmon` goroutine (or a timer interrupt) wakes it. During sleep:
- The goroutine uses **zero CPU**.
- The OS thread it was on is free to run other goroutines (Go's M:N scheduler).
- Memory is unchanged — the goroutine's stack (~2–8 KB) stays allocated.

---

## 6. `Type(text string)` — Character-by-Character Typing

```go
func (e *Element) Type(text string) error {
    // 1. Nil checks + empty text check
    if text == "" { return errors.ErrTextEmpty }

    // 2. Resolve objectID
    objectID, err = e.resolveObjectID()

    // 3. Focus and clear
    focusFunction := `function() { this.focus(); this.value = ''; return true; }`
    e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)

    // 4. Type each character: keyDown → char → keyUp
    for _, ch := range text {
        charStr := string(ch)
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyDown", "key":charStr})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"char", "text":charStr, ...})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyUp", "key":charStr})
    }
}
```

### Why Three Events Per Character? (The Playwright Pattern)

This mimics real keyboard behavior. When you physically press a key on your keyboard:

1. **`keyDown`** — The OS keyboard driver sends a "key pressed" event to Chrome. Chrome's input handler fires a `keydown` DOM event on the focused element. The `key` field identifies which key (e.g., `"a"`), but **no text is inserted yet**.

2. **`char`** — The OS input method (IME) resolves the key press into a character and sends a "character input" event. Chrome fires a `keypress` DOM event and inserts the character into the text field. The `text` field carries the actual character to insert.

3. **`keyUp`** — The OS sends a "key released" event. Chrome fires a `keyup` DOM event.

### Critical Detail: `keyDown` Must NOT Have `text`

If `keyDown` includes a `text` field, Chrome inserts the character twice — once on `keyDown` (because of `text`) and once on `char`. This is a common bug in CDP automation code. Kexas follows Playwright's pattern: `keyDown` has only `key` (no `text`), and `char` carries `text`.

### Performance

For a 20-character string, this sends **60 CDP commands** (3 per character). At ~0.5 ms per command, that's ~30 ms total. This is much slower than directly setting `element.value`, but it correctly triggers all JavaScript event handlers (`oninput`, `onchange`, etc.) that the page might rely on.

### Focus and Clear Step

Before typing, the code:
1. Calls `this.focus()` — Chrome's focus manager sets this element as the active element. This is necessary because `Input.dispatchKeyEvent` sends keys to the **currently focused element**, not to a specific element.
2. Sets `this.value = ''` — Clears any existing text. This is a direct property assignment that bypasses event handlers (intentional — we're about to trigger them properly via key events).

---

## 7. `Hover()` — Mouse Hover Simulation

```go
func (e *Element) Hover() error {
    objectID, err = e.resolveObjectID()
    return e.dispatchHoverEvent(objectID)
}

func (e *Element) dispatchHoverEvent(objectID string) error {
    var hoverFunction string = `
        function() {
            var event = new MouseEvent('mouseover', {
                'view': window, 'bubbles': true, 'cancelable': true
            });
            this.dispatchEvent(event);
            return true;
        }
    `
    result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)
}
```

### Why `dispatchEvent` Instead of `Input.dispatchMouseEvent`?

Similar to `Click()`, using JavaScript's `dispatchEvent` is more reliable than pixel-level mouse simulation:
- No need to compute the element's screen coordinates.
- Works regardless of viewport scroll position.
- Fires the same event listeners that real mouse movement would trigger.

### What `mouseover` Does in the Renderer

The `MouseEvent` with type `'mouseover'`:
1. Is dispatched on the target element.
2. **Bubbles** up the DOM tree (`bubbles: true`), triggering `mouseover` handlers on all ancestor elements.
3. May trigger CSS `:hover` pseudo-class styles (though dispatched events don't always trigger `:hover` in all browsers — this is a known limitation).
4. May trigger JavaScript handlers that show tooltips, dropdown menus, etc.

---

## 8. `GetAttribute(name string)` — Reading DOM Attributes

```go
func (e *Element) GetAttribute(name string) (string, error) {
    result, err = e.page.sendCommand(cdp.CmdDOMGetAttributes, map[string]interface{}{
        "nodeId": e.nodeID,
    })
    // Attributes are returned as alternating name/value pairs: ["class", "btn", "id", "login", ...]
    for i := 0; i < len(attributes)-1; i += 2 {
        if attributes[i] == name { return attributes[i+1], nil }
    }
    return "", nil  // attribute not found
}
```

### The Alternating Pairs Format

CDP's `DOM.getAttributes` returns a flat array, NOT a map. The format is: `[name1, value1, name2, value2, ...]`. This is a CDP design choice for efficiency — JSON arrays are cheaper to serialize than objects with dynamic keys.

The code iterates in steps of 2, comparing each name. This is O(n) where n is the number of attributes on the element (typically 2–10, so effectively O(1)).

### Note: Uses `nodeID`, Not `objectID`

Unlike `Click`/`Type`/`Hover` which use `Runtime.callFunctionOn` (requires `objectID`), `GetAttribute` uses `DOM.getAttributes` (requires `nodeID`). This is because `DOM.getAttributes` is a DOM-domain command that works directly with the DOM tree, not with JavaScript objects.

---

## 9. `GetText()` — Reading Visible Text

```go
func (e *Element) GetText() (string, error) {
    objectID, err = e.resolveObjectID()
    result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
        "functionDeclaration": `function() { return this.innerText || this.textContent || ''; }`,
        "objectId":            objectID,
        "returnByValue":       true,
    })
}
```

### `innerText` vs `textContent`

- **`innerText`** — Returns only **visible** text. Elements with `display:none` or `visibility:hidden` are excluded. It also respects CSS `text-transform` and collapses whitespace. Computing `innerText` may trigger a **layout reflow** in the renderer because it needs to know which elements are visible.
- **`textContent`** — Returns ALL text, including hidden elements. No layout reflow needed. Faster but less accurate for user-visible text.

The code tries `innerText` first (more accurate), falls back to `textContent` (broader compatibility), and finally returns `''` if both are falsy.

---

## 10. `ScrollIntoView()` — Scrolling to an Element

```go
func (e *Element) ScrollIntoView() error {
    objectID, err = e.resolveObjectID()
    scrollFunction := `
        function() {
            this.scrollIntoView({ behavior: 'instant', block: 'center', inline: 'center' });
            return true;
        }
    `
    e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)
}
```

### What Happens in the Renderer

`Element.scrollIntoView()` is a native DOM API:
1. Blink calculates the element's position relative to the viewport.
2. It computes the scroll offset needed to center the element (`block: 'center', inline: 'center'`).
3. It updates the scroll position of the nearest scrollable ancestor (usually `document.documentElement`).
4. `behavior: 'instant'` means no smooth animation — the scroll happens in one frame.
5. The compositor thread updates the scroll offset and repaints the visible area.

This is important before clicking elements that are off-screen — without scrolling, a pixel-level click would miss the element.

---

## 11. `Validate()` — Checking Element Liveness

```go
func (e *Element) Validate() error {
    result, err = e.page.sendCommand(cdp.CmdDOMDescribeNode, map[string]interface{}{
        "nodeId": e.nodeID,
    })
    // Check if the returned nodeId matches e.nodeID
}
```

### Why Elements Go Stale

DOM elements become stale when:
1. **Navigation** — The entire DOM tree is destroyed and rebuilt. All `nodeID` and `objectID` values are invalid.
2. **DOM mutation** — JavaScript removes the element from the DOM (`element.remove()` or `parent.removeChild(element)`). The `nodeID` may still be valid in Chrome's internal tracking, but the element is no longer in the document.
3. **DOM re-fetch** — If `DOM.getDocument` is called again (e.g., after `agentManager.ResetAgent(AgentDOM)`), Chrome reassigns node IDs.

`Validate` sends `DOM.describeNode` and checks if Chrome still recognizes the `nodeID` and returns the same ID back. If not, it returns `ErrElementInvalidNodeID`.

---

## 12. The Three-Tier API Pattern

Every interaction method exists in three variants:

| Variant | Wait? | Timeout | Use Case |
|---------|-------|---------|----------|
| `Click()` | No | None | Element is guaranteed to be ready |
| `WaitAndClick()` | Yes | `e.timeout` (default) | Normal usage — waits for clickability |
| `WaitAndClickFor(timeout)` | Yes | Custom | Need longer/shorter wait than default |

This pattern (also used for `Type`/`Hover`) provides flexibility:
- **Immediate** methods are fastest (1 CDP call), for performance-critical code.
- **WaitAnd** methods are safest (poll + action), for general test code.
- **WaitAndFor** methods give timeout control, for slow-loading elements.

The `WaitAndFor` variants also validate the timeout parameter: `if timeout < 1*time.Second { return errors.TimeoutInvalidFormat(timeout) }`. This prevents accidentally passing milliseconds when seconds are expected.

---

## 13. Error Handling — Sentinel Error Pattern

All Element methods use sentinel errors from `errors/errors.go`:

| Error | Condition | Meaning |
|-------|-----------|---------|
| `ErrElementNil` | `e == nil` | Nil receiver — caller has a bug |
| `ErrElementNoPage` | `e.page == nil` | Element created without a page reference |
| `ErrElementInvalidNodeID` | `e.nodeID <= 0 && e.objectID == ""` | Neither node ID nor object ID available |
| `ErrTextEmpty` | `text == ""` | Empty string passed to Type/WaitAndType |
| `ErrClickOperationFailed` | `result value != true` | Chrome reported click failure |
| `ErrHoverOperationFailed` | `result value != true` | Chrome reported hover failure |
| `ErrScrollOperationFailed` | `result value != true` | Chrome reported scroll failure |

These sentinel errors enable `errors.Is()` checks in test code:
```go
err := element.Click()
if errors.Is(err, errors.ErrElementNil) {
    // handle specifically
}
```

---

## 14. Memory and Performance Summary

| Operation | CDP Commands | Typical Latency | Memory Impact |
|-----------|-------------|-----------------|---------------|
| `NewElement` | 0 | ~0 | ~200 bytes heap |
| `resolveObjectID` (first call) | 1 (`DOM.resolveNode`) | ~0.5 ms | Caches objectID string |
| `resolveObjectID` (cached) | 0 | ~1 ns | None |
| `Click()` | 1–2 | 0.5–2 ms | Transient JSON bytes |
| `WaitAndClick()` | 2–20 (poll + click) | 100 ms–30 s | Same as Click + poll overhead |
| `Type("hello")` | 16 (1 focus + 15 key events) | ~8 ms | Transient JSON bytes |
| `GetAttribute("class")` | 1 | ~0.5 ms | Attribute string copy |
| `Screenshot()` | 1 | 50–500 ms | ~2 MB temp (base64 + PNG) |
