# Core Components

> Detailed documentation of Kexas core components

**Last Updated:** February 24, 2026

---

## 1. Browser (`browser.go`)

**Purpose**: Manages browser process lifecycle and CDP connection.

**Key Responsibilities**:
- Launch browser process via `launcher` package
- Establish CDP WebSocket connection
- Create and manage pages (tabs)
- Clean up resources on close

**Lifecycle**:
```go
Launch() → Connect CDP → Create/Attach Pages → Close()
```

**Key Methods**:
- `Launch(opts)` - Start browser and connect
- `NewPage()` - Create new tab
- `FirstPage()` - Get existing tab
- `Close()` - Terminate browser and cleanup

**Internal State**:
- `process` - Browser process handle
- `client` - CDP WebSocket client
- `ctx` - Context for cancellation
- `log` - Component logger

**Example Usage**:
```go
browser, err := kexas.Launch(&kexas.LaunchOptions{Headless: true})
if err != nil {
    log.Fatal(err)
}
defer browser.Close()

page, err := browser.NewPage()
if err != nil {
    log.Fatal(err)
}
```

---

## 2. Page (`page.go`)

**Purpose**: Represents a browser tab and provides navigation/interaction APIs.

**Key Responsibilities**:
- Navigate to URLs
- Wait for page load states
- Execute CDP commands in page session
- Capture screenshots
- Get page metadata (URL, title)

**Lifecycle**:
```go
NewPage() → Navigate() → Interact → Close()
```

**Key Methods**:
- `Navigate(url, waitUntil)` - Navigate with wait strategy
- `WaitForLoadState(strategy, timeout)` - Wait for specific load state
- `Screenshot()` - Capture page screenshot
- `URL()` - Get current URL
- `Title()` - Get page title
- `Close()` - Close tab

**Internal State**:
- `browser` - Parent browser reference
- `targetID` - CDP target identifier
- `sessionID` - CDP session identifier (for flat protocol)
- `ctx` - Context for operations
- `log` - Component logger

**Example Usage**:
```go
err := page.Navigate("https://example.com", kwait.WaitUntilLoad)
if err != nil {
    log.Fatal(err)
}

title, err := page.Title()
screenshot, err := page.Screenshot()
```

---

## 3. CDP Client (`internal/cdp/cdp.go`)

**Purpose**: Low-level Chrome DevTools Protocol communication over WebSocket.

**Key Responsibilities**:
- Establish WebSocket connection to browser
- Send CDP commands and receive responses
- Handle CDP events and dispatch to handlers
- Manage command-response matching (by ID)
- Support flat protocol (sessionId routing)

**Architecture**:
```
Client
  ├── WebSocket Connection
  ├── Command Queue (pending responses)
  ├── Event Handlers (subscriptions)
  └── Read Loop (goroutine)
```

**Key Methods**:
- `Connect(ctx, wsURL)` - Establish connection
- `Send(ctx, method, params)` - Send command to browser
- `SendToSession(ctx, sessionID, method, params)` - Send to specific page
- `On(method, handler)` - Subscribe to events
- `Close()` - Close connection

**Message Types**:
1. **Request**: Command sent to browser (has `id`, `method`, `params`, `sessionId`)
2. **Response**: Result from browser (has `id`, `result` or `error`)
3. **Event**: Notification from browser (has `method`, `params`, no `id`)

**Flat Protocol**:
- Modern CDP routing mechanism
- `sessionId` is top-level field in request (not nested)
- Simpler than legacy `Target.sendMessageToTarget` wrapping
- Used by Playwright and Puppeteer

**Example Usage**:
```go
client, err := cdp.Connect(ctx, wsURL)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

result, err := client.Send(ctx, "Page.navigate", map[string]interface{}{
    "url": "https://example.com",
})
```

---

## 4. Launcher (`internal/launcher/launcher.go`)

**Purpose**: Find, download, and launch Chromium browser process.

**Key Responsibilities**:
- Locate Chromium binary (cached, downloaded, or system)
- Download Chrome for Testing if needed
- Build command-line arguments
- Start browser process with remote debugging
- Extract WebSocket debugger URL from stdout
- Manage temporary profile directories

**Launch Flow**:
```
1. Find Chromium (cache → download → system)
2. Create temp profile directory
3. Build args (headless, port, flags)
4. Start process
5. Extract WebSocket URL from stdout
6. Return Browser handle
```

**Key Methods**:
- `Launch(ctx, opts)` - Start browser
- `findChromium()` - Locate executable
- `downloadChromium()` - Download Chrome for Testing
- `buildArgs(opts)` - Construct CLI args
- `extractDebuggerURL(file, timeout)` - Parse stdout for URL

**Temporary Profiles**:
- Each launch creates `/tmp/kexas-chrome-{timestamp}/`
- Complete isolation (no cookies, history, cache)
- Automatic cleanup on `Browser.Close()`

**Example Usage**:
```go
opts := &launcher.Options{
    Headless: true,
    Port:     9222,
}

browser, err := launcher.Launch(ctx, opts)
if err != nil {
    log.Fatal(err)
}
defer browser.Close()

wsURL := browser.WebSocketURL()
```

