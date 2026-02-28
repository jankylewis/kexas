# Element Interactions — Deep Detailed Walkthrough

Source: `kexas/element_interaction.go`, `element_hover.go`, `element_scroll.go`

This document explains every interaction method at the CDP protocol, Chrome renderer, JavaScript engine, and OS level. It covers Click, Type, Hover, and Scroll — the four fundamental ways kexas simulates user input.

---

## 1. The Interaction Pipeline — Common Architecture

Every interaction method in kexas follows a 5-stage pipeline:

```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Validation  │───►│ Object       │───►│ JavaScript   │───►│ Chrome       │───►│ Result       │
│  (Go side)   │    │ Resolution   │    │ Execution    │    │ Event System │    │ Extraction   │
│              │    │ (CDP call)   │    │ (CDP call)   │    │ (renderer)   │    │ (Go side)    │
└─────────────┘    └──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

### Stage 1: Validation (Pure Go, ~1 ns)

```go
if e == nil { return errors.ErrElementNil }
if e.page == nil { return errors.ErrElementNoPage }
if e.nodeID <= 0 && e.objectID == "" { return errors.ErrElementInvalidNodeID }
```

Three nil/zero checks. No I/O, no allocations. These are compile-time-known error values (package-level `var`), so returning them doesn't allocate.

### Stage 2: Object Resolution (CDP, ~0.5 ms first time, ~1 ns cached)

```go
objectID, err = e.resolveObjectID()
```

If `e.objectID` is already cached (non-empty string), this is a simple string return. If not, sends `DOM.resolveNode` to Chrome:

**Wire format sent:**
```json
{"id":42, "method":"DOM.resolveNode", "params":{"nodeId":15}, "sessionId":"abc-123"}
```

**Chrome's response:**
```json
{"id":42, "result":{"object":{"type":"object", "subtype":"node", "className":"HTMLInputElement", "objectId":"{\"injectedScriptId\":2,\"id\":15}"}}}
```

The `objectId` is a JSON string encoding an internal reference. It's opaque to the client — you never parse it, just pass it back in subsequent calls.

### Stage 3: JavaScript Execution (CDP, ~0.3–2 ms)

The interaction-specific JavaScript function is sent via `Runtime.callFunctionOn`:

```json
{"id":43, "method":"Runtime.callFunctionOn", "params":{
    "functionDeclaration": "function() { this.click(); return true; }",
    "objectId": "{\"injectedScriptId\":2,\"id\":15}",
    "returnByValue": true
}, "sessionId":"abc-123"}
```

Chrome's V8 engine:
1. Looks up the object by `objectId` in its `InjectedScript` table.
2. Compiles the function string to bytecode.
3. Calls the function with `this` bound to the DOM element.
4. Serializes the return value (since `returnByValue: true`).

### Stage 4: Chrome Event System (Renderer internal, ~0.01–1 ms)

The JavaScript call (e.g., `this.click()`) triggers Chrome's internal event dispatch:
1. Blink creates a synthetic `MouseEvent` (or `KeyboardEvent` for typing).
2. The event is dispatched through the DOM event phases: **capture** → **target** → **bubble**.
3. Event handlers registered on the element and its ancestors fire.
4. If the event causes DOM mutations (e.g., click on a link navigates), those mutations begin.

### Stage 5: Result Extraction (Go, ~50 ns)

```go
var resultObj map[string]interface{} = result["result"].(map[string]interface{})
var success bool
var ok bool
success, ok = resultObj["value"].(bool)
if !ok || !success {
    return errors.ErrClickOperationFailed
}
```

Two type assertions on the response map. No allocations (the map was already allocated during JSON decoding).

---

## 2. Click — Deep Protocol Walkthrough

### `Click()` — Immediate DOM-Level Click

```go
func (e *Element) Click() error {
    // validation...
    objectID, err := e.resolveObjectID()
    var clickFunction string = `function() { this.click(); return true; }`
    result, err := e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
        "functionDeclaration": clickFunction,
        "objectId":            objectID,
        "returnByValue":       true,
    })
}
```

#### `this.click()` — What Blink Does Internally

When V8 executes `this.click()` on an `HTMLElement`:

1. **`HTMLElement::click()` C++ method** is called (defined in `blink/renderer/core/html/html_element.cc`).
2. It creates a `MouseEvent` with:
   - `type: "click"`
   - `bubbles: true`
   - `cancelable: true`
   - `composed: true`
   - `button: 0` (left click)
   - `clientX/clientY: 0, 0` (synthetic — no real mouse position)
3. **Event dispatch** follows the DOM Event flow:
   - **Capture phase**: Walk from `Document` → `<html>` → ... → parent of target. Any `addEventListener(..., {capture: true})` handlers fire.
   - **Target phase**: Fire handlers on the element itself.
   - **Bubble phase**: Walk from parent → ... → `<html>` → `Document`. Any bubble-phase handlers fire.
4. **Default action**: If the event wasn't cancelled (`preventDefault()`):
   - `<a href="...">` — Triggers navigation.
   - `<button type="submit">` — Submits the parent form.
   - `<input type="checkbox">` — Toggles the checked state.
   - `<input type="radio">` — Selects the radio button.
   - `<select>` — Opens/closes the dropdown (may not work via synthetic click).

#### Limitations of `this.click()`

- **No `mouseover`/`mouseenter` events** — These fire on real mouse movement, not on `click()`.
- **No coordinate information** — `clientX`/`clientY` are 0. Websites that read click coordinates (e.g., for analytics or canvas interactions) will get wrong values.
- **No focus change** — `click()` does fire a `focus` event, but some frameworks (React, Angular) may ignore synthetic focus.
- **CSS `:active` pseudo-class** — Not triggered because there's no real mousedown/mouseup sequence.

For most web automation (form submission, button clicks, link navigation), these limitations don't matter. For edge cases, use `Hover()` + coordinate-based clicking.

### `WaitAndClick()` — With Visibility Polling

```go
func (e *Element) WaitAndClick() error {
    // validation...
    _, err := e.page.WaitForElementClickable(e.selector, e.timeout)
    if err != nil { return fmt.Errorf("element not clickable after waiting: %w", err) }
    // ... resolve + click (same as Click()) ...
}
```

#### `WaitForElementClickable` — The Poll Loop

This method repeatedly queries Chrome until the element is both **visible** and **enabled**:

```
Iteration 1 (t=0ms):     querySelector("#login") → found → isVisible? NO (display:none) → sleep 100ms
Iteration 2 (t=100ms):   querySelector("#login") → found → isVisible? NO (still loading) → sleep 100ms
Iteration 3 (t=200ms):   querySelector("#login") → found → isVisible? YES → isEnabled? YES → return!
```

Each iteration sends 1–3 CDP commands:
1. `DOM.querySelector` — Find the element.
2. `Runtime.callFunctionOn` — Check visibility (executes JS like `getComputedStyle(this).display !== 'none'`).
3. Optionally `Runtime.callFunctionOn` — Check `disabled` attribute.

At 100ms poll interval, this costs ~3 CDP round-trips per iteration ≈ 3–6 ms of actual work per 100ms. The remaining ~94–97ms is sleep time.

### `WaitAndClickFor(timeout)` — Custom Timeout Variant

```go
func (e *Element) WaitAndClickFor(timeout time.Duration) error {
    if timeout < 1*time.Second {
        return errors.TimeoutInvalidFormat(timeout)
    }
    // ... same as WaitAndClick but uses provided timeout ...
}
```

The `< 1 second` guard prevents a common mistake: passing milliseconds when seconds are expected (`WaitAndClickFor(500)` would mean 500 nanoseconds — essentially instant timeout). By requiring at least 1 second, accidental unit errors are caught.

---

## 3. Type — Character-by-Character Input Simulation

### `Type(text string)` — Immediate Typing

```go
func (e *Element) Type(text string) error {
    if text == "" { return errors.ErrTextEmpty }
    objectID, err := e.resolveObjectID()

    // Step 1: Focus + clear
    var focusFunction string = `function() { this.focus(); this.value = ''; return true; }`
    e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)

    // Step 2: Type each character
    for _, ch := range text {
        charStr := string(ch)
        // keyDown (no text!)
        e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
            "type": "keyDown", "key": charStr,
        })
        // char (has text)
        e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
            "type": "char", "text": charStr, "key": charStr, "unmodifiedText": charStr,
        })
        // keyUp
        e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
            "type": "keyUp", "key": charStr,
        })
    }
}
```

### The Three-Event Model — Why Each Exists

When you physically press and release a key on your keyboard, the OS generates this sequence:

```
Physical keyboard → OS keyboard driver → Chrome input handler → DOM events
                                                                  │
                                                     ┌────────────┼────────────┐
                                                     │            │            │
                                                  keyDown       char        keyUp
                                                     │            │            │
                                               "key pressed"  "text input"  "key released"
                                               (no text yet)  (character    (cleanup)
                                                               inserted)
