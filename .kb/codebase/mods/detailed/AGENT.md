# Agent Module — Deep Detailed Walkthrough

Source: `kexas/internal/agent/agent_manager.go`, `agent_context.go`, `agent_states.go`, `enabled_agents.go`

This document explains the Agent Manager system at the OS, process, memory, and protocol level. The Agent Manager is the invisible layer that makes CDP domain management automatic — you never manually call `DOM.enable` or `Runtime.enable`.

---

## 1. What Is a "CDP Agent"?

Chrome DevTools Protocol organizes its API into **domains** (also called "agents"). Each domain controls a specific aspect of the browser:

| Domain | What It Controls | Must Enable? |
|--------|-----------------|--------------|
| `DOM` | DOM tree inspection, node queries, mutations | Yes |
| `Runtime` | JavaScript evaluation, object inspection | Yes |
| `Page` | Navigation, screenshots, lifecycle events | Yes |
| `Input` | Keyboard/mouse event simulation | **No** (always available) |
| `Network` | HTTP request/response interception | Yes |
| `Security` | Certificate info, security state | Yes |
| `Debugger` | Breakpoints, step execution | Yes |
| `Profiler` | CPU profiling, code coverage | Yes |

Domains marked "Yes" require an explicit `<Domain>.enable` command before you can use them. Without enabling, most commands in that domain will fail, and you won't receive any events from it.

### What Happens in Chrome When You Enable a Domain

When Chrome receives `DOM.enable`:
1. Chrome's browser process routes the message to the renderer process for that session.
2. The renderer's `InspectorDOMAgent` starts observing DOM mutations.
3. It builds an initial snapshot of the document tree (node IDs, types, attributes).
4. From this point, any DOM change triggers an event (e.g., `DOM.childNodeInserted`, `DOM.attributeModified`) sent back through the CDP WebSocket.
5. This consumes memory in the renderer (tracking data structures) and CPU (event serialization).

This is why enabling a domain has a cost — and why the Agent Manager exists to do it lazily.

---

## 2. The `AgentManager` Struct — What Lives in Memory

```go
type AgentManager struct {
    enabledAgents EnabledAgents   // 8 booleans: which domains are enabled
    agentContext  AgentContext    // 8 pointers: domain-specific runtime data
    agentState    AgentStates    // 8 pointers: detailed state per domain
    cdp           *cdp.Client    // WebSocket client (shared, not owned)
    sessionID     string         // CDP session this manager is scoped to
    mutex         sync.RWMutex   // protects concurrent access
    config        *Config        // auto-enable, timeouts, debug flags
    log           *logger.Logger // structured logger scoped to "agent"
}
```

### Memory Layout

- **`enabledAgents`** — An `EnabledAgents` struct with 8 `bool` fields (DOM, Input, Runtime, etc.). Each `bool` is 1 byte in Go. Total: 8 bytes. This is the hot path — checked on every `sendCommand` call. Being a flat struct (no pointers, no heap allocation) makes it cache-friendly.

- **`agentContext`** — An `AgentContext` struct with 8 pointer fields, each pointing to a domain-specific context struct (e.g., `*DOMContext`, `*RuntimeContext`). Initially all `nil` (zero cost). Contexts are allocated only when a domain is enabled and returns initialization data. For example, `DOM.enable` returns the document root, which is stored in `DOMContext.Root`.

- **`agentState`** — An `AgentStates` struct with 8 `*AgentState` pointers. Each `AgentState` is ~56 bytes (bool + Time + interface{} + int + error). These are pre-allocated during `initAgentStates` — 8 × 56 = 448 bytes.

- **`mutex`** — A `sync.RWMutex` (24 bytes on 64-bit systems). Allows concurrent reads (`RLock`) but exclusive writes (`Lock`). In practice, reads vastly outnumber writes because most calls just check `enabledAgents.IsEnabled()`.

- **`cdp`** — 8-byte pointer to the shared `cdp.Client`. The Agent Manager does NOT own this — it's the same client used by the `Browser` and all `Page` objects.

- **`sessionID`** — 16-byte string header. This scopes all enable commands to a specific tab. When the Agent Manager sends `DOM.enable`, it includes this session ID so Chrome enables DOM only for this tab, not all tabs.

### Total Footprint

~600 bytes per `AgentManager`. Since each `Page` gets one `AgentManager`, and each parallel worker gets one `Page`, a 6-worker parallel test run allocates 6 × 600 = ~3.6 KB for agent management. Negligible.

---

## 3. `NewAgentManager` — Construction

```go
func NewAgentManager(cdpConn *cdp.Client, sessionID string) *AgentManager {
    return NewAgentManagerWithConfig(cdpConn, sessionID, DefaultConfig())
}
```

### `DefaultConfig()`

