# Element Interactions — Deep1 Level (For SE Students)

Source: `kexas/element_interaction.go`, `element_hover.go`, `element_scroll.go`

Every term defined. FAQs per section.

---

## 1. What Are "Interactions"?

Interactions are the ways your Go code simulates human actions on a web page: clicking buttons, typing text into forms, hovering over menus, scrolling to elements. Under the hood, each interaction sends one or more commands to Chrome via CDP (Chrome DevTools Protocol), which Chrome translates into the same events that a real mouse click or key press would produce.

### Key Terms

- **CDP command** — A JSON message sent to Chrome over WebSocket. Example: `{"method":"Runtime.callFunctionOn","params":{...}}`.
- **DOM event** — A notification that something happened to an HTML element. When you click a button, Chrome creates a `MouseEvent` and dispatches it to the element and its ancestors. JavaScript code can listen for these events.
- **WebSocket** — A persistent, bidirectional communication channel between your Go program and Chrome. Unlike HTTP (request → response → done), WebSocket stays open for continuous messaging.

> **FAQ: Why not just modify the DOM directly?**
> You could set `input.value = "hello"` directly, but web applications listen for events (`keydown`, `input`, `change`) to update their internal state. React, Angular, and Vue all rely on these events. If you skip them by setting values directly, the app doesn't know anything changed. Kexas simulates events to trigger the same code paths a real user would.

---

## 2. The 5-Stage Pipeline

Every interaction follows the same pattern:

```
1. Validation → 2. Object Resolution → 3. JS Execution → 4. Chrome Events → 5. Result Check
```

### Stage 1: Validation (Go side, ~1 nanosecond)

```go
if e == nil { return errors.ErrElementNil }
if e.page == nil { return errors.ErrElementNoPage }
```

Simple nil checks. No network, no allocation.

> **What is a nil check?** In Go, `nil` means "no value" for pointers, interfaces, maps, slices, channels, and functions. Calling a method on a nil pointer causes a panic (crash). Checking `e == nil` prevents this.

### Stage 2: Object Resolution (CDP, ~0.5ms first time)

```go
objectID, err := e.resolveObjectID()
```

Converts the element's `nodeID` (Chrome's DOM integer) into an `objectID` (a JavaScript handle). Cached after the first call.

> **FAQ: What's the difference between nodeID and objectID?**
> - `nodeID` — An integer assigned by Chrome's DOM inspector. Used for DOM-domain commands. Volatile — changes when Chrome re-inspects the DOM.
> - `objectID` — A string handle to a JavaScript object. Used for Runtime-domain commands (like `callFunctionOn`). Stable within the same page session.
> - Think of nodeID as a seat number (can change) and objectID as a wristband barcode (stable for the event).

### Stage 3: JavaScript Execution (CDP, ~0.3–2ms)

Chrome's V8 engine executes a JavaScript function on the element:

```json
{"method": "Runtime.callFunctionOn", "params": {
    "functionDeclaration": "function() { this.click(); return true; }",
    "objectId": "{\"injectedScriptId\":2,\"id\":15}",
    "returnByValue": true
}}
```

> **What is V8?** Google's JavaScript engine (written in C++). It compiles JavaScript to machine code for high performance. Powers Chrome, Node.js, and Deno.

> **What is `Runtime.callFunctionOn`?** A CDP command that says: "Take the object identified by `objectId`, set it as `this`, and execute this JavaScript function." It's how kexas calls methods (`.click()`, `.focus()`, `.dispatchEvent()`) on specific DOM elements.

> **What does `returnByValue: true` mean?** It tells Chrome to serialize the return value and include it in the response. Without it, Chrome returns a reference to the return value (an objectId), which would require another round-trip to read.

### Stage 4: Chrome Event System (internal, ~0.01–1ms)

The JavaScript call triggers Chrome's internal event dispatch. Events propagate through three phases:

```
Document → html → body → div → button   (Capture phase: top-down)
                                button   (Target phase: the element itself)
button → div → body → html → Document   (Bubble phase: bottom-up)
```

> **What is event propagation?** When an event occurs on an element, it doesn't just fire on that element — it travels through the DOM tree. Three phases:
> 1. **Capture** — The event travels DOWN from the document root to the target element. Handlers registered with `{capture: true}` fire here.
> 2. **Target** — The event fires on the element itself.
> 3. **Bubble** — The event travels UP from the target back to the document root. Most handlers fire here (the default).
>
> This is why you can put one click handler on a parent `<div>` and catch clicks on all children — the event bubbles up.

### Stage 5: Result Check (Go side, ~50 nanoseconds)

```go
success, ok := resultObj["value"].(bool)
if !ok || !success { return errors.ErrClickOperationFailed }
```

> **What is a type assertion?** In Go, `x.(bool)` checks if `x` (an `interface{}` value) is actually a `bool`. If it is, you get the value. If not, it returns `false, false` (with the two-value form). This is needed because CDP responses are generic maps — Go doesn't know the types until runtime.

---

## 3. Click — Deep Dive

