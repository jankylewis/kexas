# Chrome Crash Prevention

> Lessons learned from Chrome crash debugging, including zombie process cleanup, stale profile management, and mach_vm_read error prevention.

**Last Updated:** February 27, 2026

---

## Root Cause: mach_vm_read Crash

Chrome crashes with `mach_vm_read(0x..., 0x8000): (os/kern) invalid address (1)` when:

1. A previous Chrome process was killed (Ctrl+C, IDE stop, or concurrent launch)
2. Its shared memory regions (in the temp profile) still exist on disk
3. The new Chrome instance tries to read the stale shared memory → crash

This is a **macOS-specific** issue related to Crashpad (Chrome's crash reporter) accessing shared memory from a dead process.

---

## Prevention Mechanisms (launcher.go)

### 1. Kill by Port (lsof)

```go
exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
```

Finds processes bound to the debugging port (e.g., 9222) and kills them with `kill -9`.

### 2. Kill by Name (pgrep)

```go
exec.Command("pgrep", "-f", "Google Chrome for Testing")
```

Catches orphaned Chrome processes that may have released the port but are still alive (e.g., helper processes, GPU process).

### 3. Clean Stale Temp Profiles

```go
cleanStaleTempProfiles(log)
```

Removes all `kexas-chrome-*` directories from `os.TempDir()`. These contain shared memory files, lock files, and cache that cause mach_vm_read errors.

### 4. Settle Delay (500ms)

After killing processes and cleaning profiles, wait 500ms for the OS to fully release shared memory and file locks before launching Chrome.

### 5. Port Release Polling (5s)

After killing processes, poll `isPortAvailable()` every 200ms for up to 5 seconds. This handles the case where the OS takes time to reclaim the port after SIGKILL.

### 6. --disable-crash-reporter Flag

Prevents Crashpad from running, which eliminates the mach_vm_read calls that cause the crash.

### 7. --disable-dev-shm-usage Flag

Forces Chrome to use `/tmp` instead of `/dev/shm` for shared memory, avoiding contention with stale shared memory regions.

---

## Launch Sequence (Critical Order)

```
1. killExistingChromeProcesses(port)   → kill by port + kill by name
2. cleanStaleTempProfiles(log)         → remove kexas-chrome-* dirs
3. time.Sleep(500ms)                   → settle delay
4. isPortAvailable(port)               → verify port is free
5. cmd.Start()                         → launch Chrome
6. extractDebuggerURL(stdout, 10s)     → wait for ws:// URL
```

---

## Common Crash Scenarios

### Scenario 1: Ctrl+C in Terminal
- Go process dies immediately
- Chrome stays alive as orphan
- Next launch: port conflict + stale profile → crash
- **Fix:** Kill by name catches orphans

### Scenario 2: IDE Stop (GoLand/VS Code)
- IDE sends SIGKILL → Go defers don't run → browser.Close() never called
- Chrome stays alive
- **Fix:** Kill by port + kill by name on next launch
- **Better fix:** Configure IDE to send SIGINT (soft exit)

### Scenario 3: Concurrent Launches (Right-click "Go Run" twice)
- Two Go processes try to use port 9222 simultaneously
- First Chrome is killed by second launch's cleanup
- Stale shared memory from first Chrome causes mach_vm_read in second
- **Fix:** Settle delay (500ms) + stale profile cleanup

---

## Key Chrome Flags for Stability

| Flag | Purpose |
|------|---------|
| `--disable-crash-reporter` | Prevents Crashpad mach_vm_read calls |
| `--disable-dev-shm-usage` | Avoids shared memory contention |
| `--disable-gpu` | Prevents GPU process crashes |
| `--disable-infobars` | Suppresses info bars without yellow banner |
| `--no-first-run` | Skips first-run dialogs |

**NEVER use:** `--disable-blink-features=AutomationControlled` (causes yellow banner)

---

## Unit Tests

Critical unit tests in `launcher/launcher_test.go`:

- `TestCleanStaleTempProfiles_RemovesKexasProfiles` — verifies stale profile cleanup
- `TestCleanStaleTempProfiles_Idempotent` — verifies double-cleanup safety
- `TestKillExistingChromeProcesses_PortBecomesAvailable` — verifies port freed after kill
- `TestKillExistingChromeProcesses_NoProcesses` — verifies no-op safety
- `TestRapidRelaunch_NoPortConflict` — end-to-end Ctrl+C + relaunch simulation
- `TestBuildArgs_ContainsCriticalFlags` — verifies crash prevention flags present
- `TestBuildArgs_HeadlessFlag` — verifies headless mode toggle

---

## File Reference

- **`launcher/launcher.go`** — `killExistingChromeProcesses()`, `cleanStaleTempProfiles()`, `buildArgs()`, `Launch()`
- **`launcher/launcher_test.go`** — All crash prevention unit tests