```

#### `keyDown` — Key Press Without Text

```json
{"type": "keyDown", "key": "a"}
```

Chrome fires a `keydown` DOM event. At this point:
- `event.key` is `"a"`.
- `event.code` would be `"KeyA"` (physical key location).
- **No text is inserted** into the input field.
- JavaScript handlers can call `event.preventDefault()` to block the character.

**Why no `text` field?** If `keyDown` included `text`, Chrome would insert the character twice — once during `keyDown` processing and once during `char` processing. This is a known CDP behavior that trips up many automation libraries.

#### `char` — Character Input

```json
{"type": "char", "text": "a", "key": "a", "unmodifiedText": "a"}
```

Chrome fires a `keypress` DOM event (deprecated but still present) and inserts the character into the input field's value. This is where:
- `input.value` changes from `""` to `"a"`.
- The `input` DOM event fires (React's `onChange` depends on this).
- The cursor position advances by 1.

#### `keyUp` — Key Release

```json
{"type": "keyUp", "key": "a"}
```

Chrome fires a `keyup` DOM event. Most web apps ignore `keyup`, but some use it for:
- Debounced search (start searching when the user lifts their finger).
- Game controls (stop moving when key is released).
- Keyboard shortcuts (detect key combinations by tracking press/release state).

### Focus and Clear — Why Both Are Needed

```go
`function() { this.focus(); this.value = ''; return true; }`
```

1. **`this.focus()`** — Sets this element as the active element in Chrome's focus manager. `Input.dispatchKeyEvent` always sends keys to the **currently focused element**, not to a specific element. Without focus, keys go to `<body>` or the last focused element.

2. **`this.value = ''`** — Clears existing text. This is a **direct property assignment** that bypasses `input`/`change` events. This is intentional — the subsequent character-by-character typing will fire all the correct events. If we used `this.value = ''` followed by `this.dispatchEvent(new Event('input'))`, frameworks might see the clear-and-refill as two separate operations.

### Unicode and Multi-Byte Characters

The `for _, ch := range text` loop iterates over **runes** (Unicode code points), not bytes. This means:
- `"hello"` → 5 iterations (5 ASCII characters, 5 bytes).
- `"café"` → 4 iterations (4 code points, but `é` is 2 UTF-8 bytes).
- `"日本語"` → 3 iterations (3 CJK characters, 9 UTF-8 bytes).

Each rune is converted to a Go `string` via `string(ch)`, which may produce a multi-byte UTF-8 string. Chrome's Input domain handles UTF-8 correctly.

### Performance Characteristics

| Text Length | CDP Commands | Est. Latency |
|------------|-------------|--------------|
| 1 char | 4 (focus + 3 key events) | ~2 ms |
| 10 chars | 31 (focus + 30 key events) | ~15 ms |
| 50 chars | 151 (focus + 150 key events) | ~75 ms |
| 100 chars | 301 (focus + 300 key events) | ~150 ms |

For long text (passwords, URLs), this per-character approach is slow but correct. If event triggering doesn't matter, you could use `this.value = 'text'` directly (single CDP command), but kexas intentionally doesn't do this because it would skip JavaScript event handlers.

---

## 4. Hover — Mouse Event Simulation

### `Hover()` — Immediate Hover

```go
func (e *Element) Hover() error {
    objectID, err := e.resolveObjectID()
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
    result, err := e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)
}
```

### `MouseEvent('mouseover')` — What Chrome Does

When the JavaScript `new MouseEvent('mouseover', ...)` is created and dispatched:

1. **Event creation** — V8 creates a `MouseEvent` object in the JavaScript heap. The `view: window` sets the event's associated `Window` object. `bubbles: true` means the event will propagate up the DOM tree.

2. **`dispatchEvent(event)`** — Blink's event dispatcher:
   a. **Capture phase** — Walks from `Window` → `Document` → ... → parent of target.
   b. **Target phase** — Fires on the element itself.
   c. **Bubble phase** — Walks back up to `Document` → `Window`.

3. **Hover state** — CSS `:hover` pseudo-class may or may NOT activate. Synthetic `mouseover` events from JavaScript don't always update Chrome's internal hover state (which is maintained by the compositor's hit-testing system). This means:
   - **JavaScript `mouseover` handlers** — WILL fire.
   - **CSS `:hover` rules** — May NOT apply (depends on Chrome version and rendering pipeline).
   - **Tooltips** — If implemented via JavaScript `mouseover` handlers, they will appear. If implemented via CSS `:hover` or native `title` attribute, they may not.

### Why Not `Input.dispatchMouseEvent`?

`Input.dispatchMouseEvent` simulates pixel-level mouse movement:

```json
{"method": "Input.dispatchMouseEvent", "params": {
    "type": "mouseMoved", "x": 450, "y": 320
}}
```

This would correctly trigger CSS `:hover` because it goes through Chrome's compositor and hit-testing pipeline. But it requires:
1. Computing the element's bounding box (`DOM.getBoxModel`) — 1 extra CDP call.
2. Calculating the center coordinates — trivial math but adds code complexity.
3. Handling scrolled elements — coordinates must be relative to the viewport, not the document.

The JavaScript `dispatchEvent` approach avoids all of this and is sufficient for most automation needs (dropdown menus, tooltip triggers, etc.).

### `WaitAndHover()` / `WaitAndHoverFor(timeout)`

Same pattern as WaitAndClick — polls for visibility before hovering. Uses `WaitForElementVisible` instead of `WaitForElementClickable` (hover doesn't require the element to be enabled).

---

## 5. ScrollIntoView — Viewport Management

### `ScrollIntoView()` — Scroll to Element

```go
func (e *Element) ScrollIntoView() error {
    objectID, err := e.resolveObjectID()
    var scrollFunction string = `
        function() {
            this.scrollIntoView({ behavior: 'instant', block: 'center', inline: 'center' });
            return true;
        }
    `
    e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, ...)
}
```

### What `scrollIntoView` Does in the Renderer

1. **Layout calculation** — Blink computes the element's position in the document coordinate space. This may trigger a **layout reflow** if the layout is dirty (pending CSS changes haven't been computed yet). A layout reflow walks the entire render tree and can take 1–100 ms on complex pages.

2. **Scroll offset computation** — Given `block: 'center'`, Blink calculates the scroll offset needed to vertically center the element in the viewport. Given `inline: 'center'`, same for horizontal centering.

3. **Scroll execution** — With `behavior: 'instant'`:
   - The scroll container's `scrollTop` / `scrollLeft` is updated immediately.
   - No animation frames — the change happens in one compositor tick.
   - A `scroll` DOM event fires on the scroll container.

4. **Repaint** — The compositor thread updates the visible layer tiles. Only tiles that are now visible (and weren't before) need to be rasterized. Chrome's tile-based rendering means only the newly visible area is rendered — the rest stays cached.

### Why `behavior: 'instant'`?

`behavior: 'smooth'` would animate the scroll over ~300–500ms. For automation, this is wasted time. `'instant'` completes in a single frame (~16ms at 60fps), making tests faster.

### When to Use ScrollIntoView

- **Before pixel-level clicks** — If using `Input.dispatchMouseEvent` with coordinates, the element must be in the viewport.
- **Before screenshots** — To ensure the element is visible in the capture.
- **After page load** — Elements at the bottom of long pages aren't in the viewport until scrolled.

For kexas's `Click()` (which uses `this.click()`), scrolling is NOT required because DOM-level clicks work regardless of viewport position. But for visual verification and screenshots, scrolling matters.

---

## 6. The Wait Pattern — Unified Polling Architecture

All `WaitAnd*` methods use the same polling pattern:

```
Start ───► Query DOM ───► Check condition ───► Condition met? ─── YES ──► Execute action ──► Return
                                │                                  │
                                NO                                 │
                                │                                  │
                          time.Sleep(100ms)                        │
                                │                                  │
                          Timeout expired? ─── YES ──► Return error│
                                │                                  │
                                NO ◄───────────────────────────────┘
                                │
                          Loop back to Query DOM
