# Key Concepts

> Important concepts to understand when working with Kexas

**Last Updated:** February 24, 2026

---

## 1. Flat Protocol

**What**: Modern CDP routing mechanism where `sessionId` is a top-level field.

**Before (Legacy)**:
```json
{
  "id": 1,
  "method": "Target.sendMessageToTarget",
  "params": {
    "sessionId": "ABC123",
    "message": "{\"id\":2,\"method\":\"Page.navigate\",\"params\":{\"url\":\"...\"}}"
  }
}
```

**After (Flat)**:
```json
{
  "id": 1,
  "method": "Page.navigate",
  "params": {"url": "..."},
  "sessionId": "ABC123"
}
```

**Why**: 
- Simpler structure (no nested JSON strings)
- Faster (less parsing overhead)
- Less error-prone
- Used by modern tools (Playwright, Puppeteer)

**Implementation**:
```go
type Request struct {
    ID        int64                  `json:"id"`
    Method    string                 `json:"method"`
    Params    map[string]interface{} `json:"params,omitempty"`
    SessionID string                 `json:"sessionId,omitempty"` // Flat protocol
}
```

---

## 2. Target vs Session

### Target

**What**: A browser entity (page, worker, service worker, iframe).

**Properties**:
- Identified by `targetId` (string)
- Created with `Target.createTarget`
- Closed with `Target.closeTarget`
- Has type: "page", "background_page", "service_worker", etc.

**Example**:
```go
// Create new target (tab)
result, err := client.Send(ctx, "Target.createTarget", map[string]interface{}{
    "url": "about:blank",
})
targetID := result["targetId"].(string)
```

### Session

**What**: A CDP communication channel to a target.

**Properties**:
- Identified by `sessionId` (string)
- Created with `Target.attachToTarget`
- Used to send commands to specific target
- Enables flat protocol routing

**Example**:
```go
// Attach to target to get session
result, err := client.Send(ctx, "Target.attachToTarget", map[string]interface{}{
    "targetId": targetID,
    "flatten":  true,  // Enable flat protocol
})
sessionID := result["sessionId"].(string)

// Send command to session
result, err = client.SendToSession(ctx, sessionID, "Page.navigate", params)
```

### Relationship

- One target can have multiple sessions (rare, usually one)
- One session belongs to exactly one target
- Session is the communication channel, target is the entity

**Analogy**: 
- Target = Phone number
- Session = Active call to that number

---

## 3. Temporary Profiles

**What**: Each browser launch creates a fresh temporary profile directory.

**Path**: `/tmp/kexas-chrome-{timestamp}/`

**Example**: `/tmp/kexas-chrome-1708790400123456789/`

**Why**:
- **Complete isolation** between test runs
- No cookies, history, cache, or state
- Prevents test interference
- Clean slate for every test
- No need to manually clear data

**Lifecycle**:
```
Launch → Create /tmp/kexas-chrome-{timestamp}/
    ↓
Browser runs with this profile
    ↓
Close → Delete /tmp/kexas-chrome-{timestamp}/
```

**Implementation**:
```go
func buildArgs(opts *Options) ([]string, string) {
    var timestamp string = fmt.Sprintf("%d", time.Now().UnixNano())
    var userDataDir string = filepath.Join(os.TempDir(), 
        fmt.Sprintf("kexas-chrome-%s", timestamp))
    
    var args []string = []string{
        fmt.Sprintf("--user-data-dir=%s", userDataDir),
        // ... other args
    }
    
    return args, userDataDir
}
```

**Cleanup**:
```go
func (b *Browser) Close() error {
    // ... close browser process
    
    if b.userDataDir != "" {
        os.RemoveAll(b.userDataDir)  // Delete temp profile
    }
    
    return nil
}
```

---

## 4. Wait Strategies

**Philosophy**: Different scenarios need different wait conditions.

### WaitUntilCommit

**When**: Response headers received, navigation committed.

**State**: Page is still blank, no content loaded.

**Use Case**: 
- Fast redirects
- When you only care about navigation starting
- Checking if URL is reachable

**Speed**: Fastest (milliseconds)

**Example**:
```go
page.Navigate(url, kwait.WaitUntilCommit)
```

---

### WaitUntilDOMContentLoaded

**When**: HTML parsed, DOM tree built.

**State**: DOM ready, but images/CSS/async scripts may not be loaded.

**Use Case**:
- Interacting with DOM elements
- When visual appearance doesn't matter
- Fast tests that don't need images

**Speed**: Fast (hundreds of milliseconds)

**Example**:
```go
page.Navigate(url, kwait.WaitUntilDOMContentLoaded)
```

---

### WaitUntilLoad (DEFAULT)

**When**: All resources loaded (images, CSS, JS, iframes).

**State**: Page fully loaded, visually complete.

**Use Case**:
- Visual testing
- Screenshots
- Most reliable for general use
- Default behavior

**Speed**: Medium (1-3 seconds typical)

**Example**:
```go
page.Navigate(url)  // Uses WaitUntilLoad by default
page.Navigate(url, kwait.WaitUntilLoad)  // Explicit
```

---

### WaitUntilNetworkIdle

**When**: No network activity for 500ms.

**State**: All network requests completed, page settled.

**Use Case**:
- Single Page Applications (SPAs)
- Pages that fetch data after initial render
- Dynamic content loading
- Most reliable but slowest

