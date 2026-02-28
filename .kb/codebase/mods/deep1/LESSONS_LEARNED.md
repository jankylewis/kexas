# Lessons Learned — New Feature Implementation

**Scope:** Cookie Management, Storage, Multi-Tab, Video Recording, KAPI
**Date:** February 28, 2026

---

## 1. Strategy & Approach

### Feature-First, Test-Second, Docs-Third

The implementation order was:
1. **Foundation** — Add CDP constants and sentinel errors first (shared across all features)
2. **Implementation** — Build each feature file
3. **Integration** — Wire into existing systems (ensureAgentsForCommand, page tracking)
4. **Tests** — Write pure unit tests (struct tests, sentinel tests) + integration test stubs
5. **Documentation** — Three levels: high-level, deep-technical, non-technical

This order minimizes rework. Foundation changes are needed by all features, so doing them first avoids touching `const.go` and `errors.go` repeatedly.

### Modular File Structure

Each feature gets its own file:
- `cookie.go` — Cookie CRUD
- `storage.go` — localStorage/sessionStorage
- `recorder.go` — Video recording
- `kapi/flow.go` + `kapi/patterns.go` — Fluent API

Multi-tab went into `browser.go` (browser-level) and `page.go` (page-level) because it's an enhancement of existing types, not a new domain.

**Lesson:** New domains → new files. Enhancements to existing types → modify existing files.

---

## 2. Key Design Patterns Used

### Pattern: Agent Manager Auto-Enable

**Problem:** Every CDP domain (Network, Page, DOM, Runtime) must be "enabled" before use. Forgetting to enable causes silent failures.

**Solution:** `ensureAgentsForCommand()` in `page.go` maps command strings to required agents. When `sendCommand` is called, the agent is auto-enabled if needed.

**Applied to:** Cookie management (Network agent), Screencast (Page agent)

**Lesson:** Auto-enablement is superior to manual `Enable()` calls. Users never forget, and the overhead of double-enabling is zero (Agent Manager caches state).

### Pattern: Sentinel Errors

**Problem:** Need to distinguish between different error types (cookie not found vs cookie set failed vs empty name).

**Solution:** Pre-defined error variables in `errors/errors.go` that callers check with `errors.Is()`.

**Applied to:** All features — 13 new sentinel errors added.

**Lesson:** Define sentinels early (before implementation). It clarifies the error surface and makes test assertions cleaner.

### Pattern: Mutex-Protected Shared State

**Problem:** Browser's page list is accessed from multiple goroutines (parallel test workers).

**Solution:** `sync.Mutex` protecting the `pages` slice. Copy-on-read for iteration.

**Lesson:** Always copy slices before iterating if you need to call methods that might block (like CDP round-trips). Holding a lock during I/O is a concurrency antipattern.

### Pattern: Fluent Error Accumulation (KAPI)

**Problem:** Go's `(result, error)` convention breaks method chaining.

**Solution:** Errors stored internally, checked at the end with `Err()`.

**Lesson:** This trades Go's conventional error handling for readability. Acceptable for test automation, not for production code.

### Pattern: Graceful Degradation (Recorder)

**Problem:** ffmpeg is an external dependency that may not be installed.

**Solution:** Runtime detection via `exec.LookPath`. If missing, fall back to saving raw frames.

**Lesson:** Never make external tools a hard dependency. Always provide a fallback or clear error message.

---

## 3. Errors & Pitfalls to Avoid

### Pitfall: CDP Domain Not Enabled

**Symptom:** Cookie commands silently fail or return empty results.
**Root Cause:** Network domain not enabled before sending cookie commands.
**Fix:** Added Network → `ensureAgentsForCommand` mapping.
**Prevention:** Always check if a new CDP command needs a domain enabled.

### Pitfall: JavaScript Injection in Storage

**Symptom:** Storage set/get fails for values containing quotes.
**Root Cause:** Unescaped quotes break JS string literals.
**Fix:** `escapeJSString()` escapes all dangerous characters.
**Prevention:** Never concatenate user input into JS without escaping.

### Pitfall: Nil Page Panics in KAPI

**Symptom:** Nil pointer dereference when browser launch fails.
**Root Cause:** `Open()` fails, `flow.page` is nil, subsequent `Find()` dereferences it.
**Fix:** `hasBlockingError()` check at the start of every method.
**Prevention:** Always check for nil receiver/state before accessing fields.

### Pitfall: Data Races on Page List

**Symptom:** Race detector fires during parallel test execution.
**Root Cause:** Multiple goroutines accessing `browser.pages` without synchronization.
**Fix:** `sync.Mutex` on all page list operations.
**Prevention:** Any shared mutable state needs synchronization. Period.

### Pitfall: CDP Numbers Are float64

**Symptom:** Type assertion fails: `result.(int)` panics.
**Root Cause:** JSON numbers decode as `float64` in Go.
**Fix:** Always assert `float64` first, then convert to `int` if needed.
**Prevention:** Remember: JSON → Go type mapping is `number → float64`, always.

---

## 4. Best Practices

1. **Add CDP constants before implementation** — Ensures consistent command strings and catches typos at compile time
2. **Add sentinel errors before implementation** — Clarifies error surface upfront
3. **Wire agent auto-enable immediately** — Prevents "it works in tests but fails in real use" bugs
4. **Use explicit var declarations** — Kexas convention: `var x Type = value` (not `:=`) for clarity
5. **Test struct values with pure unit tests** — No browser needed to verify Cookie defaults or Config fields
6. **Mark integration tests with t.Skip** — Keeps `go test` fast; integration tests run separately
7. **Document at three levels** — High (API reference), Deep1 (how it works), Deep2 (what it does for non-engineers)
8. **Update task plans as you go** — Mark tasks completed immediately after finishing them
9. **Build after each feature** — `go build ./...` catches compilation errors early
10. **Run full test suite before moving on** — `go test ./...` ensures no regressions
