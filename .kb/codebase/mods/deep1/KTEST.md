# KTest Module — Deep1 Level (For Software Engineering Students)

Source: `kexas/ktest/` (all files), `kexas/kexas_alpha_init.go`, `kexas/kassert/`

Every technical term is defined. FAQs per section.

---

## 1. What Is ktest?

ktest is kexas's built-in test framework. It's like Go's built-in `testing` package, but designed specifically for browser automation tests. It adds:
- **AlphaInit** — Register tests without boilerplate.
- **Parallel execution** — Run tests simultaneously with isolated browsers.
- **Lifecycle hooks** — Run setup/teardown code before/after tests.
- **HTML reports** — Beautiful reports showing pass/fail status, durations, and insights.
- **kassert** — Fluent assertion library for readable test checks.

### Key Terms

- **Test framework** — A library that helps you write and run tests. It provides structure (how to define tests), execution (how to run them), and reporting (how to see results). Examples: Go's `testing`, JavaScript's Jest, Python's pytest.
- **Test runner** — The part of the framework that executes tests and collects results. In ktest, this is `ktest_runner.go`.
- **Assertion** — A statement that checks if something is true. `kassert.That(t, x).Equals(5)` asserts that `x` equals 5. If it doesn't, the test fails.
- **Hook** — Code that runs at specific points in the test lifecycle: before all tests, before each test, after each test, after all tests.

---

## 2. AlphaInit — Zero-Boilerplate Registration

### The Problem

In standard Go testing, you write `func TestXxx(t *testing.T)` and run `go test`. But kexas tests need a browser, which means manual setup:

```go
// Without AlphaInit — verbose boilerplate
func TestLogin(t *testing.T) {
    browser, _ := kexas.Launch(nil)
    defer browser.Close()
    page, _ := browser.FirstPage()
    // ... now you can test
}
```

Every test repeats the same browser setup. AlphaInit eliminates this:

```go
// With AlphaInit — clean and declarative
var _ = kexas.AlphaInit(
    ktest.Test("Login", func(page *kexas.Page, t ktest.KTestT) {
        // page is already set up for you
        page.Navigate("https://example.com/login")
    }),
)
func main() { ktest.AutoRun() }
```

### How `var _ = kexas.AlphaInit(...)` Works

> **FAQ: What does `var _ = ...` mean?**
> In Go, `var _ = expression` evaluates the expression and throws away the result (the `_` blank identifier). The expression is evaluated during **package initialization** — before `func main()` runs. This is a trick to run code at startup without writing an `init()` function.

> **FAQ: What is package initialization?**
> When a Go program starts, before `main()` runs, Go initializes each package:
> 1. Package-level variables are evaluated top-to-bottom.
> 2. `init()` functions run (if any).
> This happens in dependency order — if package A imports package B, B is initialized first.
>
> `var _ = kexas.AlphaInit(...)` runs during step 1, registering your tests into a global registry.

### Reflection — How AlphaInit Identifies Components

AlphaInit accepts `...interface{}` (any number of arguments of any type). It uses **reflection** to figure out what each argument is:

```go
func registerComponent(component interface{}) {
    switch {
    case isGroupFunction(component):   // Is it a test group?
        registerGroup(component)
    case isTestFunction(component):    // Is it a test?
        registerTest(component)
    case isHookFunction(component):    // Is it a hook (BeforeAll, etc.)?
        registerHook(component)
    }
}
```

> **What is reflection?** The ability of a program to inspect its own types and values at runtime. In Go, `reflect.TypeOf(x)` tells you the type of `x`. `reflect.ValueOf(x)` gives you access to its value. Reflection is slower than normal code (because it bypasses compile-time type checking), but it's only used once during initialization, so the cost is negligible.

> **FAQ: Why use reflection instead of typed parameters?**
> AlphaInit needs to accept groups, tests, AND hooks in a single call. In Go, you can't have a function parameter that accepts three unrelated types. `interface{}` (Go's "any type") plus reflection is the standard solution.

