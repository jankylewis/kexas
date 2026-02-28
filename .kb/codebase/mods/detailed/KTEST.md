# KTest Module — Deep Detailed Walkthrough

Source: `kexas/ktest/` (all files), `kexas/kexas_alpha_init.go`, `kexas/kassert/kassert.go`

This document explains the entire test framework at the OS, process, memory, goroutine, and protocol level. It covers AlphaInit, test registration, the runner lifecycle, parallel execution with worker pools, panic recovery, the custom `ktestT` implementation, hooks, groups, and the kassert assertion library.

---

## 1. Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                     User's main.go                            │
│                                                               │
│  var _ = kexas.AlphaInit(                                     │
│      ktest.Group("Auth", func() {                             │
│          ktest.Test("Login", func(page, t) { ... })           │
│      }),                                                      │
│  )                                                            │
│  func main() { ktest.AutoRun() }                              │
└──────────────────┬───────────────────────────────────────────┘
                   │
    ┌──────────────▼──────────────┐
    │        AlphaInit            │  (package-level init, runs at import time)
    │  ├── enforceOnePerFile      │
    │  ├── registerComponent      │
    │  │   ├── isGroupFunction    │
    │  │   ├── isTestFunction     │
    │  │   └── isHookFunction     │
    │  └── stores in global vars  │
    └──────────────┬──────────────┘
                   │
    ┌──────────────▼──────────────┐
    │         AutoRun()           │  (called from main())
    │  ├── GetRegisteredTests()   │
    │  └── runRegisteredTests()   │
    │      ├── loadConfig()       │  (reads kexas.config.json)
    │      ├── filterTests()      │  (KEXAS_TEST_RUN env var)
    │      ├── BeforeAll hook     │
    │      ├── DISPATCH:          │
    │      │   ├── sequential     │  (parallelSet=1)
    │      │   └── parallel       │  (parallelSet>1)
    │      ├── AfterAll hook      │
    │      ├── generateHTMLReport │
    │      └── printTestSummary   │
    └─────────────────────────────┘
```

---

## 2. AlphaInit — Zero-Boilerplate Test Registration

Source: `kexas/kexas_alpha_init.go`

### What Is AlphaInit?

AlphaInit is a package-level function that registers test components (groups, tests, hooks) without requiring `func init()` or `func main()` boilerplate. It's called via a package-level variable assignment:

```go
var _ = kexas.AlphaInit(
    ktest.Group("Amazon Tests", func() { ... }),
)
```

### How Go Package-Level Initialization Works (OS/Runtime Level)

When Go compiles a program, package-level `var` declarations are initialized in the order they appear, **before** `func main()` runs. The Go runtime's initialization sequence:

1. **Process start** — The OS `execve` loads the binary. The Go runtime's entry point (`_rt0_amd64_darwin` on macOS) runs first.
2. **Runtime init** — Go initializes the scheduler, GC, and goroutine system.
3. **Package init** — For each imported package (in dependency order), Go:
   a. Initializes package-level variables (top to bottom, left to right).
   b. Runs `func init()` functions (if any).
4. **`main.main()`** — Finally, the program's `main` function runs.

`var _ = kexas.AlphaInit(...)` runs during step 3a. The `_` discard means the return value is thrown away — we only care about the side effects (registration).

### `AlphaInit` Line by Line

```go
func AlphaInit(components ...interface{}) interface{} {
    var callerFile string = alphaInitCallerFile()
    if err := enforceOneAlphaInitPerFile(callerFile); err != nil {
        panic(err)
    }
    // Process each component
    for i := 0; i < len(components); i++ {
        registerComponent(components[i])
    }
    return nil
}
```

#### `alphaInitCallerFile()` — Stack Walking

```go
func alphaInitCallerFile() string {
    var _, file, _, ok = runtime.Caller(2)
    // Caller(0) = alphaInitCallerFile
    // Caller(1) = AlphaInit
    // Caller(2) = user's source file
}
```

`runtime.Caller(skip)` reads the **goroutine call stack**. In Go, each goroutine has a stack (starting at 2 KB, growing up to 1 GB). The stack contains **stack frames** — one per active function call. Each frame stores:
- Return address (program counter)
- Local variables
- Parameters

`runtime.Caller(2)` walks up 2 frames and reads the program counter. The runtime then looks up the PC in the **symbol table** (embedded in the binary by the linker) to find the source file name and line number.

This is a pure in-process operation — no syscalls. Cost: ~100 nanoseconds.

#### `enforceOneAlphaInitPerFile()` — Atomic Map

```go
var alphaInitFiles sync.Map