### `Click()` — How a DOM Click Works

```go
func (e *Element) Click() error {
    objectID, err := e.resolveObjectID()
    clickFunction := `function() { this.click(); return true; }`
    result, err := e.page.sendCommand("Runtime.callFunctionOn", map[string]interface{}{
        "functionDeclaration": clickFunction,
        "objectId":            objectID,
        "returnByValue":       true,
    })
}
```

When Chrome executes `this.click()`:

1. Chrome's Blink engine creates a synthetic `MouseEvent` with `type: "click"`.
2. Events fire in sequence: `focus` → `mousedown` → `mouseup` → `click`.
3. **Default actions** trigger based on element type:
   - `<a href="...">` → Navigation starts.
   - `<button type="submit">` → Form submits.
   - `<input type="checkbox">` → Checked state toggles.

> **What is a "synthetic" event?** An event created by code rather than by actual user input. `this.click()` creates a synthetic click — Chrome generates the event programmatically. The key difference: synthetic clicks have `event.isTrusted = false`, while real user clicks have `event.isTrusted = true`. Most web apps don't check this, but some security-sensitive sites do.

> **FAQ: Does this.click() move the mouse cursor?**
> No. `this.click()` is a DOM-level operation — it fires click events but doesn't simulate mouse movement. The `clientX`/`clientY` coordinates on the event are 0,0 (not the element's actual position). CSS `:hover` and `:active` pseudo-classes do NOT activate.

> **FAQ: When would this.click() fail?**
> - Element has been removed from the DOM (stale reference).
> - Element is inside a closed Shadow DOM (rare).
> - `event.preventDefault()` is called by a handler on a parent element (blocks the default action but the click still "fires").

### `WaitAndClick()` — The Polling Wait

Before clicking, polls Chrome every ~100ms: "Is the element visible and enabled?"

```
t=0ms:   querySelector("#btn") → found → visible? NO → sleep 100ms
t=100ms: querySelector("#btn") → found → visible? NO → sleep 100ms
t=200ms: querySelector("#btn") → found → visible? YES → enabled? YES → Click!
```

Each poll iteration sends 1-3 CDP commands (~1-3ms of actual work, ~97ms of sleep).

> **What is polling?** Repeatedly checking a condition at intervals. Like refreshing a tracking page to see if your package arrived. The alternative is event-driven (getting a notification). Polling is simpler but wastes some time sleeping between checks.

> **FAQ: Why 100ms intervals?**
> A tradeoff. Too short (10ms) = too many CDP commands, wastes CPU. Too long (1s) = slow tests. 100ms: worst case you wait 100ms extra after the element appears. Average overhead: 50ms.

### `WaitAndClickFor(timeout)` — Custom Timeout

Same as `WaitAndClick` but with a user-specified timeout. Has a `< 1 second` guard to prevent a common bug: passing milliseconds when seconds are expected (`WaitAndClickFor(500)` would be 500 nanoseconds, essentially instant).

---

## 4. Type — Character-by-Character Input

### How Typing Works

```go
func (e *Element) Type(text string) error {
    // Step 1: Focus + clear
    e.page.sendCommand("Runtime.callFunctionOn", {
        "functionDeclaration": "function() { this.focus(); this.value = ''; return true; }",
        "objectId": objectID,
    })
    
    // Step 2: Type each character
    for _, ch := range text {
        charStr := string(ch)
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyDown", "key":charStr})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"char", "text":charStr})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyUp", "key":charStr})
    }
}
```

### The Three Keyboard Events

When you physically press the "A" key:

| Event | What Happens | DOM Events Fired |
|-------|-------------|-----------------|
| `keyDown` | Key pressed down, no text yet | `keydown` |
| `char` | Character "a" inserted into field | `keypress`, `input` |
| `keyUp` | Key released | `keyup` |

> **FAQ: Why three events per character? Why not just set the value?**
> Web frameworks (React, Angular, Vue) listen for `input` and `keydown` events. React's `onChange` fires on the `input` event. Angular's `ngModel` updates on `input`. If you just set `element.value = "text"`, none of these events fire, and the app's internal state doesn't update.

> **FAQ: Why does keyDown NOT include the `text` field?**
> If keyDown included `text`, Chrome would insert the character twice — once during keyDown processing and once during the char event. This is a known CDP behavior.

> **FAQ: What about special keys like Enter, Tab, Escape?**
> Currently, kexas types printable characters only. To press Enter, you'd use `page.Evaluate("document.querySelector('#input').dispatchEvent(new KeyboardEvent('keydown', {key: 'Enter'}))")`. Support for special keys is a future feature.

### Focus and Clear

```js
this.focus(); this.value = '';
```

1. **`this.focus()`** — Makes this element the "active" element. `Input.dispatchKeyEvent` sends keys to whatever is focused. Without focus, keys go to `<body>`.

> **What is focus?** The currently "selected" element on a page. Only one element can have focus at a time. When you click a text input, it receives focus (the cursor appears). Focus determines which element receives keyboard events.

