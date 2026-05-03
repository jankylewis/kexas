# Parallel Test Execution — Use Cases

> 3 real-world scenarios explaining how kexas runs tests with `parallelSet = 2`.

---

## Scenario 1: Single File with 1 AlphaInit

**Setup**: One file `signin_tests.go` containing 1 `AlphaInit` with 2 tests.

```go
// signin_tests.go
var _ = kexas.AlphaInit(
    ktest.BeforeAll(func() { fmt.Println("setup") }),
    ktest.Group("Signin", func() {
        ktest.Test("TestLogin", TestLogin)
        ktest.Test("TestLogout", TestLogout)
    }),
    ktest.AfterAll(func() { fmt.Println("cleanup") }),
)
```

### What happens step by step

```
1. REGISTRATION
   Go compiles signin_tests.go
   └─ package-level var triggers AlphaInit
      └─ registers: BeforeAll, Group("Signin"), AfterAll
         └─ Group registers: TestLogin, TestLogout

   Registry now has: [TestLogin, TestLogout]

2. EXECUTION (parallelSet = 2)

   Main goroutine
   ├─ Global BeforeAll → "setup"
   ├─ Sort tests by priority
   ├─ Feed into testQueue channel
   │
   │   testQueue: [TestLogin, TestLogout]
   │              ─────┬──────────┬──────
   │                   │          │
   │           Worker 0          Worker 1
   │           launch browser A  launch browser B
   │           BeforeEach        BeforeEach
   │           run TestLogin     run TestLogout
   │           AfterEach         AfterEach
   │           close browser A   close browser B
   │              │          │
   │              ▼          ▼
   │           resultsChan
   │              │
   ├─ Global AfterAll → "cleanup"
   ├─ Generate HTML report
   └─ Print summary + exit
```

**Key rule**: Only 1 `AlphaInit` allowed per `.go` file. A second call panics.

---

## Scenario 2: Folder with 2 Test Files

**Setup**: `cart/` folder with `cart_a_tests.go` (2 tests) and `cart_b_tests.go` (2 tests).

```go
// cart_a_tests.go
var _ = kexas.AlphaInit(
    ktest.Group("CartA", func() {
        ktest.Test("TestAddItem", TestAddItem)
        ktest.Test("TestRemoveItem", TestRemoveItem)
    }),
)

// cart_b_tests.go
var _ = kexas.AlphaInit(
    ktest.Group("CartB", func() {
        ktest.Test("TestCheckout", TestCheckout)
        ktest.Test("TestEmptyCart", TestEmptyCart)
    }),
)
```

### What happens step by step

```
1. REGISTRATION (happens at compile/init time)
   Go loads the package containing both files.
   Each file's AlphaInit runs once (order decided by Go compiler, typically alphabetical).

   cart_a_tests.go → AlphaInit → registers CartA.TestAddItem, CartA.TestRemoveItem
   cart_b_tests.go → AlphaInit → registers CartB.TestCheckout, CartB.TestEmptyCart

   Registry now has 4 tests (single shared list):
   [CartA.TestAddItem, CartA.TestRemoveItem, CartB.TestCheckout, CartB.TestEmptyCart]

2. EXECUTION (parallelSet = 2)

   Main goroutine
   ├─ Global BeforeAll (if any)
   ├─ Sort all 4 tests by priority, then registration order
   ├─ Feed into testQueue
   │
   │   testQueue: [TestAddItem, TestRemoveItem, TestCheckout, TestEmptyCart]
   │               ─────┬──────────────────────────┬────────
   │                    │                           │
   │            Worker 0                    Worker 1
   │            ┌──────────────────┐        ┌──────────────────┐
   │            │ grab TestAddItem │        │ grab TestRemove  │
   │            │ launch browser   │        │ launch browser   │
   │            │ run it           │        │ run it           │
   │            │ close browser    │        │ close browser    │
   │            │                  │        │                  │
   │            │ grab TestCheckout│        │ grab TestEmpty   │
   │            │ launch browser   │        │ launch browser   │
   │            │ run it           │        │ run it           │
   │            │ close browser    │        │ close browser    │
   │            └──────────────────┘        └──────────────────┘
   │                    │                           │
   │                    ▼                           ▼
   │                 resultsChan (4 results)
   │
   ├─ Global AfterAll (if any)
   ├─ Generate HTML report
   └─ Print summary + exit
```

**Important**: Tests from different files can interleave across workers. Worker 0 might run a CartA test then a CartB test. This is fine because each test gets its own browser instance.

---

## Scenario 3: Running All Tests in saucelab/

