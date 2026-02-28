# Browser Module — Deep1 Level (For Software Engineering Students)

Source: `kexas/browser.go`

This document explains the Browser module as if you're a CS/SE student who knows programming basics but is new to browser automation, networking, and OS internals. Every technical term is defined. FAQs are included per section.

---

## 1. What Is This Module?

The Browser module is the **entry point** of kexas. It starts a real Google Chrome browser (as a separate program on your computer), connects to it over a special communication channel, and lets your Go code control it — navigate to websites, click buttons, type text, take screenshots.

### Key Terms

- **Module** — A self-contained unit of code that handles one responsibility. In Go, it's typically one or more `.go` files in the same package.
- **Chrome DevTools Protocol (CDP)** — A JSON-based protocol that Chrome exposes for remote control. Think of it as Chrome's "remote control API." When you open Chrome DevTools (F12), it uses CDP internally.
- **WebSocket** — A communication protocol that allows two programs to send messages back and forth over a single persistent connection. Unlike HTTP (where you send a request and get one response), WebSocket keeps the connection open for continuous bidirectional messaging.

> **FAQ: Why Chrome specifically?**
> Chrome exposes CDP on a TCP port when started with `--remote-debugging-port`. Firefox has a similar protocol, and Safari has another. Kexas currently supports Chrome only. Playwright supports all three because it maintains separate protocol adapters.

> **FAQ: What's the difference between kexas and Selenium?**
> Selenium uses the WebDriver protocol (an HTTP-based REST API) and requires a separate "driver" program (chromedriver). Kexas talks directly to Chrome via CDP over WebSocket — no middleman, lower latency, more features.

---

## 2. The `Browser` Struct — What's Stored in Memory

```go
type Browser struct {
    client   *cdp.Client         // the WebSocket connection to Chrome
    launcher *launcher.Browser   // handle to the Chrome OS process
    ctx      context.Context     // cancellation/timeout propagation
    log      *logger.Logger      // structured logger
}
```

### Field-by-Field Explanation

**`client *cdp.Client`** — This is a pointer to the CDP client object. The CDP client manages the WebSocket connection to Chrome. It can send JSON commands ("navigate to this URL") and receive JSON responses ("navigation complete").

> **What is a pointer?** In Go, `*cdp.Client` means "a memory address pointing to a `cdp.Client` value." The `Browser` struct doesn't contain the full client data — just an 8-byte address that says "the client data is over there in memory." This is important because multiple `Page` objects can share the same client without copying it.

**`launcher *launcher.Browser`** — Handle to the Chromium OS process. This contains the process ID (PID), the temp directory path, and the debugger URL. When you call `browser.Close()`, this handle is used to kill the Chrome process.

> **What is an OS process?** When you run a program (like Chrome), the operating system creates a **process** — an isolated instance of that program with its own memory space, file handles, and CPU time. Your Go program is one process; Chrome is a separate process. They communicate over the network (WebSocket), not by sharing memory.

**`ctx context.Context`** — Go's standard mechanism for cancellation and timeouts.

> **What is `context.Context`?** Think of it as a "cancellation token" that flows through your code. When you call `context.WithTimeout(parent, 30*time.Second)`, you get a new context that automatically cancels after 30 seconds. Any function that receives this context can check `ctx.Done()` to know if it should stop working. This prevents operations from hanging forever.

> **What is `context.Background()`?** It's the "root" context — it never cancels, never times out. It's used as the starting point when you don't need cancellation. Think of it as "run forever unless I explicitly stop it."

**`log *logger.Logger`** — A structured logger scoped to the "browser" component. All log messages from this module are prefixed with `[browser]` for easy filtering.

### FAQ: How much memory does a Browser struct use?

The struct itself is tiny — about 32 bytes (4 pointers × 8 bytes each on a 64-bit system). But the REAL memory cost is the Chrome process it controls: ~150–500 MB of virtual memory for Chrome itself, plus ~50 MB for your Go program. In parallel tests with 6 workers, that's ~1–3 GB total.

---

## 3. `Launch(opts)` — Starting Chrome

