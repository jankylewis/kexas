# Browser Module — Deep Detailed Walkthrough

Source: `kexas/browser.go`

This document explains every type, function, and internal mechanism in the Browser module at the OS, process, memory, and protocol level. The goal is to give you a complete mental model — no hidden magic.

---

## 1. The `Browser` Struct — What Lives in Memory

```go
type Browser struct {
    process *launcher.Browser   // handle to the OS-level Chromium child process
    client  *cdp.Client         // WebSocket connection speaking CDP JSON-RPC
    log     *logger.Logger      // structured logger scoped to "browser"
    ctx     context.Context     // shared cancellation context
}
```

### Memory Layout (What the Go Runtime Sees)

When you call `kexas.Launch(opts)`, Go allocates a `Browser` struct on the heap (because a pointer `*Browser` is returned). That struct holds:

- **`process`** — A pointer to `launcher.Browser` (itself containing `*exec.Cmd`, the PID of the Chromium child process, file descriptors for the temp log, the temp profile directory path, etc.). This is the bridge between your Go process and the Chromium OS process.
- **`client`** — A pointer to `cdp.Client`, which internally holds a `*websocket.Conn`. This is a full-duplex TCP socket upgraded to WebSocket. The kernel maintains a send buffer and receive buffer for this socket (typically 64KB–256KB each, configurable via `setsockopt`). Every CDP command flows through this single socket.
- **`log`** — Small struct holding the component name string `"browser"` and a reference to the global `slog` handler. No heavy allocation.
- **`ctx`** — A `context.Context` value. In Go, `context.Background()` returns a singleton with no deadline and no cancel function. When you use `LaunchWithContext`, the context carries a cancel channel that, when closed, propagates cancellation to all CDP calls and to the launcher process.

### Why This Matters

The `Browser` struct is the root of the object graph for one Chromium instance. If you lose this pointer (e.g., forget to `defer browser.Close()`), you leak:
1. An OS process (Chromium, typically 100–300 MB RSS).
2. A WebSocket file descriptor (TCP socket in `ESTABLISHED` state).
3. A temp profile directory on disk (10–50 MB).
4. A temp log file on disk (~1 KB).

---

## 2. `Launch(opts *LaunchOptions) (*Browser, error)` — Convenience Constructor

```go
func Launch(opts *launcher.Options) (*Browser, error) {
    return LaunchWithContext(context.Background(), opts)
}
```

### What Happens at the OS Level

This is a thin wrapper. It injects `context.Background()` — a context with no deadline, no cancel, and no values. This means:
- The Chromium process will live until `browser.Close()` is explicitly called.
- There is no automatic timeout on the launch.
- If your program panics or `os.Exit`s without calling `Close()`, the Chromium process becomes an orphan (PPID changes to 1 / launchd on macOS). The temp profile directory stays on disk until the next `cleanStaleTempProfiles` run.

### When to Use

Use `Launch` for simple scripts and test code where you control the lifecycle with `defer browser.Close()`. Use `LaunchWithContext` when you need cancellation (e.g., HTTP handler with request context, or a parent context with a timeout).

---

## 3. `LaunchWithContext(ctx, opts)` — The Real Constructor

This is where all the work happens. Let's walk through it line by line.

### Step 1: Create Logger

```go
var log *logger.Logger = logger.New("browser")
```

`logger.New` allocates a tiny `Logger` struct (component name + handler pointer). No I/O happens yet. The logger uses Go's `log/slog` package internally, which writes to stderr by default.

### Step 2: Launch Chromium Process

```go
var process *launcher.Browser
var err error
process, err = launcher.Launch(ctx, opts)
```

This is the heavyweight call. What happens inside `launcher.Launch`:

1. **Find Chromium binary** — Scans well-known paths (`/Applications/Google Chrome for Testing.app/...`, downloaded cache, etc.). This is a series of `os.Stat` syscalls — each one asks the kernel to look up the inode metadata for a path. If the binary is found, its absolute path is returned.

2. **Build CLI arguments** — Generates the `--user-data-dir=/tmp/kexas-chrome-<timestamp>-<random>` flag. The random suffix comes from `crypto/rand.Read(8 bytes)`, which reads from `/dev/urandom` on macOS/Linux (a kernel CSPRNG — no blocking, no entropy exhaustion). The temp directory is created via `os.MkdirTemp`, which calls the `mkdtemp(3)` C library function, which itself issues a `mkdir(2)` syscall.

