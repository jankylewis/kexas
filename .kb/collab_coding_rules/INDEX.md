# Kexas Go Coding Conventions

Strict rules for Go code in this project. Established 2026-05-02.

These supplement (do not replace) `.kb/codebase/infras/DESIGN_PATTERNS.md`, which already documents 16 design patterns: explicit types fail-fast, context propagation, defer cleanup, error wrapping with `%w`, goroutine-safe handlers, atomic operations, options-as-struct, variadic optional params, guard clauses, type aliases, structured logging, sentinel errors, objectId-first, retry-with-timeout, dual guard, eager-vs-lazy agent enablement.

The five rules in this folder formalize stricter requirements not fully covered there.

## Rules

| # | Rule | File |
|---|---|---|
| 01 | File size limits — 230 lines source / 350 lines test | [`01_FILE_SIZE.md`](01_FILE_SIZE.md) |
| 02 | Function size — 40 logical lines max per function | [`02_FUNCTION_SIZE.md`](02_FUNCTION_SIZE.md) |
| 03 | Explicit types — no `:=` except in 4 init/range constructs | [`03_EXPLICIT_TYPES.md`](03_EXPLICIT_TYPES.md) |
| 04 | Acronym naming — true acronyms ALL CAPS, abbreviations camelCase | [`04_NAMING_ACRONYMS.md`](04_NAMING_ACRONYMS.md) |
| 05 | Package structure — `kcore/` with root re-export shims | [`05_KCORE_STRUCTURE.md`](05_KCORE_STRUCTURE.md) |
| 06 | Markdown files — <200 lines, ALL_UPPERCASED stem | [`06_MARKDOWN_FILES.md`](06_MARKDOWN_FILES.md) |
| 07 | KB live-sync — memory writes mirror to `.kb/` same turn | [`07_KB_LIVE_SYNC.md`](07_KB_LIVE_SYNC.md) |

## Enforcement

These rules are enforced by review, not yet by tooling. Future:

- `gofmt` and `go vet` are baseline (independent of these rules)
- A `golangci-lint` config could enforce 230/350 line limits and the explicit-types preference
- Until automated, judgement + reviewer flag

## Scope

Rules apply to code under `github.com/jankylewis/kexas`. They are deliberately stricter than idiomatic Go in places — particularly Rule 03 (explicit types). Why: fail-fast culture and fast readability. The cost is verbosity; we accept it consciously.

## Order of authority

When a conflict arises:

1. **These rules** (specific, project-wide)
2. **`.kb/codebase/infras/DESIGN_PATTERNS.md`** (architectural patterns)
3. **Effective Go / standard Go community conventions** (default)

Standard Go applies wherever this folder and DESIGN_PATTERNS are silent.