```

### Poll Interval Tradeoff

The 100ms poll interval is a tradeoff:
- **Too short** (10ms) — Excessive CDP commands. 100 commands/second per element, wasting CPU on both Go and Chrome sides. Network loopback overhead dominates.
- **Too long** (1s) — Slow tests. If an element becomes visible 50ms after the first poll, you wait 950ms unnecessarily.
- **100ms** — Good balance. Worst case: 100ms wasted. Average case: 50ms wasted. CDP cost: ~10 commands/second.

### Timeout Enforcement

```go
var deadline time.Time = time.Now().Add(timeout)
for time.Now().Before(deadline) {
    // poll...
    time.Sleep(100 * time.Millisecond)
}
return fmt.Errorf("element %s not found after %v", selector, timeout)
```

`time.Now()` calls the OS clock (`clock_gettime(CLOCK_MONOTONIC)` on Linux/macOS). Monotonic clock is used automatically by Go's `time` package — it's immune to NTP adjustments and system clock changes.

---

## 7. Error Propagation — How Failures Flow

```
Element.WaitAndClick()
  └── WaitForElementClickable() → timeout error
        └── fmt.Errorf("element not clickable after waiting: %w", err)
              └── caller receives: "element not clickable after waiting: timeout waiting for #login after 10s"
