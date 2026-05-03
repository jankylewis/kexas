# 02 — Function size

## The rule

**Each function: 40 logical lines max.**

When exceeded, split into multiple smaller functions with meaningful names.

## How "logical lines" are counted

- **Count:** code statements and braces
- **Don't count:** blank lines, comment-only lines, embedded string literals (CSS / JS / HTML strings inside Go function bodies)

The intent is to limit function-body *logic*. A function whose body is mostly a long string literal isn't actually doing 100 things — it's setting up data.

## Why 40

- Easier to read end-to-end without scrolling
- Each function does one thing — a 100-line function usually does several
- Easier to test in isolation (smaller surface area)
- Forces extraction of helpers, which improves overall code organization

## How to split

When a function exceeds 40 lines, look for:

1. **Sequential phases** — "do A, then B, then C" → extract each phase as a helper function. Example: `Launch` (~120 lines) → split into `findChromiumOrFail`, `cleanStaleProfilesOnce`, `startChromeProcess`, `extractWebSocketURL`.
2. **Branching logic** — long switch/case blocks → extract per-case handlers.
3. **Setup / execution / teardown** — extract setup or teardown into separate functions.
4. **Repeated patterns** — if you see two similar 30-line blocks, extract their shared core.

## Embedded string-literal exemption

If a function has a long embedded string literal (e.g., 50+ lines of JavaScript injected via `Page.addScriptToEvaluateOnNewDocument`), the string itself doesn't count toward the 40-line cap. Move the string to a `const` or top-level `var` if it's reusable across functions; otherwise inline is fine.

Example: `attachToPage()` in `browser.go` has the 7-step stealth JS as an inline string literal — that's exempt. The Go logic *surrounding* the string must still fit within 50 lines.

## When NOT to split

- Don't split if it forces 5+ helper functions that only the original function calls (over-fragmentation).
- Don't split linear arrange-act-assert test functions when the 40-line limit isn't exceeded.
- Don't split for the sake of hitting the rule — the rule exists to improve readability, not to maximize the number of small functions.

## Existing offenders to fix during refactor

These are known violators (lines from before refactor, may have changed):

- `attachToPage` in `browser.go` (~100 lines, mostly stealth JS — embedded string is exempt; Go logic should still fit)
- `Launch` in `klauncher/launcher.go` (~120 lines — split into phase helpers)
- `runTestsSequentialRegistered` in `ktest/ktest_runner.go` (~80 lines)
- `executeSingleTest` in `ktest/ktest_parallel.go` (~70 lines)

## Audit checklist

- [ ] No function body exceeds 40 logical lines
- [ ] Embedded string literals not counted toward limit
- [ ] Split functions have meaningful names (not `helper1`, `helper2`)
- [ ] Helpers extracted are usable by future callers, not just one-off