**Speed**: Slowest (3-10 seconds typical)

**Example**:
```go
page.Navigate(url, kwait.WaitUntilNetworkIdle)
```

---

### Choosing a Strategy

```
Fast ←─────────────────────────────────────→ Reliable

Commit → DOMContentLoaded → Load → NetworkIdle

Use Commit:          Redirects, URL checks
Use DOMContentLoaded: DOM manipulation, fast tests
Use Load:            Most cases (DEFAULT)
Use NetworkIdle:     SPAs, dynamic content
```

---

## 5. Chrome DevTools Protocol (CDP)

**What**: Binary protocol for communicating with Chromium-based browsers.

**Transport**: WebSocket (JSON messages over TCP)

**Message Types**:

1. **Command** (Request): Client → Browser
   ```json
   {
     "id": 1,
     "method": "Page.navigate",
     "params": {"url": "https://example.com"}
   }
   ```

2. **Response**: Browser → Client
   ```json
   {
     "id": 1,
     "result": {"frameId": "..."}
   }
   ```

3. **Event**: Browser → Client (no request)
   ```json
   {
     "method": "Page.loadEventFired",
     "params": {"timestamp": 123456}
   }
   ```

**Domains**: CDP is organized into domains (Page, Network, DOM, Runtime, etc.)

**Documentation**: https://chromedevtools.github.io/devtools-protocol/

---

## 6. Context Cancellation

**What**: Go's mechanism for canceling operations and propagating cancellation.

**Why**:
- Graceful shutdown
- Timeout handling
- Resource cleanup
- Cancellation propagation

**Pattern**:
```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Pass context through call chain
browser, err := kexas.LaunchWithContext(ctx, opts)
page, err := browser.NewPage()
err = page.Navigate(url)  // Uses browser's context

// If timeout expires, all operations are canceled
```

**Checking Cancellation**:
```go
select {
case result := <-resultChan:
    return result, nil
case <-ctx.Done():
    return nil, ctx.Err()  // context.DeadlineExceeded or context.Canceled
}
```

---

## 7. Goroutines and Channels

**Goroutines**: Lightweight threads managed by Go runtime.

**Use Cases in Kexas**:
- CDP message reading (readLoop)
- Event handler execution
- Concurrent operations

**Example**:
```go
// Start background goroutine
go client.readLoop()

// Execute handlers concurrently
for _, handler := range handlers {
    go handler(event.Params)
}
```

**Channels**: Communication between goroutines.

**Use Cases in Kexas**:
- Command-response matching
- Event notification
- Synchronization

**Example**:
```go
// Create channel for response
respChan := make(chan *Response, 1)
pending[id] = respChan

// Wait for response
select {
case resp := <-respChan:
    return resp.Result, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

---

## 8. Atomic Operations

**What**: Lock-free operations on shared variables.

**Why**:
- High performance (no mutex overhead)
- Prevents race conditions
- Simple for counters and flags

**Use Cases in Kexas**:
- Command ID generation
- Connection state (open/closed)
- Concurrent counters

**Example**:
```go
type Client struct {
    nextID atomic.Int64
    closed atomic.Bool
}

// Generate unique ID (thread-safe)
id := c.nextID.Add(1)

// Check if closed (thread-safe)
if c.closed.Load() {
    return ErrConnectionClosed
}

// Close once (thread-safe)
if !c.closed.CompareAndSwap(false, true) {
    return nil // Already closed
}
```

---

## 9. Error Wrapping

**What**: Adding context to errors while preserving the original error.

**Why**:
- Provides context at each layer
- Preserves error chain
- Enables `errors.Is()` and `errors.As()`

**Pattern**:
```go
func (p *Page) Navigate(url string) error {
    result, err := p.sendCommand("Page.navigate", params)
    if err != nil {
        return fmt.Errorf("kexas: navigation failed: %w", err)
    }
    return nil
}
```

**Unwrapping**:
```go
// Check if error is specific type
if errors.Is(err, cdp.ErrTimeout) {
    // Handle timeout
}

// Extract specific error type
var navErr *NavigationError
if errors.As(err, &navErr) {
    fmt.Println("Failed to navigate to:", navErr.URL)
}
```

---

## 10. Internal Packages

**What**: Go packages in `internal/` directory.

**Rule**: Cannot be imported by external packages.

**Why**:
- Hide implementation details
- Prevent external dependencies on internals
- Freedom to refactor without breaking users
- Clear API boundary

**Structure**:
```
kexas/
├── browser.go          # Public API
├── page.go             # Public API
└── internal/           # Private implementation
    ├── cdp/            # Cannot be imported externally
    ├── launcher/       # Cannot be imported externally
    └── ...
```

**Exposing Types**:
```go
// options.go - Public API
package kexas

import "github.com/jankylewis/kexas/internal/launcher"

// Type alias exposes internal type cleanly
type LaunchOptions = launcher.Options
```

---

## Summary

Understanding these concepts is essential for:
- Working with Kexas codebase
- Contributing features
- Debugging issues
- Understanding design decisions

Key takeaways:
- **Flat protocol** simplifies CDP communication
- **Temporary profiles** ensure test isolation
- **Wait strategies** balance speed and reliability
- **Context** enables cancellation and timeouts
- **Goroutines** enable concurrency
- **Internal packages** hide implementation