### One AlphaInit Per File

```go
var alphaInitFiles sync.Map

func enforceOneAlphaInitPerFile(filename string) error {
    _, loaded := alphaInitFiles.LoadOrStore(filename, true)
    if loaded { return fmt.Errorf("only one AlphaInit per file") }
    return nil
}
```

> **What is `sync.Map`?** A thread-safe map from Go's standard library. Unlike regular Go maps (which are NOT safe for concurrent access), `sync.Map` can be read and written by multiple goroutines simultaneously without a mutex. `LoadOrStore` atomically checks if a key exists and inserts it if not — like a "get or create" operation.

> **FAQ: Why only one AlphaInit per file?**
> To keep test organization clean. Each file should define one logical group of tests. Multiple AlphaInit calls in the same file would be confusing and likely a copy-paste mistake.

---

## 3. Test Groups — Hierarchical Organization

```go
var _ = kexas.AlphaInit(
    ktest.Group("Auth", func() {
        ktest.Group("Login", func() {
            ktest.Test("should succeed with valid credentials", func(page *kexas.Page, t ktest.KTestT) { ... }),
            ktest.Test("should fail with invalid password", func(page *kexas.Page, t ktest.KTestT) { ... }),
        }),
        ktest.Group("Logout", func() {
            ktest.Test("should redirect to home", func(page *kexas.Page, t ktest.KTestT) { ... }),
        }),
    }),
)
```

This creates a tree structure:
```
Auth
├── Login
│   ├── should succeed with valid credentials
│   └── should fail with invalid password
└── Logout
    └── should redirect to home
```

Test names are built from the group path: `"Auth.Login.should succeed with valid credentials"`.

### The Group Stack

Groups use a **stack** to track nesting:

> **What is a stack?** A data structure where items are added and removed from the top only (last-in, first-out, LIFO). Like a stack of plates — you put plates on top and take plates from the top. When `Group("Auth", func() { ... })` is called:
> 1. "Auth" is pushed onto the stack.
> 2. The inner function runs (which may push more groups).
> 3. "Auth" is popped from the stack.
>
> Any `Test()` call inside the function sees "Auth" on the stack and registers the test under that group.

### Group-Level Hooks

Groups can have their own `BeforeAll` and `AfterAll`:

```go
ktest.Group("Auth", func() {
    ktest.BeforeAll(func() {
        // Seed the database with test users
    })
    ktest.Test("Login", ...)
    ktest.AfterAll(func() {
        // Clean up test users
    })
})
```

These run once per group, not per test.

---

## 4. KTestT — The Custom Test Interface

### Why Not Use `*testing.T`?

Go's `*testing.T` is created by the `go test` command. If you run your program with `go run ./main.go` (no `go test`), there's no `*testing.T` available. Kexas provides `KTestT` — an interface that `*testing.T` satisfies but that can also be implemented independently.

