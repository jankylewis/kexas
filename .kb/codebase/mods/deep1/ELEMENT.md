# Element Module — Deep1 Level (For Software Engineering Students)

Source: `kexas/element.go`, `element_interaction.go`, `element_hover.go`, `element_methods.go`, `element_scroll.go`

Every technical term is defined. FAQs per section.

---

## 1. What Is an Element?

An `Element` represents a single HTML element on a web page — a `<button>`, an `<input>`, a `<div>`, a `<link>`. After you find an element with `page.Find("#login")`, you get an `Element` object that lets you interact with it: click it, type text into it, read its attributes, check if it's visible.

### Key Terms

- **HTML Element** — A building block of a web page. Written as tags: `<button id="submit">Click</button>`. The browser parses these tags into a tree structure called the DOM.
- **Node** — Any item in the DOM tree. Elements are one type of node. Text content, comments, and the document itself are also nodes.
- **Attribute** — A name-value pair on an HTML tag: `<input type="text" id="username" class="form-field">` has three attributes: `type`, `id`, and `class`.

---

## 2. The `Element` Struct

```go
type Element struct {
    page     *Page
    selector string
    nodeID   cdp.NodeID   // an integer
    objectID string
    timeout  time.Duration
}
```

### `nodeID` vs `objectID` — The Most Important FAQ

> **FAQ: What is the difference between nodeID and objectID?**
>
> This is the most confusing part of CDP. Here's the clear explanation:
>
> **`nodeID` (integer)** — Chrome's internal DOM node identifier. Think of it as a row number in a spreadsheet. When Chrome builds the DOM tree, it assigns each node a sequential integer: the `<html>` element might be nodeID 1, `<head>` might be 2, `<body>` might be 3, etc.
>
> - **Pros**: Simple integer, used by DOM-domain commands (`DOM.querySelector`, `DOM.getAttributes`).
> - **Cons**: **Volatile** — node IDs are reassigned when Chrome re-inspects the DOM (after navigation, after `DOM.getDocument` is called again). A nodeID that was valid 5 seconds ago may now refer to a different element or be completely invalid.
>
> **`objectID` (string)** — A JavaScript-level handle to the element. It's like a ticket number at a deli counter — you hand it back to Chrome and say "give me the object associated with this ticket." The objectID looks like `{"injectedScriptId":2,"id":15}`.
>
> - **Pros**: Stable within a page session. Used by Runtime-domain commands (`Runtime.callFunctionOn`) which let you call JavaScript functions directly on the element.
> - **Cons**: Invalidated when the page navigates (the JavaScript context is destroyed). More complex than an integer.
>
> **When is each used?**
> - `nodeID` → `DOM.getAttributes`, `DOM.describeNode` (DOM-domain commands)
> - `objectID` → `Runtime.callFunctionOn("this.click()")` (Runtime-domain commands — all interactions)
>
> **Analogy**: `nodeID` is like a seat number in a movie theater (can change if people reshuffle). `objectID` is like a wristband with a barcode (stable for the event, invalid after).

### Other Fields

**`page *Page`** — Back-pointer to the Page that contains this element. The Element uses this to send CDP commands. If the page is closed, the element becomes useless.

**`selector string`** — The CSS selector used to find this element (e.g., `"#login"`, `".submit-btn"`). Stored for two reasons:
1. **Debugging** — Error messages include the selector: `"element #login not clickable after 10s"`.
2. **Re-querying** — `WaitAndClick` needs to re-find the element to check visibility, using this stored selector.

**`timeout time.Duration`** — Default timeout for wait operations (typically 10–30 seconds).

> **What is `time.Duration`?** Go's type for representing a length of time. Internally, it's an `int64` counting nanoseconds. `10 * time.Second` = 10,000,000,000 nanoseconds. Common constants: `time.Second`, `time.Millisecond`, `time.Minute`.

### FAQ: How much memory does an Element use?

About 200 bytes total: 8 bytes for the page pointer, ~16+36 bytes for the selector string, 8 bytes for nodeID, ~16+80 bytes for objectID, 8 bytes for timeout. Creating 100 elements costs ~20 KB — negligible.

---

## 3. `resolveObjectID()` — Lazy Resolution

