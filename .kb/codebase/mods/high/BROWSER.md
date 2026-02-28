# Browser (High-Level)

Source: `kexas/browser.go`

## Mission

Bridge the gap between the OS-level Chromium launcher and the high-level Page/Element APIs. `kexas.Browser` owns the Chromium OS process, the CDP WebSocket connection, and all tab sessions derived from it.

## What It Owns

| Resource | Type | Cleanup |
|----------|------|---------|
| Chromium OS process | `*os.Process` (via launcher) | `SIGKILL` on `Close()` |
| WebSocket connection | `*cdp.Client` (TCP socket + goroutines) | `client.Close()` → `close(2)` syscall |
| Temp profile directory | `/tmp/kexas-chrome-*` | `os.RemoveAll` on `Close()` |
| Page sessions | `[]Page` (via `attachToPage`) | Invalidated when WebSocket dies |

## Key Functions

- **`Launch(opts)`** — Spawns Chromium via `launcher.Launch`, connects CDP WebSocket, returns `*Browser`. This is the main entry point.
- **`LaunchWithContext(ctx, opts)`** — Same as `Launch` but with a caller-provided `context.Context` for cancellation control.
- **`NewPage()`** — Creates a new tab via `Target.createTarget`, attaches a CDP session, applies stealth settings, returns `*Page`.
- **`FirstPage()`** — Discovers the default "about:blank" tab that Chromium opens on startup, attaches to it. Avoids creating an extra tab.
- **`Close()`** — Closes the CDP client, kills the Chromium process, removes temp profile. Idempotent — safe to call multiple times.

## Lifecycle

```
kexas.Launch(opts)
  │
  ├── 1. launcher.Launch(ctx, opts)     → Chromium OS process + debugger URL
  ├── 2. cdp.Connect(ctx, debuggerURL)  → WebSocket client (bidirectional)
  ├── 3. return &Browser{client, launcher, ctx, log}
  │
  ▼
browser.NewPage() / browser.FirstPage()
  │
  ├── 4. Target.createTarget / Target.getTargets  → targetID
  ├── 5. Target.attachToTarget(targetID)           → sessionID
  ├── 6. Apply stealth (user-agent, navigator patches)
  ├── 7. return &Page{browser, targetID, sessionID, agentManager}
  │
  ▼
browser.Close()
  │
  ├── 8. client.Close()                → WebSocket TCP FIN
  ├── 9. launcher.Close()              → SIGKILL Chromium + rm temp dir
  └── done
```

## Concurrency Model

- One `Browser` = one Chromium process = one WebSocket connection.
- Multiple `Page` objects can share one `Browser` (multiplexed via `sessionId` in CDP messages).
- In parallel test mode (`ktest`), each worker gets its **own** `Browser` — complete process-level isolation. No shared state between workers.

## Error Handling

- `Launch` errors are fatal — caller cannot proceed without a browser.
- `NewPage`/`FirstPage` errors are wrapped with context (`"failed to create new page: ..."`) and propagated.
- `Close` errors are logged but not propagated — teardown is best-effort.

## Why It Matters

Every test begins with `Launch` and ends with `Close`. The Browser struct is the root of the entire object graph (`Browser → Page → Element`). Getting its lifecycle right — especially cleanup — prevents leaked Chromium processes, orphaned temp directories, and port exhaustion in CI environments.
