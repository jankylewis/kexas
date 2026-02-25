# Kexas

> High-performance browser automation for Go

Kexas is a modern browser automation library built for Go, leveraging the Chrome DevTools Protocol (CDP) for blazing-fast browser control with Go's powerful concurrency model.

## Features

- 🚀 **High Performance** — Direct CDP communication over WebSocket
- ⚡ **True Concurrency** — Leverage goroutines for massive parallelization
- 🎯 **Simple API** — Clean, idiomatic Go interface
- 🔧 **Headless & Headed** — Run with or without UI
- 📦 **Single Binary** — No runtime dependencies
- 🧪 **Zero-Boilerplate Testing** — AlphaInit pattern for instant test setup
- ✅ **Rich Assertions** — kassert library with fluent API
- 📸 **Screenshot Support** — Automatic screenshots on test failure
- ⚙️ **Configuration-Driven** — JSON config file support

## Installation

```bash
go get github.com/kexas-project/kexas
```

## Quick Start

### Basic Browser Automation

```go
package main

import (
    "log"
    "github.com/kexas-project/kexas"
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
    "github.com/kexas-project/kexas"
    "github.com/kexas-project/kexas/ktest"
    "github.com/kexas-project/kexas/kassert"
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

🚧 **Early Development** — v0.1.0-dev

### ✅ Currently Implemented
- Browser launcher (Chrome for Testing)
- CDP WebSocket client with flat protocol support
- Page navigation with wait strategies
- Playwright-standard wait mechanisms (commit, domcontentloaded, load, networkidle)
- Temporary browser profiles with auto-cleanup
- Structured logging system
- **Testing framework (ktest)** with AlphaInit zero-boilerplate pattern
- **Assertion library (kassert)** with fluent API
- **Configuration system** with JSON config file support
- **Screenshot functionality** with automatic capture on test failure
- **Sequential test execution** (one browser per test)
- **Interface compatibility** between ktest and kassert

### 🚧 Coming Soon
- Element selection and interaction
- Screenshots and PDFs
- Network interception
- Cookie management
- Multi-platform support (Linux, Windows)

## Architecture

```
kexas/
├── kexas.go              # Package documentation and version
├── kexas_alpha_init.go   # AlphaInit zero-boilerplate system
├── browser.go            # Browser management
├── page.go               # Page operations
├── element.go            # Element (placeholder)
├── locator.go            # Locator (placeholder)
├── options.go            # Launch options

├── internal/              # Framework internals
│   ├── cdp/               # CDP protocol client
│   ├── logger/            # Internal logging system
│   └── errors/            # Error types

├── ktest/                 # Public testing framework
├── kassert/               # Public assertion library
├── kwait/                 # Public wait strategies
├── launcher/              # Public browser launcher
├── errors/                # Public error types
├── tests/                 # Unit tests
└── .kb/                   # Knowledge base
```

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
    "github.com/kexas-project/kexas"
    "github.com/kexas-project/kexas/ktest"
    "github.com/kexas-project/kexas/kassert"
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

- **Sequential execution** - Tests run one by one
- **One browser per test** - Each test gets isolated browser instance
- **Automatic cleanup** - Browser closes after each test
- **Screenshot on failure** - Automatic capture when tests fail

## Wait Strategies

Kexas implements Playwright-standard wait strategies for page navigation:

```go
import "github.com/kexas-project/kexas/kwait"

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

## Development

```bash
# Run tests
go test ./tests/... -v

# Build library
go build .

# Run client example (if available)
cd ../amz && go run nav_tests.go
```

## Philosophy

Kexas is built on the belief that modern test automation infrastructure should be:
- **Performance-first** — Built for speed and scale
- **Cloud-native** — Designed for containerized environments
- **Developer-friendly** — Simple, predictable APIs
- **Zero-boilerplate** — Minimal setup for maximum productivity

## License

MIT

## Contributing

Contributions welcome! This project is in early development.
