# Kexas

> High-performance, parallel browser automation for Go

Kexas is a modern browser automation platform for Go. It drives Chromium directly through the Chrome DevTools Protocol (CDP), launches every test in an isolated browser profile, and ships a batteries-included test framework, assertion library, recorder, and HTML reporting pipeline.

## Features

- 🚀 **Direct CDP control** — Raw WebSocket transport, no external drivers
- ⚡ **Parallel orchestration** — `KEXAS_WORKERS=N` fans out isolated browsers with automatic retries
- 🧪 **Zero-boilerplate tests** — `ktest` fluent API + `.WithPriority()` + Playwright-style `ktest.Step()` instrumentation
- ✅ **kassert** — Fluent assertions, structured errors, log capture, extended helpers
- � **Grafana-inspired HTML report** — Expandable rows, structured steps, per-test console logs, donut + slowest-test insights
- 🕹️ **Recorder & multitab APIs** — Generate scripts, control multiple targets, advanced storage + cookie helpers
- 📦 **Single binary runtime** — Fresh Chromium profile per launch, auto cleanup, optional packaged browser
- 📸 **Artifacts** — Screenshots on failure, per-test stdout capture, console section in reports
- ⚙️ **Configurable** — JSON config, env overrides, headless/headed toggle, timeouts, custom Chrome executable

## Installation

```bash
go get github.com/jankylewis/kexas
```

## Quick Start

### Basic Browser Automation

```go
package main

import (
    "log"
    "github.com/jankylewis/kexas"
)

func main() {
    // Launch browser
    browser, err := kexas.Launch()
    if err != nil {
        log.Fatal(err)
    }
    defer browser.Close()

    // Create page and navigate
    page, err := browser.NewPage()
    if err != nil {
        log.Fatal(err)
    }
    defer page.Close()

    // Navigate to website
    err = page.Navigate("https://example.com")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Navigation successful!")
}
```

### Zero-Boilerplate Testing with AlphaInit

```go
package main

import (
    "github.com/jankylewis/kexas"
    "github.com/jankylewis/kexas/ktest"
    "github.com/jankylewis/kexas/kassert"
)

var _ = kexas.AlphaInit(
    ktest.Test("HomepageLoad", func(page *kexas.Page, t ktest.KTestT) {
        err := page.Navigate("https://example.com")
        kassert.ThatError(t, err).IsNil()
        
        title, err := page.Title()
        kassert.That(t, title).Contains("Example Domain")
    }),
    
    ktest.Test("PageTitle", func(page *kexas.Page, t ktest.KTestT) {
        title, err := page.Title()
        kassert.ThatError(t, err).IsNil()
        kassert.That(t, title).IsNotEmpty()
    }),
)

func main() {
    ktest.AutoRun()  // Automatically discovers and runs all tests
}
```

## Project Status

🚧 **Building fast** — `v0.1.0-dev`

### ✅ Currently Implemented
- Chromium launcher with self-assigned port + exponential backoff retries (5 attempts)
- Fresh temp user-data-dir per launch with automatic cleanup + stale-profile sweeper
- CDP client, multitab sessions, storage, cookies, recorder, scroll APIs
- Page + element interactions, strict finders, waits, kwait helpers
- `ktest` (groups, hooks, priorities A–Z, parallel runner, structured steps)
- `kassert` (core + extended matchers, structured logging)
- HTML report generator + embedded CSS/JS + per-test console + donut insights
- Saucelab sample suite (100 tests) exercising every subsystem

### 🚧 Coming Soon
- Network interception + HAR capture
- Video artifacts and PDF export
- Windows / Linux release builds

## Architecture (Reindexed Feb 28, 2026)

Every file under `kexas/`, grouped by package:

