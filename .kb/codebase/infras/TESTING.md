# Testing Strategy

> Testing conventions and best practices

**Last Updated:** February 24, 2026

---

## Test Execution Model

### Sequential Test Execution (One Browser Per Test)

**Principle**: Each test runs in complete isolation with its own browser instance.

**Execution Flow**:

#### Test #1
1. **Launch new browser** (Chrome for Testing)
2. **Single tab only** - no additional tabs unless explicitly created by test
3. **Execute test** on the single tab
4. **Graceful browser cleanup** - close browser regardless of test status
5. **Test #1 completes**

#### Test #2
1. **Launch new browser** (fresh instance)
2. **Single tab only** - no additional tabs unless explicitly created by test
3. **Execute test** on the single tab
4. **Graceful browser cleanup** - close browser regardless of test status
5. **Test #2 completes**

#### ... continue for all tests

**Key Requirements**:

- **One browser per test** - No browser reuse between tests
- **One tab per test** - Default single tab, additional tabs only if test explicitly creates them
- **Sequential execution** - Tests run one after another, not in parallel
- **Guaranteed cleanup** - Browser must close gracefully after each test
- **Status-agnostic cleanup** - Browser closes whether test passes, fails, or panics

**Why This Model**:

1. **Isolation** - Tests cannot interfere with each other
2. **Clean state** - Each test starts with fresh browser state
3. **Reliability** - No cross-test contamination
4. **Debugging** - Failed tests don't affect subsequent tests

**Implementation Notes**:

```go
// ✅ CORRECT - One browser per test
func TestSomething(t *testing.T) {
    browser, err := kexas.Launch(opts)
    if err != nil {
        t.Fatal(err)
    }
    defer browser.Close() // Guaranteed cleanup
    
    page, err := browser.NewPage()
    if err != nil {
        t.Fatal(err)
    }
    defer page.Close()
    
    // Test logic here...
}

// ❌ WRONG - Browser reuse between tests
var globalBrowser *kexas.Browser

func TestSomething1(t *testing.T) {
    if globalBrowser == nil {
        globalBrowser, _ = kexas.Launch(opts)
    }
    // Test logic...
}

func TestSomething2(t *testing.T) {
    // Reusing browser from TestSomething1 - BAD!
    page, _ := globalBrowser.NewPage()
    // Test logic...
}
```

---

## Test Organization

### Location

**All tests in `tests/` directory.**

**Naming**: `tests/<package>_test.go`

**Examples**:
- `tests/browser_test.go` - Tests for browser.go
- `tests/page_test.go` - Tests for page.go
- `tests/cdp_test.go` - Tests for internal/cdp/
- `tests/launcher_test.go` - Tests for internal/launcher/

---

## Test Types

### 1. Unit Tests

**Purpose**: Test individual functions/methods in isolation.

**Characteristics**:
- Fast (milliseconds)
- No external dependencies
- No real browser
- Run on every commit

**Example**:
```go
func TestResponseError_Error(t *testing.T) {
    var err *cdp.ResponseError = &cdp.ResponseError{
        Code:    -32000,
        Message: "Invalid parameters",
    }
    
    var expected string = "cdp error -32000: Invalid parameters"
    if err.Error() != expected {
        t.Errorf("expected %q, got %q", expected, err.Error())
    }
}
```

---

### 2. Integration Tests

**Purpose**: Test with real browser, full system.

**Characteristics**:
- Slow (seconds)
- Requires browser
- Tests real scenarios
- Run manually or in CI

**Pattern**: Mark with `t.Skip()` for manual execution.

**Example**:
```go
func TestPage_Navigate(t *testing.T) {
    t.Skip("requires real browser - integration test")
    
    browser, err := kexas.Launch(&kexas.LaunchOptions{Headless: true})
    if err != nil {
        t.Fatal(err)
    }
    defer browser.Close()
    
    page, err := browser.NewPage()
    if err != nil {
        t.Fatal(err)
    }
    defer page.Close()
    
    err = page.Navigate("https://example.com")
    if err != nil {
        t.Errorf("navigation failed: %v", err)
    }
}
```

