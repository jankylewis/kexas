# Element Finding on WebUI

> How Kexas locates DOM elements via CDP, including selector strategies, retry logic, and the critical nodeId vs objectId distinction.

**Last Updated:** February 27, 2026

---

## Overview

Element finding is implemented in `page_find.go`. The public entry point is `Page.Find(selector)`, which auto-detects the selector type, delegates to the appropriate CDP strategy, and retries for up to 10 seconds to handle SPA rendering delays.

```
page.Find(selector)
    │
    ├─ starts with "#"   → findByID(id)
    ├─ starts with "//"  → findByXPath(xpath)
    ├─ starts with "("   → findByXPath(xpath)
    └─ otherwise         → findByCSS(selector)
```

---

## Selector Strategies

### 1. CSS Selector (`findByCSS`)

**CDP Commands Used:** `Runtime.evaluate` + objectId extraction

```go
// JavaScript executed in the page context
document.querySelector("a[data-nav-ref='nav_ya_signin']")
```

**Flow:**
1. `Runtime.evaluate` with `document.querySelector(selector)`, `returnByValue: false`
2. Extract `objectId` from the result's `result` object
3. Check for `subtype == "null"` or `type == "undefined"` → element not found
4. Check for `exceptionDetails` → execution context destroyed (page navigated)
5. Create `Element` with `NewElementWithObject(page, selector, 0, objectID, 10s)`

**Why Runtime.evaluate instead of DOM.querySelector:**
- `DOM.querySelector` requires a parent `nodeId` (the document root), which can become stale after navigation
- `Runtime.evaluate` works directly in the current execution context
- Returns an `objectId` immediately, no extra `DOM.requestNode` step needed

### 2. ID Selector (`findByID`)

**CDP Commands Used:** `Runtime.evaluate` + objectId extraction

```go
// JavaScript executed in the page context
document.getElementById("ap_email_login")
```

**Flow:** Same as CSS but uses `document.getElementById`. The `#` prefix is stripped before calling.

**Why not DOM.performSearch:**
- `DOM.performSearch` returns `nodeIds` as numbers (float64 in JSON), which caused type assertion bugs early on
- `document.getElementById` via `Runtime.evaluate` is simpler and returns `objectId` directly

### 3. XPath Selector (`findByXPath`)

**CDP Commands Used:** `DOM.performSearch` → `DOM.getSearchResults`

```go
// XPath query
DOM.performSearch({query: "//button[@id='submit']"})
DOM.getSearchResults({searchId, fromIndex: 0, toIndex: 1})
```

**Flow:**
1. `DOM.performSearch` returns `searchId` and `resultCount`
2. Check `resultCount > 0`, otherwise element not found
3. `DOM.getSearchResults` returns `nodeIds` array
4. **Critical:** `nodeIds` are `float64` in JSON, not strings — must type-assert as `float64` then cast to `int64`
5. Create `Element` with `NewElement(page, xpath, nodeID, 10s)` (nodeId-based, no objectId)

**XPath Limitation:** This path only gets a `nodeId`, not an `objectId`. The element will need to call `DOM.resolveNode` later for `Runtime.callFunctionOn` operations.

---

## Retry Mechanism

```go
var timeout time.Duration = 10 * time.Second
var pollInterval time.Duration = 200 * time.Millisecond

for time.Since(start) < timeout {
    elem, err = findByCSS(selector)  // or findByID/findByXPath
    if err == nil && elem != nil {
        return elem, nil
    }
    lastErr = err
    time.Sleep(pollInterval)
}
```

- **Timeout:** 10 seconds (hardcoded, handles SPA rendering delays)
- **Poll interval:** 200ms
- **On timeout:** Logs page title for debugging context, returns last error

**Why 10 seconds:** Amazon and similar SPAs can take 2-5 seconds to render content after navigation. The 10s timeout provides a comfortable margin.

---

## nodeId vs objectId: The Critical Distinction

This is the single most important architectural lesson from the kexas element-finding journey.

### nodeId (CDP DOM domain)

- **What:** An integer assigned by the DOM agent to a node in the DOM tree
- **Scope:** Valid only within the current DOM agent session
- **Staleness:** Becomes invalid when the page navigates, DOM is rebuilt, or the DOM agent is re-enabled
- **Use cases:** `DOM.getAttributes`, `DOM.getBoxModel`, `DOM.describeNode`, `DOM.getOuterHTML`

### objectId (CDP Runtime domain)

- **What:** A string handle (`{"injectedScriptId":N,"id":N}`) to a JavaScript object in the V8 heap
- **Scope:** Valid within the current execution context
- **Staleness:** Becomes invalid only when the execution context is destroyed (page navigation)
- **Use cases:** `Runtime.callFunctionOn`, `Runtime.getProperties`