---

## 5. Wait Strategies (`internal/kwait/kwait.go`)

**Purpose**: Smart waiting mechanisms for reliable automation.

**Key Responsibilities**:
- Implement Playwright-standard wait strategies
- Poll conditions with configurable timeout/interval
- Handle page load states
- Provide generic wait utilities

**Wait Strategies** (from fastest to slowest):

1. **WaitUntilCommit**: Response headers received (page still blank)
2. **WaitUntilDOMContentLoaded**: HTML parsed, DOM ready
3. **WaitUntilLoad**: All resources loaded (images, CSS, JS) - **DEFAULT**
4. **WaitUntilNetworkIdle**: No network activity for 500ms

**Key Functions**:
- `For(ctx, condition, opts)` - Generic condition polling
- `UntilReady[T](ctx, fn, opts)` - Wait for non-nil result
- `ForPageLoad(ctx, getTitleFn, waitUntil, timeout)` - Page load waiting

**Example Usage**:
```go
// Navigate with specific wait strategy
page.Navigate(url, kwait.WaitUntilDOMContentLoaded)

// Or use default (load)
page.Navigate(url)

// Manual wait
page.WaitForLoadState(kwait.WaitUntilNetworkIdle, 30*time.Second)

// Generic condition wait
err := kwait.For(ctx, func() (bool, error) {
    title, _ := page.Title()
    return title != "", nil
}, kwait.DefaultOptions())
```

---

## 6. Logger (`internal/logger/logger.go`)

**Purpose**: Structured, leveled logging with component tagging.

**Key Responsibilities**:
- Provide Debug, Info, Warn, Error levels
- Color-coded terminal output
- Component-based organization
- Key-value pair formatting
- Thread-safe logging

**Log Levels**:
- **Debug**: Verbose development messages (CDP commands, internal state)
- **Info**: Operational messages (navigation, page creation)
- **Warn**: Potentially harmful situations (timeouts, fallbacks)
- **Error**: Error conditions (failures, exceptions)
- **Silent**: Disable all logging

**Example Usage**:
```go
log := logger.New("cdp")
log.Debug("sending command", "method", "Page.navigate", "url", url)
log.Info("navigated successfully", "elapsed", "1.2s")
log.Error("navigation failed", "err", err)

// Change log level
log = log.WithLevel(logger.LevelDebug)
```

**Output Format**:
```
2026-02-24T22:00:00+07:00 [INFO ] [page] Navigated to https://example.com
2026-02-24T22:00:01+07:00 [DEBUG] [cdp] → Page.navigate {"url":"https://example.com"}
2026-02-24T22:00:02+07:00 [ERROR] [page] Navigation failed err=timeout
```

---

## Component Interaction

```
User Code
    │
    ├─> Browser.Launch()
    │       │
    │       ├─> Launcher.Launch() → Start process
    │       └─> CDP.Connect() → WebSocket
    │
    ├─> Browser.NewPage()
    │       │
    │       └─> CDP.Send("Target.createTarget")
    │
    ├─> Page.Navigate()
    │       │
    │       ├─> CDP.SendToSession("Page.navigate")
    │       └─> kwait.ForPageLoad()
    │
    └─> Browser.Close()
            │
            ├─> CDP.Close() → Close WebSocket
            └─> Launcher.Close() → Kill process
```

---

## 7. Element (`element.go`, `element_methods.go`)

**Purpose**: Represents a DOM element and provides query methods.

**Last Updated:** February 27, 2026

**Key Responsibilities**:
- Store element reference (objectId and/or nodeId)
- Resolve objectId from nodeId when needed (lazy, cached)
- Check element visibility via JavaScript
- Retrieve text content and attributes

**Internal State**:
- `page` - Parent page reference
- `selector` - Original selector used to find the element
- `nodeID` - CDP node identifier (may be 0 for objectId-only elements)
- `objectID` - CDP remote object ID (preferred handle for all interactions)
- `timeout` - Default timeout for operations

**Key Methods**:
- `IsVisible()` - Check visibility via `Runtime.callFunctionOn` with computed style
- `GetText()` - Get `innerText` via `Runtime.callFunctionOn`
- `GetAttribute(name)` - Get attribute via `DOM.getAttributes`
- `Validate()` - Check if node still exists via `DOM.describeNode`
- `resolveObjectID()` - Resolve and cache objectId from nodeId

**Constructors**:
- `NewElement(page, selector, nodeID, timeout)` - Legacy, nodeId-based
- `NewElementWithObject(page, selector, nodeID, objectID, timeout)` - Preferred, objectId-based

---

## 8. Element Interaction (`element_interaction.go`)

**Purpose**: Action methods for interacting with DOM elements.

**Last Updated:** February 27, 2026

**Key Responsibilities**:
- Click elements via `Runtime.callFunctionOn` with `this.click()`
- Type text via `Input.dispatchKeyEvent` (per-character keyboard events)
- Hover over elements via `MouseEvent` dispatch
- Wait for element states before interaction

