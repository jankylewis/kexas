# Page (High-Level)

Source: `kexas/page.go`, `page_find.go`, `page_wait.go`, `page_scroll.go`

## Mission

Represent a single browser tab. `kexas.Page` is where test authors spend most of their time — navigating to URLs, finding elements, running JavaScript, taking screenshots, and waiting for page readiness.

## What It Holds

| Field | Purpose |
|-------|---------|
| `browser` | Back-pointer to parent `Browser` (shared WebSocket client) |
| `targetID` | Chrome's unique ID for this tab (maps to a renderer process) |
| `sessionID` | CDP session ID — multiplexes commands over the shared WebSocket |
| `agentManager` | Tracks which CDP domains (DOM, Runtime, Page, etc.) are enabled for this session |
| `ctx` | Context for cancellation propagation |
| `log` | Structured logger scoped to `"page"` |

## Key Functions

### Navigation
- **`Navigate(url, waitUntil...)`** — Sends `Page.navigate`, polls `document.readyState` until `"complete"` (or custom wait condition). Resets DOM agent caches after success.
- **`WaitForLoadState(waitUntil, timeout)`** — Polls `document.readyState` until it reaches `"interactive"` or `"complete"`.
- **`WaitForNavigationCompleted(timeout)`** — Convenience wrapper for `WaitForLoadState(WaitUntilLoad, timeout)`.

### Element Discovery
- **`Find(selector, opts...)`** — Locates the first matching element via CSS/XPath. Returns `*Element` with nodeID + objectID.
- **`FindAll(selector, opts...)`** — Returns all matching elements.
- **`WaitAndFind(selector)`** — Polls until the element appears or timeout. Preferred for test code.

### Utilities
- **`Evaluate(expression)`** — Runs arbitrary JavaScript via `Runtime.evaluate`. Returns the result.
- **`Screenshot()`** — Captures the viewport as PNG bytes via `Page.captureScreenshot`.
- **`SetContent(html)`** — Replaces the page's HTML without a network request. Ideal for unit-testing DOM interactions.
- **`URL()` / `Title()`** — Reads current URL or document title via JavaScript.
- **`Close()`** — Closes this tab via `Page.close`. The renderer process is killed or recycled.

### The CDP Gateway
- **`sendCommand(method, params)`** — Every Page operation goes through this. It calls `ensureAgentsForCommand(method)` to lazily enable the required CDP domain, then proxies the command to `browser.client.SendToSession`.

## Lifecycle

```
Browser.attachToPage(targetID)
  │
  ├── Target.attachToTarget → sessionID
  ├── Create AgentManager(client, sessionID)
  ├── Apply stealth settings
  └── return &Page{...}

page.Navigate("https://example.com")
  │
  ├── sendCommand("Page.navigate", {url})
  ├── Poll document.readyState until "complete"
  └── Reset DOM agent caches

page.Find("#login")
  │
  ├── sendCommand("DOM.querySelector", {selector})
  └── return &Element{nodeID, objectID, page}

page.Close()
  │
  └── sendCommand("Page.close") → renderer process torn down
```

## Concurrency Model

- A `Page` is **NOT thread-safe**. One goroutine per page.
- In parallel mode, each worker gets its own `Browser` + `Page`. No sharing.
- `sendCommand` is sequential within a session — CDP processes commands in order.

## Error Handling

- Navigation timeouts include the duration: `"page load timeout after 30s"`.
- All CDP errors are wrapped with `%w` for error chain inspection.
- Sentinel errors (`ErrTextEmpty`, `ErrElementNoPage`) propagate from Element operations.

## Why It Matters

`Page` is where business flows live. Every test scenario — login, checkout, form submission — is a sequence of `Navigate → Find → Click/Type → Assert`. The Page module centralizes CDP complexity (agent management, readyState polling, session multiplexing) so test authors write clean, readable code.
