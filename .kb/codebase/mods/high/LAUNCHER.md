# Launcher (High-Level)

Source: `kexas/launcher/launcher.go`

## Mission

Spawn throwaway Chromium instances with complete isolation. The launcher hides OS-level complexity (process management, temp directories, port allocation, Chrome flags) so higher layers call `launcher.Launch()` and get a ready-to-use debugger URL.

## What It Creates Per Launch

| Artifact | Location | Cleanup |
|----------|----------|---------|
| Chromium OS process | PID in process table | `SIGKILL` on `Close()` |
| Temp profile directory | `/tmp/kexas-chrome-<timestamp>-<random>/` | `os.RemoveAll` on `Close()` |
| Temp log file | `/tmp/kexas-chrome-*.log` | `os.Remove` on `Close()` |
| TCP port | `127.0.0.1:<port>` | Released when process dies |

## Key Functions

- **`Launch(ctx, opts)`** — Main entry point. Finds Chromium, allocates a port, creates temp dir, starts the process, extracts the debugger URL from Chrome's stdout.
- **`Close()`** — Kills the Chromium process (`SIGKILL`), removes temp profile and log file.
- **`DefaultOptions()`** — Returns sane defaults: headless, 1280×720 viewport, standard Chrome flags.
- **`buildArgs(opts)`** — Converts `Options` into Chrome CLI arguments (`--headless=new`, `--user-data-dir=...`, `--remote-debugging-port=...`, etc.).
- **`findFreePort()`** — Scans for an available TCP port by binding and immediately releasing.
- **`findChromium()`** — Discovers Chrome binary: checks `opts.ExecutablePath`, then `~/.kexas/browsers/`, then system paths.
- **`extractDebuggerURL(logFile)`** — Tails Chrome's stderr log until `DevTools listening on ws://...` appears.
- **`cleanStaleTempProfiles()`** — Deletes leftover `/tmp/kexas-chrome-*` dirs from crashed runs. Runs **once** per process via `sync.Once`.

## Lifecycle

```
launcher.Launch(ctx, opts)
  │
  ├── 1. sync.Once: cleanStaleTempProfiles()   ← first launch only
  ├── 2. findChromium()                         → binary path
  ├── 3. findFreePort()                         → port number
  ├── 4. Create /tmp/kexas-chrome-<unique>/     → user-data-dir
  ├── 5. buildArgs(opts)                        → CLI argument slice
  ├── 6. exec.CommandContext(ctx, binary, args)  → start OS process
  ├── 7. extractDebuggerURL(logFile)            → ws://127.0.0.1:<port>/devtools/browser/<guid>
  └── return &launcher.Browser{process, debuggerURL, tempDir, logFile}
```

## The sync.Once Fix

Without `sync.Once`, every parallel worker's `Launch` call would run `cleanStaleTempProfiles()`, which deletes ALL `/tmp/kexas-chrome-*` directories — including ones actively being used by other workers' Chrome processes. This caused the infamous SingletonLock crashes.

With `sync.Once`, cleanup runs exactly once (on the first `Launch` call), before any worker has created a temp directory.

## Parallel Safety

- Each `Launch` call creates a **unique** temp directory using timestamp + random suffix — no collisions.
- Each `Launch` allocates a **unique** TCP port via `findFreePort()` — no port conflicts.
- Stale cleanup runs **once** before any workers start — no deletion of active directories.

## Why It Matters

The launcher is the foundation of test isolation. Every flaky test, every "SingletonLock" crash, every "address already in use" error traces back to launcher-level concerns. Getting process spawning, port allocation, and cleanup right is essential for reliable parallel execution in CI.
