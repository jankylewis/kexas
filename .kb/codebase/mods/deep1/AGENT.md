# Agent Manager Module — Deep1 Level (For Software Engineering Students)

Source: `kexas/internal/agent/agent_manager.go`, `agent_context.go`, `agent_states.go`, `enabled_agents.go`

Every technical term is defined. FAQs per section.

---

## 1. What Is the Agent Manager?

Chrome DevTools Protocol (CDP) organizes its features into "domains" — groups of related commands. For example:
- The **DOM domain** handles everything related to the page's HTML structure (finding elements, reading attributes).
- The **Runtime domain** handles JavaScript execution (running code, inspecting objects).
- The **Page domain** handles navigation, screenshots, and page lifecycle events.
- The **Input domain** handles keyboard and mouse simulation.

Most domains must be **explicitly enabled** before you can use them. If you try to call `DOM.querySelector` without first calling `DOM.enable`, Chrome will return an error.

The Agent Manager automates this. It sits between your code and Chrome, and every time you send a command, it checks: "Does this command need a domain that isn't enabled yet? If so, enable it first." You never have to think about domain enablement.

### Key Terms

- **Domain (Agent)** — A group of related CDP commands. "Domain" and "agent" are used interchangeably. Chrome calls them "domains" in documentation; kexas calls them "agents" in code.
- **Enable** — Sending a `<Domain>.enable` command to Chrome to activate a domain. After enabling, Chrome starts tracking state for that domain and accepting its commands.
- **Lazy enablement** — Don't enable a domain until someone actually needs it. This saves resources.

> **FAQ: Why does Chrome require explicit enablement?**
> Performance. When you enable the DOM domain, Chrome starts tracking every DOM mutation (element added, attribute changed, text modified) and maintaining a mirror of the DOM tree. Enabling the Network domain makes Chrome report every HTTP request. If all domains were always enabled, Chrome would waste CPU and memory tracking things nobody asked for. Explicit enablement means you only pay for what you use.

> **FAQ: What is `sync/atomic`?**
> `sync/atomic` is a Go standard library package that provides atomic operations — operations that complete in a single, uninterruptible step. For example, `atomic.AddInt64(&counter, 1)` increments a counter safely even if multiple goroutines call it simultaneously. Kexas uses `sync.RWMutex` (from the `sync` package) instead of `sync/atomic` because it needs to protect multiple fields at once, not just a single integer.

---

## 2. The `AgentManager` Struct

```go
type AgentManager struct {
    enabledAgents EnabledAgents   // 8 booleans: which domains are on
    agentContext  AgentContext    // cached data from domain enables
    agentState    AgentStates    // detailed state per domain
    cdp           *cdp.Client    // WebSocket connection to Chrome
    sessionID     string         // which tab this manager belongs to
    mutex         sync.RWMutex   // concurrency protection
    config        *Config        // settings (timeouts, debug flags)
    log           *logger.Logger // structured logger
}
```

### `EnabledAgents` — The Boolean Tracker

```go
type EnabledAgents struct {
    DOM, Input, Runtime, Network, Page, Security, Debugger, Profiler bool
}
```

Eight boolean fields — one per domain. When `DOM` is `true`, it means the DOM domain has been enabled for this session.

> **What is a boolean?** A type with only two possible values: `true` or `false`. In Go, a `bool` takes 1 byte of memory. Eight booleans = 8 bytes total — tiny.

The most important method: `IsEnabled(agentName string) bool` — a `switch` statement that returns the appropriate boolean.

> **What is a switch statement?** Go's way of selecting between multiple options based on a value. Like a multi-way if-else. `switch agentName { case "DOM": return ea.DOM; case "Runtime": return ea.Runtime; ... }`. This is called on every single `sendCommand` call, so it needs to be fast. For 8 cases, it takes ~5-20 nanoseconds.

### `AgentContext` — Cached Domain Data

When you enable some domains, Chrome returns useful data. For example, `DOM.enable` returns the root node of the document. Rather than re-fetching this data every time, the Agent Manager caches it.

```go
type AgentContext struct {
    DOM      *DOMContext      // e.g., root node ID from DOM.enable
    Runtime  *RuntimeContext  // e.g., execution context info
    // ... one per domain
}
```