> **What is an interface in Go?** A set of method signatures. Any type that implements all the methods "satisfies" the interface. `KTestT` requires methods like `Log()`, `Error()`, `Fatal()`, `Failed()`, `Name()`. Both `*testing.T` (from Go's standard library) and ktest's internal `ktestT` satisfy this interface.

```go
type KTestT interface {
    Helper()
    Log(args ...interface{})
    Logf(format string, args ...interface{})
    Error(args ...interface{})
    Errorf(format string, args ...interface{})
    Fatal(args ...interface{})
    Fatalf(format string, args ...interface{})
    Failed() bool
    Name() string
    Run(name string, f func(KTestT)) bool
}
```

### `ktestT` — The Internal Implementation

```go
type ktestT struct {
    name   string
    failed bool
    depth  int
    mu     sync.Mutex  // protects 'failed'
}
```

> **FAQ: Why is there a mutex on just the `failed` field?**
> In parallel mode, when a child test fails, it calls `parentT.Errorf(...)` which sets `parentT.failed = true`. If two workers fail simultaneously, both might try to set `failed` at the same time. Without the mutex, this is a **data race** — two goroutines writing to the same memory location simultaneously. The mutex ensures only one goroutine writes at a time.

> **What is a data race?** When two goroutines access the same variable concurrently and at least one access is a write. Data races cause undefined behavior — the program might work fine, produce wrong results, or crash, depending on timing. Go's race detector (`go test -race`) finds these bugs.

### `Fatal` vs `Error`

- **`Error(msg)`** — Marks the test as failed and logs the message. The test **continues running** (subsequent assertions still execute).
- **`Fatal(msg)`** — Marks the test as failed, logs the message, and **immediately stops** the test (via `runtime.Goexit()` in standard `testing.T`, or by setting a flag in `ktestT`).

> **FAQ: When should I use Fatal vs Error?**
> Use `Fatal` when the failure makes further testing pointless. Example: if `Navigate` fails, there's no page to test — use `Fatal`. Use `Error` when you want to check multiple things even if one fails. Example: checking 5 form fields — use `Error` so you see all failures, not just the first.

---

## 5. The Runner — How Tests Execute

### `AutoRun()` — The Entry Point

```go
func main() {
    ktest.AutoRun()
}
```

`AutoRun()`:
1. Gets all tests registered via AlphaInit.
2. Falls back to reflection-based discovery (finding `func Test*` functions) if no AlphaInit tests exist.
3. Calls `runRegisteredTests()`.

### `runRegisteredTests()` — The Orchestrator

```
1. Create a KTestT instance
2. Load config from kexas.config.json
3. Filter tests (KEXAS_TEST_RUN environment variable)
4. Sort tests by priority (High → Normal → Low)
5. Run BeforeAll hook
6. DISPATCH: sequential or parallel
7. Run AfterAll hook
8. Generate HTML report
9. Print summary
10. Exit (code 0 if all pass, code 1 if any fail)
```

> **What is an environment variable?** A named value set in the shell (terminal) before running a program. `KEXAS_TEST_RUN=Login go run ./main.go` sets `KEXAS_TEST_RUN` to `"Login"`. The program reads it with `os.Getenv("KEXAS_TEST_RUN")`. Environment variables are used for configuration that changes between runs without modifying code.

> **What is an exit code?** A number (0-255) that a program returns to the operating system when it finishes. `0` means success. Any non-zero value means failure. CI systems (GitHub Actions, Jenkins) check exit codes to determine if a build passed or failed.

---

## 6. Parallel Execution — Worker Pool Pattern

### How It Works

```go
func runTestsParallelWorkers(tests []NamedTest, config *Config, parallelSet int) []testResult {
    testQueue := make(chan NamedTest, len(tests))     // buffered channel
    resultsChan := make(chan testResult, len(tests))   // buffered channel
    
    // Fill the queue
    for _, test := range tests { testQueue <- test }
    close(testQueue)
    
    // Launch workers
    var wg sync.WaitGroup
    for i := 0; i < parallelSet; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for test := range testQueue {        // grab next test from queue
                result := executeSingleTest(workerID, test, config)
                resultsChan <- result            // send result
            }
        }(i)
    }
    
    // Wait and collect
    go func() { wg.Wait(); close(resultsChan) }()
    var results []testResult
    for result := range resultsChan { results = append(results, result) }
    return results
}
```

### Key Concepts

> **What is a channel?** Go's mechanism for goroutines to communicate. A channel is like a pipe — one goroutine sends data in, another receives data out. `ch <- value` sends, `value := <-ch` receives. Channels are thread-safe by design.

> **What is a buffered channel?** A channel with a capacity. `make(chan int, 10)` can hold 10 values before blocking. Without a buffer, sends block until a receiver is ready. With a buffer, sends succeed immediately (until the buffer is full).

> **What is `sync.WaitGroup`?** A counter that waits for a group of goroutines to finish. `wg.Add(1)` increments the counter. `wg.Done()` decrements it. `wg.Wait()` blocks until the counter reaches zero. It's the standard way to wait for N concurrent tasks to complete.

> **What is the "worker pool" pattern?** A fixed number of goroutines (workers) that pull tasks from a shared queue. Instead of creating one goroutine per task (which might create thousands), you create a fixed number (e.g., 6) that process tasks one at a time. This controls resource usage.

> **What is "work stealing"?** When multiple workers read from the same channel, Go guarantees each message goes to exactly one worker. The first idle worker "steals" the next task from the queue. Faster workers naturally get more tasks — no manual load balancing needed.

### Process Isolation

Each worker launches its **own** Chromium process:

```go
func executeSingleTest(workerID int, test NamedTest, config *Config) testResult {
    browser, page, err := launchIsolatedBrowser(config)
    defer browser.Close()
    // ... run test ...
}
```

This means with `parallelSet: 6`, you have 6 separate Chrome processes running simultaneously. Each has its own:
- Memory space (~150 MB each)
- Temp profile directory
- Cookies, localStorage, cache
- TCP port

> **FAQ: Why not share one browser with multiple tabs?**
> Isolation. If tests share a browser, they share cookies, localStorage, and cache. Test A logging in might affect Test B's expectations. Separate browsers guarantee zero cross-contamination.

> **FAQ: Isn't 6 Chrome processes a lot of memory?**
> Yes — about 1 GB total. For CI servers with 4+ GB RAM, this is fine. For laptops, reduce `parallelSet` to 2-3. The config file (`kexas.config.json`) controls this.

---

## 7. Panic Recovery — Why Tests Don't Crash Each Other

```go
func runTestWithRecovery(childT *ktestT, test NamedTest, page *kexas.Page, ...) {
    defer func() {
        if r := recover(); r != nil {
            childT.Errorf("test panicked: %v", r)
        }
    }()
    test.Func(page, childT)  // run the actual test
}
```

> **What is a panic?** Go's equivalent of an unhandled exception. When code calls `panic("something went wrong")` or causes a nil pointer dereference, the goroutine starts unwinding — it stops executing, runs all `defer` functions in reverse order, and then crashes the program (if not recovered).

> **What is `recover()`?** A built-in Go function that catches a panic. It can ONLY be called inside a `defer` function. If a panic is in progress, `recover()` returns the panic value and stops the unwinding. If there's no panic, `recover()` returns `nil`.

> **FAQ: Why is panic recovery important for parallel tests?**
> Without `recover()`, a panic in one test kills the entire goroutine. In parallel mode, that worker goroutine dies, and its remaining queued tests never run. Worse, if the panic propagates to the main goroutine, ALL workers crash. With `recover()`, the panicking test is marked as failed, and the worker continues to the next test.

> **FAQ: What causes panics in browser tests?**
> Common causes:
> - Nil pointer dereference: calling a method on a nil `*Element` (e.g., `page.Find()` returned nil and you forgot to check).
> - Index out of bounds: accessing `elements[5]` when only 3 elements exist.
> - Type assertion failure: `result["value"].(string)` when the value is actually an `int`.

---

## 8. Lifecycle Hooks

```go
var _ = kexas.AlphaInit(
    ktest.BeforeAll(func() {
        // Runs once before ALL tests (e.g., seed database)
    }),
    ktest.BeforeEach(func(page *kexas.Page) {
        // Runs before EACH test (e.g., navigate to login page)
        page.Navigate("https://example.com/login")
    }),
    ktest.AfterEach(func(page *kexas.Page) {
        // Runs after EACH test (e.g., clear cookies)
    }),
    ktest.AfterAll(func() {
        // Runs once after ALL tests (e.g., cleanup database)
    }),
)
```

### Execution Order

```
BeforeAll()            ← main goroutine, once
  │
  ├── Worker 1:
  │   ├── BeforeEach(page1)
  │   ├── Test A (page1)
  │   ├── AfterEach(page1)
  │   ├── BeforeEach(page1)
  │   ├── Test D (page1)
  │   └── AfterEach(page1)
  │
  ├── Worker 2:
  │   ├── BeforeEach(page2)
  │   ├── Test B (page2)
  │   ├── AfterEach(page2)
  │   ...
  │
AfterAll()             ← main goroutine, once
```

> **FAQ: Is BeforeEach thread-safe?**
> It must be! In parallel mode, multiple workers call BeforeEach simultaneously, each with their own `page`. The hook function itself must not modify shared state (global variables). Since each worker gets its own `page`, operations on `page` are safe.

---

## 9. kassert — Fluent Assertions

### What Makes Assertions "Fluent"?

Fluent API means you chain method calls to read like English:

```go
kassert.That(t, title).Named("page title").Equals("Google")
// Reads as: "Assert that [page title] equals 'Google'"

kassert.That(t, count).IsGreaterThan(0).IsLessThan(100)
// Reads as: "Assert that count is greater than 0 AND less than 100"
```

### How Chaining Works

Each method returns `*Assertion`, allowing the next method call:

```go
func (a *Assertion) Equals(expected interface{}) *Assertion {
    if !reflect.DeepEqual(a.actual, expected) {
        a.t.Errorf("Expected %s to equal %v, got %v", a.name, expected, a.actual)
    }
    return a  // ← this enables chaining
}
```

> **What is `reflect.DeepEqual`?** A function that compares two values recursively. For simple types (int, string), it checks equality. For slices, it checks element-by-element. For structs, it checks field-by-field. For maps, it checks key-by-key. For pointers, it follows the pointer and compares the pointed-to values. This is more thorough than `==`, which doesn't work for slices and maps in Go.

### Common Assertions

```go
kassert.That(t, value).Equals(expected)        // deep equality
kassert.That(t, value).IsNil()                  // value == nil
kassert.That(t, value).IsNotNil()               // value != nil
kassert.That(t, value).IsTrue()                 // value == true
kassert.That(t, value).IsFalse()                // value == false
kassert.That(t, str).Contains("hello")          // string contains substring
kassert.That(t, slice).HasLength(5)             // len(slice) == 5
kassert.That(t, num).IsGreaterThan(0)           // numeric comparison
kassert.That(t, num).IsLessThan(100)            // numeric comparison
```

> **FAQ: Why not just use `if x != 5 { t.Error(...) }`?**
> You can! But kassert provides:
> 1. Better error messages — `"Expected page title to equal 'Google', got 'Yahoo'"` vs your manual `"not equal"`.
> 2. Method chaining — Check multiple conditions in one line.
> 3. Consistency — All tests use the same assertion style.
> 4. Statistics — kassert tracks pass/fail counts globally for reporting.

---

## 10. HTML Reports

After all tests complete, ktest generates an HTML report:

```
kexas-report/
├── report.html                    ← latest report (overwritten each run)
└── report-20260227-143022.html   ← versioned copy (never overwritten)
```

The report includes:
- Pass/fail counts with a visual donut chart.
- Each test's name, status (✓/✗), duration, and worker ID.
- Screenshot links for failed tests (if `screenshotOnFail` is enabled).

> **FAQ: Why version reports?**
> So you can compare results across runs. If Monday's run had 2 failures and Tuesday's has 5, you can see what changed. The versioned copy uses a timestamp in the filename, so each run creates a unique file.

---

## 11. Summary — The Complete Test Lifecycle

```
1. Package init: AlphaInit registers groups, tests, hooks
2. main(): AutoRun() is called
3. Config: Load kexas.config.json (parallelSet, headless, etc.)
4. Filter: KEXAS_TEST_RUN env var filters tests by name
5. Sort: Tests ordered by priority (High → Normal → Low)
6. BeforeAll: Global setup hook runs
7. Workers: N goroutines launched, each with own Chrome process
8. Per test:
   a. BeforeEach(page)
   b. Test function runs (with panic recovery)
   c. AfterEach(page)
   d. Screenshot if failed
   e. Browser closed
9. AfterAll: Global teardown hook runs
10. Report: HTML report generated
11. Summary: Pass/fail counts printed
12. Exit: Code 0 (success) or 1 (failure)
```
