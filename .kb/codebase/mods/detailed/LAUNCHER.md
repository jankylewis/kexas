# Launcher Module — Detailed Walkthrough

This document dissects every exported/public-facing type and helper inside `kexas/klauncher`. For each function we capture:
- **Purpose** — why it exists.
- **Inputs** — parameters and expectations.
- **Outputs** — return values and error conditions.
- **Important Calls** — downstream helpers invoked.
- **Notes** — gotchas, concurrency rules, OS behavior.

---
## Types & Constants

### `type Options`
| Field | Type | Purpose |
|-------|------|---------|
| `Headless` | `bool` | Request Chrome's `--headless=new` mode. |
| `Port` | `int` | Remote debugging port. `0` means "ask OS for any free port". |
| `Args` | `[]string` | Extra Chromium CLI flags appended verbatim. |
| `ExecutablePath` | `string` | Override auto-discovery with a specific binary path. |
| `WindowWidth` | `int` | Used in `--window-size`. Defaults to `1280`. |
| `WindowHeight` | `int` | Used in `--window-size`. Defaults to `720`. |

### `type Browser`
Wraps the spawned Chromium process.
| Field | Type | Description |
|-------|------|-------------|
| `cmd` | `*exec.Cmd` | Running process handle. |
| `wsURL` | `string` | Full `ws://…/devtools/browser/...` debugger URL scraped from stdout. |
| `log` | `*logger.Logger` | Namespaced logger (`launcher`). |
| `cancelFunc` | `context.CancelFunc` | Cancels `CommandContext`. |
| `userDataDir` | `string` | Unique temp profile directory to delete on close. |
| `logFile` | `string` | Temp stdout/stderr log file (used by `extractDebuggerURL`). |
| `port` | `int` | Debugging port reserved for this Chrome instance. |

---
## Functions

### `func DefaultOptions() *Options`
- **Purpose**: Provide sane defaults so most callers can simply tweak a few fields.
- **Inputs**: None.
- **Outputs**: Pointer to `Options` (Headless=true, Port=0, empty Args, 1280×720).
- **Notes**: Always called when `Launch(nil)` is used.

### `func Launch(ctx context.Context, opts *Options) (*Browser, error)`
- **Purpose**: Entry point for starting Chromium.
- **Inputs**:
  - `ctx`: Cancel context for the process. If `nil`, callers usually go through `kexas.Launch` which injects `context.Background()`.
  - `opts`: Launch configuration. Nil is auto-replaced with `DefaultOptions()`.
- **Outputs**: `*Browser` wrapping the process and debugger URL; error if anything fails.
- **Important Calls**:
  1. `findChromium()` – locate Chrome binary if `ExecutablePath` empty.
  2. `buildArgs(opts)` – compute CLI flags and `userDataDir`.
  3. `os.CreateTemp` – allocate log file for stdout/stderr.
  4. `findFreePort()` – when `opts.Port == 0`.
  5. `cleanOnce.Do(cleanStaleTempProfiles)` – one-time stale profile cleanup.
  6. `killExistingChromeProcesses` / `isPortAvailable` – only when using fixed port 9222.
  7. `cmd.Start()` – spawn Chromium.
  8. `extractDebuggerURL(stdout, 10s)` – poll the log for DevTools URL.
- **Notes**:
  - `cleanOnce` ensures the cleanup race from prior debugging is eliminated.
  - On error after `cmd.Start`, the process is killed and context cancelled.

### `func (b *Browser) WebSocketURL() string`
- **Purpose**: Expose the debugger endpoint to CDP clients.
- **Inputs**: Receiver `*Browser`.
- **Outputs**: `wsURL` string.
- **Notes**: No blocking, just returns cached string.

### `func (b *Browser) Close() error`
- **Purpose**: Tear down Chromium and clean artifacts.
- **Inputs**: Receiver `*Browser`.
- **Outputs**: `error` if killing process fails; otherwise nil.
- **Important Calls**:
  1. `b.cancelFunc()` – cancel context so Chrome gets SIGKILL when necessary.
  2. `b.cmd.Process.Kill()` + `cmd.Wait()` – ensure process exit.
  3. Poll loop using `isPortAvailable` (up to 3s) – warn if port never frees.
  4. `os.RemoveAll(b.userDataDir)` – delete temp profile dir.
  5. `os.Remove(b.logFile)` – delete temp log (ignore ENOENT).
- **Notes**: `Close` must be called even if `Launch` failed after start; use `defer` in callers to avoid leaks.