**Setup**: `saucelab/` contains multiple `.go` files across the package:
- `signin_tests.go` → 3 tests
- `add_to_cart_tests.go` → 2 tests
- (any other `*_tests.go` files)

### What happens step by step

```
1. REGISTRATION (all files in the same Go package)
   Go loads ALL .go files in saucelab/ as one package.
   Each file's AlphaInit runs once during package init.

   signin_tests.go     → AlphaInit → registers 3 tests
   add_to_cart_tests.go → AlphaInit → registers 2 tests
   ...more files...     → AlphaInit → registers N tests

   Registry: [all tests from all files, in registration order]

2. SORTING — which test runs first?

   Tests are sorted by TWO criteria:
   ┌─────────────────────────────────────────────┐
   │  1st: Priority   (High=0 → Normal=1 → Low=2) │
   │  2nd: Order      (registration order)         │
   └─────────────────────────────────────────────┘

   Example with 5 tests:
     TestLogin         (Normal, order=1, from signin_tests.go)
     TestLogout        (Normal, order=2, from signin_tests.go)
     TestLockedUser    (High,   order=3, from signin_tests.go)
     TestAddToCart     (Normal, order=4, from add_to_cart_tests.go)
     TestRemoveFromCart(Normal, order=5, from add_to_cart_tests.go)

   After sorting:
     TestLockedUser    ← High priority, runs first
     TestLogin         ← Normal, order=1
     TestLogout        ← Normal, order=2
     TestAddToCart     ← Normal, order=4
     TestRemoveFromCart← Normal, order=5

3. EXECUTION (parallelSet = 2)

   testQueue (sorted): [Locked, Login, Logout, AddCart, RemoveCart]

   Worker 0                         Worker 1
   ┌────────────────────┐           ┌────────────────────┐
   │ TestLockedUser     │           │ TestLogin           │
   │ (browser, run, close)│         │ (browser, run, close)│
   │                    │           │                      │
   │ TestLogout         │           │ TestAddToCart         │
   │ (browser, run, close)│         │ (browser, run, close)│
   │                    │           │                      │
   │ TestRemoveFromCart │           │  (idle, done)         │
   │ (browser, run, close)│         │                      │
   └────────────────────┘           └────────────────────┘
           │                                │
           ▼                                ▼
        resultsChan (5 results)
           │
   Collector (mutex-protected)
           │
   HTML Report → test-results/report.html
                 test-results/report-20260227-200000.html (archived)
```

**Who runs first?** Priority decides. Within same priority, whichever test was registered first (file init order × registration order inside that AlphaInit).

**Who runs on which worker?** Whichever worker finishes its current test first grabs the next one from the queue. It's a race — first-come-first-served.

---

## Summary Table

| Scenario | Files | AlphaInits | Tests | Queue |
|----------|-------|------------|-------|-------|
| 1 file   | 1     | 1          | 2     | 2 tests, 2 workers each get 1 |
| 1 folder, 2 files | 2 | 2   | 4     | 4 tests, 2 workers each get ~2 |
| Full project | N   | N          | M     | M tests sorted by priority, 2 workers pull from shared queue |

**Golden rules**:
- 1 `AlphaInit` per file (enforced, panics otherwise).
- All tests land in one shared queue regardless of source file.
- Priority sorts the queue; workers pull next available test.
- Each test gets its own browser — zero shared state between workers.

---

## Troubleshooting Parallel Worker Errors

> Documented bugs discovered during parallel test execution with multiple
> worker goroutines. Each entry includes the symptom, root cause, fix, and
> a regression test that guards against reintroduction.

---

### Bug A — cleanStaleTempProfiles Race Deletes Active Worker Dirs

**Symptom**: `Failed to create SingletonLock: No such file or directory (2)`
errors when 2+ workers launch browsers. Chrome aborts because its temp profile
directory was deleted out from under it. Massive stderr log flooding follows
as Chrome retries and fails in a loop.

**Root cause**: `cleanStaleTempProfiles()` was called on **every** `Launch()`
call. It removes ALL `kexas-chrome-*` directories from the temp folder. When
workers run in parallel, Worker B's cleanup deletes Worker A's active temp
profile directory before Worker A's Chrome process can write its `SingletonLock`.

```
Worker 0: buildArgs() → dir = kexas-chrome-AAA
Worker 0: cmd.Start() → Chrome creates kexas-chrome-AAA/SingletonLock
Worker 1: buildArgs() → dir = kexas-chrome-BBB
Worker 1: cleanStaleTempProfiles() → deletes ALL kexas-chrome-* INCLUDING AAA
Worker 0: Chrome tries to use AAA → "No such file or directory"
```

