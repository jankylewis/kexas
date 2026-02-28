# Parallel Chrome Management: kexas vs Playwright vs go-rod

> Date: 2026-02-28
> Purpose: Deep comparison of how each framework handles parallel browser instances

---

## 1. Port Allocation — Avoiding the TOCTOU Race

The #1 pitfall in parallel Chrome management is port collision. If you find a free
port and then tell Chrome to use it, another process can grab it in between.

### Playwright

- **Default**: Uses `--remote-debugging-pipe` — no port at all. Communication
  happens over stdio file descriptors (fd 3 and fd 4), completely eliminating
  port-based races.
- **Fallback** (when `cdpPort` is specified by user): passes the user's port
  directly. Does NOT auto-find ports.
- Source: `packages/playwright-core/src/server/chromium/chromium.ts`

### go-rod

- **Default**: `defaults.Port = "0"` → `--remote-debugging-port=0`
- Chrome picks its own port atomically via the OS kernel's `SO_REUSEADDR` path.
- The actual port is extracted from the `ws://127.0.0.1:<PORT>/devtools/...`
  line printed to stderr.
- Source: `lib/launcher/launcher.go`, line ~200

### kexas (now fixed)

- **Before**: `findFreePort()` → `net.Listen("tcp", ":0")` → close → pass port
  to Chrome. Classic TOCTOU race.
- **After**: `--remote-debugging-port=0` (same as go-rod). Port extracted from
  WebSocket URL via `extractPortFromWSURL()`.

### Critical Pitfall

```
❌ NEVER do this:
   port = findFreePort()      // Step 1: find
   chrome --port=<port>       // Step 2: use  ← RACE between steps

✅ ALWAYS do this:
   chrome --port=0            // Chrome picks atomically
   port = parseFromStderr()   // Read what Chrome chose
```

---

## 2. Stderr/URL Extraction — Getting the WebSocket Endpoint

Chrome prints `DevTools listening on ws://127.0.0.1:<PORT>/devtools/browser/<ID>`
to **stderr** (not stdout). How you read this line determines reliability.

### Playwright

- Uses `stdio: 'pipe'` in Node.js `child_process.spawn()` options.
- `browserProcess.stderr` is piped to a `BrowserLogsCollector`.
- A `ManualPromise<string>` is resolved when `DevTools listening on` is found.
- **Event-driven**: no polling, no file I/O. Callback fires immediately when
  the line arrives.
- Also detects `"Failed to create a ProcessSingleton"` for profile-in-use errors.
- Source: `packages/playwright-core/src/server/browserType.ts`

### go-rod

- `URLParser` struct implements `io.Writer` interface.
- Attached via `cmd.Stderr = io.MultiWriter(stderrLogger, urlParser)`.
- Inside `Write()`, it buffers bytes, scans for the WebSocket URL regex.
- When found, sends URL through `urlParser.URL` channel (`chan string`).
- **Zero-copy streaming**: bytes flow directly from OS pipe to parser.
- Source: `lib/launcher/url_parser.go`

### kexas (now fixed)

- **Before**: `os.CreateTemp()` → `cmd.Stderr = file` → poll file with
  `file.Seek(0,0)` + `bufio.Scanner` in a loop. File buffering caused the
  `"DevTools listening on"` line to arrive after the 10s timeout under load.
- **After**: `cmd.StderrPipe()` → goroutine with `bufio.Scanner` → sends URL
  through `chan string`. Similar pattern to go-rod's channel approach.

### Critical Pitfall

```
❌ NEVER do this:
   file, _ = os.CreateTemp(...)
   cmd.Stderr = file
   // poll file.Seek(0,0) in a loop ← OS buffering delays output

✅ DO this:
   pipe, _ = cmd.StderrPipe()       // direct pipe, no file buffering
   go scanForURL(pipe, resultChan)   // goroutine reads in real-time
```

---

## 3. Process Isolation — User Data Directories

Each parallel Chrome instance needs its own user data directory. If two
instances share a directory, Chrome's `SingletonLock` mechanism will cause
one to fail.