func enforceOneAlphaInitPerFile(filename string) error {
    var _, loaded = alphaInitFiles.LoadOrStore(filename, true)
    if loaded {
        return fmt.Errorf("kexas: only one AlphaInit allowed per file, but '%s' called it twice", filename)
    }
    return nil
}
```

`sync.Map` is a concurrent-safe map optimized for two cases: (1) keys are written once and read many times, (2) different goroutines access different keys. `LoadOrStore` atomically checks if the key exists and inserts it if not. If the key was already present, `loaded` is `true` → panic.

This prevents accidental double-registration. Since AlphaInit runs during package init (single-threaded), contention is impossible, but `sync.Map` is used for correctness guarantees.

#### `registerComponent()` — Reflection-Based Type Dispatch

```go
func registerComponent(component interface{}) error {
    switch {
    case isGroupFunction(component):
        return registerGroup(component)
    case isTestFunction(component):
        return registerTest(component)
    case isHookFunction(component):
        return registerHook(component)
    default:
        return fmt.Errorf("unsupported component type: %T", component)
    }
}
```

Each `is*Function` uses `reflect.TypeOf(component)` to inspect the component's type at runtime. For example:

```go
func isTestFunction(component interface{}) bool {
    var componentType reflect.Type = reflect.TypeOf(component)
    // Check: func(string, func(*Page, KTestT))
    if componentType.NumIn() == 2 &&
       componentType.In(0).Kind() == reflect.String &&
       componentType.In(1).Kind() == reflect.Func { ... }
}
```

**Reflection cost**: `reflect.TypeOf` returns a cached `*rtype` pointer (no allocation). `NumIn()`, `In(0)`, `Kind()` are simple field reads on the type descriptor. Total cost: ~50 nanoseconds per check. This runs once during init, not in hot paths.

---

## 3. Test Registration — Groups and Named Tests

Source: `kexas/ktest/ktest_groups.go`, `ktest_registration.go`

### `NamedTest` — The Test Unit

```go
type NamedTest struct {
    Name     string                        // full dotted name: "Auth.Login.should_work"
    Func     func(*kexas.Page, KTestT)     // the actual test function
    Filename string                        // source file (e.g., "auth_test")
    Priority TestPriority                  // 0=High, 1=Normal, 2=Low
    Order    int                           // registration order for stable sort
}
```

Each registered test becomes a `NamedTest` stored in either:
- `registeredTests` (package-level `[]NamedTest`) — for root-level tests.
- `GroupContext.Tests` (per-group `[]NamedTest`) — for grouped tests.

### `GroupContext` — Test Grouping

```go
type GroupContext struct {
    Name      string
    Parent    *GroupContext       // nil for root groups
    BeforeAll func()
    AfterAll  func()
    Tests     []NamedTest
    Children  []*GroupContext
    mu        sync.RWMutex
}
```

Groups form a tree. The current group is tracked via a **stack** (`groupStack []*GroupContext`). When `ktest.Group("Auth", func() { ... })` is called:

1. `createGroupContext("Auth")` allocates a new `GroupContext`.
2. `setGroupParent(group)` links it to the current group (or adds it to `rootGroups` if no parent).
3. `executeGroupFunction(group, groupFunc)` pushes the group onto the stack, calls `groupFunc()`, then pops it.
4. During `groupFunc()`, any `ktest.Test(...)` calls see the current group on the stack and register tests under it.

### Stack-Based Nesting

```go
func executeGroupFunction(group *GroupContext, groupFunc func()) {
    groupStack = append(groupStack, group)
    currentGroup = group
    groupFunc()  // <-- Test() calls inside here see currentGroup
    popGroupFromStack()
}
```

This is a classic **context stack** pattern. It supports unlimited nesting:

```go
ktest.Group("Auth", func() {
    ktest.Group("Login", func() {
        ktest.Test("should_work", ...)  // name: "Auth.Login.should_work"
    })
})
```

The group path is built by walking up the `Parent` chain: `GetGroupPath()` returns `"Auth.Login"`.

### Caller File Detection

```go
for depth := 0; depth <= 10; depth++ {
    var _, file, _, ok = runtime.Caller(depth)
    if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
        callerFilename = file
        break
    }
}
```

This walks the call stack up to 10 frames, looking for the first frame that's NOT in the ktest/kexas internal packages. This finds the user's source file, which is used for display names like `<auth_test.should_login>`. The filtering prevents internal framework files from being reported as the test's source.

---

## 4. `KTestT` — The Custom Test Interface

Source: `kexas/ktest/ktest_types.go`

### Why Not Use `*testing.T`?

Go's `*testing.T` is tightly coupled to `go test`. If you want to run tests via `go run ./myapp` (no `go test`), you can't create a `*testing.T`. Kexas provides `KTestT`, an interface that `*testing.T` satisfies but that can also be implemented independently:

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

### `ktestT` — The Custom Implementation

```go
type ktestT struct {
    name   string
    failed bool
    depth  int
    mu     sync.Mutex  // protects 'failed'
}
```

#### Thread Safety

The `mu sync.Mutex` is critical. In parallel mode, multiple worker goroutines call `Errorf` on the **parent** `ktestT` to propagate failures upward:

```go
func (t *ktestT) Errorf(format string, args ...interface{}) {
    t.mu.Lock()
    fmt.Printf(format+"\n", args...)
    t.failed = true
    t.mu.Unlock()
}
```

Without the mutex, two goroutines could simultaneously read-modify-write `t.failed`, causing a **data race**. Go's race detector (`go test -race`) would flag this.

#### `Run` — Subtest Simulation

```go
func (t *ktestT) Run(name string, f func(KTestT)) bool {
    var childT *ktestT = &ktestT{name: name, failed: false, depth: t.depth + 1}
    f(childT)
    if childT.Failed() {
        t.mu.Lock()
        t.failed = true
        t.mu.Unlock()
    }
    return !childT.Failed()
}
```

This mirrors `testing.T.Run()` — creates a child T, runs the function, and propagates failure upward. But unlike `testing.T.Run()`, this is synchronous (no goroutine), making it simpler to reason about.

### `testingTWrapper` — Adapting `*testing.T`

```go
type testingTWrapper struct { t *testing.T }
func (w *testingTWrapper) Errorf(format string, args ...interface{}) { w.t.Errorf(format, args...) }
// ... all methods delegate to w.t
```

This wrapper lets the legacy `Run(t *testing.T, suite)` API work with the same internal code as `Main()`. It's a thin adapter — no behavior change, just interface satisfaction.

---

## 5. The Runner — Test Lifecycle

Source: `kexas/ktest/ktest_runner.go`

### `runRegisteredTests()` — The Orchestrator

```go
func runRegisteredTests(tests []NamedTest) {
    var t KTestT = newKTestT()
    kassert.ResetStats()
    var filteredTests []NamedTest = filterTests(tests)
    var config *Config = loadConfig()

    ExecuteGlobalBeforeAll()                // main goroutine

    var results []testResult
    if config.ParallelSet > 1 {
        results = runTestsParallelWorkers(t, filteredTests, config, config.ParallelSet)
    } else {
        results = runTestsSequentialRegistered(t, filteredTests, config)
    }

    ExecuteGlobalAfterAll()                 // main goroutine

    kassert.PrintStats()
    generateHTMLReport(results, config)
    printTestSummary(results, filteredTests)
}
```

#### Execution Order (Critical!)

1. **`BeforeAll`** — Runs on the **main goroutine**, before any browser is launched. Safe to do one-time setup (DB seed, env config).
2. **Test execution** — Either sequential or parallel (see below).
3. **`AfterAll`** — Runs on the **main goroutine**, after all tests complete. Safe to do cleanup.
4. **Report generation** — Collects results, generates HTML.
5. **Summary and exit** — Prints pass/fail counts, exits with code 1 if any failures.

### `loadConfig()` — Configuration from JSON

```go
func loadConfig() *Config {
    return LoadConfigFromFile("./kexas.config.json")
}
```

Reads `kexas.config.json` from the current working directory. Uses Go's `os.ReadFile` (which calls `open(2)` + `read(2)` + `close(2)`) and `json.Unmarshal`. If the file doesn't exist, returns `DefaultConfig()` silently.

Key config fields:
- `parallelSet: 6` → 6 worker goroutines, each with its own browser.
- `headless: true` → Chrome runs without a visible window.
- `screenshotOnFail: true` → Captures PNG on test failure.

### `filterTests()` — KEXAS_TEST_RUN Filter

```go
var testFilter string = os.Getenv("KEXAS_TEST_RUN")
```

`os.Getenv` reads from the process's environment block (set by the shell before `execve`). If set, only tests whose name contains the filter string are kept. This enables running a single test during development.

---

## 6. Parallel Execution — The Worker Pool

Source: `kexas/ktest/ktest_parallel.go`

### `runTestsParallelWorkers()` — The Dispatcher

```go
func runTestsParallelWorkers(parentT KTestT, tests []NamedTest, config *Config, parallelSet int) []testResult {
    SortTestsByPriority(tests)

    var workerCount int = parallelSet
    if workerCount > len(tests) { workerCount = len(tests) }

    var testQueue chan NamedTest = make(chan NamedTest, len(tests))
    var resultsChan chan testResult = make(chan testResult, len(tests))

    // Feed all tests into queue
    for _, test := range tests { testQueue <- test }
    close(testQueue)

    // Launch workers
    var wg sync.WaitGroup
    for workerID := 0; workerID < workerCount; workerID++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            runWorker(id, parentT, testQueue, resultsChan, config)
        }()
    }

    // Wait and collect
    go func() { wg.Wait(); close(resultsChan) }()
    var results []testResult
    for result := range resultsChan { results = append(results, result) }
    return results
}
```

### What Happens at the Goroutine/OS Level

1. **Channel creation** — `make(chan NamedTest, len(tests))` allocates a buffered channel. Internally, this is a ring buffer in heap memory, protected by a mutex. The buffer size equals the number of tests, so all tests can be enqueued without blocking.

2. **Enqueue** — All tests are pushed into `testQueue` before any worker starts. The channel acts as a thread-safe work queue. No locks needed by the caller — the channel's internal mutex handles it.

3. **Worker goroutines** — Each `go func() { ... }()` creates a new goroutine. In Go, goroutines start with a **2 KB stack** (vs. OS threads which get 1–8 MB). The Go scheduler multiplexes goroutines onto OS threads (M:N scheduling). With 6 workers, you might see 6 OS threads if all workers are busy with syscalls (WebSocket I/O), or fewer if some are blocked.

4. **Work stealing** — `for test := range testQueue` reads from the channel. Multiple goroutines can read from the same channel — Go guarantees each message is delivered to exactly one reader. This is the **work stealing** pattern: the first idle worker picks up the next test.

5. **Result collection** — Workers send `testResult` structs to `resultsChan`. The main goroutine reads from it via `for result := range resultsChan`. The `close(resultsChan)` (called after `wg.Wait()`) terminates the range loop.

### Memory Model: No Shared Mutable State

Each worker goroutine is **completely isolated**:
- Own `Browser` (own Chromium OS process)
- Own `Page` (own CDP session)
- Own `AgentManager`
- Own `childT` (`ktestT` for this specific test)

The ONLY shared mutable state is:
- `parentT.failed` — protected by `sync.Mutex`.
- `testQueue` channel — protected by channel internal mutex.
- `resultsChan` channel — same.

This design eliminates data races by construction.

### `SortTestsByPriority` — Dispatch Order

```go
func SortTestsByPriority(tests []NamedTest) {
    sort.SliceStable(tests, func(i, j int) bool {
        if tests[i].Priority != tests[j].Priority {
            return tests[i].Priority < tests[j].Priority  // High=0 first
        }
        return tests[i].Order < tests[j].Order  // then by registration order
    })
}
```

`sort.SliceStable` uses a merge sort variant (stable = equal elements keep their original order). Priority 0 (High) tests are dispatched first, then 1 (Normal), then 2 (Low). Within the same priority, registration order is preserved.

### `executeSingleTest()` — One Test's Lifecycle

```go
func executeSingleTest(workerID int, parentT KTestT, test NamedTest, config *Config) testResult {
    var testStart time.Time = time.Now()

    // 1. Create isolated T
    var childT *ktestT = &ktestT{name: test.Name, failed: false, depth: 0}

    // 2. Launch isolated browser (own Chromium process!)
    browser, page, launchErr = launchIsolatedBrowser(config)
    if launchErr != nil {
        childT.Errorf("failed to launch browser: %v", launchErr)
        return buildResult(test, childT, testStart, workerID)
    }
    defer browser.Close()
    defer page.Close()

    // 3. BeforeEach hook
    ExecuteGlobalBeforeEach(page)

    // 4. Run test with panic recovery
    runTestWithRecovery(childT, test, page, shortName, workerID, config)

    // 5. AfterEach hook
    ExecuteGlobalAfterEach(page)

    // 6. Propagate failure to parent
    if childT.Failed() { parentT.Errorf("test %s failed", shortName) }

    return buildResult(test, childT, testStart, workerID)
}
```

#### Step 2: `launchIsolatedBrowser` — OS Process Isolation

```go
func launchIsolatedBrowser(config *Config) (*kexas.Browser, *kexas.Page, error) {
    var opts *launcher.Options = launcher.DefaultOptions()
    opts.Headless = config.Headless
    if config.BrowserExecutable != "" { opts.ExecutablePath = config.BrowserExecutable }
    browser, err = kexas.Launch(opts)
    page, err = browser.FirstPage()
    return browser, page, nil
}
```

Each call to `kexas.Launch` spawns a **new Chromium OS process** with a **unique temp profile directory**. With 6 workers, you get 6 Chromium processes, each with its own:
- PID
- Virtual address space (~500 MB–2 GB)
- Temp profile dir (`/tmp/kexas-chrome-<timestamp>-<random>`)
- WebSocket debugger port
- Renderer processes (one per tab)

This is **process-level isolation** — the strongest form. Cookies, localStorage, cache, and all browser state are completely independent.

#### Step 4: `runTestWithRecovery` — Panic Safety

```go
func runTestWithRecovery(childT *ktestT, test NamedTest, page *kexas.Page, ...) {
    defer func() {
        if r := recover(); r != nil {
            childT.Errorf("worker[%d] test %s panicked: %v", workerID, shortName, r)
            if config.ScreenshotOnFail {
                takeScreenshot(childT, page, screenshotName, config.ScreenshotDir)
            }
        }
    }()
    test.Func(page, childT)  // <-- actual test execution
    if childT.Failed() && config.ScreenshotOnFail {
        takeScreenshot(childT, page, screenshotName, config.ScreenshotDir)
    }
}
```

**Why `recover()`?** In Go, a `panic` unwinds the goroutine's stack, running deferred functions in LIFO order. If no `recover()` catches it, the entire program crashes. In a parallel test worker, a panic in one test would kill ALL workers.

`recover()` inside a `defer` catches the panic value, marks the test as failed, and allows the worker goroutine to continue processing the next test from the queue. This is essential for test isolation — one crashing test must not affect others.

**Screenshot on panic**: After recovery, the page may still be in a usable state (the browser process is separate). The code attempts a screenshot for debugging. If the browser crashed too, the screenshot call will fail gracefully (logged but ignored).

---

## 7. Hooks — BeforeAll, BeforeEach, AfterEach, AfterAll

Source: `kexas/ktest/ktest_hooks.go`

### Global Hook Storage

```go
var (
    globalBeforeAll  func()
    globalAfterAll   func()
    globalBeforeEach func(*kexas.Page)
    globalAfterEach  func(*kexas.Page)
)
```

These are package-level function pointers. Each is 8 bytes (pointer to a function value, which itself holds a function pointer + closure variables). When nil, the hook is skipped.

### Registration

```go
func BeforeEach(hookFunc func(*kexas.Page)) interface{} {
    globalBeforeEach = hookFunc
    return nil
}
```

Simple assignment. Returns `nil` so it can be used in `AlphaInit(...)` which expects `interface{}`.

### Execution Order

```
Main Goroutine:
  ExecuteGlobalBeforeAll()        ← once, before any tests