### `func buildArgs(opts *Options) ([]string, string)`
- **Purpose**: Assemble hardened Chromium arguments and unique profile paths.
- **Inputs**: `opts` containing headless flag, window size, etc.
- **Outputs**: `(args []string, userDataDir string)`.
- **Important Calls**: `cryptorand.Read` for 8 random bytes; `time.Now().UnixNano()` for timestamp component; `filepath.Join(os.TempDir(), ...)` to place dirs under `/tmp` (or OS equivalent).
- **Notes**: Many `--disable-*` flags are set to reduce background noise and improve stability.

### `func extractDebuggerURL(file *os.File, timeout time.Duration, log *logger.Logger) (string, error)`
- **Purpose**: Scrape `DevTools listening on ws://…` from Chrome’s stdout/stderr log.
- **Inputs**: Temp log file handle, total timeout (10s in `Launch`), logger.
- **Outputs**: WebSocket URL or `ErrNoDebuggerURL`.
- **Important Calls**: Rewinds file via `Seek(0,0)` each iteration; uses regex `ws://[^\s]+`; logs Chrome `ERROR/ FATAL` lines for easier diagnosis.
- **Notes**: On timeout, dumps last ~20 lines to logs for debugging.

### `func findFreePort() (int, error)`
- **Purpose**: Let the OS choose a safe TCP port.
- **Inputs**: None.
- **Outputs**: Port number; error if `net.Listen` fails.
- **Notes**: Binds to `127.0.0.1:0`, reads `Addr().(*net.TCPAddr).Port`, closes listener immediately.

### `func isPortAvailable(port int) bool`
- **Purpose**: Probe whether a port is immediately bindable.
- **Inputs**: Port number.
- **Outputs**: `true` if `net.Listen` succeeds; `false` otherwise.
- **Notes**: Used by both the launcher (pre-start) and `Browser.Close` (post-stop).

### `func killExistingChromeProcesses(port int)`
- **Purpose**: Clean up manual/legacy Chrome instances occupying a fixed debugging port (primarily 9222).
- **Inputs**: Port number.
- **Outputs**: None (best-effort, logs warnings on failures).
- **Important Calls**:
  1. `exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))` to find processes bound to the port.
  2. `kill -9` each PID.
  3. `pgrep -f "Google Chrome for Testing"` to find orphan processes by name when port probing fails.
  4. Poll `isPortAvailable` up to 5 seconds to confirm release.
- **Notes**: Only triggered when `opts.Port == 9222` to avoid unnecessary work for auto ports.

### `func cleanStaleTempProfiles(log *logger.Logger)` *(guarded by `cleanOnce`)*
- **Purpose**: Delete leftover `/tmp/kexas-chrome-*` directories from past crashes.
- **Inputs**: Logger for diagnostics.
- **Outputs**: None.
- **Important Calls**: `os.ReadDir(os.TempDir())`, `os.RemoveAll` on matching entries, logging successes/failures.
- **Notes**: Runs exactly once per process; prevents the "SingletonLock deleted mid-launch" race.

### `func findChromium() (string, error)`
- **Purpose**: Locate or download the Chromium binary.
- **Inputs**: None.
- **Outputs**: Absolute path to executable, or `ErrChromiumNotFound`.
- **Important Calls**: `isChromiumInstalled`, `getChromiumPath`, `downloadChromium`, `os.Stat` scanning well-known macOS paths.
- **Notes**: Logging describes whether it used cached, downloaded, or system Chrome, aiding support triage.

---
## Control Flow Summary
```
Launch()
 ├─ if opts nil → DefaultOptions
 ├─ findChromium → execPath
 ├─ buildArgs → args, userDataDir
 ├─ create temp stdout log
 ├─ optionally findFreePort + rebuild args
 ├─ cleanOnce.Do(cleanStaleTempProfiles)
 ├─ optional killExistingChromeProcesses → isPortAvailable
 ├─ cmd.Start
 ├─ extractDebuggerURL (10s timeout)
 └─ return &Browser{cmd, wsURL, ...}

Browser.Close()
 ├─ cancel context
 ├─ cmd.Process.Kill + Wait
 ├─ poll isPortAvailable
 ├─ RemoveAll(userDataDir)
 └─ Remove(logFile)
```

---
## Testing References
`kexas/klauncher/launcher_test.go` covers:
- `TestBuildArgs_UniqueProfileDirs_Concurrent`
- `TestCleanStaleTempProfiles_DeletesActiveWorkerDir`
- `TestCleanOnce_RunsExactlyOnce`
- Log-file cleanup, port availability, stale profile cleanup, etc.

Use those tests as executable documentation when reasoning about future changes.
