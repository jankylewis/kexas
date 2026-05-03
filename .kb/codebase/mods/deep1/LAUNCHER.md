# Launcher Module — Deep1 Level (For SE Students)

Source: `kexas/klauncher/launcher.go`

---

## 1. What Is the Launcher?

The Launcher starts a real Chrome browser as a separate OS process. It finds the Chrome binary, creates an isolated temp directory, allocates a network port, starts Chrome, and extracts the debugging URL for kexas to connect.

### Key Terms

- **Binary** — An executable program file (e.g., `Google Chrome` on macOS).
- **OS process** — A running instance of a program with its own memory and CPU time.
- **Debugging URL** — A WebSocket URL Chrome exposes: `ws://127.0.0.1:9222/devtools/browser/abc-123`.

---

## 2. What It Creates Per Launch

| Artifact | Example | Cleanup |
|----------|---------|---------|
| Chrome process | PID 12345 | SIGKILL on Close() |
| Temp profile dir | `/tmp/kexas-chrome-1709012345-abc/` | os.RemoveAll on Close() |
| Temp log file | `/tmp/kexas-chrome-*.log` | os.Remove on Close() |
| TCP port | `127.0.0.1:9222` | Released when process dies |

---

## 3. Launch Steps

### Step 1: Clean Stale Profiles (sync.Once)

> **What is sync.Once?** Ensures a function runs exactly once, no matter how many goroutines call it. Uses an atomic flag internally.

> **What is a stale temp profile?** If your program crashes, `Close()` never runs, leaving temp dirs behind. `cleanStaleTempProfiles` deletes all `/tmp/kexas-chrome-*` from previous crashed runs.

> **FAQ: Why sync.Once?** In parallel mode, 6 workers call `Launch()` simultaneously. Without sync.Once, Worker 1's cleanup deletes Worker 0's active directory — causing the infamous SingletonLock crash. sync.Once ensures cleanup runs ONCE before any worker creates its directory.

### Step 2: Find Chrome Binary

Searches in order: (1) user-specified path, (2) `~/.kexas/browsers/`, (3) system-installed Chrome.

> **What is `~`?** Shorthand for your home directory (`/Users/yourname` on macOS).

### Step 3: Allocate Free Port

```go
listener, _ := net.Listen("tcp", "127.0.0.1:0")  // :0 = OS picks a free port
port := listener.Addr().(*net.TCPAddr).Port
listener.Close()  // release it so Chrome can use it
```

> **What is a TCP port?** A number (0-65535) identifying a specific service. Like an apartment number in a building (IP address). `127.0.0.1` means localhost — your own computer.

> **FAQ: Isn't there a race condition?** Theoretically yes — between releasing the port and Chrome binding to it, another program could grab it. In practice, this almost never happens on localhost.

### Step 4: Create Temp Directory

`/tmp/kexas-chrome-<timestamp>-<random>/` — unique per launch. Permission `0700` (owner-only access).

> **What is 0700?** Unix file permission. Owner gets read+write+execute (7), group and others get nothing (0,0).

### Step 5: Build Chrome Arguments

Key flags:

| Flag | Purpose |
|------|---------|
| `--headless=new` | No visible window (works on servers without a display) |
| `--remote-debugging-port=N` | Enable CDP on this port |
| `--user-data-dir=path` | Isolated profile (fresh cookies, cache, history) |
| `--window-size=1280,720` | Viewport size for screenshots |
| `--no-first-run` | Skip welcome dialogs |
| `--disable-gpu` | Avoid GPU issues on headless servers |

> **What is headless mode?** Running Chrome without a GUI. It still loads pages, runs JS, renders layouts — just doesn't display on screen. Essential for CI servers.

### Step 6: Start Chrome Process

```go
cmd := exec.CommandContext(ctx, chromePath, args...)
cmd.Start()  // starts Chrome without waiting for it to finish
```

Under the hood, Go calls `fork()` + `execve()`:

> **What is fork()?** Unix syscall that creates a copy of the current process. Now two processes exist.

> **What is execve()?** Unix syscall that replaces the child process's code with Chrome's code. The child is now running Chrome.

> **What are stdout/stderr?** Standard output streams. stdout (fd 1) for normal output, stderr (fd 2) for errors. Chrome prints its debugger URL to stderr.

### Step 7: Extract Debugger URL

The launcher tails Chrome's log file until it finds: `DevTools listening on ws://127.0.0.1:9222/devtools/browser/abc-123`

This URL is returned to callers so they can connect via WebSocket.

---

## 4. `Close()` — Cleanup

1. Send **SIGKILL** to Chrome process (immediate termination, cannot be ignored).
2. Remove temp profile directory (`os.RemoveAll`).
3. Remove temp log file.

> **What is SIGKILL?** A Unix signal that forcibly terminates a process. Unlike SIGTERM (polite request), SIGKILL cannot be caught or ignored. Used because Chrome sometimes ignores SIGTERM.

> **FAQ: What if Close() is never called?** Chrome keeps running as an orphan process. Temp dir stays on disk. Always use `defer browser.Close()` after Launch.

---

## 5. Parallel Safety

| Concern | Solution |
|---------|----------|
| Stale cleanup deleting active dirs | `sync.Once` — cleanup runs once before any worker |
| Port collisions between workers | Each `findFreePort()` gets a unique port from the OS |
| Profile directory collisions | Timestamp + random suffix = unique path |
| One worker's crash affecting others | Each worker has its own Chrome process — complete isolation |

---

## 6. Complete Flow

```
launcher.Launch(ctx, opts)
  ├── sync.Once: clean stale /tmp/kexas-chrome-* dirs
  ├── findChromium() → /path/to/chrome
  ├── findFreePort() → 52431
  ├── mkdir /tmp/kexas-chrome-1709012345-abc/
  ├── buildArgs(opts) → ["--headless=new", "--remote-debugging-port=52431", ...]
  ├── exec.CommandContext(chrome, args...).Start()
  ├── tail log until "DevTools listening on ws://..."
  └── return &Browser{process, debuggerURL, tempDir}
```

> **FAQ: How long does Launch take?** Typically 500ms–2 seconds. Chrome startup is the bottleneck. In parallel mode, all 6 workers launch simultaneously, so total wall-clock time is still ~2 seconds.
