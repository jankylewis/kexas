# ktest Runner (High-Level)

Source: `kexas/ktest/` (all files), `kexas/kexas_alpha_init.go`, `kexas/kassert/`

## Mission

Provide a batteries-included test framework for kexas. `ktest` handles test registration, parallel execution with isolated browsers, lifecycle hooks, panic recovery, HTML reporting, and fluent assertions — so test authors focus on scenarios, not infrastructure.

## Key Components

| Component | Source | Purpose |
|-----------|--------|---------|
| AlphaInit | `kexas_alpha_init.go` | Zero-boilerplate test registration via package-level `var` |
| Groups | `ktest_groups.go` | Hierarchical test organization (nested groups with hooks) |
| Registration | `ktest_registration.go` | `Test()`, `TestWithPriority()` — stores tests with metadata |
| Runner | `ktest_runner.go` | Orchestrates config loading, filtering, dispatch, reporting |
| Parallel | `ktest_parallel.go` | Worker pool with channel-based work stealing |
| Hooks | `ktest_hooks.go` | `BeforeAll`, `BeforeEach`, `AfterEach`, `AfterAll` |
| KTestT | `ktest_types.go` | Thread-safe `testing.T` replacement with mutex-protected state |
| kassert | `kassert/kassert.go` | Fluent assertions: `kassert.That(t, val).Equals(expected)` |

## Registration Flow

```go
var _ = kexas.AlphaInit(
    ktest.Group("Auth", func() {
        ktest.BeforeEach(func(page *kexas.Page) { page.Navigate("/login") }),
        ktest.Test("Login", func(page *kexas.Page, t ktest.KTestT) {
            // test code
        }),
    }),
)

func main() { ktest.AutoRun() }
```

AlphaInit runs at package init time (before `main`). It uses reflection to identify groups, tests, and hooks, then stores them in global registries. One `AlphaInit` per file is enforced via `sync.Map`.

## Execution Lifecycle

```
main() → ktest.AutoRun()
  │
  ├── 1. Load config from kexas.config.json
  ├── 2. Filter tests (KEXAS_TEST_RUN env var)
  ├── 3. Sort by priority (High → Normal → Low)
  ├── 4. ExecuteGlobalBeforeAll()        ← main goroutine
  │
  ├── 5. DISPATCH (parallelSet > 1 → parallel, else sequential)
  │     ├── Create buffered channel (test queue)
  │     ├── Launch N worker goroutines
  │     │     └── Per test:
  │     │           ├── launchIsolatedBrowser() → own Chromium process
  │     │           ├── BeforeEach(page)
  │     │           ├── runTestWithRecovery(test) ← panic-safe
  │     │           ├── AfterEach(page)
  │     │           ├── Screenshot on failure
  │     │           └── browser.Close()
  │     └── Collect results via result channel
  │
  ├── 6. ExecuteGlobalAfterAll()         ← main goroutine
  ├── 7. Generate HTML report (+ versioned copy)
  └── 8. Print summary + exit code
```

## Parallel Execution Model

- Each worker goroutine gets its **own Chromium OS process** — complete process-level isolation.
- Tests are distributed via a buffered channel (work-stealing pattern).
- The ONLY shared state is: `parentT.failed` (mutex-protected), the test queue channel, and the result channel.
- Panics in one test are caught by `recover()` — they don't crash other workers.

## kassert — Fluent Assertions

```go
kassert.That(t, title).Named("page title").Equals("Google")
kassert.That(t, count).IsGreaterThan(0)
kassert.That(t, items).HasLength(5)
kassert.That(t, err).IsNil()
```

- Uses `reflect.DeepEqual` for equality checks.
- Returns `*Assertion` for method chaining.
- Works with both `*testing.T` and `ktest.KTestT`.

## Why It Matters

Without `ktest`, every project would reinvent parallel execution, browser isolation, panic recovery, screenshot capture, and HTML reporting. Centralizing it ensures consistent behavior, high parallelism, and a single place to add features like retry logic, flaky test detection, or CI integration.