---

## Test Naming Convention

**Format**: `Test<Function>_<Scenario>`

**Examples**:
```go
func TestBrowserLaunch_Success(t *testing.T) { ... }
func TestBrowserLaunch_ChromiumNotFound(t *testing.T) { ... }
func TestPageNavigate_InvalidURL(t *testing.T) { ... }
func TestCDPClient_SendToSession_Timeout(t *testing.T) { ... }
```

**Rules**:
- Start with `Test`
- Function/type name in PascalCase
- Scenario describes what's being tested
- Use underscores to separate parts

---

## Testing Framework

### For Unit Tests

**Use**: Standard Go `testing` package.

**Why**: Fast, simple, no dependencies.

**Example**:
```go
func TestThat_Equals_Success(t *testing.T) {
    kassert.That(t, 42).Equals(42)
}

func TestThat_Equals_Failure(t *testing.T) {
    // This test expects failure
    var mockT *testing.T = &testing.T{}
    kassert.That(mockT, 42).Equals(99)
    // Verify mockT.Failed() == true
}
```

---

### For Integration Tests

**Use**: `ktest` framework.

**Why**: Provides browser setup/teardown, config loading.

**Example**:
```go
import (
    "testing"
    "github.com/kexas-project/kexas/internal/ktest"
    "github.com/kexas-project/kexas/internal/kassert"
)

type MySuite struct {
    ktest.Suite
}

func (s *MySuite) TestNavigation() {
    page := s.Page()
    err := page.Navigate("https://example.com")
    kassert.ThatError(s.T(), err).IsNil()
    
    title, err := page.Title()
    kassert.ThatError(s.T(), err).IsNil()
    kassert.That(s.T(), title).Contains("Example")
}

func TestMySuite(t *testing.T) {
    ktest.Run(t, &MySuite{})
}
```

---

## Assertion Library

**Use**: `kassert` package.

**Why**: Fluent API, clear error messages, chainable.

### Value Assertions

```go
kassert.That(t, value).Equals(expected)
kassert.That(t, value).NotEquals(unexpected)
kassert.That(t, value).IsNil()
kassert.That(t, value).IsNotNil()
kassert.That(t, value).IsTrue()
kassert.That(t, value).IsFalse()
```

### String Assertions

```go
kassert.That(t, str).Contains("substring")
kassert.That(t, str).NotContains("substring")
kassert.That(t, str).StartsWith("prefix")
kassert.That(t, str).EndsWith("suffix")
kassert.That(t, str).IsEmpty()
kassert.That(t, str).IsNotEmpty()
```

### Collection Assertions

```go
kassert.That(t, slice).HasLength(5)
kassert.That(t, slice).IsEmpty()
kassert.That(t, slice).IsNotEmpty()
```

### Numeric Assertions

```go
kassert.That(t, num).IsGreaterThan(10)
kassert.That(t, num).IsLessThan(100)
kassert.That(t, num).IsGreaterThanOrEqual(10)
kassert.That(t, num).IsLessThanOrEqual(100)
```

### Error Assertions

```go
kassert.ThatError(t, err).IsNil()
kassert.ThatError(t, err).IsNotNil()
kassert.ThatError(t, err).HasMessage("expected error")
```

---

## Test Structure

### Arrange-Act-Assert Pattern

```go
func TestPageNavigate_Success(t *testing.T) {
    // Arrange - Set up test data and dependencies
    browser, _ := kexas.Launch(&kexas.LaunchOptions{Headless: true})
    defer browser.Close()
    page, _ := browser.NewPage()
    defer page.Close()
    
    // Act - Execute the operation being tested
    err := page.Navigate("https://example.com")
    
    // Assert - Verify the results
    kassert.ThatError(t, err).IsNil()
}
```

---

## Test Coverage

### Critical Paths First