```
kexas/                              Root package — public API surface
├── kexas.go                        Package doc + Version constant ("0.1.0-dev")
├── kexas_alpha_init.go             AlphaInit() zero-boilerplate test registration trigger
├── options.go                      LaunchOptions type alias → launcher.Options
├── browser.go                      Browser struct, Launch(), NewPage(), Close(), multitab attach
├── page.go                         Page struct, Navigate(), Title(), EvalJS(), Screenshot(), Close()
├── page_find.go                    Find(), FindAll(), FindStrict() — CSS/XPath element selection
├── page_wait.go                    WaitForElementVisible(), WaitForSelector() with timeout
├── page_scroll.go                  ScrollToTop(), ScrollToBottom(), ScrollBy() — page-level scrolling
├── element.go                      Element struct, objectId-based interaction via Runtime.callFunctionOn
├── element_hover.go                Hover(), HoverWithOffset() — mouse hover via Input.dispatchMouseEvent
├── element_interaction.go          Click(), Type(), Clear(), SelectOption(), Press() — user actions
├── element_methods.go              GetAttribute(), GetText(), GetInnerHTML(), IsVisible(), IsEnabled()
├── element_scroll.go               ScrollIntoView() — element-level scrollIntoViewIfNeeded
├── cookie.go                       Cookie struct, GetCookies(), SetCookie(), DeleteCookies(), ClearCookies()
├── storage.go                      LocalStorage/SessionStorage — Get/Set/Remove/Clear/GetAll/Has/Length
├── recorder.go                     Recorder — frame capture, save to disk, custom config, start/stop
├── go.mod                          Module: github.com/jankylewis/kexas
├── go.sum                          Dependency checksums
│
├── errors/                         Public error types
│   └── errors.go                   Sentinel errors (ErrElementNotFound, ErrTimeout, …) + structured error types
│
├── internal/                       Framework internals (not importable by consumers)
│   ├── cdp/                        Chrome DevTools Protocol client
│   │   ├── cdp.go                  WebSocket client, Send/Subscribe, command-response matching, event bus
│   │   ├── commands.go             Command struct — method, description, category metadata
│   │   └── const.go                CDP method constants (CmdPageNavigate, CmdDOMQuerySelector, …)
│   ├── agent/                      CDP agent lifecycle manager (Playwright-style smart enablement)
│   │   ├── enabled_agents.go       Agent name constants (DOM, Input, Runtime, Network, Page, …)
│   │   ├── agent_manager.go        AgentManager — enable/disable agents, track state, mutex-guarded
│   │   ├── agent_states.go         AgentStates — per-agent runtime state (enabled, lastUsed, context)
│   │   ├── agent_context.go        AgentContext — runtime context structs (DOM, Input, Network, …)
│   │   ├── context_init.go         initDOMContext(), initRuntimeContext(), … — per-agent init routines
│   │   ├── operations.go           IsEnabled(), Enable(), Disable() — agent state operations
│   │   └── smart_enablement.go     EnsureAgent() — lazy enable on first use (Playwright pattern)
│   └── logger/                     Structured leveled logging
│       └── logger.go               Logger — Debug/Info/Warn/Error, colored output, component tags
│
├── kapi/                           HTTP API client (like Playwright's APIRequestContext)
│   ├── client.go                   NewClient(), Get/Post/Put/Patch/Delete, headers, auth, timeouts
│   └── response.go                 Response struct — StatusCode, Headers, Body, JSON helpers
│
├── kassert/                        Fluent assertion library
│   ├── kassert.go                  That() + Assertion — Equals, Contains, IsNil, HasLength, IsBetween, …
│   ├── kassert_errors.go           ThatError() + ErrorAssertion — IsNil, HasMessage, IsType
│   ├── kassert_extended.go         MatchesRegex, IsEmpty, HasPrefix, HasSuffix, IsOneOf, …
│   └── kassert_logging.go          AssertionStats (pass/fail counters), configurable LogFunc
│
├── ktest/                          Test framework (Playwright Test + TestNG inspired)
│   ├── ktest.go                    Package doc, Suite struct, Config, Main() entry point
│   ├── ktest_autorun.go            AutoRun() — auto-discover Test* functions, launch, run
│   ├── ktest_registration.go       Test() — fluent test registration with .WithPriority()
│   ├── ktest_groups.go             Group() — hierarchical test organization, sorted execution
│   ├── ktest_hooks.go              BeforeAll/AfterAll/BeforeEach/AfterEach — global lifecycle hooks
│   ├── ktest_types.go              ktestT — thread-safe KTestT impl with log capture (Log/Logf/Error/Errorf)
│   ├── ktest_impl.go              discoverTestFunctions(), internal runner helpers
│   ├── ktest_runner.go             RunAll() — sequential/parallel dispatch, report generation, exit code
│   ├── ktest_parallel.go           Parallel worker pool, isolated browser per test, retry with backoff (5 attempts)
│   ├── ktest_steps.go              Step() — Playwright-style structured steps with timing + status
│   ├── ktest_steps_test.go         Unit tests for Step() logic
│   └── report/                     HTML report generator (Grafana/Prometheus-inspired)
│       ├── report_model.go         TestReport, TestSuiteResult, TestCaseResult, TestStep, TestStatus
│       ├── report_collector.go     Collector — thread-safe result accumulator, BuildReport(), grouping
│       ├── report_html.go          Generate(), buildHTML(), writeTestRow/Steps/Error/Logs, formatters
│       ├── report_assets.go        Embedded CSS — dark/light theme, donut chart, step-list, log-section
│       ├── report_script.go        Embedded JS — theme toggle, tab filtering, search, expand/collapse
│       ├── report_steps_test.go    Unit tests — steps rendering, error display, worker tags, duration
│       ├── report_filtering_test.go Unit tests — data attributes, detail adjacency, JS regression guards
│       └── report_logs_test.go     Unit tests — console log rendering, escaping, nil/empty handling
│
├── kwait/                          Wait strategies
│   └── kwait.go                    WaitUntil enum (Commit, DOMContentLoaded, Load, NetworkIdle) + helpers
│
├── launcher/                       Chromium process management
│   ├── launcher.go                 Launch(), Close(), buildArgs(), findChromium(), extractDebuggerURL()
│   ├── downloader.go               downloadChromium() — fetch Chrome for Testing, unzip, cache
│   ├── launcher_cleanup.go         cleanStaleTempProfiles() (sync.Once), killExistingChromeProcesses()
│   ├── launcher_test.go            Unit tests — launch, close, port, args, profile cleanup
│   └── launcher_parallel_test.go   Parallel launch stress tests — concurrent workers, port isolation
│
└── tests/                          Integration + regression test suites (31 files)
    ├── README.md                   Test suite documentation and run instructions
    ├── agent_test.go               Agent manager enable/disable/smart-enablement tests
    ├── alpha_init_test.go          AlphaInit registration + auto-discovery tests
    ├── browser_test.go             Browser launch, connect, close tests
    ├── cdp_test.go                 CDP client send/subscribe/event tests
    ├── config_timeout_test.go      Configuration loading + timeout behavior tests
    ├── cookie_test.go              Cookie get/set/delete/clear tests
    ├── critical_selection_test.go  Element selection edge cases (strict, missing, multiple)
    ├── critical_unit_test.go       Core API unit tests (Navigate, Title, EvalJS, Screenshot)
    ├── critical_wait_test.go       Wait strategy + timeout tests
    ├── element_enhanced_test.go    Extended element methods (attributes, visibility, enabled)
    ├── element_interaction_test.go Click, type, clear, select, press interaction tests
    ├── element_test.go             Element struct + objectId lifecycle tests
    ├── errors_test.go              Error types + sentinel error tests
    ├── kapi_test.go                HTTP client get/post/put/delete + auth + timeout tests
    ├── kassert_extended_test.go    Extended assertion tests (regex, empty, prefix, oneOf)
    ├── kassert_test.go             Core assertion tests (equals, contains, nil, length, between)
    ├── ktest_parallel_test.go      Parallel runner, worker isolation, retry logic tests
    ├── ktest_test.go               Test registration, groups, hooks, priority tests
    ├── kwait_test.go               Wait strategy enum + helper tests
    ├── launcher_test.go            Launcher integration tests (findChromium, port extraction)
    ├── logger_test.go              Logger level, color, component tag tests
    ├── multitab_test.go            Multi-page attach/switch/close tests
    ├── page_find_strict_test.go    Strict finder (exactly one match or error) tests
    ├── page_find_test.go           Find/FindAll CSS + XPath tests
    ├── page_test.go                Page navigate, title, eval tests
    ├── page_wait_test.go           WaitForElementVisible + WaitForSelector tests
    ├── recorder_test.go            Recorder start/stop/save/config tests
    ├── report_html_test.go         Full HTML report generation + structure tests
    ├── report_test.go              Report model + collector + stats tests
    ├── scroll_test.go              Page + element scroll tests
    └── storage_test.go             LocalStorage + SessionStorage tests
```

