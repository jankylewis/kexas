# Kexas Tests

Test files are organized by source-package mirror — each `tests/<X>/` folder contains tests for the corresponding `kexas/<X>/` source package or sub-system.

## Running Tests

```bash
# All unit tests across every subdir (default — fast, no Chrome needed)
go test ./tests/... -count=1

# All integration tests (requires Chrome). Use -p 1 for stability.
go test -tags=integration -p 1 ./tests/... -count=1

# Single-folder targeting
go test ./tests/kcore/ -count=1
go test -tags=integration ./tests/kcore/ -count=1

# Parallel with race detector
go test ./tests/... -count=1 -race
```

> **`-p 1` for integration runs:** when invoked via `./tests/...`, Go test runs each package binary in parallel by default. With 10 packages each launching Chrome, the parallel browser load can saturate system resources and hang. Use `-p 1` to serialize package execution — adds ~10s but eliminates the flake. Single-folder runs (`./tests/kcore/` etc.) don't need this.
>
> **`./tests/` (no `...`) doesn't work:** `tests/` itself contains no `.go` files now, only sub-packages. Always use `./tests/...` to recurse.

## Build Tags

| Tag | Description | Requires Chrome? |
|-----|-------------|-----------------|
| *(none)* | Unit tests — run by default | No |
| `integration` | Browser integration tests | Yes |

Files with `//go:build integration` at the top are excluded from default runs.

## Folder Layout

Each subdirectory is its own Go test package (`package <name>_test`).

| Folder | Files | Tests | Purpose |
|---|---|---|---|
| `kcore/` | 15 | 201 | Core engine — browser, page, element, browser-state (cookie/storage/multitab/recorder) |
| `ktest/` | 10 | 70 | Test framework + HTML report (config, suite, registration, parallel, html_*) |
| `kassert/` | 4 | 82 | Assertion library (basic/compare + extended_matchers/extended_stats) |
| `kapi/` | 6 | 44 | HTTP API client (helpers, client_construction, http_methods, request_builder, response, runtime) |
| `critical/` | 6 | 52 | Critical-path scenarios — selection / wait / unit, each split into basic+edge/timeout |
| `internal/` | 3 | 24 | Internal packages (agent, cdp, logger) |
| `kwait/` | 2 | 16 | Wait strategies (for + pageload) |
| `errors/` | 1 | 9 | Error types and sentinels |
| `launcher/` | 1 | 8 | Launcher integration (port, cleanup) |
| `config/` | 1 | 5 | Configuration loading + timeout |

**Total: 49 files, 511 tests.**

The `launcher/` source package itself also contains internal tests (`launcher/launcher_test.go`, `launcher/launcher_parallel_test.go`) that test unexported internals. Those are separate from `tests/launcher/`.

## File-size convention

Per project rules:

- **`.go` source files**: 230 lines max
- **`_test.go` test files**: 350 lines max
- When exceeded, split into multiple meaningful files within the same package (e.g., `kapi_test.go` → `helpers_test.go`, `client_construction_test.go`, `http_methods_test.go`, `request_builder_test.go`, `response_test.go`, `runtime_test.go`).

Same-package splits preserve cross-file mock/helper visibility (Go: identifiers in one file in `package X` are visible to all other files in the same `package X`), so type definitions and helpers placed in any file are usable across all files in that folder.
