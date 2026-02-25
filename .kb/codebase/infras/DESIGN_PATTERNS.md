# Design Patterns

> Common patterns and best practices used in Kexas

**Last Updated:** February 24, 2026

---

## 1. Explicit Types (Fail-Fast)

**Rule**: Always specify types explicitly. Never rely on type inference for complex types.

```go
// ❌ BAD - Implicit, ambiguous
result := doSomething()
data := make(map[string]interface{})

// ✅ GOOD - Explicit types
var result CommandResponse = doSomething()
var data map[string]EventPayload = make(map[string]EventPayload)
```

**Why**: Catches type errors at compile time, improves readability, prevents bugs.

**Where Used**:
- All function parameters and return values
- All struct fields
- All variable declarations (except simple cases like `i := 0`)

---

## 2. Context Propagation

**Pattern**: Pass `context.Context` through the call chain for cancellation and timeouts.

```go
// Browser stores context
type Browser struct {
    ctx context.Context
    // ...
}

// Page inherits context from browser
type Page struct {
    ctx context.Context  // From browser
    // ...
}

// All operations use context
func (p *Page) Navigate(url string) error {
    result, err := p.sendCommand(p.ctx, "Page.navigate", params)
    // ...
}
```

**Why**: Enables graceful shutdown, timeout handling, and cancellation propagation.

**Where Used**:
- Browser launch and lifecycle
- Page operations
- CDP commands
- Wait conditions

---

## 3. Resource Cleanup (defer)

**Pattern**: Use `defer` for guaranteed resource cleanup.

```go
func Launch(opts *LaunchOptions) (*Browser, error) {
    process, err := launcher.Launch(ctx, opts)
    if err != nil {
        return nil, err
    }
    
    client, err := cdp.Connect(ctx, wsURL)
    if err != nil {
        process.Close()  // Clean up on error
        return nil, err
    }
    
    return &Browser{process: process, client: client}, nil
}

// User code
browser, err := kexas.Launch(opts)
if err != nil {
    log.Fatal(err)
}
defer browser.Close()  // Guaranteed cleanup
```

**Why**: Prevents resource leaks, ensures cleanup even on panic.

**Where Used**:
- Browser and page lifecycle
- File handles
- WebSocket connections
- Temporary directories

---

## 4. Error Wrapping

**Pattern**: Wrap errors with context using `fmt.Errorf` and `%w`.

```go
func (p *Page) Navigate(url string) error {
    result, err := p.sendCommand("Page.navigate", params)
    if err != nil {
        return fmt.Errorf("kexas: navigation failed: %w", err)
    }
    return nil
}
```

**Why**: Preserves error chain, adds context, enables `errors.Is()` and `errors.As()`.

**Convention**:
- Prefix with package name: `"kexas: ..."` or `"cdp: ..."`
- Add operation context: `"navigation failed"`, `"failed to connect"`
- Use `%w` to wrap underlying error

**Where Used**:
- All public API methods
- Internal package boundaries
- Error propagation

---

## 5. Goroutine-Safe Event Handling

**Pattern**: Use channels and mutexes for concurrent event handling.

```go
type Client struct {
    handlers   map[string][]EventHandler
    handlersMu sync.RWMutex
    // ...
}

func (c *Client) On(method string, handler EventHandler) {
    c.handlersMu.Lock()
    defer c.handlersMu.Unlock()
    c.handlers[method] = append(c.handlers[method], handler)
}

func (c *Client) handleEvent(event *Event) {
    c.handlersMu.RLock()
    handlers := c.handlers[event.Method]
    c.handlersMu.RUnlock()
    
    for _, handler := range handlers {
        go handler(event.Params)  // Concurrent execution
    }
}
```

**Why**: Thread-safe, non-blocking, leverages Go's concurrency.

**Where Used**:
- CDP event handling
- Page load events
- Network events (future)

---

## 6. Atomic Operations

**Pattern**: Use `sync/atomic` for lock-free state management.

```go
type Client struct {
    nextID atomic.Int64
    closed atomic.Bool
    // ...
}

func (c *Client) Send(method string, params map[string]interface{}) error {
    if c.closed.Load() {
        return ErrConnectionClosed
    }
    
    id := c.nextID.Add(1)
    // ...
}

func (c *Client) Close() error {
    if !c.closed.CompareAndSwap(false, true) {
        return nil // Already closed
    }
    // ...
}
```

**Why**: Lock-free, high performance, prevents race conditions.

**Where Used**:
- CDP command ID generation
- Connection state management
- Concurrent counters