### Test execution flow

1. `launcher` spins up Chromium with a fresh temp profile (`kexas-chrome-<nano>-<random>`) and `--remote-debugging-port=0`.
2. `internal/cdp` connects to the DevTools websocket. `internal/agent` lazily enables CDP agents on first use.
3. `browser.go` wraps the CDP client into `Browser`; `page.go` creates `Page` instances per tab.
4. `ktest` dispatches tests to workers, each with an isolated browser. Steps, logs, and errors are captured per test.
5. `ktest/report` collects results → groups by file → emits self-contained `test-results/report.html` + timestamped archive.

## Configuration

Kexas supports optional configuration via `kexas.config.json` at your project root:

```json
{
  "headless": true,
  "timeout": 30000,
  "retries": 0,
  "parallel": false,
  "screenshotOnFail": true,
  "screenshotDir": "./test-results/screenshots",
  "videoDir": "./test-results/videos",
  "slowMo": 0,
  "baseURL": "",
  "browserExecutable": ""
}
```

### Testing with Configuration

When using the testing framework, config is automatically loaded:

```go
package main

import (
    "github.com/jankylewis/kexas"
    "github.com/jankylewis/kexas/ktest"
    "github.com/jankylewis/kexas/kassert"
)

var _ = kexas.AlphaInit(
    ktest.Test("ConfigExample", func(page *kexas.Page, t ktest.KTestT) {
        // Test with config automatically loaded
        err := page.Navigate("https://example.com")
        kassert.ThatError(t, err).IsNil()
    }),
)

func main() {
    ktest.AutoRun()  // Auto-loads kexas.config.json
}
```