```

All errors are wrapped with `%w`, creating an error chain. The caller can:
- **Print it** — Gets the full chain: `element not clickable after waiting: timeout waiting for #login after 10s`.
- **Check root cause** — `errors.Is(err, ErrTimeout)`.
- **Extract details** — `errors.As(err, &timeoutErr)`.

### Sentinel Errors vs Wrapped Errors

| Error | Type | When |
|-------|------|------|
| `ErrElementNil` | Sentinel | `e == nil` |
| `ErrElementNoPage` | Sentinel | `e.page == nil` |
| `ErrTextEmpty` | Sentinel | `text == ""` in Type |
| `ErrClickOperationFailed` | Sentinel | Chrome returned `false` |
| `"element not clickable..."` | Wrapped | WaitAndClick timeout |
| `"failed to resolve object..."` | Wrapped | DOM.resolveNode failed |

Sentinel errors are package-level `var` values — comparing them is a pointer comparison (`==`), which is O(1).

---

## 8. Interaction Summary Matrix

| Method | CDP Commands | Requires Visibility? | Triggers DOM Events? | Triggers CSS :hover? |
|--------|-------------|---------------------|---------------------|---------------------|
| `Click()` | 1–2 | No | click, focus, mousedown, mouseup | No |
| `WaitAndClick()` | 3–20+ | Yes (polls) | Same as Click | No |
| `Type("abc")` | 10 | No | keydown, keypress, keyup, input, change | No |
| `WaitAndType("abc")` | 12–30+ | Yes (polls) | Same as Type | No |
| `Hover()` | 1–2 | No | mouseover | Partial |
| `WaitAndHover()` | 3–20+ | Yes (polls) | Same as Hover | Partial |
| `ScrollIntoView()` | 1–2 | No | scroll | N/A |
| `GetAttribute("x")` | 1 | No | None | No |
| `GetText()` | 1–2 | No | None | No |
| `Validate()` | 1 | No | None | No |