Worker Goroutine (per test):
  ExecuteGlobalBeforeEach(page)   ← before each test, receives the test's Page
  test.Func(page, t)              ← the actual test
  ExecuteGlobalAfterEach(page)    ← after each test

Main Goroutine:
  ExecuteGlobalAfterAll()         ← once, after all tests
```

### Thread Safety of BeforeEach

In parallel mode, `ExecuteGlobalBeforeEach(page)` is called by multiple worker goroutines, each with a different `page`. The hook function itself must be **thread-safe** and **idempotent**. Example safe hook:

```go
ktest.BeforeEach(func(page *kexas.Page) {
    page.Navigate("https://example.com/login")  // each page is independent
})
```

Example UNSAFE hook:

```go
var count int  // shared mutable state!
ktest.BeforeEach(func(page *kexas.Page) {
    count++  // DATA RACE
})
```

---

## 8. The Suite API — Legacy Test Organization

Source: `kexas/ktest/ktest.go`

### `Suite` Struct

```go
type Suite struct {
    Browser *kexas.Browser
    Page    *kexas.Page
    t       KTestT
    config  *Config
}
```

The Suite pattern uses struct embedding:

```go
type MySuite struct {
    ktest.Suite  // embedded — MySuite "inherits" Browser, Page, T()
}
func (s *MySuite) TestLogin() {
    s.Page.Navigate("https://example.com")
}
```

### `RunWithConfigInternal` — Suite Execution

```go
func RunWithConfigInternal(t KTestT, suite interface{}, config *Config, testFilter string) {
    baseSuite = validateAndSetupSuite(t, suiteValue)  // reflection: find embedded Suite
    browser, page = setupTestEnvironment(t, config)   // launch browser
    baseSuite.Browser = browser
    baseSuite.Page = page
    runTestLifecycle(t, suiteValue, suiteType, baseSuite, config, testFilter)
}
```

### `findTestMethods` — Reflection Discovery

```go
func findTestMethods(suiteType reflect.Type, testFilter string) []reflect.Method {
    for i := 0; i < suiteType.NumMethod(); i++ {
        var method reflect.Method = suiteType.Method(i)
        if strings.HasPrefix(method.Name, "Test") { methods = append(methods, method) }
    }
}
```

Go's `reflect.Type.NumMethod()` reads from the type's **method table** — a compile-time structure embedded in the binary. No runtime allocation. `Method(i)` returns a `reflect.Method` with the method name and a `reflect.Value` that can be `.Call(nil)`'d to invoke it.

### `runTestAttempt` — Test With Timeout

```go
var done chan bool = make(chan bool, 1)
go func() {
    defer func() { if r := recover(); r != nil { ... }; done <- true }()
    testMethod.Call(nil)
}()

