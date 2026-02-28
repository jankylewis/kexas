# Kexas Naming Philosophy — Alphabetical Over Numerical

> Date: 2026-02-28
> Status: Active design principle

---

## Core Principle

Kexas uses **alphabetical ordering (A B C)** instead of **numerical ordering (1 2 3)**
throughout its API design. This applies to priorities, initialization phases, and
any future ordering concepts.

### Why Alphabetical?

1. **Natural language** — "Priority A" reads better than "Priority 1"
2. **Semantic meaning** — Alpha/Beta/Gamma carry lifecycle semantics
3. **Brand consistency** — `kexas.AlphaInit` established this pattern
4. **IDE discoverability** — `ktest.Priority.` triggers autocomplete for A–E

---

## Applied In Practice

### Test Priorities

```go
// ✅ Preferred — alphabetical via Priority struct
ktest.Test("Login", fn).WithPriority(ktest.Priority.A)  // runs first
ktest.Test("Cart", fn).WithPriority(ktest.Priority.B)   // runs second
ktest.Test("Cleanup", fn).WithPriority(ktest.Priority.E) // runs last

// ❌ Deprecated — numerical naming
ktest.TestWithPriority("Login", ktest.PriorityHigh, fn)
```

### Initialization Phases

```go
// AlphaInit — the first/primary initialization phase
var _ = kexas.AlphaInit(
    ktest.Test("TestLogin", fn),
)
```

Alpha (A) = first phase. Future phases could be BetaInit, GammaInit if needed.

---

## Priority Mapping (A–Z)

All 26 letters are available. Common usage:

| Priority | Value | Typical Usage |
|----------|-------|---------------|
| `Priority.A` | 0 | Critical path / smoke tests |
| `Priority.B` | 1 | Normal (default for new tests) |
| `Priority.C` | 2 | Lower / URL checks |
| `Priority.D` | 3 | Background |
| `Priority.E` | 4 | Deferred |
| `Priority.F`–`Priority.Z` | 5–25 | Available for fine-grained ordering |

Default priority for tests without `.WithPriority()` is `Priority.B` (normal).

### Legacy Aliases (Deprecated)

| Old | New |
|-----|-----|
| `PriorityHigh` | `Priority.A` |
| `PriorityNormal` | `Priority.B` |
| `PriorityLow` | `Priority.C` |

---

## Sort Behavior

Tests are sorted by priority (A first), then by **registration order** within
the same priority level. This is a stable sort — tests registered earlier run
earlier when priorities are equal.