2. **`this.value = ''`** — Clears existing text via direct property assignment. This doesn't fire events — intentional, because the subsequent character-by-character typing fires all needed events.

### Unicode Support

`for _, ch := range text` iterates over Unicode **runes** (code points), not bytes:
- `"café"` → 4 runes: `c`, `a`, `f`, `é` (even though `é` is 2 bytes in UTF-8).
- `"日本語"` → 3 runes (even though each is 3 bytes in UTF-8).

> **What is a rune?** Go's name for a Unicode code point — a single "character" in the Unicode standard. Internally, it's an `int32`. Go's `range` on a string iterates by runes, not bytes.

> **What is UTF-8?** A variable-length encoding where ASCII characters use 1 byte, accented characters use 2 bytes, CJK characters use 3 bytes, and emojis use 4 bytes. Go strings are UTF-8 by default.

---

## 5. Hover — Mouse Event Simulation

```go
func (e *Element) Hover() error {
    objectID, _ := e.resolveObjectID()
    hoverFunction := `function() {
        var event = new MouseEvent('mouseover', {view:window, bubbles:true, cancelable:true});
        this.dispatchEvent(event);
        return true;
    }`
    e.page.sendCommand("Runtime.callFunctionOn", ...)
}
```

> **What is dispatchEvent?** A JavaScript method that manually fires an event on an element. `element.dispatchEvent(new MouseEvent('mouseover'))` is like moving your mouse over the element — but done programmatically.

> **What is MouseEvent?** A JavaScript constructor that creates a mouse event object. Parameters: event type (`mouseover`, `click`, `mousedown`), options (`bubbles`, `cancelable`, `view`).

> **FAQ: Does Hover() trigger CSS :hover styles?**
> **Partially.** JavaScript `mouseover` handlers fire correctly (dropdowns, tooltips). But CSS `:hover` pseudo-class may NOT activate because Chrome's hover tracking is separate from JavaScript events. For most automation (testing tooltip visibility, dropdown menus), this is fine.

---

## 6. ScrollIntoView — Viewport Management

```go
func (e *Element) ScrollIntoView() error {
    scrollFunction := `function() {
        this.scrollIntoView({behavior:'instant', block:'center', inline:'center'});
        return true;
    }`
    e.page.sendCommand("Runtime.callFunctionOn", ...)
}
```

> **What is the viewport?** The visible portion of the web page. If a page is 5000px tall and your browser window shows 720px, the viewport is 720px. Elements outside the viewport exist in the DOM but aren't visible.

> **What does `behavior:'instant'` mean?** No smooth scrolling animation — jump immediately. `'smooth'` would animate over ~300ms, which wastes time in tests.

> **What does `block:'center'` mean?** Center the element vertically in the viewport. Alternatives: `'start'` (top), `'end'` (bottom), `'nearest'` (minimum scroll).

> **FAQ: Do I need ScrollIntoView before Click?**
> No. `this.click()` is a DOM-level click that works regardless of viewport position. But scroll before screenshots (to make the element visible in the capture).

---

## 7. GetAttribute / GetText — Reading Data

### GetAttribute

```go
value, err := elem.GetAttribute("href")
```

Uses `DOM.getAttributes` which returns an alternating array: `["class","btn","id","login","href","/dashboard"]`. Kexas iterates in steps of 2 to find the name match.

> **FAQ: What's the difference between an HTML attribute and a JavaScript property?**
> - **Attribute**: What's in the HTML source: `<input value="hello">`. Read with `getAttribute("value")` → always `"hello"`.
> - **Property**: Current runtime value: `element.value`. Changes when the user types. After typing "world": property = `"world"`, attribute = still `"hello"`.
> - `GetAttribute` reads attributes (original HTML values). For current values, use `Evaluate`.

### GetText

```go
text, err := elem.GetText()
```

Evaluates: `this.innerText || this.textContent || ''`

> **FAQ: innerText vs textContent?**
> - `innerText` — Only visible text. Respects CSS (`display:none` text is excluded). May trigger a layout computation.
> - `textContent` — ALL text including hidden. Faster (no layout needed).
> - Kexas tries innerText first (more accurate for user-visible content), falls back to textContent.

---

## 8. Error Handling Summary

| Error | When | Action |
|-------|------|--------|
| `ErrElementNil` | Element is nil | Fix your Find() call |
| `ErrElementNoPage` | Page reference gone | Element is orphaned |
| `ErrTextEmpty` | `Type("")` called | Check your test data |
| `ErrClickOperationFailed` | Chrome returned false | Element may be detached |
| Timeout error | WaitAnd* timed out | Increase timeout or check element exists |

All errors use Go's `%w` wrapping for error chain inspection with `errors.Is()` and `errors.As()`.

> **What is error wrapping?** `fmt.Errorf("click failed: %w", originalErr)` creates a new error containing the original. Callers can unwrap: `errors.Is(err, ErrClickOperationFailed)` returns true even through multiple wrapping layers. Like nested envelopes — open each to find the original letter.