3. **Create temp log file** — `os.CreateTemp("", "kexas-chrome-*.log")` creates a file in `/tmp` with a unique name. The kernel allocates an inode, a directory entry, and returns a file descriptor (integer) to the Go runtime.

4. **Spawn Chromium** — `exec.CommandContext(ctx, execPath, args...).Start()`. Under the hood:
   - Go calls `fork(2)` to create a child process (exact copy of the Go process's address space via copy-on-write page tables).
   - Immediately calls `execve(2)` to replace the child's memory image with the Chromium binary. The kernel loads the Mach-O (macOS) or ELF (Linux) binary, sets up the stack, heap, and program counter, then starts execution.
   - The child process inherits file descriptors for stdout/stderr (redirected to the temp log file via `cmd.Stdout = logFile`).
   - The child process's PID is stored in `cmd.Process.Pid`.

5. **Extract debugger URL** — Chromium writes `DevTools listening on ws://127.0.0.1:<port>/devtools/browser/<guid>` to stderr. `extractDebuggerURL` polls the temp log file every 100ms (up to 10 seconds) by seeking to offset 0 and re-reading. This is a `lseek(2)` + `read(2)` loop. A regex extracts the `ws://...` URL.

6. **One-time cleanup** — `cleanOnce.Do(cleanStaleTempProfiles)` uses `sync.Once` to ensure stale temp directories from previous crashed runs are deleted exactly once per process. This prevents the race where Worker 1's cleanup deletes Worker 0's active profile directory.

If any of these steps fail, the function kills the child process (`cmd.Process.Kill()` → `kill(2)` sending SIGKILL), cancels the context, removes temp files, and returns an error.

### Step 3: Get WebSocket URL

```go
var wsURL string = process.WebSocketURL()
```

Simple getter — returns the string that was scraped from Chromium's stderr. No I/O.

### Step 4: Connect CDP Over WebSocket

```go
var client *cdp.Client
client, err = cdp.Connect(ctx, wsURL)
```

What happens at the network/OS level:

1. **DNS resolution** — The URL is `ws://127.0.0.1:<port>/...`, so no DNS lookup is needed. `127.0.0.1` maps directly to the loopback interface.

2. **TCP connection** — `net.Dial("tcp", "127.0.0.1:<port>")` issues a `connect(2)` syscall. The kernel creates a socket, performs the TCP three-way handshake (SYN → SYN-ACK → ACK) entirely within the loopback stack (no actual network traffic — the kernel short-circuits loopback packets). This is extremely fast (~microseconds).

3. **WebSocket upgrade** — An HTTP `GET` with `Upgrade: websocket` header is sent over the TCP socket. Chrome's DevTools server responds with `101 Switching Protocols`. From this point, the connection speaks the WebSocket framing protocol (2-byte header + payload).

4. **CDP client setup** — The `cdp.Client` spawns a background goroutine that continuously reads WebSocket frames, dispatches responses to waiting callers (matched by JSON-RPC message ID), and processes CDP events.

If CDP connect fails, the function closes the launcher process to avoid leaking the Chromium child.

### Step 5: Build and Return Browser

```go
return &Browser{
    process: process,
    client:  client,
    log:     log,
    ctx:     ctx,
}, nil
```

The `&Browser{...}` literal causes a heap allocation. The Go compiler's escape analysis determines that this struct escapes the function (returned as a pointer), so it goes on the heap rather than the stack. The garbage collector will manage this memory until all references are dropped and `Close()` has been called.

---

## 4. `NewPage() (*Page, error)` — Creating a New Tab

### What Happens at the Chrome Process Level

```go
result, err = b.client.Send(b.ctx, "Target.createTarget", map[string]interface{}{
    "url": "about:blank",
})
```

1. **CDP command** — The Go process serializes `{"id":N, "method":"Target.createTarget", "params":{"url":"about:blank"}}` to JSON, wraps it in a WebSocket text frame, and writes it to the socket.

2. **Chrome receives it** — Chrome's browser process (the main/broker process) creates a new renderer process (or reuses one from the renderer pool). On macOS, this means Chrome calls `posix_spawn` or `fork+exec` to create a child process of itself. This new process gets its own virtual address space, V8 JavaScript engine instance, and Blink rendering engine instance.

3. **Target ID returned** — Chrome responds with `{"id":N, "result":{"targetId":"<GUID>"}}`. The target ID is a UUID string identifying this tab.

4. **Attach to target** — `b.attachToPage(targetID)` sends `Target.attachToTarget` with `flatten: true`. This tells Chrome to create a **flattened CDP session** — meaning all CDP messages for this tab go through the same WebSocket connection (multiplexed by session ID), rather than opening a second WebSocket. This is critical for performance: one TCP connection serves all tabs.

### What `attachToPage` Does in Detail

1. **Session creation** — Chrome returns a `sessionId` (another UUID). From now on, all CDP commands for this tab include this session ID.

2. **Domain enablement** — Sends `Network.enable` and `Page.enable` through the session. These are like "subscriptions" — telling Chrome to start sending events (e.g., `Network.requestWillBeSent`, `Page.loadEventFired`) for this tab. Without enabling a domain, you get no events and some commands may fail.

3. **Stealth script injection** — `Page.addScriptToEvaluateOnNewDocument` injects JavaScript that runs **before** any page script. This JavaScript:
   - Deletes `navigator.webdriver` (which is `true` for automated browsers).
   - Overrides `navigator.plugins` to report fake plugins (non-automated browsers have plugins).
   - Patches `Notification.permission` to return `"default"` instead of `"denied"`.
   - Overrides `navigator.languages` to report realistic values.
   - Patches `chrome.runtime` to look like a normal Chrome extension API.
   - Handles iframe `contentWindow` access to prevent detection via cross-origin checks.

   These scripts are stored in Chrome's browser process and re-injected on every navigation. They live in memory for the lifetime of the session.

4. **Agent Manager creation** — `agent.NewAgentManager(b.client, sessionID)` creates a struct that tracks which CDP domains (DOM, Runtime, Input, etc.) are enabled for this specific session. This avoids sending redundant `DOM.enable` commands.

5. **Page struct** — A new `Page` is allocated on the heap with references to the browser, session ID, target ID, logger, context, and agent manager.

---

## 5. `FirstPage() (*Page, error)` — Attaching to the Default Tab

When Chromium starts, it always creates one initial tab (the `about:blank` page). `FirstPage` finds it:

1. **Enumerate targets** — `Target.getTargets` returns all targets (pages, service workers, browser, etc.).
2. **Filter** — Loop through `targetInfos`, looking for the first entry where `type == "page"`.
3. **Attach** — Pass the target ID to `attachToPage` (same flow as above).

### Why FirstPage Instead of NewPage?

`NewPage` creates a **second** tab, leaving the initial tab unused (wasting a renderer process). `FirstPage` reuses the already-existing tab, saving ~50–100 MB of memory per unused tab.

---

## 6. `Close() error` — Tearing Everything Down

```go
func (b *Browser) Close() error {
    b.client.Close()      // close WebSocket
    err := b.process.Close()  // kill Chromium, remove temp files
    return err
}
```

### What Happens at the OS Level

1. **`b.client.Close()`** — Sends a WebSocket close frame (`opcode 0x8`), waits for the peer's close frame, then calls `conn.Close()` which issues `close(2)` on the file descriptor. The kernel transitions the TCP socket through `FIN_WAIT_1 → FIN_WAIT_2 → TIME_WAIT` (or `CLOSE_WAIT` if Chrome closes first). The socket stays in `TIME_WAIT` for ~60 seconds (configurable via `net.ipv4.tcp_fin_timeout` on Linux, `net.inet.tcp.msl` on macOS) to handle late-arriving packets.

2. **`b.process.Close()`** — This is the launcher's `Close` method:
   - **Cancel context** — `b.cancelFunc()` closes the context's `Done()` channel. Any goroutines blocked on `<-ctx.Done()` wake up and can clean up.
   - **Kill process** — `cmd.Process.Kill()` sends `SIGKILL` (signal 9) to the Chromium browser process. `SIGKILL` cannot be caught or ignored — the kernel immediately terminates the process, reclaims its memory pages, closes its file descriptors, and reaps its child processes (renderer processes). On macOS, the renderer processes are children of the browser process, so they also get `SIGKILL` via process group cleanup.
   - **Wait for exit** — `cmd.Wait()` blocks until the kernel reports the process has exited (via `waitpid(2)`). This collects the exit status and prevents the process from becoming a zombie.
   - **Port release poll** — Polls `isPortAvailable(port)` for up to 3 seconds. Even after the process exits, the kernel may hold the port in `TIME_WAIT`. This poll ensures the port is truly free before returning, which matters for rapid re-launches.
   - **Delete temp profile** — `os.RemoveAll(userDataDir)` recursively deletes the temp profile directory. This calls `unlink(2)` for each file and `rmdir(2)` for each directory. The kernel decrements inode reference counts; when they reach zero, the disk blocks are freed.
   - **Delete temp log** — `os.Remove(logFile)` deletes the stdout/stderr log file.

### Memory Reclamation

After `Close()`:
- Chromium's virtual memory (typically 500 MB–2 GB virtual, 100–300 MB RSS) is fully reclaimed by the kernel.
- The Go-side `Browser`, `cdp.Client`, and `launcher.Browser` structs become eligible for garbage collection once no more references exist. The GC will reclaim them during the next collection cycle.
- Temp files on disk are deleted synchronously — disk space is freed immediately (unless another process holds an open file descriptor to them, which is unlikely).

---

## 7. Error Handling Patterns — Deep Explanation

### Type Assertion Safety

Every CDP response is a `map[string]interface{}` (JSON decoded). The code uses the two-value type assertion pattern:

```go
var targetID string
var ok bool
targetID, ok = result["targetId"].(string)
if !ok {
    return nil, fmt.Errorf("invalid target ID response")
}
```

**Why two-value?** In Go, a single-value type assertion (`x := val.(string)`) panics if the assertion fails. The two-value form (`x, ok := val.(string)`) never panics — it returns the zero value and `false`. Since CDP responses come from an external process (Chrome), we must **never** trust their shape.

### Error Wrapping

```go
return nil, fmt.Errorf("failed to launch browser: %w", err)
```

The `%w` verb wraps the original error, creating an error chain. Callers can use `errors.Is(err, target)` or `errors.As(err, &target)` to inspect the root cause. This is critical for debugging: a test failure log might show `failed to launch browser: failed to find Chromium: ErrChromiumNotFound`, giving the full causal chain.

### Stealth Errors Are Intentionally Ignored

```go
_, _ = client.SendToSession(ctx, sessionID, "Network.setUserAgentOverride", params)
```

Stealth commands (user agent override, script injection) ignore errors because:
1. They are **best-effort** — the browser works fine without them.
2. Some Chrome versions may not support all stealth commands.
3. Failing on stealth would prevent legitimate testing when anti-detection isn't needed.

---

## 8. Process Tree Visualization

```
Your Go Process (PID 1234)
├── goroutine: main test runner
├── goroutine: cdp.Client read loop (reads WebSocket frames)
├── goroutine: cdp.Client write loop (sends WebSocket frames)
└── child process: Chromium browser (PID 5678)
    ├── renderer process (PID 5679) — Tab 1
    ├── renderer process (PID 5680) — Tab 2 (if NewPage called)
    ├── GPU process (PID 5681)
    ├── utility process (PID 5682) — network service
    └── ... (Chrome spawns many helper processes)
```

The Go process communicates with the **browser process** (PID 5678) over WebSocket. The browser process internally routes CDP commands to the correct renderer process via IPC (Mojo on modern Chrome). Your Go code never directly talks to renderer processes.

---

## 9. WebSocket Connection — Protocol Deep Dive

```
Go Process                          Chrome Browser Process
    │                                       │
    │  ──── TCP SYN ─────────────────────►  │  (connect to 127.0.0.1:<port>)
    │  ◄─── TCP SYN-ACK ────────────────  │
    │  ──── TCP ACK ─────────────────────►  │
    │                                       │
    │  ──── HTTP GET /devtools/browser/... ►│  (WebSocket upgrade request)
    │  ◄─── HTTP 101 Switching Protocols ── │
    │                                       │
    │  ──── WS Frame: {"id":1, "method":   │  (CDP command)
    │       "Target.getTargets"} ─────────► │
    │  ◄─── WS Frame: {"id":1, "result":  │  (CDP response)
    │       {"targetInfos":[...]}} ──────── │
    │                                       │
    │  ──── WS Frame: close ──────────────► │  (browser.Close())
    │  ◄─── WS Frame: close ────────────── │
    │                                       │
    │  ──── TCP FIN ─────────────────────►  │
    │  ◄─── TCP FIN-ACK ────────────────── │
```

All CDP traffic is multiplexed over this single WebSocket. Session-scoped commands include a `sessionId` field in the JSON payload to route them to the correct tab's renderer.

---

## 10. Dependencies

| Dependency | Role | OS-Level Impact |
|-----------|------|-----------------|
| `launcher` | Spawns Chromium child process | `fork+exec`, file I/O for temp dirs/logs |
| `cdp` | WebSocket JSON-RPC client | TCP socket, goroutines for read/write loops |
| `internal/agent` | Tracks enabled CDP domains per session | Pure memory (maps and booleans) |
| `internal/logger` | Structured logging via `slog` | Writes to stderr fd |
