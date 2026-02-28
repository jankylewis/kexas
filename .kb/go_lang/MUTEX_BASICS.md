# Mutex Basics in Go

> Reference sheet on `sync.Mutex`, locking patterns, and best practices.

---

## What is a Mutex?
- A **mutex (mutual exclusion lock)** ensures only one goroutine can access a critical section of code at a time.
- Provided by `sync.Mutex` (binary lock) and `sync.RWMutex` (read/write lock).
- Zero value is ready to use; do not copy mutexes after first use.

---

## Core API
```go
type Counter struct {
    mu sync.Mutex
    n  int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.n++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.n
}
```

### Rules
1. **Lock before shared state access**; unlock in `defer` to avoid leaks.
2. **Keep lock scope small**—do only the necessary work while locked.
3. **Never double-lock** the same mutex on the same goroutine.
4. **No copy**: store mutexes by pointer or struct, but once in use, the struct must not be copied.

---

## When to Use `sync.RWMutex`
- Many readers, few writers.
- `RLock()` allows concurrent readers; `Lock()` excludes everyone.
- Be careful to avoid reader/writer starvation—Go's RWMutex favors writers.

---

## Common Patterns
| Pattern | Description |
| --- | --- |
| Protect struct fields | Embed `mu sync.Mutex` and guard all exported methods. |
| Guard maps | Go maps are not thread-safe; wrap with a mutex or use `sync.Map`. |
| Condition variables | Combine `Mutex` + `sync.Cond` for producer/consumer coordination. |

---

## Deadlock Checklist
- Ensure every `Lock()` has a matching `Unlock()` even on error paths.
- Avoid lock ordering cycles (A locks B while B locks A).
- Prefer channel communication when possible; mutexes are for state protection.

---

## In kexas
- `ktestT` uses a mutex to guard `failed` state and logs so parallel tests don't interleave output.
- Reporter collector wraps its slice with a mutex to support concurrent worker writes.

Mutexes are simple but powerful—use them to guard shared state; pair with channels for higher-level coordination.