## Testing Framework

### AlphaInit Pattern

The AlphaInit pattern provides zero-boilerplate test setup:

```go
var _ = kexas.AlphaInit(
    // Test registration
    ktest.Test("MyTest", func(page *kexas.Page, t ktest.KTestT) {
        // Test logic here
    }),
    
    // Group-based organization
    ktest.Group("Authentication", func() {
        ktest.Test("Login", func(page *kexas.Page, t ktest.KTestT) {
            // Login test
        }),
        
        ktest.Test("Logout", func(page *kexas.Page, t ktest.KTestT) {
            // Logout test
        }),
    }),
    
    // Lifecycle hooks
    ktest.BeforeAll(func() {
        // Global setup
    }),
    
    ktest.BeforeEach(func(page *kexas.Page) {
        // Per-test setup
    }),
)
```

### Assertions with kassert

Rich assertion library with fluent API:

```go
// Basic assertions
kassert.That(t, value).Equals(expected)
kassert.That(t, value).IsNotNil()
kassert.That(t, str).Contains("substring")

// Error assertions
kassert.ThatError(t, err).IsNil()
kassert.ThatError(t, err).HasMessage("timeout")

// Numeric assertions
kassert.That(t, count).IsGreaterThan(0)
kassert.That(t, score).IsBetween(0, 100)

// String assertions
kassert.That(t, title).StartsWith("Welcome")
kassert.That(t, url).EndsWith(".com")
```

### Test Execution Model

- **Parallel or sequential** — set `KEXAS_WORKERS` (default 1). Each worker isolates its own browser.
- **Retries with backoff** — launcher attempts up to 5 times with exponential delay and line-level logging.
- **Automatic cleanup** — browsers close per test, temp profiles removed, ports released.
- **Artifacts** — screenshots on failure, per-test console captured, fluent step timeline in report.

## Wait Strategies

Kexas implements Playwright-standard wait strategies for page navigation:

```go
import "github.com/jankylewis/kexas/kwait"

// Wait until response headers received (fastest)
page.Navigate(url, kwait.WaitUntilCommit)

// Wait until DOM is ready
page.Navigate(url, kwait.WaitUntilDOMContentLoaded)

// Wait until all resources loaded (default)
page.Navigate(url, kwait.WaitUntilLoad)

// Wait until no network activity for 500ms (slowest, most reliable)
page.Navigate(url, kwait.WaitUntilNetworkIdle)
```

## Requirements

- Go 1.25.5 or later
- Chrome or Chromium browser installed

## Development & Reindexing Checklist

```
# Run all core unit tests (kexas package)
go test ./...

# Focused subsystem suites
go test ./launcher -run . -v
go test ./ktest/report -run . -v

# Saucelab sample suite (100 real-world tests + HTML report)
cd saucelab
KEXAS_WORKERS=4 go run .

# Preview latest report locally
python3 -m http.server 8923 --directory test-results
open test-results/report.html
```

## Philosophy

Kexas is built on the belief that modern automation infra should be:
- **Performance-first** — CDP, goroutines, minimized overhead
- **Resilient-by-default** — Temp profiles, retries, structured telemetry
- **Developer-friendly** — Fluent APIs, recorder, zero-boilerplate AlphaInit
- **Observable** — Detailed HTML reports, per-test logs, Grafana-inspired UX

## License

MIT

## Contributing

Contributions welcome! This project is in early development.
