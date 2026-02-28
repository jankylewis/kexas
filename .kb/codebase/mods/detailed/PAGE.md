# Page Module — Deep Detailed Walkthrough

Source: `kexas/page.go` and related files (`page_find.go`, `page_wait.go`)

This document explains every type, function, and internal mechanism in the Page module at the OS, process, memory, and protocol level.

---

## 1. The `Page` Struct — What Lives in Memory

```go
type Page struct {
    browser      *Browser              // back-pointer to the parent Browser
    targetID     string                // Chrome target UUID for this tab
    sessionID    string                // CDP session UUID attached to the target
    log          *logger.Logger        // structured logger scoped to "page"
    ctx          context.Context       // shared cancellation context
    agentManager *agent.AgentManager   // lazy CDP domain enabler
}
```

### Memory Layout

A `Page` is heap-allocated (returned as `*Page`). It holds:

- **`browser`** — A pointer back to the `Browser` struct. This is how `Page` accesses the WebSocket client (`browser.client`). It does NOT own the browser — multiple `Page` objects can share the same `Browser`. In Go, this is just an 8-byte pointer (on 64-bit systems).

- **`targetID`** — A string like `"7A3B2C1D-E4F5-6789-ABCD-EF0123456789"`. In Go, a string is a 16-byte header (pointer to backing byte array + length integer). The backing array (~36 bytes for a UUID) lives on the heap. This ID maps 1:1 to a Chrome renderer process — Chrome's browser process maintains an internal `TargetInfo` map keyed by this ID.

- **`sessionID`** — Another UUID string, same memory layout. The session ID is how CDP multiplexes commands over a single WebSocket. When `sendCommand` fires a CDP message, the JSON includes `"sessionId": "<this-value>"`, and Chrome routes it to the correct renderer.

- **`agentManager`** — A pointer to `agent.AgentManager`, which holds:
  - `EnabledAgents` struct: 8 boolean fields (DOM, Input, Runtime, etc.), each 1 byte.
  - `AgentStates` struct: 8 pointers to `AgentState` structs.
  - `AgentContext` struct: 8 pointers to domain-specific context structs.
  - A reference to the `cdp.Client` and the session ID.
  
  Total footprint: ~200–400 bytes on the heap. This is tiny, but its role is critical — it prevents redundant CDP `enable` commands.

### Ownership Graph

```
Browser (1)
├── cdp.Client (WebSocket, shared by all pages)
├── Page A (session "abc")
│   └── AgentManager (tracks which CDP domains are enabled for session "abc")
└── Page B (session "xyz")
    └── AgentManager (independent tracking for session "xyz")
```

Each `Page` has its own `AgentManager` because CDP domain enablement is **per-session**. Enabling `DOM` on Page A does NOT enable it on Page B. Chrome tracks this server-side too.

---

## 2. `sendCommand` — The Universal CDP Gateway

```go
func (p *Page) sendCommand(method string, params map[string]interface{}) (map[string]interface{}, error) {
    p.ensureAgentsForCommand(method)
    return p.browser.client.SendToSession(p.ctx, p.sessionID, method, params)
}
```

Every single Page operation goes through `sendCommand`. This is the **choke point** between your Go code and Chrome.

### What Happens at the Wire Level

1. **Agent check** — `ensureAgentsForCommand` looks at the method prefix (e.g., `"DOM."`, `"Runtime."`) and, if the corresponding domain isn't enabled yet, sends an enable command first. For example, if you call `DOM.querySelector` but DOM isn't enabled, it first sends `{"method":"DOM.enable","sessionId":"abc"}`. Chrome responds with `{"id":N,"result":{}}`, and the `AgentManager` flips `EnabledAgents.DOM = true`. Subsequent DOM commands skip this step (pure Go boolean check, ~1 nanosecond).

2. **Serialization** — `SendToSession` builds a JSON object:
   ```json
   {"id": 42, "method": "DOM.querySelector", "params": {"nodeId": 1, "selector": "#login"}, "sessionId": "abc-123"}
   ```
   This is serialized via `encoding/json.Marshal`, which uses reflection to walk the `map[string]interface{}` and produce a `[]byte`. The JSON bytes are then wrapped in a WebSocket text frame (2–10 byte header depending on payload size).

3. **Write to socket** — The WebSocket frame is written to the TCP socket via `write(2)` syscall. The kernel copies the bytes into the socket's send buffer. If the buffer is full (unlikely for localhost), the call blocks until space is available.

