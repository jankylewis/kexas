# Agent Manager (High-Level)

Source: `kexas/internal/agent/agent_manager.go`, `agent_context.go`, `agent_states.go`, `enabled_agents.go`

## Mission

Automatically manage CDP domain enablement so that `Page` and `Element` code never has to manually call `DOM.enable` or `Runtime.enable`. The Agent Manager is a transparent optimization layer — it intercepts every CDP command, checks if the required domain is enabled, and enables it on demand.

## The 8 CDP Domains

| Domain | What It Controls | Auto-Enabled By |
|--------|-----------------|-----------------|
| `DOM` | DOM tree queries, node inspection | `DOM.*` commands |
| `Runtime` | JavaScript evaluation, object references | `Runtime.*` commands |
| `Page` | Navigation, screenshots, lifecycle | `Page.*` commands |
| `Input` | Keyboard/mouse simulation | Always available (no enable needed) |
| `Network` | HTTP request/response monitoring | Not auto-enabled (future feature) |
| `Security` | Certificate/security state | Not auto-enabled |
| `Debugger` | Breakpoints, stepping | Not auto-enabled |
| `Profiler` | CPU profiling, coverage | Not auto-enabled |

## What It Holds

| Component | Purpose |
|-----------|---------|
| `EnabledAgents` | 8 booleans — which domains are currently enabled (the hot-path check) |
| `AgentContext` | Cached data from domain enables (e.g., DOM root node from `DOM.enable`) |
| `AgentStates` | Per-domain metadata: last used time, memory estimate, error state |
| `sync.RWMutex` | Protects all mutable state for concurrent access |
| `cdp.Client` + `sessionID` | Where to send enable commands |

## How It Works

```
page.Find("#login")
  └── sendCommand("DOM.querySelector", ...)
        └── ensureAgentsForCommand("DOM.querySelector")
              └── agentManager.EnsureAgent("DOM")
                    ├── IsEnabled("DOM")? → YES → return (1 nanosecond)
                    └── IsEnabled("DOM")? → NO  → send "DOM.enable" → flip to true → return
```

- **First call to a domain**: 1 extra CDP round-trip (~0.5 ms) to send `<domain>.enable`.
- **All subsequent calls**: Pure boolean check (~1 ns). No I/O.

## Key Functions

- **`EnsureAgent(name)`** — Check if enabled; if not, send enable command. Thread-safe (RWMutex).
- **`ResetAgent(name)`** — Mark domain as disabled, clear cached context. Called after navigation invalidates DOM state.
- **`GetContext(name)` / `SetContext(name, data)`** — Read/write domain-specific cached data.
- **`ResetAll()`** — Disable all domains and clear all context. Called during session teardown.

## Concurrency

- Each `Page` gets its own `AgentManager` — no sharing between pages.
- In parallel mode, each worker has its own `Browser → Page → AgentManager` chain. Zero contention.
- The `sync.RWMutex` allows concurrent reads (most calls) with exclusive writes (rare enable operations).

## Why It Matters

Without the Agent Manager, every `sendCommand` call would need to manually check and enable the required CDP domain — or just enable everything upfront (wasteful). The lazy pattern saves both developer effort and Chrome resources, especially in parallel tests where 6 workers × 8 domains = 48 potential enable calls that may never be needed.