---

## 7. Builder Pattern (Options)

**Pattern**: Use struct literals for configuration.

```go
type Options struct {
    Headless       bool
    Port           int
    Args           []string
    ExecutablePath string
}

func DefaultOptions() *Options {
    return &Options{
        Headless: true,
        Port:     9222,
        Args:     []string{},
    }
}

// Usage
opts := &kexas.LaunchOptions{
    Headless: false,
    Port:     9223,
}
browser, err := kexas.Launch(opts)
```

**Why**: Clear, explicit, easy to extend, no method chaining needed.

**Where Used**:
- Browser launch options
- Wait options
- Test configuration

---

## 8. Variadic Optional Parameters

**Pattern**: Use variadic parameters for optional arguments.

```go
func (p *Page) Navigate(url string, waitUntil ...kwait.WaitUntil) error {
    // Determine wait strategy (default to load)
    var strategy kwait.WaitUntil = kwait.WaitUntilLoad
    if len(waitUntil) > 0 {
        strategy = waitUntil[0]
    }
    
    // Use strategy
    err := p.WaitForLoadState(strategy, timeout)
    // ...
}

// Usage
page.Navigate(url)  // Uses default
page.Navigate(url, kwait.WaitUntilDOMContentLoaded)  // Custom
```

**Why**: Clean API, backward compatible, optional parameters.

**Where Used**:
- Page navigation wait strategies
- Optional timeouts
- Optional configurations

---

## 9. Guard Clauses (Early Returns)

**Pattern**: Check error conditions first, return early.

```go
// ❌ BAD - Nested conditions
func process(page *Page) error {
    if page != nil {
        if page.IsLoaded {
            if page.HasContent {
                // deeply nested logic
            }
        }
    }
    return nil
}

// ✅ GOOD - Guard clauses
func process(page *Page) error {
    if page == nil {
        return ErrNilPage
    }
    if !page.IsLoaded {
        return ErrPageNotLoaded
    }
    if !page.HasContent {
        return ErrNoContent
    }
    
    // Flat logic here
    return nil
}
```

**Why**: Reduces nesting, improves readability, follows fail-fast principle.

**Where Used**:
- Input validation
- State checking
- Error handling

---

## 10. Type Aliases for Clarity

**Pattern**: Use type aliases to expose internal types cleanly.

```go
// options.go
package kexas

import "github.com/kexas-project/kexas/internal/launcher"

// LaunchOptions configures browser launch behavior.
type LaunchOptions = launcher.Options

// DefaultLaunchOptions returns sensible default launch options.
func DefaultLaunchOptions() *LaunchOptions {
    return launcher.DefaultOptions()
}
```

**Why**: Clean public API, hides internal package structure, easy to refactor.

**Where Used**:
- Launch options
- Wait strategies
- Configuration types

---

## 11. Structured Logging

**Pattern**: Use key-value pairs for structured logs.

```go
log := logger.New("cdp")
log.Debug("sending command", "method", "Page.navigate", "url", url, "sessionId", sessionID)
log.Info("navigated successfully", "elapsed", elapsed, "url", url)
log.Error("navigation failed", "err", err, "url", url)
```

**Why**: Machine-parseable, easy to filter, provides context.

**Convention**:
- First parameter: message string
- Remaining parameters: key-value pairs
- Keys: lowercase, no spaces
- Values: any type

**Where Used**:
- All components (Browser, Page, CDP, Launcher)
- Debug information
- Error context

---

## 12. Sentinel Errors

**Pattern**: Define package-level error variables for common errors.

```go
// Common CDP errors.
var (
    ErrConnectionClosed error = errors.New("cdp: connection closed")
    ErrTimeout          error = errors.New("cdp: operation timed out")
    ErrInvalidMessage   error = errors.New("cdp: invalid message format")
)

// Usage
if err == cdp.ErrConnectionClosed {
    // Handle closed connection
}

// Or with errors.Is()
if errors.Is(err, cdp.ErrTimeout) {
    // Handle timeout
}
```

**Why**: Enables error comparison, clear error types, reusable.

**Where Used**:
- CDP errors
- Launcher errors
- Wait errors

---

## Summary

These patterns are used consistently throughout Kexas to ensure:
- **Reliability**: Fail-fast, explicit types, proper cleanup
- **Concurrency**: Thread-safe, atomic operations, goroutines
- **Maintainability**: Clear code, guard clauses, structured logging
- **Usability**: Clean API, optional parameters, good defaults

See `CODING_RULES.md` for enforcement of these patterns.