```go
func Launch(opts *Options) (*Browser, error) {
    return LaunchWithContext(context.Background(), opts)
}

func LaunchWithContext(ctx context.Context, opts *Options) (*Browser, error) {
    // 1. Start Chrome as a separate OS process
    launcherBrowser, err := launcher.Launch(ctx, launcherOpts)
    
    // 2. Connect to Chrome via WebSocket
    client, err := cdp.Connect(ctx, launcherBrowser.DebuggerURL())
    
    // 3. Return the Browser struct
    return &Browser{client: client, launcher: launcherBrowser, ctx: ctx, log: log}, nil
}
```

### Step 1: Starting Chrome (What Happens at the OS Level)

When `launcher.Launch` is called, here's what happens under the hood:

1. **Find Chrome binary** — The launcher searches for Chrome on your system (e.g., `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome` on macOS).

2. **Create temp directory** — A unique directory like `/tmp/kexas-chrome-1709012345-abc123/` is created. This is Chrome's "user data directory" — where it stores cookies, cache, history, etc. By using a temp directory, each test run starts with a completely fresh browser state.

> **What is a temp directory?** A directory in the system's temporary file area (`/tmp/` on macOS/Linux, `%TEMP%` on Windows). Files here are expected to be short-lived and may be cleaned up by the OS on reboot.

3. **Allocate a TCP port** — The launcher finds a free port (e.g., `9222`) for Chrome's debugging interface.

> **What is a TCP port?** Think of your computer's IP address as a building's street address, and a port as an apartment number. Port numbers range from 0 to 65535. When Chrome listens on port 9222, it means "I'm accepting connections at 127.0.0.1:9222." Your Go program connects to that specific port to talk to Chrome.

> **What is `127.0.0.1`?** It's the "loopback" address — it always refers to your own computer. Traffic to `127.0.0.1` never leaves your machine; it goes through the kernel's network stack but doesn't touch the physical network card. Also called `localhost`.

4. **Start the process** — Go's `exec.Command` calls the OS's `fork()` + `exec()` system calls:
   - **`fork()`** — Creates a copy of the current process (your Go program). Now there are two processes with the same code.
   - **`exec()`** — The child process replaces its code with Chrome's code. Now the child is running Chrome.

> **What is a system call (syscall)?** A request from your program to the operating system's kernel. Your program can't directly access hardware (disk, network, screen) — it must ask the OS via syscalls. `fork()`, `exec()`, `open()`, `read()`, `write()`, `close()` are common syscalls.

5. **Extract debugger URL** — Chrome prints a line like `DevTools listening on ws://127.0.0.1:9222/devtools/browser/abc-123` to its error output. The launcher reads this line to get the WebSocket URL.

### Step 2: Connecting via WebSocket

```
cdp.Connect(ctx, "ws://127.0.0.1:9222/devtools/browser/abc-123")
```

This establishes a WebSocket connection. Here's what happens at the network level:

1. **TCP handshake (3-way)** — Your Go program and Chrome establish a reliable connection:
   - **SYN** — Your program sends a "synchronize" packet to Chrome: "I want to connect."
   - **SYN-ACK** — Chrome responds: "I acknowledge your request and I also want to connect."
   - **ACK** — Your program responds: "I acknowledge your acknowledgment. We're connected."

> **What is TCP?** Transmission Control Protocol. It guarantees that data arrives in order and without errors. If a packet is lost, TCP automatically retransmits it. This is important because CDP commands must arrive intact and in order.

> **What is SYN, ACK?** They're flags (single bits) in the TCP packet header. SYN means "synchronize sequence numbers" (start a connection). ACK means "acknowledgment" (I received your message). The 3-way handshake ensures both sides are ready before sending data.

2. **HTTP Upgrade** — WebSocket starts as an HTTP request with a special header:
   ```
   GET /devtools/browser/abc-123 HTTP/1.1
   Upgrade: websocket
   Connection: Upgrade
   Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
   ```
   Chrome responds with HTTP 101 (Switching Protocols). From this point on, the connection speaks WebSocket, not HTTP.

> **What is the HTTP Upgrade?** HTTP normally follows a request-response pattern (you ask, server answers, connection may close). WebSocket needs a persistent, bidirectional channel. The Upgrade mechanism reuses the existing TCP connection but switches the protocol from HTTP to WebSocket.

3. **Full-duplex communication** — Now both sides can send messages at any time, independently.

