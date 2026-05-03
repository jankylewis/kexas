# Parallel Test Execution — Lessons Learned & Architecture

> Knowledge base entry for the ktest parallel execution system.
> Completed: 2026-02-27

---

## Architecture Decision: Launch-Per-Test vs Browser Pool

### What we chose: Launch-Per-Test
Each worker goroutine launches its own browser, runs one test, closes it, then loops.

### Why NOT a Browser Pool
- Pool adds complexity: acquire/release lifecycle, pool exhaustion handling, browser health checks.
- Isolation is weaker — a crashed browser poisons the pool slot.
- Launch-per-test gives **100% isolation** with zero shared mutable state.
- Chrome launch overhead (~1-2s) is acceptable for E2E tests that take 5-30s each.

### When to reconsider
If test count grows to 50+ and launch overhead becomes >20% of total time, implement browser pooling as an optimization layer on top of the existing worker model.

---

## Critical Pitfall: Port Conflicts in Parallel

### The Bug
`DefaultOptions()` hardcoded `Port: 9222`. When 2 workers launched simultaneously, both tried to bind Chrome to port 9222. Worker 0 succeeded; worker 1 got "port already in use."

### The Fix
- Changed `DefaultOptions().Port` from `9222` to `0` (dynamic).
- Added `findFreePort()` — binds to `:0`, reads OS-assigned port, releases.
- `Launch()` resolves `Port=0` → real free port before spawning Chrome.
- Skip expensive `killExistingChromeProcesses()` for dynamic ports (OS guarantees free).

### Lesson
**Never hardcode ports when parallelism is possible.** Always use dynamic port assignment unless the user explicitly specifies a fixed port.

---

## Thread Safety Strategy

### What needs guarding
1. **`ktestT.failed` field** — written by test goroutine, read by collector. Guarded by `sync.Mutex`.
2. **`ktestT.Log/Error/Fatal` output** — interleaved output from N workers. Guarded by same mutex.
3. **Test results** — each worker writes to a buffered `chan testResult`. No mutex needed (channels are goroutine-safe).
4. **`kassert` stats** — already use `sync/atomic` internally. Safe.

### What does NOT need guarding
- `Config` — read-only after `loadConfig()`.
- `[]NamedTest` slice — read-only after registration phase.
- Global hooks (`BeforeAll`, `AfterAll`) — run on main goroutine only, before/after all workers.
- `BeforeEach`/`AfterEach` — run inside each worker's goroutine, not shared.

### Pattern: Child T per Test
Each parallel test gets its own `*ktestT` child. Failure propagates to parent T only after test completes. This avoids lock contention on a shared T.

---

## Worker Pool Design

```
Main goroutine:
  1. Sort tests by priority
  2. Feed tests into buffered channel (len=testCount)
  3. Close channel
  4. Spawn N workers
  5. Wait (via sync.WaitGroup) for all workers to finish
  6. Close results channel
  7. Collect results

Worker goroutine:
  for test := range testQueue {
      browser, page := launch()
      BeforeEach(page)
      run test with panic recovery
      AfterEach(page)
      browser.Close()
      resultsChan <- result
  }
```

### Key design choices
- **Buffered channels** for both testQueue and resultsChan — no goroutine blocks.
- **Cap workers to test count** — `min(parallelSet, len(tests))` prevents idle workers.
- **Panic recovery per test** — one test panic doesn't kill the worker.
- **Deferred browser.Close()** — browser always cleaned up, even on panic.

### Visual: 2 Workers, 2 Tests

```
                   +--------------------+
                   |   testQueue (chan) |
                   |  [T1] -> [T2]      |
                   +---------+----------+
                             |
                +------------+-------------+
                |                          |
        Worker 0 goroutine           Worker 1 goroutine
        (isolated browser A)         (isolated browser B)
                |                          |
   1. recv T1 from channel       1. recv T2 from channel
   2. launch browser A           2. launch browser B
   3. BeforeEach hooks           3. BeforeEach hooks
   4. run T1                     4. run T2
   5. AfterEach hooks            5. AfterEach hooks
   6. send result to resultsChan 6. send result to resultsChan

                             |
                   +---------v----------+
                   | resultsChan (chan) |
                   +---------+----------+
                             |
                   +---------v----------+
                   |   Collector        |
                   | (mutex-protected)  |
                   +--------------------+
```

Notes:
- testQueue is closed once all tests are enqueued, so both workers exit naturally.
- Each worker owns its browser lifecycle; no browser state is shared.
- Collector records arrival order, not start order; suitable for HTML report generation.