4. **Chrome processes it** — The browser process receives the frame, demultiplexes by `sessionId`, and forwards to the correct renderer process via Mojo IPC. The renderer executes the command (e.g., runs a CSS selector query on its DOM tree) and sends the result back through the same IPC channel.

5. **Response arrives** — The CDP client's background read goroutine receives the response frame, matches it by `id` to the waiting caller (using a `map[int]chan` internally), and sends the result through the channel. `SendToSession` receives it and returns the `map[string]interface{}`.

### Timing

For localhost CDP calls, the round-trip is typically **0.1–5 ms**:
- Serialization: ~10–50 μs
- Socket write + kernel processing: ~5–20 μs  
- Chrome processing: 50 μs – 5 ms (depends on command complexity)
- Socket read + deserialization: ~10–50 μs

DOM queries on large pages can take longer because the renderer must walk the DOM tree.

---

## 3. `ensureAgentsForCommand` — Lazy Domain Enablement

```go
func (p *Page) ensureAgentsForCommand(method string) error {
    // Maps CDP method prefixes to required agents
    switch {
    case strings.HasPrefix(method, "DOM."):
        return p.agentManager.EnsureAgent(agent.AgentDOM)
    case strings.HasPrefix(method, "Runtime."):
        return p.agentManager.EnsureAgent(agent.AgentRuntime)
    case strings.HasPrefix(method, "Page."):
        return p.agentManager.EnsureAgent(agent.AgentPage)
    case strings.HasPrefix(method, "Input."):
        // Input domain doesn't need explicit enabling
        return nil
    default:
        return nil
    }
}
```

### Why Lazy?

CDP domains are expensive to enable. When you enable `DOM`, Chrome starts tracking every DOM mutation and maintaining a mirror of the DOM tree. Enabling `Network` makes Chrome report every HTTP request. If you enable all domains upfront, Chrome sends a flood of events you might never use, consuming CPU and memory in both Chrome and Go.

The lazy pattern means: "Don't enable DOM until someone actually needs a DOM command." This is especially important in parallel tests where each worker has its own session — enabling 8 domains × 6 workers = 48 enable commands that might never be needed.

### `EnsureAgent` Internals

`AgentManager.EnsureAgent(agentName)` checks `EnabledAgents.IsEnabled(agentName)`. If already enabled, returns immediately (boolean check, no I/O). If not enabled:
1. Sends `"<agentName>.enable"` via CDP.
2. On success, calls `EnabledAgents.Enable(agentName)` to flip the boolean.
3. If the enable returns context data (e.g., `DOM.enable` returns the document root node), stores it in `AgentContext`.

---

## 4. `Navigate(url string, ...)` — Deep Walkthrough

### Step 1: Send Navigation Command

```go
_, err = p.sendCommand(cdp.CmdPageNavigate, map[string]interface{}{"url": url})
```

This sends `Page.navigate` to Chrome. What Chrome does internally:

1. **URL parsing** — Chrome's `GURL` class parses the URL, validates the scheme, extracts host/port/path.
2. **Network request** — The browser process's network service initiates a TCP connection to the target server (DNS lookup → TCP handshake → TLS handshake if HTTPS → HTTP request). This all happens in Chrome's network process.
3. **Response processing** — HTML bytes arrive, Chrome's HTML parser (in the renderer process) begins building the DOM tree incrementally. CSS is parsed into a CSSOM. JavaScript is parsed and compiled by V8.
4. **Frame lifecycle** — The renderer process fires lifecycle events: `DOMContentLoaded` when the HTML is fully parsed, `load` when all sub-resources (images, stylesheets, scripts) have finished loading.

### Step 2: Poll `document.readyState`

```go
for {
    result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
        "expression": "document.readyState",
    })
    // Check if result is "complete"
    if readyState == "complete" { break }
    time.Sleep(pollInterval)
}
```

`document.readyState` is a DOM property that reflects the page's loading phase:
- `"loading"` — HTML is still being parsed. The DOM tree is incomplete.
- `"interactive"` — HTML parsing is done (`DOMContentLoaded` fired), but sub-resources (images, iframes, stylesheets) may still be loading.
- `"complete"` — All sub-resources are loaded (`load` event fired).

Each poll sends a `Runtime.evaluate` command, which:
1. Enters the renderer's V8 isolate.
2. Evaluates `document.readyState` (a native getter on the `Document` DOM object, implemented in C++ in Blink).
3. Returns the string result.