### Playwright

- `fs.promises.mkdtemp(path.join(os.tmpdir(), 'playwright_'))` — random suffix
  via Node.js built-in.
- Detects `"Failed to create a ProcessSingleton"` error and retries.
- Source: `packages/playwright-core/src/server/browserType.ts`

### go-rod

- `MkdirTemp("", "rod-user-data-*")` with `utils.RandString(8)` for extra
  uniqueness.
- Source: `lib/launcher/launcher.go`

### kexas

- `os.MkdirTemp("", "kexas-chrome-*")` — OS provides random suffix.
- Additional crypto-random 16-char hex suffix for defense-in-depth.

### Critical Pitfall — Cleanup Race

```
❌ NEVER do this:
   // On every Launch():
   cleanStaleTempProfiles()  // deletes ALL kexas-chrome-* dirs
   //                           ↑ including ones used by other workers!

✅ DO this:
   var cleanOnce sync.Once
   cleanOnce.Do(func() { cleanStaleTempProfiles() })
   // Only runs once, before any worker has created its dir
```

---

## 4. Retry / Recovery Strategies

### Playwright

- Retries on `"Inconsistency detected by ld.so"` (glibc race in Linux).
- Single retry via `_innerLaunchWithRetries`.
- Detects `ProcessSingleton` failure for profile-in-use and gives clear error.
- Source: `packages/playwright-core/src/server/browserType.ts`

### go-rod

- No built-in retry in the launcher itself.
- Relies on `rod.Browser.Connect()` with timeout and context cancellation.
- Users implement retry at the application level.

### kexas

- 3 retries with exponential backoff (500ms → 1s → 2s) in
  `launchIsolatedBrowser()`.
- Stagger delay (300ms × workerID) between worker launches to reduce OS
  resource contention.

---

## 5. Process Cleanup

### Playwright

- Tracks all launched browsers in a `Set<BrowserProcess>`.
- `process.on('exit', ...)` + `process.on('SIGINT', ...)` handlers kill all.
- `browserProcess.kill()` sends SIGKILL after a grace period.

### go-rod

- `leakless` package: wraps Chrome in a parent process that monitors the Go
  process. If Go crashes, the parent kills Chrome. Cross-platform.
- Source: `lib/launcher/launcher.go` (`leakless.New()`)

### kexas

- `sync.Once`-guarded cleanup of stale temp profiles.
- `killExistingChromeProcesses(port)` via `lsof` for port-specific cleanup.
- `defer browser.Close()` in each worker goroutine.

---

## Summary Matrix

| Concern | Playwright | go-rod | kexas |
|---------|-----------|--------|-------|
| **Port allocation** | Pipe (no port) or user-specified | `port=0` | `port=0` |
| **Stderr reading** | `stdio: 'pipe'` + event callback | `io.Writer` + channel | `StderrPipe` + goroutine + channel |
| **User data dir** | `mkdtemp` + random | `MkdirTemp` + `RandString` | `MkdirTemp` + crypto-random hex |
| **Retry** | 1× on glibc race | None (user-level) | 3× exponential backoff |
| **Cleanup** | Process set + SIGKILL | `leakless` parent process | `sync.Once` + `lsof` |
| **Stagger** | None | None | 300ms × workerID |
| **Singleton detection** | Detects + clear error | Not explicit | `sync.Once` prevents |

---

## Key Lessons

1. **Pipe-based communication beats file-based** — all three frameworks use
   pipes or streaming interfaces, never file polling.

2. **Let the OS assign ports** — both go-rod and kexas use `port=0`. Playwright
   goes further by avoiding ports entirely (`--remote-debugging-pipe`).

3. **Temp directory isolation is non-negotiable** — every framework creates
   unique directories. The cleanup of stale dirs must never touch active ones.

4. **Leakless process management** (go-rod) is the gold standard for ensuring
   Chrome doesn't outlive its parent. kexas and Playwright use signal handlers.

5. **Retry with backoff** is kexas's unique strength — neither Playwright nor
   go-rod retry Chrome launches by default.