### Why objectId is Preferred (go-rod Pattern)

```
nodeId path:  Find → DOM.querySelector → nodeId → DOM.resolveNode → objectId → Runtime.callFunctionOn
objectId path: Find → Runtime.evaluate → objectId → Runtime.callFunctionOn
```

- **Fewer round-trips:** objectId is obtained directly from `Runtime.evaluate`
- **More reliable:** No intermediate `nodeId` that can go stale
- **go-rod/Playwright pattern:** Both store `objectId` on the Element struct and use it for all interactions

### Kexas Implementation

```go
type Element struct {
    page     *Page
    selector string
    nodeID   cdp.NodeID    // May be 0 for objectId-only elements
    objectID string        // Preferred handle for all interactions
    timeout  time.Duration
}
```

- `findByCSS` and `findByID` create elements with `objectID` only (nodeID = 0)
- `findByXPath` creates elements with `nodeID` only (objectID resolved lazily via `resolveObjectID()`)
- All interaction methods (`Click`, `Type`, `IsVisible`, `GetText`) use `objectID` via `resolveObjectID()`

### Guard Pattern

All interaction methods use a dual guard:

```go
if e.nodeID <= 0 && e.objectID == "" {
    return errors.ErrElementInvalidNodeID
}
```

This accepts elements that have *either* a valid nodeId *or* a valid objectId.

---

## Constructors

### `NewElement(page, selector, nodeID, timeout)`

Legacy constructor. Creates element with nodeId only. Used by `findByXPath`.

### `NewElementWithObject(page, selector, nodeID, objectID, timeout)`

Preferred constructor. Creates element with objectId (and optionally nodeId). Used by `findByCSS` and `findByID`.

---

## Agent Enablement

Element finding requires CDP agents to be enabled. The `Page.sendCommand()` method calls `ensureAgentsForCommand()` before every command:

| Command | Required Agent |
|---------|---------------|
| `Runtime.evaluate` | Runtime |
| `Runtime.callFunctionOn` | Runtime |
| `DOM.performSearch` | DOM |
| `DOM.getSearchResults` | DOM |
| `DOM.requestNode` | DOM |
| `DOM.resolveNode` | DOM |

The `AgentManager.EnsureAgent()` handles lazy enablement — agents are enabled on first use and cached.

---

## Diagnostic on Failure

When `Find()` times out, it logs the page title for debugging context:

```go
// Evaluates document.title for context
p.log.Info("element not found after timeout",
    "selector", selector,
    "timeout", timeout,
    "elapsed", time.Since(start),
    "pageTitle", pageTitle,
)
```

This helps diagnose issues like:
- Bot detection redirects (page title: "Authentication required")
- Unexpected navigation (page title doesn't match expected page)
- Slow loading (timeout too short for the content)

---

## Common Pitfalls and Lessons Learned

### 1. nodeId Type Assertion Bug

`DOM.getSearchResults` returns `nodeIds` as JSON numbers, which Go unmarshals as `float64`, not `int64` or `string`:

```go
// WRONG: nodeIDs[0].(string) → panic
// WRONG: nodeIDs[0].(int64) → panic
// CORRECT:
var nodeIDFloat float64
nodeIDFloat, ok = nodeIDs[0].(float64)
var nodeID cdp.NodeID = cdp.NodeID(int64(nodeIDFloat))
```

### 2. Stale nodeId After Navigation

After `page.Navigate()`, all previously obtained `nodeId` values become invalid. The DOM agent's internal node map is rebuilt. Using a stale `nodeId` results in "Could not find node with given id" errors.

**Solution:** Always re-find elements after navigation, or use `objectId`-based elements (which are also invalidated by navigation but fail more gracefully).

### 3. Execution Context Destroyed

If the page navigates while `Runtime.evaluate` is executing, the result will contain `exceptionDetails` with a "context destroyed" error. The retry loop handles this by catching the error and retrying.

### 4. SPA Content Rendering Delays

Single Page Applications (like Amazon) may show the page structure immediately but populate form fields and buttons asynchronously. The 10-second retry loop handles this, but selectors must target the *final* rendered element, not intermediate loading states.

---

## File Reference

- **`page_find.go`** — `Find()`, `findByID()`, `findByCSS()`, `findByXPath()`, `FindByXPath()`
- **`element.go`** — `Element` struct, `NewElement()`, `NewElementWithObject()`, `resolveObjectID()`, `IsVisible()`
- **`page.go`** — `ensureAgentsForCommand()`, `sendCommand()`
- **`internal/agent/`** — `AgentManager`, `EnsureAgent()`, `EnabledAgents`