The `time.Sleep(pollInterval)` between polls is a Go-level sleep — the goroutine is parked by the Go scheduler and woken after the interval. During this time, the goroutine consumes no CPU. The OS thread it was running on is free to run other goroutines.

### Step 3: Return

On success, `Navigate` returns `nil`. On timeout, it returns an error with the duration. On CDP failure, it wraps the error.

---

## 5. `Close()` — Page Teardown

```go
func (p *Page) Close() error {
    _, err = p.sendCommand(cdp.CmdPageClose, nil)
    return err
}
```

### What Happens in Chrome

`Page.close` tells Chrome to close this target (tab). Chrome:
1. Fires the `beforeunload` event in the renderer (if the page registered a handler).
2. Tears down the renderer process (or marks it for reuse if Chrome's process-per-site model allows).
3. Removes the target from its internal target list.
4. Detaches the CDP session — any further commands with this `sessionId` will fail.

### Memory Impact

After `page.Close()`:
- Chrome's renderer process for this tab is killed or recycled. This frees ~50–150 MB of virtual memory.
- The Go-side `Page` struct and its `AgentManager` become GC-eligible once no references remain.
- The `sessionID` is invalidated server-side — Chrome will reject any future commands using it.

---

## 6. `URL()`, `Title()` — DOM Property Queries

```go
func (p *Page) URL() (string, error) {
    result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
        "expression": "window.location.href",
    })
    // Extract result["result"]["value"].(string)
}
```

### What Happens in the Renderer Process

`Runtime.evaluate` with `expression: "window.location.href"`:
1. V8 enters the JavaScript execution context for this page.
2. `window` is the global object. `location` is a `Location` DOM object (backed by C++ in Blink). `href` is a getter that reads the current URL from the `FrameLoader` in Blink.
3. The string is serialized as a CDP result and sent back.

These calls are very fast (~0.1–0.5 ms) because they don't trigger any layout or rendering — they just read a cached property.

### Type Assertion Chain

```go
url, ok = result["result"].(map[string]interface{})["value"].(string)
```

This is a **chained type assertion**. First, `result["result"]` is asserted to `map[string]interface{}` (the CDP `RemoteObject`). Then `["value"]` is asserted to `string`. If either assertion fails, `ok` is `false` and `url` is `""`. This two-step chain is safe because Go evaluates left-to-right and the two-value form never panics.

---

## 7. `Screenshot()` — Screen Capture

```go
func (p *Page) Screenshot() ([]byte, error) {
    result, err = p.sendCommand(cdp.CmdPageCaptureScreenshot, map[string]interface{}{
        "format": "png",
    })
    // result["data"] is a base64-encoded PNG string
    screenshot, err = base64.StdEncoding.DecodeString(data)
}
```

### What Happens in Chrome

1. **Compositing** — Chrome's compositor thread renders all layers of the page into a single bitmap (in GPU memory or CPU memory depending on `--disable-gpu`).
2. **PNG encoding** — The bitmap is encoded to PNG format (lossless compression). For a 1280×720 viewport, the raw bitmap is ~3.7 MB (1280 × 720 × 4 bytes RGBA). After PNG compression, it's typically 100 KB–2 MB depending on content.
3. **Base64 encoding** — The PNG bytes are base64-encoded (increases size by ~33%) and sent as a string in the CDP response.
4. **Base64 decoding** — On the Go side, `base64.StdEncoding.DecodeString` converts the string back to raw PNG bytes. This allocates a new `[]byte` on the heap.

### Memory Impact

A single screenshot temporarily uses:
- **Chrome side**: ~4 MB for the bitmap + ~1 MB for the PNG buffer.
- **Go side**: ~1.3 MB for the base64 string (in the CDP response) + ~1 MB for the decoded PNG bytes.
- After the function returns, the base64 string is GC-eligible; only the decoded `[]byte` is kept.

---

## 8. `SetContent(html string)` — Injecting HTML Without Navigation

```go
func (p *Page) SetContent(html string) error {
    // Step 1: Get the frame ID
    frameResult, err = p.sendCommand("Page.getFrameTree", nil)
    // Step 2: Extract frameID from the response
    // Step 3: Set document content
    _, err = p.sendCommand(cdp.CmdPageSetDocumentContent, map[string]interface{}{
        "frameId": frameID,
        "html":    html,
    })
}
```

### Why Frame ID Is Required

CDP's `Page.setDocumentContent` requires a `frameId` because a page can contain multiple frames (iframes). You must specify which frame's document to replace. The main frame's ID is found via `Page.getFrameTree`, which returns a tree structure. The code extracts `frameTree.frame.id`.

### What Happens in the Renderer

1. Chrome's Blink engine receives the new HTML string.
2. The existing DOM tree is completely discarded — all nodes are freed from Blink's memory arena.
3. The HTML parser creates a new DOM tree from the provided string.
4. No network requests are made (unlike `Navigate`), making this much faster.
5. `document.readyState` goes to `"complete"` almost immediately since there are no sub-resources to wait for (unless the HTML contains `<script src="...">` or `<img src="...">`).

This method is ideal for unit-testing DOM interactions without hitting a real server.

---

## 9. `Evaluate(expression string)` — Running Arbitrary JavaScript

```go
func (p *Page) Evaluate(expression string) (interface{}, error) {
    result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
        "expression":    expression,
        "returnByValue": true,
    })
    return resultObj["value"], nil
}
```

### `returnByValue: true` — What This Means

When `returnByValue` is `true`, Chrome serializes the result as a JSON value and sends it inline. This means:
- Primitives (string, number, boolean) are sent directly.
- Objects and arrays are JSON-serialized (deep copy).
- DOM nodes, functions, and symbols **cannot** be returned by value — they would fail.

When `returnByValue` is `false` (or omitted), Chrome returns a `RemoteObject` reference (an `objectId` string). You can then use `Runtime.callFunctionOn` or `Runtime.getProperties` to interact with the object without serializing it. This is how `Element.resolveObjectID` works.

### Security

`Runtime.evaluate` runs JavaScript in the **page's execution context**. It has full access to the page's DOM, cookies, localStorage, etc. This is by design — kexas is an automation tool. But it means you should never pass unsanitized user input as the `expression`.

---

## 10. Error Conventions

### Sentinel Errors

The Page module propagates sentinel errors from `errors/errors.go`:
- `ErrTextEmpty` — Returned when `WaitAndType("")` is called.
- `ErrElementNoPage` — Returned when an Element's `page` reference is nil (shouldn't happen in normal use).
- `ErrElementInvalidNodeID` — Returned when a stale Element is used after the DOM was refreshed.