```go
func DefaultConfig() *Config {
    return &Config{
        AutoEnable:     true,           // lazily enable domains on first use
        EnableTimeout:  5 * time.Second, // timeout for enable commands
        ContextTimeout: 30 * time.Minute, // how long to cache context data
        Debug:          false,
    }
}
```

- **`AutoEnable: true`** — The key flag. When `true`, calling `EnsureAgent("DOM")` will automatically send `DOM.enable` if DOM isn't enabled yet. When `false`, it would return an error instead.
- **`EnableTimeout`** — If `DOM.enable` takes longer than 5 seconds, the enable call times out. This can happen if Chrome is overloaded.
- **`ContextTimeout`** — Context data (like the DOM root node) is considered valid for 30 minutes. After that, a re-enable would refresh it. In practice, navigation invalidates context much sooner.

### `initAgentStates()`

```go
func (am *AgentManager) initAgentStates() {
    var commonAgents []string = []string{
        AgentDOM, AgentInput, AgentRuntime, AgentNetwork,
        AgentPage, AgentSecurity, AgentDebugger, AgentProfiler,
    }
    for i := 0; i < len(commonAgents); i++ {
        var agentState *AgentState = &AgentState{
            Enabled: false, LastUsed: time.Time{}, Context: nil, MemoryUsage: 0, Error: nil,
        }
        am.agentState.SetState(agent, agentState)
    }
}
```

This pre-allocates an `AgentState` struct for each of the 8 known domains. All start with `Enabled: false`. The `LastUsed` field is the zero `time.Time` (January 1, year 1), meaning "never used." This pre-allocation avoids nil checks later — `GetState("DOM")` always returns a valid pointer.

---

## 4. `EnabledAgents` — The Hot-Path Boolean Tracker

```go
type EnabledAgents struct {
    DOM, Input, Runtime, Network, Page, Security, Debugger, Profiler bool
}
```

### `IsEnabled(agentName string) bool`

A `switch` statement on the agent name string, returning the corresponding boolean. This is called on **every** `sendCommand` call (via `ensureAgentsForCommand` → `EnsureAgent` → `IsEnabled`). 

**Performance analysis**: The `switch` compiles to a series of string comparisons. For 8 cases, Go's compiler may generate a jump table or a binary search depending on optimization level. In practice, the most common agents (DOM, Runtime, Page) are checked first due to the `switch` ordering. The entire check takes ~5–20 nanoseconds.

### `Enable(agentName string)` / `Disable(agentName string)`

Same `switch` pattern, but sets the boolean to `true` / `false`. These are called only when an enable command succeeds (rare — once per domain per page lifetime).

### `List() []string`

Returns a slice of enabled agent names. Allocates a new `[]string` each call (because it appends dynamically). Used for debugging/logging, not in hot paths.

---

## 5. `AgentContext` — Domain-Specific Runtime Data

```go
type AgentContext struct {
    DOM      *DOMContext
    Input    *InputContext
    Runtime  *RuntimeContext
    Network  *NetworkContext
    Page     *PageContext
    Security *SecurityContext
    Debugger *DebuggerContext
    Profiler *ProfilerContext
}
```

Each context struct holds data returned by the domain's `enable` command:

### `DOMContext`

```go
type DOMContext struct {
    Root *DOMRoot `json:"root"`
}
type DOMRoot struct {
    NodeID   int    `json:"nodeId"`
    NodeType int    `json:"nodeType"`
    NodeName string `json:"nodeName"`
}
```

When `DOM.enable` succeeds, Chrome returns the document root node. `DOMContext.Root` caches this so subsequent DOM commands can use the root node ID without re-querying. The root node is typically `nodeId: 1`, `nodeType: 9` (Document node), `nodeName: "#document"`.

### `InputContext`

```go
type InputContext struct {
    Ready     bool      `json:"ready"`
    EnabledAt time.Time `json:"enabled_at"`
}
```

Input doesn't require explicit enabling, but the context tracks when it was first used for observability.

### `GetContext` / `SetContext` / `ClearContext`

These use the same `switch` pattern as `EnabledAgents`. `GetContext` returns `interface{}` (any type) because each domain has a different context struct. Callers must type-assert:

```go
var domCtx *DOMContext = am.agentContext.GetContext("DOM").(*DOMContext)
```

`ClearContext` sets the pointer to `nil`, making the context GC-eligible. This is called during `ResetAgent` (e.g., after navigation invalidates the DOM tree).

---

## 6. `AgentStates` — Detailed Per-Agent Tracking

```go
type AgentState struct {
    Enabled     bool        // is this domain currently enabled?
    LastUsed    time.Time   // when was the last command using this domain?
    Context     interface{} // domain-specific cached data
    MemoryUsage int         // estimated memory used by this domain's tracking
    Error       error       // last error from this domain (nil if healthy)
}
```