**Key insight**: The dirs WERE unique (crypto-random suffixes worked). The error
was `"No such file or directory"`, NOT a collision. Another worker's cleanup
deleted the directory.

**Fix**: Wrap `cleanStaleTempProfiles` in `sync.Once` so it runs exactly once
per process — on the very first `Launch()` call, before any worker has created
its temp dir:
```go
var cleanOnce sync.Once
// In Launch():
cleanOnce.Do(func() { cleanStaleTempProfiles(log) })
```

**Regression tests**:
- `TestCleanStaleTempProfiles_DeletesActiveWorkerDir` — proves the race
- `TestCleanOnce_RunsExactlyOnce` — proves sync.Once guard works

**File**: `kexas/klauncher/launcher.go` → `Launch()` + `cleanOnce`

---

### Bug A2 — Timestamp-Only Temp Dir Names (Defense in Depth)

**Symptom**: Potential for identical temp dir paths if two goroutines call
`buildArgs` at the same nanosecond (theoretical, fixed proactively).

**Root cause**: `buildArgs()` originally used only `time.Now().UnixNano()`.

**Fix**: Append 8 crypto-random bytes to the directory suffix:
```
kexas-chrome-<unixnano>-<16 hex chars>
```

**Regression test**: `TestBuildArgs_UniqueProfileDirs_Concurrent` — spawns 50
concurrent goroutines and asserts zero duplicate dirs.

**File**: `kexas/klauncher/launcher.go` → `buildArgs()`

---

### Bug B — WaitAndType Rejects Empty String

**Symptom**: `TestSauceLoginPasswordRequired` fails with `text cannot be empty`
when `fillCredentials(page, t, sauceUser, "")` calls `WaitAndType("")`.

**Root cause**: `WaitAndType` (and `Type`) correctly validate that text is
non-empty — this is intentional API behavior. The bug was in the test helper
`fillCredentials()` which unconditionally called `WaitAndType` even when the
password was an empty string.

**Fix**: Skip the `WaitAndType` call when `user` or `pass` is empty:
```go
if pass != "" {
    passwordInput.WaitAndType(pass)
}
```
When testing "password required", the correct approach is to leave the field
untouched (no typing at all), not to type an empty string.

**Lesson**: Never pass empty strings to `WaitAndType` / `Type`. These methods
reject empty input by design (`ErrTextEmpty`). To test "missing field"
scenarios, simply skip typing into that field.

**File**: `saucelab/signin_tests.go` → `fillCredentials()`

---

### Bug C — Stale Log File Accumulation

**Symptom**: After many parallel test runs (especially crashed ones), the
system temp directory fills with `kexas-chrome-*.log` files. These stale files
can cause stderr-like noise when Chrome processes inherit them or when disk
space runs low.

**Root cause**: `Browser.Close()` cleaned up the temp profile directory
(`userDataDir`) but did not clean up the temp log file created by
`os.CreateTemp("", "kexas-chrome-*.log")`.

**Fix**: Added a `logFile` field to the `Browser` struct and added cleanup
logic in `Close()`:
```go
if b.logFile != "" {
    os.Remove(b.logFile)
}
```

**Regression test**: `TestBrowserClose_CleansUpLogFile` — creates a temp log
file, simulates cleanup, verifies removal, and confirms idempotent
double-remove.

**File**: `kexas/klauncher/launcher.go` → `Browser` struct + `Close()`

---

### Errors to Avoid (Quick Reference)

| Mistake | Consequence | Prevention |
|---------|------------|------------|
| Calling `cleanStaleTempProfiles` on every `Launch()` | Deletes active worker dirs → SingletonLock failure | Use `sync.Once` — clean only on first launch |
| Timestamp-only temp dir names | Potential dir collision | Always use `UnixNano + crypto/rand` |
| Calling `WaitAndType("")` | `ErrTextEmpty` error | Skip typing when value is empty |
| Not cleaning temp log files | Disk clutter, stale stderr | Clean `logFile` in `Close()` |
| Sharing browser state between workers | Flaky tests, data races | Each worker launches its own browser |
| Non-idempotent `BeforeEach` hooks | Race conditions in parallel | Keep hooks stateless or use per-test state |

---

### Debug Checklist for Parallel Failures

1. **Check temp directory** — `ls /tmp/kexas-chrome-*` — are there stale profiles?
2. **Check port conflicts** — `lsof -i :<port>` — is a zombie Chrome holding a port?
3. **Run with `-race`** — `go test -race ./...` — any data races in shared state?
4. **Check worker count** — is `parallelSet` ≤ available CPU cores?
5. **Check hooks** — do `BeforeEach`/`AfterEach` modify shared globals?
6. **Check test isolation** — does each test start from a fresh page state?