select {
case <-done:
    return !t.Failed()
case <-time.After(config.Timeout):
    t.Errorf("test timeout after %v", config.Timeout)
    return false
}
```

The test runs in a **separate goroutine** so it can be timed out. The `select` statement blocks until either:
- The test finishes (sends to `done`).
- The timeout fires (`time.After` returns a channel that receives after `config.Timeout`).

If the timeout fires first, the test goroutine is **NOT killed** — Go has no way to forcibly stop a goroutine. The goroutine continues running but its result is ignored. This is a known limitation. In practice, the deferred `browser.Close()` at the call site will kill the Chromium process, which will cause any pending CDP calls to fail, effectively unblocking the stuck goroutine.

---

## 9. kassert — The Fluent Assertion Library

Source: `kexas/kassert/kassert.go`

### Design

kassert follows the **fluent assertion** pattern (inspired by FluentAssertions in C# and Playwright assertions):

```go
kassert.That(t, actualValue).Equals(expectedValue)
kassert.That(t, text).Named("page title").Contains("Google")
```

### `Assertion` Struct

```go
type Assertion struct {
    t      TestingT     // the test context (for reporting failures)
    actual interface{}  // the value being asserted
    name   string       // human-readable name for error messages
}
```

### `That()` — Entry Point

```go
func That(t TestingT, actual interface{}) *Assertion {
    return &Assertion{t: t, actual: actual, name: "value"}
}
```

Returns a pointer so methods can be chained: `That(t, x).Named("x").Equals(5).IsGreaterThan(0)`.

### `Equals()` — Deep Comparison

```go
func (a *Assertion) Equals(expected interface{}) *Assertion {
    if !reflect.DeepEqual(a.actual, expected) {
        a.t.Helper()
        a.t.Errorf("Expected %s to equal:\n  %v\nbut got:\n  %v", a.name, expected, a.actual)
    }
    return a  // allows chaining
}
```

`reflect.DeepEqual` recursively compares two values:
- For primitives: value equality.
- For slices/arrays: element-by-element deep equality.
- For maps: key-by-key deep equality.
- For structs: field-by-field deep equality.
- For pointers: follows the pointer and compares the pointed-to values.

Cost: O(n) where n is the total size of the data structure. For simple values (string, int), it's ~50 nanoseconds. For large slices, it can be slower.

`a.t.Helper()` marks this function as a test helper. When `t.Errorf` reports a failure, Go's testing framework skips helper functions in the stack trace, showing the user's test code as the failure location instead of the assertion library code.

### Numeric Assertions — `toFloat64`

```go
func toFloat64(value interface{}) (float64, bool) {
    var val reflect.Value = reflect.ValueOf(value)
    switch val.Kind() {
    case reflect.Int, reflect.Int8, ..., reflect.Int64:
        return float64(val.Int()), true
    case reflect.Uint, ..., reflect.Uint64:
        return float64(val.Uint()), true
    case reflect.Float32, reflect.Float64:
        return val.Float(), true
    default:
        return 0, false
    }
}
```

This converts any numeric type to `float64` for comparison. Go's type system doesn't allow comparing `int` and `int64` directly — you need explicit conversion. `toFloat64` uses reflection to handle all numeric types uniformly.

**Precision note**: Converting `int64` to `float64` can lose precision for values > 2^53 (9007199254740992). For test assertions, this is almost never an issue.

### `TestingT` Interface

```go
type TestingT interface {
    Helper()
    Error(args ...interface{})
    Errorf(format string, args ...interface{})
}
```

This is kassert's own interface — a subset of `testing.T`. Both `*testing.T` and ktest's `ktestT` satisfy it. This makes kassert usable with both `go test` and `ktest.AutoRun`.

---

## 10. HTML Report Generation

Source: `kexas/ktest/report/`

### Flow

```go
func generateHTMLReport(results []testResult, config *Config) {
    collector := report.NewCollector(config.ParallelSet)
    for _, r := range results {
        collector.Add(report.TestCaseResult{Name: r.name, Status: ..., Duration: r.elapsed, ...})
    }
    htmlReport := collector.BuildReport(projectName)
    outputPath, versionedPath, err := report.GenerateFiles(htmlReport, config.ReportDir)
}
```

`report.GenerateFiles` writes two files:
1. `report.html` — The main report (overwritten each run).
2. `report-<timestamp>.html` — A versioned copy (never overwritten) for historical comparison.

Both files are written via `os.WriteFile` (`open(2)` + `write(2)` + `close(2)`). The versioned copy uses `time.Now().Format("20060102-150405")` for the timestamp.

---

## 11. Process Tree in Parallel Mode

```
main process (Go binary, PID 100)
├── goroutine: main (loadConfig, BeforeAll, AfterAll, report)
├── goroutine: worker 0
│   └── child process: Chromium (PID 201, profile /tmp/kexas-chrome-xxx-aaa)
│       ├── renderer (PID 202)
│       └── GPU process (PID 203)
├── goroutine: worker 1
│   └── child process: Chromium (PID 301, profile /tmp/kexas-chrome-xxx-bbb)
│       ├── renderer (PID 302)
│       └── GPU process (PID 303)
├── goroutine: worker 2
│   └── child process: Chromium (PID 401, profile /tmp/kexas-chrome-xxx-ccc)
│       └── ...
├── goroutine: result collector (range resultsChan)
└── goroutine: wg.Wait → close(resultsChan)
```

Total memory with 6 workers: Go process (~50 MB) + 6 × Chromium (~150 MB each) ≈ **~1 GB**.

---

## 12. The sync.Once Fix — Why Cleanup Runs Only Once

```go
var cleanOnce sync.Once

func Launch(ctx context.Context, opts *Options) (*Browser, error) {
    cleanOnce.Do(func() { cleanStaleTempProfiles(log) })
    // ... launch Chromium ...
}
```

Without `sync.Once`, every worker's `Launch` call would run `cleanStaleTempProfiles`, which deletes ALL `/tmp/kexas-chrome-*` directories. In a 6-worker setup:
1. Worker 0 launches Chromium → creates `/tmp/kexas-chrome-xxx-aaa`.
2. Worker 1 calls `Launch` → `cleanStaleTempProfiles` runs → deletes Worker 0's directory!
3. Worker 0's Chromium crashes with "No such file or directory" for SingletonLock.

With `sync.Once`, cleanup runs exactly once (on the first `Launch` call), before any worker has created a temp directory. Subsequent `Launch` calls skip cleanup entirely. `sync.Once` uses an atomic flag + mutex internally — the flag is checked atomically (1 CPU cycle), and the mutex is only acquired on the first call.