### `CountEnabled() int`

Iterates all 8 agents, counting those with `Enabled == true`. Used for diagnostics.

### `TotalMemoryUsage() int`

Sums `MemoryUsage` across all agents. This is an **estimate** — the actual Chrome-side memory for DOM tracking, network interception, etc. is much larger and not directly measurable from Go.

### `ResetAll()`

Sets all agents to disabled state, clears context and error, resets `LastUsed` to zero time. Called during page close or session teardown. After `ResetAll`, the next `EnsureAgent` call will re-enable from scratch.

---

## 7. The Lazy Enable Flow — End to End

Here's what happens when you call `page.Find("#login")` for the first time on a fresh page:

```
page.Find("#login")
  └── page.sendCommand("DOM.querySelector", ...)
        └── page.ensureAgentsForCommand("DOM.querySelector")
              └── strings.HasPrefix("DOM.querySelector", "DOM.") → true
                    └── agentManager.EnsureAgent("DOM")
                          ├── am.mutex.RLock()
                          ├── am.enabledAgents.IsEnabled("DOM") → false (first time)
                          ├── am.mutex.RUnlock()
                          ├── am.mutex.Lock()  (upgrade to write lock)
                          ├── am.cdp.SendToSession(ctx, sessionID, "DOM.enable", nil)
                          │     └── [CDP WebSocket round-trip: ~1 ms]
                          ├── am.enabledAgents.Enable("DOM") → DOM = true
                          ├── am.agentState.GetState("DOM").Enabled = true
                          ├── am.agentState.GetState("DOM").LastUsed = time.Now()
                          ├── am.mutex.Unlock()
                          └── return nil
        └── browser.client.SendToSession(ctx, sessionID, "DOM.querySelector", params)
              └── [CDP WebSocket round-trip: ~1 ms]
```

On the second `Find` call:

```
page.Find("#password")
  └── page.sendCommand("DOM.querySelector", ...)
        └── page.ensureAgentsForCommand("DOM.querySelector")
              └── agentManager.EnsureAgent("DOM")
                    ├── am.mutex.RLock()
                    ├── am.enabledAgents.IsEnabled("DOM") → true (cached!)
                    ├── am.mutex.RUnlock()
                    └── return nil  (no CDP call, ~5 ns)
        └── browser.client.SendToSession(ctx, sessionID, "DOM.querySelector", params)
```

The second call skips the enable step entirely — just a boolean check behind a read lock.

---

## 8. Concurrency Safety

The `AgentManager` uses `sync.RWMutex` to protect all mutable state:

- **Read operations** (`IsEnabled`, `GetState`, `GetContext`) acquire `RLock()` — multiple goroutines can read simultaneously.
- **Write operations** (`Enable`, `SetState`, `SetContext`) acquire `Lock()` — exclusive access, blocks all readers and writers.

### Why Mutex Instead of Atomic?

`enabledAgents` has 8 booleans. Using `sync/atomic` would require 8 separate atomic operations or a single `uint64` with bitmasking. The mutex approach is simpler and the performance difference is negligible since:
1. Lock contention is rare — each parallel worker has its OWN `AgentManager` (no sharing).
2. The critical section is tiny (boolean check → return).
3. Read locks are non-blocking when there's no writer.

In the parallel test scenario, each worker has an independent `Browser → Page → AgentManager` chain. There is **zero contention** between workers' agent managers.

---

## 9. Agent Constants

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

These constants are used throughout the codebase as the canonical names for CDP domains. Using constants instead of raw strings prevents typos and enables IDE auto-completion.

---

## 10. How Agent Manager Fits in the Architecture

```
┌─────────────────────────────────────────────────┐
│                   Your Test Code                 │
│  page.Navigate(url)  page.Find(sel)  elem.Click()│
└───────────────┬─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────┐
│                  Page.sendCommand                │
│  "Is the required domain enabled?"               │
│  ┌──────────────────────────┐                    │
│  │     AgentManager         │                    │
│  │  enabledAgents.DOM=true? │──── yes → skip     │
│  │  enabledAgents.DOM=false?│──── send DOM.enable│
│  └──────────────────────────┘                    │
│  Then: browser.client.SendToSession(...)         │
└───────────────┬─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────┐
│              CDP WebSocket Client                │
│  Serialize JSON → write to TCP socket            │
└───────────────┬─────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────┐
│           Chrome Renderer Process                │
│  Execute command → return result                 │
└─────────────────────────────────────────────────┘
```

The Agent Manager is a **transparent optimization layer**. Remove it, and everything still works — you'd just send redundant `enable` commands on every CDP call, wasting ~0.5 ms each time. With it, the overhead drops to ~5 nanoseconds (a boolean check) for all but the first call to each domain.