---

## Advanced Example: AlphaInit + Groups + Hooks (2 Tests)

### Registration Phase (AlphaInit)

```
func AlphaInit() {
    ktest.Group("checkout", func(g *ktest.GroupContext) {
        g.BeforeAll(func() { log.Info("checkout setup") })
        g.AfterAll(func() { log.Info("checkout cleanup") })
        g.BeforeEach(func(page *kexas.Page) { seedCart(page) })
        g.AfterEach(func(page *kexas.Page) { clearCart(page) })

        g.Test("TestGuestCheckout", TestGuestCheckout)
        g.Test("TestMemberCheckout", TestMemberCheckout)
    })
}
```

`AlphaInit()` runs once at startup, registering:

1. A group named **checkout** with its own hooks.
2. Two `ktest.Test` entries that ultimately land in the global `registeredTests` slice.

### Execution Timeline (ParallelSet = 2)

```
Global BeforeAll (suite-wide)
  └─ checkout.BeforeAll
       ├─ Worker 0 handles TestGuestCheckout
       │    1. checkout.BeforeEach
       │    2. TestGuestCheckout (isolated browser A)
       │    3. checkout.AfterEach
       └─ Worker 1 handles TestMemberCheckout
            1. checkout.BeforeEach
            2. TestMemberCheckout (isolated browser B)
            3. checkout.AfterEach
  └─ checkout.AfterAll
Global AfterAll
```

Key points:

- **Hook layering**: global hooks wrap the entire run; group hooks wrap tests inside that group; per-test hooks (BeforeEach/AfterEach) execute inside the worker goroutine before and after each `ktest.Test`.
- **Parallel execution**: even though both tests share the checkout hooks, each worker receives its own browser/page, so mutations inside `BeforeEach` affect only that worker.
- **Results collection**: once each worker finishes AfterEach, it pushes its outcome into `resultsChan`; the mutex-guarded collector then feeds stats + HTML report generation.

Diagram with queue + hooks:

```
AlphaInit
  └─ register Group("checkout") with hooks + tests

testQueue: [checkout.TestGuestCheckout, checkout.TestMemberCheckout]

Global BeforeAll
  ↓
checkout.BeforeAll
  ↓                            ↓ (executed concurrently)
Worker0 ── checkout.BeforeEach ── run TestGuest ── checkout.AfterEach ──▶ resultsChan
Worker1 ── checkout.BeforeEach ── run TestMember ── checkout.AfterEach ─▶ resultsChan
  ↓
checkout.AfterAll
  ↓
Global AfterAll
```

This example shows how AlphaInit, groups, and hooks compose with the worker pool: hooks run on the main goroutine except BeforeEach/AfterEach, which execute inside each worker right before/after the test function.

---

## Priority Sorting

Tests are sorted before dispatch: `PriorityHigh(0) > PriorityNormal(1) > PriorityLow(2)`.
Within same priority, stable sort preserves registration order.

This ensures critical tests start first and get results earliest.

---

## Common Errors & Misunderstandings

| Error | Explanation |
|-------|-------------|
| Port conflict with `parallelSet > 1` | Fixed by dynamic port assignment (`Port: 0`) |
| Race on `ktestT.failed` | Fixed by `sync.Mutex` on all read/write to `failed` |
| Interleaved console output | Fixed by mutex-guarded `Log/Logf/Error/Errorf` |
| Global hooks running in parallel | Wrong — `BeforeAll`/`AfterAll` run on main goroutine only |
| Browser pool needed for parallelism | Not needed — launch-per-test is simpler and more isolated |
| `os.Exit(1)` in `Fatal()` kills all workers | Known — `Fatal` should be avoided in parallel tests. Use `Error` instead. |

---

## Testing Strategy

- 21 unit tests in `tests/ktest_parallel_test.go`
- All run with `-race` flag — zero data races
- Tests cover: priority sorting, config parsing, thread-safe T, worker capping, channel isolation
- 4 additional launcher tests for `findFreePort` uniqueness and port bindability

---

## Config

```json
{
  "parallelSet": 2,
  "headless": false,
  "timeout": 60000,
  "screenshotOnFail": true
}
```

- `parallelSet: 1` = sequential (default, backward-compatible)
- `parallelSet: N` = N worker goroutines
- Automatically sets `parallel: true` when `parallelSet > 1`

---

**Files**: `ktest/ktest_parallel.go`, `ktest/ktest_types.go`, `ktest/ktest_groups.go`, `ktest/ktest_runner.go`, `klauncher/launcher.go`