### Timeout Errors

Timeout errors always include the duration for debugging:
```go
return fmt.Errorf("page load timeout after %v", timeout)
```

### CDP Error Wrapping

All CDP failures are wrapped with context:
```go
return fmt.Errorf("failed to close page: %w", err)
```

The `%w` verb creates an error chain. A caller can use `errors.Is(err, someTarget)` to check the root cause without string matching.

---

## 11. Concurrency Model

A `Page` is **NOT thread-safe** by design. In kexas:
- In sequential mode, one goroutine uses one Page — no contention.
- In parallel mode, each worker gets its own isolated `Browser` + `Page` (via `launchIsolatedBrowser`). No two goroutines share a `Page`.

The `sendCommand` method sends CDP commands sequentially (one at a time per page) because CDP itself is sequential within a session. If you sent overlapping commands from multiple goroutines, the responses would be matched correctly by message ID, but Chrome might process them out of order, leading to unpredictable behavior.

---

## 12. Process Flow: A Complete Navigate + Find + Click

```
Your Go Code                    Page.sendCommand           CDP Client           Chrome
     │                               │                        │                    │
     ├── page.Navigate(url) ────────►│                        │                    │
     │                               ├── ensureAgents ───────►│                    │
     │                               │   (Page.enable if needed)                   │
     │                               ├── Page.navigate ──────►│── WS frame ──────► │
     │                               │                        │◄── WS frame ────── │ (nav started)
     │                               ├── poll readyState ────►│── WS frame ──────► │
     │                               │   (loop until "complete")                   │
     │                               │◄────────── nil error ──│                    │
     │◄── nil ─────────────────────── │                        │                    │
     │                               │                        │                    │
     ├── page.Find("#login") ───────►│                        │                    │
     │                               ├── ensureAgents ───────►│                    │
     │                               │   (DOM.enable if needed)                    │
     │                               ├── DOM.querySelector ──►│── WS frame ──────► │
     │                               │                        │◄── WS frame ────── │ (nodeId: 42)
     │◄── *Element{nodeID:42} ────── │                        │                    │
     │                               │                        │                    │
     ├── elem.Click() ─────────────►(Element calls page.sendCommand)               │
     │                               ├── Runtime.callFunctionOn►── WS frame ─────► │
     │                               │   (objectId, "this.click()")                │
     │                               │                        │◄── WS frame ────── │ (true)
     │◄── nil ─────────────────────── │                        │                    │
```