> **What is full-duplex?** It means both sides can send and receive simultaneously. Like a phone call where both people can talk at once. The opposite is half-duplex (walkie-talkie: one talks, the other listens) or simplex (one-way only, like a TV broadcast).

### Step 3: Return the Browser

The `Browser` struct is created on the heap and a pointer is returned. From this point, the caller can create pages, navigate, and run tests.

### FAQ: What happens if Chrome fails to start?

`Launch` returns an error. Common causes:
- Chrome binary not found → install Chrome or set `ExecutablePath` in options.
- Port already in use → another Chrome instance is running on that port. The launcher handles this by finding a different free port.
- Insufficient disk space for temp profile → rare, but the error message will mention the temp directory.

### FAQ: What's the difference between `Launch` and `LaunchWithContext`?

`Launch` uses `context.Background()` (never cancels). `LaunchWithContext` lets you pass a context with a timeout or cancellation — useful if you want to abort the launch after N seconds.

---

## 4. `NewPage()` — Creating a New Tab

```go
func (b *Browser) NewPage() (*Page, error) {
    // 1. Ask Chrome to create a new tab
    result, err := b.client.Send(b.ctx, "Target.createTarget", map[string]interface{}{
        "url": "about:blank",
    })
    // 2. Extract the target ID
    targetID := result["targetId"].(string)
    // 3. Attach a CDP session to the new tab
    return b.attachToPage(targetID)
}
```

### What Happens in Chrome

1. **`Target.createTarget`** — This CDP command tells Chrome's browser process: "Create a new tab." Chrome:
   - Allocates a new **renderer process** (or reuses one, depending on site isolation settings).
   - Creates an internal `TargetInfo` entry with a unique UUID (the `targetID`).
   - Navigates the new tab to `about:blank` (an empty page).

> **What is a target?** In CDP terminology, a "target" is anything that can be debugged — a tab, an iframe, a service worker, a web worker. Each target has a unique ID. When we say "target," we usually mean "browser tab."

> **What is a renderer process?** Chrome uses a multi-process architecture. The **browser process** handles UI, networking, and coordination. Each tab (or group of tabs) gets its own **renderer process** that parses HTML, runs JavaScript, and paints pixels. If one tab crashes, others survive.

2. **`Target.attachToTarget`** — This creates a **CDP session** for the new tab. A session is a logical channel multiplexed over the single WebSocket. Each session has a unique `sessionId` string.

> **FAQ: What's the difference between a targetID and a sessionID?**
> - `targetID` identifies the **tab** in Chrome's internal registry. It's permanent for the tab's lifetime.
> - `sessionID` identifies the **debugging session** attached to that tab. You can detach and reattach, getting a new sessionID each time.
> - Think of `targetID` as a room number and `sessionID` as your visitor badge for that room. The room exists even without your badge.

3. **`attachToPage`** — Creates an `AgentManager` for the session, applies stealth settings (user-agent override, navigator patches), and returns a `*Page` struct.

### FAQ: Can I have multiple pages from one browser?

Yes. `NewPage()` can be called multiple times. Each call creates a new tab with its own session. All tabs share the same WebSocket connection — CDP demultiplexes by `sessionId`.

---

## 5. `FirstPage()` — Using the Default Tab

```go
func (b *Browser) FirstPage() (*Page, error) {
    // 1. List all existing targets
    result, err := b.client.Send(b.ctx, "Target.getTargets", nil)
    // 2. Find the first "page" type target
    // 3. Attach to it
    return b.attachToPage(targetID)
}
```

When Chrome starts, it always opens one default tab (usually `about:blank`). `FirstPage()` finds that tab and attaches to it, avoiding the overhead of creating an extra tab.

### FAQ: Why not just always use NewPage()?

`NewPage()` creates an **additional** tab. If you do `Launch()` then `NewPage()`, you now have 2 tabs — the default blank one and the new one. `FirstPage()` reuses the existing blank tab, so you have exactly 1 tab. This saves ~50 MB of memory (one fewer renderer process).

---

## 6. `Close()` — Shutting Down

```go
func (b *Browser) Close() error {
    // 1. Close the CDP WebSocket connection
    b.client.Close()
    // 2. Kill the Chrome process and clean up temp files
    b.launcher.Close()
}
```

### What Happens at the OS Level

