# New Feature Module Interactions — Deep Technical (Student Level)

---

## How Features Plug Into the Existing Architecture

Kexas has a layered architecture. New features slot into specific layers:

```
Layer 4 (User API):     kapi/flow.go, kapi/patterns.go
Layer 3 (Feature API):  cookie.go, storage.go, recorder.go
Layer 2 (Core API):     page.go, browser.go, element.go
Layer 1 (Infrastructure): internal/cdp/, internal/agent/, internal/logger/
Layer 0 (Transport):    WebSocket connection to Chrome
```

Each layer only calls the layer directly below it. KAPI (Layer 4) calls Feature API (Layer 3) calls Core API (Layer 2) calls Infrastructure (Layer 1).

---

## Cookie → Agent Manager Interaction

### The Problem
The Network CDP domain must be enabled before any cookie command works. If you send `Network.getCookies` without first sending `Network.enable`, Chrome ignores it.

### How It's Solved

```
cookie.go: SetCookie(cookie)
  │
  ▼
page.go: sendCommand("Network.setCookie", params)
  │
  ├── ensureAgentsForCommand("Network.setCookie")
  │     │
  │     ▼
  │   agent_manager.go: EnsureAgent(AgentNetwork)
  │     │
  │     ├── Already enabled? → return nil (cached)
  │     └── Not enabled? → Send("Network.enable", {}) → mark as enabled
  │
  └── cdp_client: Send("Network.setCookie", params)
```

**Key insight:** `ensureAgentsForCommand` is a **switch statement** in `page.go` that maps command strings to required agents. Adding a new CDP domain means adding cases to this switch.

### What's in the switch?

```go
case cdp.CmdNetworkGetCookies, cdp.CmdNetworkSetCookie,
    cdp.CmdNetworkDeleteCookies, cdp.CmdNetworkClearBrowserCookies:
    return p.agentManager.EnsureAgent(agent.AgentNetwork)
```

All four Network cookie commands map to the same agent. The Agent Manager deduplicates — calling `EnsureAgent(AgentNetwork)` 100 times only sends `Network.enable` once.

---

## Storage → Runtime.evaluate Interaction

Storage doesn't use its own CDP domain. Instead, it builds JavaScript strings and sends them through the existing `page.Evaluate()` method, which uses `Runtime.evaluate`.

```
storage.go: Set("token", "abc123")
  │
  ├── escapeJSString("token") → "token"
  ├── escapeJSString("abc123") → "abc123"
  ├── Build: "localStorage.setItem('token', 'abc123')"
  │
  ▼
page.go: Evaluate("localStorage.setItem('token', 'abc123')")
  │
  ▼
page.go: sendCommand("Runtime.evaluate", {expression: ...})
  │
  ├── ensureAgentsForCommand("Runtime.evaluate")
  │     → EnsureAgent(AgentRuntime)
  │
  └── cdp_client → WebSocket → Chrome executes JS → response
```

**Why this works:** `Runtime.evaluate` is Chrome's "run arbitrary JavaScript" command. It can do anything that `document.querySelector()` or `localStorage.setItem()` can do. Storage operations are pure JS — no special CDP domain needed.

**Trade-off:** Using JS evaluation means the storage operations run in the page's JavaScript context. If the page has overridden `localStorage.setItem`, our call would use the overridden version. In practice, no legitimate application does this.

---

## Multi-Tab → Browser + Page Bidirectional Interaction

Multi-tab is unique because it modifies **two existing modules** rather than creating a new file.

### Browser → Page (creation)

```
browser.go: NewPage()
  │
  ├── Send("Target.createTarget", {url: "about:blank"})
  ├── attachToPage(targetID) → *Page
  │
  ├── pagesMu.Lock()
  ├── pages = append(pages, page)  ← Track the new page
  └── pagesMu.Unlock()
```

### Page → Browser (destruction)

```
page.go: Close()
  │
  ├── Send("Page.close")
  ├── p.closed = true
  │
  ├── p.browser.removePage(p)      ← Remove from tracking
  │     ├── pagesMu.Lock()
  │     ├── splice page out of slice
  │     └── pagesMu.Unlock()
```

**Bidirectional dependency:** Browser creates Pages and tracks them. Pages notify Browser when they close. This is a controlled circular reference — the Page holds a `*Browser` pointer set during `attachToPage`.

### Thread Safety Model

```
Goroutine A (test worker 1)       Goroutine B (test worker 2)
        │                                  │
        ├── browser.NewPage()              ├── browser.PageCount()
        │   ├── Lock()                     │   ├── Lock()  ← BLOCKS until A releases
        │   ├── append(pages, p1)          │   │
        │   └── Unlock()                   │   │
        │                                  │   ├── len(pages)  ← now sees p1
        │                                  │   └── Unlock()
```

Without the mutex, goroutine B might see a half-written slice header (data race).

---

## Recorder → Page.Screenshot Interaction

The recorder doesn't use CDP screencast events (which would require event subscription infrastructure). Instead, it polls:

```
recorder.go: collectFrames() [background goroutine]
  │
  loop:
  ├── select:
  │   ├── case <-stopCh: return      ← Stop signal
  │   └── default:
  │       ├── page.Screenshot()       ← Full CDP round-trip
  │       │   └── sendCommand("Page.captureScreenshot")
  │       │       └── Chrome renders → encodes PNG → base64 → response
  │       ├── mu.Lock()
  │       ├── frames = append(frames, Frame{data, timestamp, index})
  │       ├── mu.Unlock()
  │       └── time.Sleep(100ms)       ← ~10 fps
```

**Why mutex for frames?** The collector goroutine writes frames while `SaveFrames()`/`SaveVideo()`/`FrameCount()` may read them from a different goroutine.

### The Stop Channel Pattern

```go
// Start:
r.stopCh = make(chan struct{})
go r.collectFrames()

// Stop:
close(r.stopCh)  // ALL readers of stopCh immediately unblock
```

`close(ch)` is a broadcast — every goroutine doing `<-ch` or `select { case <-ch: }` wakes up simultaneously. This is different from sending a value (which only one receiver gets).

---

## KAPI → Everything Interaction

KAPI is the top layer that delegates to all other modules:

```
kapi.Open(url)
  ├── kexas.Launch() → browser.go
  ├── browser.FirstPage() → browser.go
  └── page.Navigate() → page.go

flow.Find(selector)
  └── page.Find() → page_find.go

flow.Click()
  └── element.Click() → element.go

flow.SetCookie(name, value, domain)
  └── page.SetCookie() → cookie.go

flow.SetLocalStorage(key, value)
  └── page.LocalStorage().Set() → storage.go

flow.FillForm(fields)
  └── for each field: flow.Find(selector).Type(value)
      └── page.Find() + element.Type()
```

**KAPI never calls internal/ packages directly.** It only uses the public API of kexas types (Browser, Page, Element, Cookie, Storage). This means internal changes don't break KAPI.

---

## Error Propagation Across Layers

### Traditional (raw kexas):
```
cdp_client error → page.sendCommand wraps it → cookie.go wraps it → user sees full chain
Example: "set cookie failed for 'token': Network.setCookie: websocket closed"
```

### KAPI:
```
cdp_client error → page error → cookie error → flow.addError(wrapped) → user calls flow.Err()
Example: "setCookie 'token' failed: set cookie failed for 'token': Network.setCookie: websocket closed"
```

Each layer wraps the error with its own context using `fmt.Errorf("... : %w", err)`. The `%w` verb preserves the error chain so `errors.Is()` still works through all layers.
