# 05 — Package structure: `kcore/` with root re-export shims

## The rule

The browser-engine implementation lives in `kexas/kcore/`. The root `kexas` package contains **thin re-export shims** that preserve the public API path — `kexas.Browser`, `kexas.Launch()`, etc.

## Layout

```
kexas/
├── kexas.go               # version + package doc only
├── kexas_alpha_init.go    # AlphaInit (root-level entry point)
├── options.go             # LaunchOptions type alias to launcher.Options
│
├── browser.go             # re-exports: type Browser = kcore.Browser; func Launch(...) { ... }
├── page.go                # re-exports
├── element.go             # re-exports
├── cookie.go              # re-exports
├── storage.go             # re-exports
├── recorder.go            # re-exports
│
├── kcore/                 # implementation
│   ├── browser.go         # actual Browser struct + Launch + ...
│   ├── page.go            # actual Page + Navigate + ...
│   ├── element.go         # actual Element + ...
│   └── ... (all engine code)
│
├── internal/              # stays at project root (NOT under kcore/)
│   ├── cdp/
│   ├── agent/
│   └── logger/
│
├── ktest/, kassert/, kapi/, kwait/, launcher/, errors/    # supporting packages, unchanged
```

## Re-export pattern

### Types via aliases

```go
// kexas/browser.go
package kexas

import "github.com/jankylewis/kexas/kcore"

type Browser = kcore.Browser
type LaunchOptions = kcore.LaunchOptions
```

Type aliases make `kexas.Browser` and `kcore.Browser` the **same type** — no conversion, no wrapping, fully transparent.

### Functions via forwarding wrappers

```go
// kexas/browser.go
func Launch(opts *LaunchOptions) (*Browser, error) {
    return kcore.Launch(opts)
}

func LaunchWithContext(ctx context.Context, opts *LaunchOptions) (*Browser, error) {
    return kcore.LaunchWithContext(ctx, opts)
}
```

Forwarding wrappers (not `var Launch = kcore.Launch`) so `go doc kexas Launch` shows the signature properly.

## Why option B (and not A or C)

We chose option B over staying-at-root (A) and breaking-into-subpackages (C):

| Option | API impact | Visual structure | Cost |
|---|---|---|---|
| A (stay at root) | None | Flat 17 files | Idiomatic but cluttered |
| **B (kcore/ + shims)** ✅ | None | Clear engine boundary | One indirection layer |
| C (true sub-packages) | **Breaking** (`kbrowser.Launch()`) | Cleanest | Hurts ergonomics |

B preserves the `kexas.Launch()` ergonomics while giving `kcore/` as a visible "engine" boundary.

## What stays at root

These do not move into `kcore/`:

- `kexas.go` — version constant + package doc
- `kexas_alpha_init.go` — `AlphaInit` registration entry point (orthogonal to the engine; ties into ktest)
- `options.go` — `LaunchOptions` type alias to `launcher.Options`

These are API-surface helpers, not engine internals.

## What stays under `internal/`

`internal/cdp/`, `internal/agent/`, `internal/logger/` stay at the project root, **NOT** under `kcore/internal/`.

Reason: Go's `internal/` rule limits `kexas/kcore/internal/X` to importers under `kexas/kcore/*` only — which would block `launcher`, `ktest`, `kapi` from using these shared packages. Project-root `internal/` is the right scope.

## Tests

- **External tests** (`tests/kcore/*_test.go`, `package kcore_test`) — test through the public API via `import "github.com/jankylewis/kexas"`. Already in place after the test reorganization on 2026-05-02.
- **Internal / white-box tests** (testing unexported fields) — colocate with the implementation in `kcore/` as `kcore/<file>_test.go` with `package kcore`. Move with the file when refactoring an unexported method.

## Migration order (when executing the refactor)

Refactor in this order to keep tests green throughout:

1. Create `kexas/kcore/` directory.
2. For each file moving (e.g., `browser.go`):
   a. Move implementation file to `kcore/browser.go`, change package decl to `package kcore`.
   b. Create new root-level `browser.go` with re-export shims (type aliases + forwarding wrappers).
   c. Run tests — fix imports if needed.
3. Repeat for `page.go`, all `page_*.go`, `element.go`, all `element_*.go`, `cookie.go`, `storage.go`, `recorder.go`.
4. Internal tests (if any exist) move with their files.

Because `tests/kcore/` already imports `kexas` (root) and not `kcore` directly, no test changes needed during the move — the re-export shims keep the same public API.

## Audit checklist

- [ ] All engine files (browser/page/element/cookie/storage/recorder/page_*/element_*) live under `kcore/`
- [ ] Root-level files for those names exist as re-export shims (type aliases + forwarding wrappers)
- [ ] `kexas.go`, `kexas_alpha_init.go`, `options.go` remain at project root
- [ ] `internal/` remains at project root (not under `kcore/`)
- [ ] `tests/kcore/*` continues to import `kexas` (the root package), not `kcore` directly