> **FAQ: What is a "context" here (not to be confused with Go's context.Context)?**
> In this code, "context" means "domain-specific runtime data." `DOMContext` stores the DOM root node. `RuntimeContext` stores JavaScript execution environment info. This is completely unrelated to Go's `context.Context` (which is for cancellation/timeouts). Naming collision, unfortunately.

### `AgentStates` — Per-Domain Metadata

```go
type AgentState struct {
    Enabled     bool        // is this domain currently on?
    LastUsed    time.Time   // when was it last used?
    Context     interface{} // cached data
    MemoryUsage int         // estimated memory (Chrome-side)
    Error       error       // last error, nil if healthy
}
```

This is for observability — you can check which domains are active, when they were last used, and if any have errors. Useful for debugging flaky tests.

### `sync.RWMutex` — Concurrency Protection

> **What is a mutex?** Short for "mutual exclusion." A lock that prevents multiple goroutines from accessing shared data simultaneously. When goroutine A acquires the lock (`mutex.Lock()`), goroutine B must wait until A releases it (`mutex.Unlock()`). This prevents data races — situations where two goroutines read/write the same data at the same time, producing corrupted results.

> **What is a RWMutex?** A "readers-writer" mutex. It allows:
> - **Multiple readers simultaneously** — If goroutines only need to read data (not modify it), they can all hold the read lock at the same time. `mutex.RLock()` / `mutex.RUnlock()`.
> - **Exclusive writer** — If a goroutine needs to modify data, it gets an exclusive write lock. No readers or writers can proceed until it's done. `mutex.Lock()` / `mutex.Unlock()`.
>
> This is perfect for the Agent Manager: most calls just check `IsEnabled()` (read), and only the rare enable call modifies data (write).

> **FAQ: Why does the Agent Manager need a mutex if each Page has its own?**
> In practice, there's almost no contention — each worker has its own AgentManager. But the mutex is there for correctness guarantees. If someone accidentally shared a Page between goroutines (a bug), the mutex prevents data corruption instead of producing mysterious failures.

---

## 3. `EnsureAgent` — The Core Method

```go
func (am *AgentManager) EnsureAgent(agentName string) error {
    am.mutex.RLock()
    if am.enabledAgents.IsEnabled(agentName) {
        am.mutex.RUnlock()
        return nil  // already enabled — fast path
    }
    am.mutex.RUnlock()
    
    // Slow path: need to enable
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    // Double-check (another goroutine might have enabled it while we waited for the lock)
    if am.enabledAgents.IsEnabled(agentName) {
        return nil
    }
    
    // Send enable command to Chrome
    _, err := am.cdp.SendToSession(am.ctx, am.sessionID, agentName+".enable", nil)
    if err != nil {
        return fmt.Errorf("failed to enable agent %s: %w", agentName, err)
    }
    
    am.enabledAgents.Enable(agentName)
    am.agentState.GetState(agentName).Enabled = true
    am.agentState.GetState(agentName).LastUsed = time.Now()
    return nil
}
```

### The Double-Check Pattern

Notice the code checks `IsEnabled` twice — once with a read lock, once with a write lock. This is a well-known pattern called **double-checked locking**:

1. **First check (read lock)** — Fast path. If the domain is already enabled, return immediately. Most calls take this path.
2. **Second check (write lock)** — After acquiring the write lock, check again. Between releasing the read lock and acquiring the write lock, another goroutine might have enabled the domain. Without this second check, we'd send `DOM.enable` twice, which is wasteful (though not harmful).

> **FAQ: What is a "fast path" and "slow path"?**
> - **Fast path** — The common case that executes quickly. Here: domain already enabled → return immediately (read lock only, ~5 ns).
> - **Slow path** — The rare case that does more work. Here: domain not enabled → acquire write lock, send CDP command (~0.5 ms), update state.
> - Good code optimizes for the fast path because it runs thousands of times more often.

### What Happens When a Domain Is Enabled

When Chrome receives `DOM.enable`:
1. Chrome's renderer process activates its `InspectorDOMAgent`.
2. The agent starts observing DOM mutations (elements added/removed, attributes changed).
3. Chrome sends back the document root node information.
4. From now on, Chrome accepts DOM-domain commands (`DOM.querySelector`, `DOM.getAttributes`, etc.) for this session.

---

## 4. `ResetAgent` — Invalidating After Navigation

```go
func (am *AgentManager) ResetAgent(agentName string) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    am.enabledAgents.Disable(agentName)
    am.agentContext.ClearContext(agentName)
    am.agentState.GetState(agentName).Enabled = false
}
```

After a page navigation, the DOM tree is completely destroyed and rebuilt. The cached DOM root node is invalid. `ResetAgent("DOM")` marks the DOM domain as disabled and clears its cached context. The next `sendCommand` call that needs DOM will re-enable it and get fresh data.

> **FAQ: Why not just keep the domain enabled?**
> The domain is still enabled in Chrome's view, but the cached data (root node ID, execution contexts) is stale. By marking it as "disabled" locally, we force `EnsureAgent` to re-run the enable flow, which refreshes the cached data. It's a pragmatic approach — re-enabling is cheap (~0.5 ms), and the alternative (tracking exactly which cached data is stale) would be much more complex.

---

## 5. How It Fits Together

Here's the complete flow when you call `page.Find("#login")` for the first time:

```
Your code: page.Find("#login")
  │
  ├── page.sendCommand("DOM.querySelector", {selector: "#login"})
  │     │
  │     ├── page.ensureAgentsForCommand("DOM.querySelector")
  │     │     │
  │     │     └── Is the method prefix "DOM."? → YES
  │     │           │
  │     │           └── agentManager.EnsureAgent("DOM")
  │     │                 │
  │     │                 ├── RLock → IsEnabled("DOM")? → NO (first time!)
  │     │                 ├── RUnlock
  │     │                 ├── Lock (exclusive)
  │     │                 ├── Double-check: IsEnabled("DOM")? → still NO
  │     │                 ├── Send "DOM.enable" to Chrome via WebSocket ← ~0.5ms
  │     │                 ├── Chrome activates DOM tracking
  │     │                 ├── Set EnabledAgents.DOM = true
  │     │                 ├── Unlock
  │     │                 └── return nil
  │     │
  │     └── browser.client.SendToSession("DOM.querySelector", ...) ← ~1ms
  │           Chrome searches its DOM tree for "#login"
  │           Returns nodeId: 42
  │
  └── return &Element{nodeID: 42, selector: "#login", page: p}
```

Second call to `page.Find("#password")`:

```
Your code: page.Find("#password")
  │
  ├── page.sendCommand("DOM.querySelector", {selector: "#password"})
  │     │
  │     ├── page.ensureAgentsForCommand("DOM.querySelector")
  │     │     │
  │     │     └── agentManager.EnsureAgent("DOM")
  │     │           │
  │     │           ├── RLock → IsEnabled("DOM")? → YES (cached!)
  │     │           ├── RUnlock
  │     │           └── return nil  ← ~5 nanoseconds, NO network call
  │     │
  │     └── browser.client.SendToSession("DOM.querySelector", ...) ← ~1ms
  │
  └── return &Element{nodeID: 47, selector: "#password", page: p}
```

The second call skips the enable step entirely. Over a test with 50 DOM operations, that's 49 skipped enable calls — saving ~25ms.

---

## 6. The Agent Constants

```go
const (
    AgentDOM      = "DOM"
    AgentInput    = "Input"
    AgentRuntime  = "Runtime"
    AgentNetwork  = "Network"
    AgentPage     = "Page"
    AgentSecurity = "Security"
    AgentDebugger = "Debugger"
    AgentProfiler = "Profiler"
)
```

> **FAQ: Why use constants instead of just writing `"DOM"` everywhere?**
> 1. **Typo prevention** — If you write `AgentDOm` (typo), the compiler catches it. If you write `"DOm"` (string), nothing catches it until runtime.
> 2. **IDE support** — Constants show up in autocomplete. Strings don't.
> 3. **Refactoring** — If Chrome renames a domain, you change one constant instead of searching for strings across the codebase.

---

## 7. Why It Matters — The Big Picture

Without the Agent Manager, Page code would look like:

```go
// BAD — manual domain management
func (p *Page) Find(selector string) (*Element, error) {
    if !p.domEnabled {
        _, err := p.sendCommand("DOM.enable", nil)
        p.domEnabled = true
    }
    result, err := p.sendCommand("DOM.querySelector", {selector})
    // ...
}
```

Every function would need its own domain check. With 6 domains and 20+ functions, that's a lot of repeated boilerplate. The Agent Manager centralizes it:

```go
// GOOD — Agent Manager handles it
func (p *Page) Find(selector string) (*Element, error) {
    result, err := p.sendCommand("DOM.querySelector", {selector})
    // sendCommand automatically calls ensureAgentsForCommand → EnsureAgent
}
```

Clean, simple, and correct. The complexity lives in one place (AgentManager), not scattered across every function.
