# Kexas Tests

## Running Tests

```bash
# Unit tests only (default — fast, no Chrome needed)
go test ./tests/ -count=1

# Integration tests only (requires Chrome)
go test -tags=integration ./tests/ -count=1

# All tests
go test -tags=integration ./tests/ -count=1

# Launcher tests (separate package)
go test ./launcher/ -count=1

# Parallel with race detector
go test ./tests/ -count=1 -race
```

## Build Tags

| Tag | Description | Requires Chrome? |
|-----|-------------|-----------------|
| *(none)* | Unit tests — run by default | No |
| `integration` | Browser integration tests | Yes |

## Test File Groups

### Core (browser, page, navigation)
- `alpha_init_test.go` — AlphaInit enforcement
- `browser_test.go` — Browser launch/close *(integration)*
- `page_test.go` — Page navigation, title, URL *(integration)*
- `page_find_test.go` — Page.Find selector logic
- `page_find_strict_test.go` — Strict find mode *(integration)*
- `page_wait_test.go` — WaitForElement* methods

### Element (interaction, scroll, hover)
- `element_test.go` — Element struct, IsVisible
- `element_enhanced_test.go` — Enhanced element methods
- `element_interaction_test.go` — Click, Type, Hover
- `scroll_test.go` — ScrollIntoView, page scroll

### Features (cookie, storage, multitab, recorder)
- `cookie_test.go` — Cookie management *(integration)*
- `storage_test.go` — Local/session storage *(integration)*
- `multitab_test.go` — Multi-tab management *(integration)*
- `recorder_test.go` — Video recording *(integration)*

### Infrastructure (agent, CDP, errors, config, logger)
- `agent_test.go` — Agent manager
- `cdp_test.go` — CDP command registry
- `config_timeout_test.go` — Timeout configuration
- `errors_test.go` — Error types and sentinels
- `logger_test.go` — Logger output
- `launcher_test.go` — Launcher utilities (port, cleanup)

### KAPI (HTTP API client)
- `kapi_test.go` — HTTP client, request builder, response parsing

### Test Framework (ktest, kassert, kwait)
- `ktest_test.go` — Test registration *(integration)*
- `ktest_parallel_test.go` — Parallel execution
- `kassert_test.go` — Assertion helpers
- `kassert_extended_test.go` — Extended assertions
- `kwait_test.go` — Wait utilities

### Report
- `report_test.go` — Test report generation
- `report_html_test.go` — HTML report output

### Critical
- `critical_unit_test.go` — Critical unit scenarios
- `critical_selection_test.go` — Critical selector scenarios
- `critical_wait_test.go` — Critical wait scenarios