```go
func (e *Element) resolveObjectID() (string, error) {
    if e.objectID != "" {
        return e.objectID, nil  // already cached
    }
    // Send CDP command to resolve
    result, err := e.page.sendCommand("DOM.resolveNode", {"nodeId": e.nodeID})
    e.objectID = extractedObjectID  // cache it
    return e.objectID, nil
}
```

### What "Lazy Resolution" Means

"Lazy" means "don't do work until it's actually needed." When an Element is first created by `page.Find()`, it may only have a `nodeID`. The first time you call `Click()`, `Type()`, or any interaction method, `resolveObjectID()` converts the `nodeID` into an `objectID` by asking Chrome: "Give me a JavaScript handle for DOM node #42."

After the first resolution, the `objectID` is **cached** in the struct. All subsequent calls return instantly (string comparison, no network I/O).

> **What is caching?** Storing a computed result so you don't have to recompute it. Like writing down a phone number instead of looking it up every time you want to call someone. The tradeoff: the cached value might become outdated (stale).

### FAQ: When does the cached objectID become stale?

When the page navigates to a new URL. Navigation destroys the JavaScript execution context, invalidating all objectIDs. After navigation, you must call `page.Find()` again to get a fresh Element.

### FAQ: What is `DOM.resolveNode`?

A CDP command that says: "I have a DOM node by its nodeID. Please create a JavaScript wrapper object for it and give me a handle (objectID) so I can call JavaScript functions on it." Chrome creates an `HTMLElement` JavaScript object backed by the C++ DOM node and returns a reference.

---

## 4. Click — How Clicking Works

### `Click()` — Immediate

```go
func (e *Element) Click() error {
    objectID, err := e.resolveObjectID()
    result, err := e.page.sendCommand("Runtime.callFunctionOn", {
        "functionDeclaration": "function() { this.click(); return true; }",
        "objectId": objectID,
        "returnByValue": true,
    })
}
```

### What `Runtime.callFunctionOn` Does

This CDP command says: "Take the JavaScript object identified by `objectID`, and call this function with `this` set to that object." It's like saying: "Hey Chrome, find the `<button>` element I pointed to earlier, and run `this.click()` on it."

> **What is `this` in JavaScript?** Inside a function, `this` refers to the object the function is being called on. When you write `element.click()`, `this` inside `click()` refers to `element`. CDP's `callFunctionOn` explicitly sets `this` to the DOM element identified by `objectID`.

### What `this.click()` Does in Chrome

When Chrome executes `this.click()` on an HTML element, it fires a sequence of DOM events:

