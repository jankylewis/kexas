# Kexas Infrastructure Overview

> High-level architecture and design principles

**Last Updated:** February 24, 2026

---

## Architecture Overview

Kexas is a **high-performance browser automation library** built on three foundational pillars:

```
┌─────────────────────────────────────────────────────────┐
│                    User Application                      │
│              (Test code, automation scripts)             │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   Kexas Public API                       │
│         Browser, Page, Element (future), Locator         │
└─────────────────────────────────────────────────────────┘
                            │
                ┌───────────┼───────────┐
                ▼           ▼           ▼
        ┌──────────┐  ┌─────────┐  ┌────────┐
        │   CDP    │  │ Launcher│  │  Wait  │
        │  Client  │  │ Process │  │ Logic  │
        └──────────┘  └─────────┘  └────────┘
                │           │
                ▼           ▼
        ┌──────────────────────────┐
        │   Chrome/Chromium        │
        │   (Browser Process)      │
        └──────────────────────────┘
```

---

## Key Design Principles

1. **Separation of Concerns**: Public API (Browser, Page) vs Internal implementation (CDP, Launcher)
2. **Fail-Fast**: Explicit types, immediate error handling, no silent failures
3. **Concurrency-First**: Built for Go's goroutines and channels
4. **Protocol-Driven**: Direct CDP communication over WebSocket (no middleware)
5. **Minimal Dependencies**: Only essential external packages

---

## Project Structure

```
kexas/
├── kexas.go              # Package entry point, version constant
├── browser.go            # Browser type and lifecycle management
├── page.go               # Page type and navigation/interaction
├── element.go            # Element type (placeholder, Phase 6)
├── locator.go            # Locator type (placeholder, Phase 6)
├── options.go            # Launch options (type alias)
│
├── internal/             # Internal packages (not exposed to users)
│   ├── cdp/              # Chrome DevTools Protocol client
│   ├── launcher/         # Browser process launcher
│   ├── kwait/            # Wait strategies and polling
│   ├── ktest/            # Testing framework
│   ├── kassert/          # Assertion library
│   ├── logger/           # Structured logging
│   └── errors/           # Error types and handling
│
├── cmd/                  # CLI tools
│   └── kexas/            # Main CLI entry point
│
├── tests/                # Unit and integration tests
│   ├── browser_test.go
│   ├── page_test.go
│   ├── cdp_test.go
│   └── ...
│
├── examples/             # Usage examples
│   ├── basic_usage.go
│   └── ...
│
├── .kb/                  # Knowledge base (documentation)
│   ├── browsers/         # Browser-specific docs
│   ├── codebase/         # Codebase infrastructure docs
│   └── ...
│
├── .tasks/               # Task tracking and planning
│   └── 02242026/         # Daily task logs
│
└── .rules/               # Coding rules and standards
    └── CODING_RULES.md
```

---

## File Organization Rules

- **Root level**: Public API only (Browser, Page, Element, Locator, Options)
- **internal/**: Implementation details, not accessible to users
- **tests/**: All unit tests, organized by package (`<package>_test.go`)
- **examples/**: Runnable examples for documentation
- **.kb/**: Knowledge base for documentation and guides
- **.tasks/**: Project management and progress tracking

---

## Documentation Structure

- `OVERVIEW.md` - This file, high-level architecture
- `CORE_COMPONENTS.md` - Detailed component documentation
- `DESIGN_PATTERNS.md` - Common patterns and best practices
- `DATA_FLOW.md` - How data flows through the system
- `KEY_CONCEPTS.md` - Important concepts to understand
- `INTERNAL_PACKAGES.md` - Internal package documentation
- `TESTING.md` - Testing strategy and conventions
- `DEVELOPMENT.md` - Development workflow and guidelines

---

## Quick Start for Contributors

1. Read `CODING_RULES.md` in `.rules/` directory
2. Review `CORE_COMPONENTS.md` to understand main types
3. Check `DESIGN_PATTERNS.md` for coding patterns
4. Explore `examples/` for usage patterns
5. Look at `tests/` for test patterns
6. Check `.tasks/` for current work

---

## Next Steps

- **New to Kexas?** Start with `CORE_COMPONENTS.md`
- **Want to contribute?** Read `DEVELOPMENT.md`
- **Need to understand flow?** Check `DATA_FLOW.md`
- **Looking for patterns?** See `DESIGN_PATTERNS.md`