**Key Methods**:
- `Click()` / `WaitAndClick()` / `WaitAndClickFor(timeout)` - Click actions
- `Type(text)` - Type using CDP keyboard events (SPA-compatible)
- `WaitAndType(text)` / `WaitAndTypeFor(text, timeout)` - Type with JS value assignment
- `Hover()` / `WaitAndHover()` / `WaitAndHoverFor(timeout)` - Hover actions

**Critical Design Decision**: `Type()` uses `Input.dispatchKeyEvent` per character (like go-rod/Playwright) instead of JavaScript `this.value = text`. This is required for SPA frameworks (React, Angular) that listen to real keyboard events.

---

## 9. Page Find (`page_find.go`)

**Purpose**: Element finding with auto-retry for SPA rendering delays.

**Last Updated:** February 27, 2026

**Key Responsibilities**:
- Auto-detect selector type (CSS, ID, XPath)
- Retry finding for up to 10 seconds (200ms poll interval)
- Log page title on timeout for debugging

**Key Methods**:
- `Find(selector)` - Auto-detect and retry
- `FindByXPath(xpath)` - Public XPath wrapper
- `findByCSS(selector)` - `Runtime.evaluate` + `document.querySelector`
- `findByID(id)` - `Runtime.evaluate` + `document.getElementById`
- `findByXPath(xpath)` - `DOM.performSearch` + `DOM.getSearchResults`

**objectId-first pattern**: `findByCSS` and `findByID` return elements with `objectId` only (no nodeId). `findByXPath` returns elements with `nodeId` only (objectId resolved lazily).

---

## 10. Agent Manager (`internal/agent/`)

**Purpose**: Lazy CDP domain enablement with state caching.

**Last Updated:** February 27, 2026

**Key Responsibilities**:
- Track which CDP agents are enabled per session
- Enable agents on demand (lazy enablement)
- Cache agent context (document root node, execution context ID)
- Thread-safe operations via `sync.RWMutex`

**Key Files**:
- `agent_manager.go` - `AgentManager` struct and initialization
- `smart_enablement.go` - `EnsureAgent()`, `EnsureAgents()`, `WaitForAgentReady()`
- `enabled_agents.go` - `EnabledAgents` tracking struct with agent constants
- `operations.go` - `Disable()`, `DisableAll()`, `ResetAgent()`, `Stats()`
- `agent_context.go` - Context types for DOM, Input, Runtime, etc.

**Key Methods**:
- `EnsureAgent(name)` - Enable agent if not already enabled (core method)
- `EnsureAgents(names...)` - Enable multiple agents in parallel
- `IsAgentReady(name)` - Check if agent is enabled and has context
- `ResetAgent(name)` - Force re-enablement on next use (after navigation)

**Agent Constants**: `DOM`, `Input`, `Runtime`, `Network`, `Page`, `Security`, `Debugger`, `Profiler`

**Integration**: `Page.sendCommand()` calls `ensureAgentsForCommand(method)` which maps CDP command names to required agents and calls `AgentManager.EnsureAgent()`.

---

## Component Interaction (Updated)

```
User Code
    │
    ├─> Browser.Launch()
    │       │
    │       ├─> Launcher.Launch() → Start process
    │       └─> CDP.Connect() → WebSocket
    │
    ├─> Browser.NewPage()
    │       │
    │       ├─> CDP.Send("Target.createTarget")
    │       ├─> CDP.Send("Target.attachToTarget") → sessionId
    │       ├─> Stealth: Network.enable → Network.setUserAgentOverride
    │       ├─> Stealth: Page.enable → Page.addScriptToEvaluateOnNewDocument
    │       └─> Create AgentManager(cdp, sessionId)
    │
    ├─> Page.Navigate()
    │       │
    │       ├─> ensureAgentsForCommand("Page.navigate") → EnsureAgent("Page")
    │       ├─> CDP.SendToSession("Page.navigate")
    │       └─> kwait.ForPageLoad()
    │
    ├─> Page.Find(selector)
    │       │
    │       ├─> Retry loop (10s, 200ms poll)
    │       ├─> ensureAgentsForCommand("Runtime.evaluate") → EnsureAgent("Runtime")
    │       ├─> Runtime.evaluate(document.querySelector) → objectId
    │       └─> Return Element{objectId}
    │
    ├─> Element.Click() / Element.Type(text)
    │       │
    │       ├─> resolveObjectID() → cached or DOM.resolveNode
    │       ├─> Runtime.callFunctionOn(objectId, ...) [Click/Focus]
    │       └─> Input.dispatchKeyEvent per char [Type only]
    │
    └─> Browser.Close()
            │
            ├─> CDP.Close() → Close WebSocket
            └─> Launcher.Close() → Kill process → 2s port release wait
```

---

## Next Steps

- See `DESIGN_PATTERNS.md` for common patterns
- See `DATA_FLOW.md` for detailed flow diagrams
- See `KEY_CONCEPTS.md` for important concepts
- See `.kb/browsers/ELEMENT_FINDING.md` for detailed element finding docs
- See `.kb/browsers/ELEMENT_INTERACTION.md` for detailed interaction docs
- See `.kb/browsers/STEALTH_ANTI_DETECTION.md` for stealth techniques