1. **`focus`** — The element receives keyboard focus (if it's focusable).
2. **`mousedown`** — Simulates pressing the mouse button down.
3. **`mouseup`** — Simulates releasing the mouse button.
4. **`click`** — The actual click event. This is what triggers `<a href>` navigation, `<button>` form submission, and any `addEventListener("click", ...)` handlers.

> **What is a DOM event?** A notification that something happened. When a user clicks a button, the browser creates a `MouseEvent` object and "dispatches" it — sending it to the element and all its ancestors. JavaScript code can "listen" for events with `element.addEventListener("click", handler)`.

> **What is event bubbling?** After an event fires on the target element, it "bubbles up" to parent elements. A click on `<button>` also triggers click handlers on the `<div>` containing it, then `<body>`, then `<html>`, then `document`. This is why you can put a single click handler on a parent `<div>` and it catches clicks on all children.

### FAQ: Why use `this.click()` instead of simulating mouse coordinates?

`this.click()` is a **DOM-level** click — it fires click events directly on the element, regardless of where the element is on screen. The alternative (`Input.dispatchMouseEvent` with x,y coordinates) is a **pixel-level** click that simulates real mouse movement.

DOM-level is preferred because:
- No need to calculate element coordinates (saves one CDP call).
- Works even if the element is off-screen (scrolled out of view).
- More reliable — no issues with overlapping elements blocking the click target.

### `WaitAndClick()` — With Visibility Wait

```go
func (e *Element) WaitAndClick() error {
    _, err := e.page.WaitForElementClickable(e.selector, e.timeout)
    // ... then click ...
}
```

Before clicking, this method **polls** Chrome every 100ms: "Is the element visible? Is it enabled (not `disabled`)?" Once both conditions are true, it proceeds with the click.

> **What does "enabled" mean for an HTML element?** An `<input>` or `<button>` can have a `disabled` attribute: `<button disabled>Can't click me</button>`. Disabled elements don't respond to clicks. `WaitAndClick` waits until the `disabled` attribute is removed.

### `WaitAndClickFor(timeout)` — Custom Timeout

Same as `WaitAndClick` but you specify the timeout:

```go
elem.WaitAndClickFor(30 * time.Second)  // wait up to 30 seconds
```

There's a guard: `if timeout < 1*time.Second` returns an error. This prevents a common mistake — passing milliseconds when seconds are expected.

### FAQ: When should I use Click() vs WaitAndClick()?

- **`Click()`** — When you KNOW the element is ready (e.g., you just did a `WaitAndFind` that already confirmed visibility).
- **`WaitAndClick()`** — Default choice for test code. Handles timing automatically.
- **`WaitAndClickFor(t)`** — When the element is expected to take longer than the default timeout (e.g., waiting for an AJAX request to complete).

---

## 5. Type — Character-by-Character Typing

### `Type(text string)` — How Keyboard Input Works

```go
func (e *Element) Type(text string) error {
    // 1. Focus the element and clear existing text
    e.page.sendCommand("Runtime.callFunctionOn", {
        "functionDeclaration": "function() { this.focus(); this.value = ''; return true; }",
        "objectId": objectID,
    })
    
    // 2. Type each character
    for _, ch := range text {
        charStr := string(ch)
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyDown", "key":charStr})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"char", "text":charStr})
        e.page.sendCommand("Input.dispatchKeyEvent", {"type":"keyUp", "key":charStr})
    }
}
```

### Why Three Events Per Character?

When you physically press a key on your keyboard, the operating system generates three separate events:

1. **`keyDown`** — "A key was pressed down." The browser fires a `keydown` DOM event. **No text is inserted yet** — this is just the notification that a key is being held.

2. **`char`** — "A character should be typed." The browser fires a `keypress` event and inserts the character into the text field. This is where `input.value` actually changes.

3. **`keyUp`** — "The key was released." The browser fires a `keyup` DOM event.

> **FAQ: Why does kexas type one character at a time instead of pasting the whole string?**
> Many web applications use JavaScript event handlers on `keydown`, `keypress`, `keyup`, and `input` events. For example, React's `onChange` handler fires on each `input` event. If we just set `element.value = "text"` directly, those handlers would never fire, and the application might not update its state correctly. Character-by-character typing ensures all event handlers fire properly.

> **FAQ: Why does `keyDown` NOT include the `text` field?**
> If `keyDown` included `text`, Chrome would insert the character twice — once during `keyDown` processing and once during `char` processing. This is a known CDP behavior. Kexas follows Playwright's pattern: `keyDown` has only `key` (which key was pressed), and `char` carries `text` (what character to insert).

### Performance

Each character requires 3 CDP commands. For "hello" (5 characters), that's 15 commands + 1 focus command = 16 total. At ~0.5ms per command, typing "hello" takes about 8ms. Typing a 100-character password takes ~150ms.

---

## 6. Hover — Mouse Hover Simulation

```go
func (e *Element) Hover() error {
    objectID, err := e.resolveObjectID()
    // Dispatch a mouseover event
    e.page.sendCommand("Runtime.callFunctionOn", {
        "functionDeclaration": `function() {
            var event = new MouseEvent('mouseover', {view:window, bubbles:true, cancelable:true});
            this.dispatchEvent(event);
            return true;
        }`,
        "objectId": objectID,
    })
}
```

### What Is `dispatchEvent`?

`element.dispatchEvent(event)` manually fires an event on an element. It's like pressing a doorbell button programmatically instead of physically.

> **What is a MouseEvent?** A JavaScript object representing a mouse action. `new MouseEvent('mouseover', {bubbles:true})` creates a "mouse moved over this element" event. The `bubbles:true` option means the event also fires on parent elements (event bubbling).

### FAQ: Does this trigger CSS `:hover` styles?

**Partially.** JavaScript `mouseover` events fire correctly, so JavaScript-based hover effects (tooltips, dropdown menus) work. But CSS `:hover` pseudo-class styles may NOT activate because Chrome's internal hover tracking is separate from JavaScript events. For most automation needs, this is fine.

---

## 7. ScrollIntoView — Making Elements Visible

```go
func (e *Element) ScrollIntoView() error {
    e.page.sendCommand("Runtime.callFunctionOn", {
        "functionDeclaration": `function() {
            this.scrollIntoView({behavior:'instant', block:'center', inline:'center'});
            return true;
        }`,
    })
}
```

> **What is `scrollIntoView`?** A native JavaScript method that scrolls the page (or a scrollable container) until the element is visible in the viewport. `behavior:'instant'` means no smooth animation — just jump immediately. `block:'center'` centers the element vertically.

> **What is the viewport?** The visible area of the web page in the browser window. If a page is 3000 pixels tall but your window is only 720 pixels tall, only 720 pixels are in the viewport at any time. The rest is off-screen until you scroll.

### FAQ: Do I need to scroll before clicking?

**No** — kexas uses `this.click()` which is a DOM-level click that works regardless of viewport position. But scrolling is useful before taking screenshots (to ensure the element appears in the captured image).

---

## 8. GetAttribute / GetText — Reading Element Data

### GetAttribute

```go
value, err := elem.GetAttribute("href")
// Returns the attribute value, or "" if not found
```

Uses `DOM.getAttributes`, which returns attributes as alternating name-value pairs: `["class", "btn", "id", "login", "href", "/dashboard"]`. The code iterates in steps of 2 to find the matching name.

> **FAQ: What's the difference between an attribute and a property?**
> - **Attribute** — What's written in the HTML source: `<input value="hello">`. The attribute is `"hello"`.
> - **Property** — The current runtime value in JavaScript: `element.value`. If the user types "world" into the input, the property changes to `"world"`, but the attribute stays `"hello"`.
> - `GetAttribute` reads the HTML attribute. To read the property, use `page.Evaluate("document.querySelector('#input').value")`.

### GetText

```go
text, err := elem.GetText()
// Returns the visible text content
```

Executes `this.innerText || this.textContent || ''`.

> **FAQ: What's the difference between `innerText` and `textContent`?**
> - **`innerText`** — Returns only **visible** text. Elements with `display:none` are excluded. May trigger a layout calculation (slower).
> - **`textContent`** — Returns ALL text, including hidden elements. Faster because no layout needed.
> - Kexas tries `innerText` first (more accurate for what the user sees), falls back to `textContent`.

---

## 9. Validate — Checking If an Element Is Still Valid

```go
func (e *Element) Validate() error {
    result, err := e.page.sendCommand("DOM.describeNode", {"nodeId": e.nodeID})
}
```

> **FAQ: Why would an element become invalid?**
> 1. **Page navigation** — The entire DOM is destroyed and rebuilt. All nodeIDs and objectIDs become stale.
> 2. **JavaScript removes it** — `element.remove()` or `parent.removeChild(element)` in page JavaScript.
> 3. **Dynamic page update** — A React/Angular/Vue component re-renders, replacing the old DOM node with a new one.
>
> After any of these, calling `Click()` or `Type()` on the old Element will fail. Use `Validate()` to check, or simply call `page.Find()` again to get a fresh Element.

---

## 10. The Three-Tier Pattern — Summary

| Method | Waits? | Timeout | Best For |
|--------|--------|---------|----------|
| `Click()` | No | N/A | Element is guaranteed ready |
| `WaitAndClick()` | Yes | Default (10-30s) | Normal test code |
| `WaitAndClickFor(t)` | Yes | Custom `t` | Slow-loading elements |

This same pattern applies to `Type`/`WaitAndType`/`WaitAndTypeFor` and `Hover`/`WaitAndHover`/`WaitAndHoverFor`.

### FAQ: Why three tiers instead of just one method with an optional timeout?

Explicitness. When you read test code and see `Click()`, you immediately know "no waiting." When you see `WaitAndClick()`, you know "it polls first." When you see `WaitAndClickFor(60*time.Second)`, you know "it's expected to be slow." This makes test intent clear at a glance.