1. **WebSocket close** — Sends a WebSocket Close frame, then calls `close(2)` on the TCP socket file descriptor.

> **What is a file descriptor (fd)?** In Unix/macOS/Linux, everything is a file — including network connections. When your program opens a TCP connection, the OS assigns a small integer (e.g., fd=5) that represents that connection. You use the fd to read from and write to the connection. `close(2)` tells the OS: "I'm done with fd 5, free the resources."

> **What is a WS (WebSocket) frame?** WebSocket messages are wrapped in "frames" — small headers (2–14 bytes) prepended to the message payload. The header contains: opcode (text/binary/close/ping/pong), payload length, and an optional masking key. A Close frame (opcode 0x8) signals a graceful shutdown.

2. **Kill Chrome** — Sends `SIGKILL` to the Chrome process.

> **What is SIGKILL?** A Unix signal that immediately terminates a process. Unlike `SIGTERM` (polite request to exit), `SIGKILL` cannot be caught or ignored — the OS forcibly destroys the process. We use `SIGKILL` instead of `SIGTERM` because Chrome sometimes ignores `SIGTERM` or takes too long to shut down.

3. **Remove temp directory** — `os.RemoveAll("/tmp/kexas-chrome-xxx/")` recursively deletes the profile directory.

> **What is a stale temp profile?** If your program crashes (or you Ctrl+C during a test), `Close()` never runs, leaving the temp directory behind. These orphaned directories are "stale." The launcher's `cleanStaleTempProfiles()` function (run via `sync.Once` on the first launch) cleans them up.

4. **TCP connection cleanup** — After the process dies, the OS releases the TCP port. There may be a brief `TIME_WAIT` state (typically 60 seconds) where the port can't be reused.

> **What is TIME_WAIT?** After a TCP connection closes, the OS keeps the port in a `TIME_WAIT` state for a short time. This ensures any delayed packets from the old connection don't get confused with a new connection on the same port. In kexas, this is rarely a problem because each launch uses a different port.

### FAQ: What if Close() is never called?

The Chrome process keeps running as an orphan. It consumes memory and CPU. The temp directory stays on disk. In CI, this can lead to resource exhaustion. Always use `defer browser.Close()` immediately after `Launch()`.

### FAQ: Is Close() safe to call multiple times?

Yes. It's designed to be idempotent — calling it twice does nothing harmful.

---

## 7. The Ownership Chain

```
Browser (your Go code)
  │
  ├── Owns: cdp.Client (WebSocket connection)
  ├── Owns: launcher.Browser (Chrome OS process + temp dir)
  │
  ├── Creates: Page A (sessionID "abc")
  │     └── Owns: AgentManager A (tracks which CDP domains are enabled)
  │
  └── Creates: Page B (sessionID "xyz")
        └── Owns: AgentManager B (independent from A)
```

**Key rule**: When `Browser.Close()` is called, ALL pages become invalid. The WebSocket is gone, so any `sendCommand` call will fail. Always close pages before (or let the browser close handle it).

---

## 8. How It All Connects — Complete Flow

```
1. Your test calls kexas.Launch()
     │
     ▼
2. Launcher finds Chrome binary, allocates port, creates temp dir
     │
     ▼
3. Launcher starts Chrome via fork() + exec()
     Chrome prints: "DevTools listening on ws://127.0.0.1:9222/devtools/browser/guid"
     │
     ▼
4. cdp.Connect() does: TCP handshake → HTTP Upgrade → WebSocket established
     Now your Go program and Chrome can talk bidirectionally
     │
     ▼
5. browser.FirstPage() → Target.getTargets → Target.attachToTarget → sessionID
     Now you have a Page object bound to a specific tab
     │
     ▼
6. page.Navigate("https://example.com") → sends CDP command over WebSocket
     Chrome loads the page (DNS → TCP → TLS → HTTP → HTML parsing)
     │
     ▼
7. page.Find("#login") → sends DOM.querySelector over WebSocket
     Chrome returns nodeID → kexas wraps it in an Element struct
     │
     ▼
8. element.Click() → sends Runtime.callFunctionOn("this.click()") over WebSocket
     Chrome clicks the element, fires DOM events
     │
     ▼
9. browser.Close() → close WebSocket → SIGKILL Chrome → rm temp dir
     Everything is cleaned up
```