**Priority 1**: Core functionality
- Browser launch and close
- Page creation and navigation
- CDP command sending
- Wait mechanisms

**Priority 2**: Error handling
- Invalid inputs
- Timeouts
- Connection failures
- Browser crashes

**Priority 3**: Edge cases
- Empty strings
- Nil pointers
- Concurrent operations
- Resource exhaustion

---

## Running Tests

### All Tests

```bash
go test ./tests/... -v
```

### Specific Test

```bash
go test ./tests -run TestBrowserLaunch_Success -v
```

### With Coverage

```bash
go test ./tests/... -cover
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests Only

```bash
# Remove skip statements temporarily
go test ./tests/... -v -tags=integration
```

### Parallel Execution

```bash
go test ./tests/... -v -parallel=4
```

---

## Test Helpers

### Common Setup

```go
// tests/tests.go
package tests

import "testing"

func SetupBrowser(t *testing.T) *kexas.Browser {
    t.Helper()
    
    browser, err := kexas.Launch(&kexas.LaunchOptions{Headless: true})
    if err != nil {
        t.Fatalf("failed to launch browser: %v", err)
    }
    
    t.Cleanup(func() {
        browser.Close()
    })
    
    return browser
}
```

### Usage

```go
func TestSomething(t *testing.T) {
    browser := SetupBrowser(t)
    page, _ := browser.NewPage()
    // ... test code
}
```

---

## Best Practices

### 1. Use t.Helper()

```go
func assertNoError(t *testing.T, err error) {
    t.Helper()  // Makes error point to caller, not this function
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### 2. Use t.Cleanup()

```go
func TestSomething(t *testing.T) {
    browser, _ := kexas.Launch(opts)
    t.Cleanup(func() {
        browser.Close()  // Guaranteed cleanup
    })
    
    // Test code
}
```

### 3. Use Subtests

```go
func TestBrowser(t *testing.T) {
    t.Run("Launch", func(t *testing.T) {
        // Test launch
    })
    
    t.Run("NewPage", func(t *testing.T) {
        // Test page creation
    })
}
```

### 4. Table-Driven Tests

```go
func TestNavigate(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr bool
    }{
        {"valid URL", "https://example.com", false},
        {"invalid URL", "not-a-url", true},
        {"empty URL", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := page.Navigate(tt.url)
            if (err != nil) != tt.wantErr {
                t.Errorf("Navigate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run unit tests
        run: go test ./tests/... -v
      
      - name: Run integration tests
        run: go test ./tests/... -v -tags=integration
      
      - name: Coverage
        run: |
          go test ./tests/... -coverprofile=coverage.out
          go tool cover -func=coverage.out
```

---

## Early Bailout Pattern (Updated February 27, 2026)

Integration tests that involve multi-step browser flows should **abort immediately on first failure** instead of continuing through doomed subsequent steps.

### Why Not `t.Fatal()` / `t.Fatalf()`

The `KTestT` interface's `Fatal` and `Fatalf` call `os.Exit(1)`, which:
- Kills the entire process immediately
- Skips `defer` cleanup (browser close, temp file removal)
- Leaves orphan Chromium processes and occupied ports

### Recommended Pattern: `t.Errorf()` + `return`

```go
func TestAmazonSignInComplete(t ktest.KTestT) {
    // Step 1: Navigate
    err = page.Navigate("https://www.amazon.com/")
    if err != nil {
        t.Errorf("Step 1 failed: %v", err)
        return  // Bail out, defer cleanup still runs
    }

    // Step 2: Find sign-in link
    signInLink, err := page.Find("a[data-nav-ref='nav_ya_signin']")
    if err != nil {
        t.Errorf("Step 2 failed: %v", err)
        return
    }

    // ... subsequent steps follow same pattern
}
```

**Benefits**:
- `defer` blocks execute normally (browser closes, ports released)
- Clear error message identifies exactly which step failed
- No wasted 10-second `Find()` timeout per subsequent doomed step

---

## Port Conflict Between Sequential Tests (Updated February 27, 2026)

When running multiple browser tests sequentially, the OS may not release the debugging port (default: 9222) immediately after the browser process is killed.

### Symptom
```
failed to extract debugger URL: ...
```

### Root Cause
The previous test's `Browser.Close()` → `Launcher.Close()` kills the Chromium process, but the OS may hold the port in `TIME_WAIT` state for several hundred milliseconds.

### Fix
`launcher.Close()` uses an adaptive polling mechanism to wait for the port to be released:

```go
func (b *Browser) Close() error {
    // Kill browser process
    b.cmd.Process.Kill()
    b.cmd.Wait()
    
    // Poll for port release with 3s timeout and 150ms intervals
    var portTimeout time.Duration = 3 * time.Second
    var pollInterval time.Duration = 150 * time.Millisecond
    
    for time.Since(start) < portTimeout {
        if isPortAvailable(b.port) {
            break // Port released, continue immediately
        }
        time.Sleep(pollInterval)
    }
    
    // Clean up temp profile
    os.RemoveAll(b.userDataDir)
}
```

**Benefits**:
- **Faster cleanup**: If the port releases quickly (e.g., 300ms), the test continues immediately instead of waiting the full 2 seconds
- **Adaptive timeout**: Waits up to 3 seconds if needed, but typically much less
- **Better reliability**: Actively checks port availability instead of guessing

### Test Selector Best Practices

When writing selectors for real-world sites:
- **Avoid hardcoded IDs** that change between site updates (e.g., `#continue`, `#auth-signin-button`)
- **Prefer attribute selectors** that are more stable (e.g., `input[type='submit']`)
- **Use data attributes** when available (e.g., `a[data-nav-ref='nav_ya_signin']`)

---

## Test Filtering with KEXAS_TEST_RUN (Updated February 27, 2026)

For AlphaInit/Group registered tests, you can filter which tests to run using the `KEXAS_TEST_RUN` environment variable.

### Usage

```bash
# Run a specific test by exact name
KEXAS_TEST_RUN=TestAmazonSignInComplete go run signin_tests.go

# Run tests matching a partial string (contains)
KEXAS_TEST_RUN=SignInComplete go run signin_tests.go
KEXAS_TEST_RUN=LoginOnly go run signin_tests.go

# Run all tests (no filter)
go run signin_tests.go
```

### Behavior

- **Exact match**: If the filter exactly matches a test name, only that test runs
- **Partial match**: If the filter is contained within a test name, that test runs
- **No matches**: If no tests match, the framework lists available tests and exits
- **No filter**: If `KEXAS_TEST_RUN` is not set, all registered tests run

### Example Output

```bash
$ KEXAS_TEST_RUN=TestAmazonSignInComplete go run signin_tests.go
🔍 KEXAS_TEST_RUN filter: TestAmazonSignInComplete
📋 Running 1 filtered test(s):
  - Amazon Sign-In Flow.TestAmazonSignInComplete
```

### Supported Test Types

This filtering works for:
- AlphaInit registered tests (`ktest.Test("name", ...)`)
- Group-based tests (`ktest.Group(...)` with nested tests)
- Both root-level and grouped tests

---

## Summary

Testing strategy:
- **Unit tests**: Fast, isolated, no browser
- **Integration tests**: Real browser, full system
- **Naming**: `Test<Function>_<Scenario>`
- **Framework**: Standard `testing` for units, `ktest` for integration
- **Assertions**: Use `kassert` for clear, fluent assertions
- **Early bailout**: Use `t.Errorf()` + `return`, never `t.Fatal()` (preserves defer cleanup)
- **Port conflicts**: Adaptive port polling (3s timeout, 150ms intervals) in `Browser.Close()` between sequential tests
- **Test filtering**: Use `KEXAS_TEST_RUN` environment variable to run specific tests
- **Coverage**: Critical paths first, then errors, then edge cases

