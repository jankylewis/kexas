# Chrome Parallel Launch Fix — Lessons Learned

> Date: 2026-02-28
> Status: RESOLVED — 97/100 saucelab tests pass (3 failures are test-logic, 0 launch failures)

---

## The Problem

When running browser tests in parallel (3 workers), Chrome instances failed to launch
with `"could not find debugger URL"` errors and exactly **10.2s durations** — matching
the 10-second extraction timeout + 200ms overhead.

### Symptoms

- 4+ tests failing with identical 10.2s duration
- `ErrNoDebuggerURL` from `extractDebuggerURL`
- Failures only under parallel load, never in serial execution

---

## Root Causes (3 bugs, layered)

### Bug 1: TOCTOU Race in Port Allocation

`findFreePort()` binds to port 0, reads the assigned port, then **releases it**.
Between release and Chrome binding, another worker (or OS process) can grab the port.

```
Worker A: findFreePort() → port 51234 → listener.Close()
                                         ↑ RACE WINDOW
Worker B: findFreePort() → port 51234 → listener.Close()
Worker A: Chrome tries to bind 51234 → FAILS (Worker B took it)
```

### Bug 2: File-Based Stderr Polling Has Buffering Lag

The old code wrote Chrome's stderr to a temp file, then polled it with
`file.Seek(0,0)` + `bufio.Scanner`. Under parallel OS load, file I/O buffering
caused the `"DevTools listening on"` line to arrive **after** the 10s timeout.

### Bug 3: `cleanStaleTempProfiles()` Deleted Active Worker Dirs

Called on every `Launch()` call, it deleted ALL `kexas-chrome-*` temp dirs — including
ones belonging to other workers whose Chrome processes were still starting up.
Error: `"No such file or directory (2)"` (not `"already locked"`).

---

## The Fix (3 parts, referencing Playwright & go-rod)

### Fix 1: Chrome Self-Assigned Port (`--remote-debugging-port=0`)

**How Playwright does it:**
- Default: `--remote-debugging-pipe` (stdio pipe transport, no port at all)
- Fallback (Selenium/CDP): `--remote-debugging-port=0`
- Source: `packages/playwright-core/src/server/chromium/chromium.ts`

**How go-rod does it:**
- Default: `defaults.Port = "0"` → `--remote-debugging-port=0`
- Source: `lib/launcher/launcher.go`

**Our fix:**
Pass `--remote-debugging-port=0` directly. Chrome atomically picks a free port
and prints it to stderr. We extract the actual port from the WebSocket URL.

```go
// Before (TOCTOU race):
freePort, _ = findFreePort()  // port released here
opts.Port = freePort
// Chrome may fail to bind — port grabbed by another process

// After (atomic):
// opts.Port stays 0 → --remote-debugging-port=0
// Chrome picks its own port, prints: DevTools listening on ws://127.0.0.1:<PORT>/...
actualPort = extractPortFromWSURL(wsURL)
```

### Fix 2: Pipe-Based Real-Time Stderr Reading

**How Playwright does it:**
- `stdio: 'pipe'` in `launchProcess()`
- `browserLogsCollector.onMessage(callback)` — event-driven, not polling
- `ManualPromise` resolves when `DevTools listening on` is found
- Source: `packages/playwright-core/src/server/browserType.ts`

**How go-rod does it:**
- `URLParser` implements `io.Writer` interface
- Piped as `cmd.Stderr = io.MultiWriter(logger, parser)`
- `parser.URL` is a `chan string` — channel-based, zero-copy
- Source: `lib/launcher/url_parser.go`

**Our fix:**
Replace `os.CreateTemp` + file polling with `cmd.StderrPipe()` + goroutine scanner.

```go
// Before (file-based polling with buffering lag):
stdout, _ = os.CreateTemp("", "kexas-chrome-*.log")
cmd.Stderr = stdout
wsURL, _ = extractDebuggerURL(stdout, 10*time.Second, log)  // polls file

// After (pipe-based real-time):
stderrPipe, _ = cmd.StderrPipe()
cmd.Stdout = nil
wsURL, _ = extractDebuggerURLFromPipe(stderrPipe, 10*time.Second, log)
```

The pipe reader uses a goroutine + channel pattern (similar to go-rod's `URLParser`):
1. Goroutine reads lines from pipe via `bufio.Scanner`
2. Sends result to a `chan result` when URL found or error detected
3. Main goroutine does `select` with timeout

### Fix 3: `sync.Once` for Stale Profile Cleanup

Wrap `cleanStaleTempProfiles()` in `sync.Once` so it only runs on the first
`Launch()` call — before any worker has created its temp dir.

```go
var cleanOnce sync.Once
// In Launch():
cleanOnce.Do(func() { cleanStaleTempProfiles(log) })
```

---

## Defense-in-Depth Measures

### Retry with Exponential Backoff

In `ktest/ktest_parallel.go`, `launchIsolatedBrowser` retries up to 3 times
with 500ms → 1s → 2s backoff. Catches transient OS-level failures.

### Stagger Worker Startup

Workers launch with a 300ms stagger delay to reduce OS resource contention
when multiple Chrome instances start simultaneously.

```go
if id > 0 {
    time.Sleep(time.Duration(id) * 300 * time.Millisecond)
}
```

### Crypto-Random User Data Dirs

Each launch creates a user data dir with a crypto-random 16-char hex suffix
(not timestamp-based), eliminating any possibility of directory name collision.

---

## Key Takeaways

1. **Never use find-then-use for ports** — always let the process self-assign.
   Both Playwright and go-rod use `port=0` as the default.

2. **Never poll files for process output** — use pipes or `io.Writer` interfaces.
   Both Playwright (`stdio: 'pipe'`) and go-rod (`io.MultiWriter`) stream in real-time.

3. **Cleanup must be idempotent and safe** — `sync.Once` for shared cleanup,
   crypto-random dirs for isolation.

4. **The 10.2s clue was diagnostic gold** — matching timeout + overhead proves
   the failure is in the extraction path, not in Chrome itself.

---

## Files Changed

| File | Change |
|------|--------|
| `launcher/launcher.go` | Port=0 self-assign, pipe-based stderr, extractPortFromWSURL |
| `launcher/launcher_cleanup.go` | Extracted cleanup/port utilities (500-line limit) |
| `launcher/launcher_parallel_test.go` | 12 critical tests for pipe extraction, port parsing, TOCTOU |
| `ktest/ktest_parallel.go` | Retry with backoff, stagger delay |

## Test Results

- **Unit tests:** 26/26 pass (launcher package)
- **Saucelab e2e:** 97/100 pass (3 failures are test-logic in sort dropdown, 0 launch failures)
