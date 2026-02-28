# Multi-Tab Management — Deep Technical Guide (Student Level)

**Files:** `kexas/browser.go`, `kexas/page.go`

---

## Why Multi-Tab?

Real web applications frequently open new tabs: OAuth login popups, `target="_blank"` links, payment gateways, PDF previews. A testing framework that can only control one tab cannot test these workflows.

### The CDP Target Model

In Chrome DevTools Protocol, every tab is a **Target**. Each target has:

| Field | Description |
|-------|-------------|
| **targetId** | Unique identifier (UUID string) for this target |
| **type** | What kind of target: `"page"`, `"background_page"`, `"service_worker"`, etc. |
| **url** | The current URL of the target |
| **attached** | Whether a CDP session is connected to this target |

To interact with a tab, you must **attach** to it — this creates a CDP **session** (identified by `sessionId`). All commands to that tab go through its session.

### FAQ: What is a CDP session?

When you attach to a target, Chrome creates a session — a logical channel within the single WebSocket connection. Each session has a unique `sessionId`. Commands sent with that sessionId go to that specific tab. This is how one WebSocket connection controls multiple tabs.

---

## Page Tracking Architecture

### The Problem

Before multi-tab support, `Browser` had no memory of pages it created. `NewPage()` returned a `*Page` and forgot about it. This made it impossible to enumerate open tabs or find a specific one.

### The Solution: pages Slice + Mutex

```go
type Browser struct {
    // ...existing fields...
    pages   []*Page
    pagesMu sync.Mutex
}
```

**Why sync.Mutex?** In parallel test execution, multiple goroutines (test workers) may call `NewPage()`, `Close()`, or `Pages()` simultaneously. Without a mutex, concurrent slice modifications would cause a **data race** — a bug where two goroutines read/write the same memory simultaneously, producing corrupted or inconsistent data.

**What is a data race?** Imagine two goroutines both trying to append to the same slice at the same time. Go slices are backed by an array with a length and capacity. If both goroutines read length=5 at the same time, both try to write to index 5, and only one write survives. Worse, the slice header might end up in an inconsistent state. Go's race detector (`go test -race`) catches these bugs.

### Lock Discipline

Every access to `b.pages` is protected:

```go
// Read: lock, copy, unlock, then iterate the copy
b.pagesMu.Lock()
var pageCopy []*Page = make([]*Page, len(b.pages))
copy(pageCopy, b.pages)
b.pagesMu.Unlock()
// Now safe to iterate pageCopy without holding the lock

// Write: lock, modify, unlock
b.pagesMu.Lock()
b.pages = append(b.pages, page)
b.pagesMu.Unlock()
```

**Why copy before iterating?** If you hold the lock while iterating and calling `page.URL()` (which does a CDP round-trip), you block all other goroutines from accessing the pages list for the entire duration of those network calls. Copying the slice and releasing the lock immediately is much better for concurrency.

### FAQ: Why sync.Mutex and not sync.RWMutex?

`sync.RWMutex` allows multiple concurrent readers but exclusive writers. For our case, page list operations are fast (microseconds) and writes (NewPage, Close) happen frequently. The overhead of RWMutex's more complex locking isn't worth it for such short critical sections. Simple Mutex is clearer and sufficient.

---

## WaitForNewPage: The Polling Pattern

### How It Works

1. **Snapshot** — Record current page count under lock
2. **Execute action** — Call the user's function (e.g., click a link that opens a new tab)
3. **Poll** — Every 100ms, call `Target.getTargets` and count page-type targets
4. **Detect** — When count exceeds snapshot, find the new targetId
5. **Attach** — Call `attachToPage(newTargetId)`, add to pages list, return

```go
func (b *Browser) WaitForNewPage(action func(), timeout time.Duration) (*Page, error) {
    // 1. Snapshot
    b.pagesMu.Lock()
    var beforeCount int = len(b.pages)
    b.pagesMu.Unlock()

    // 2. Action
    action()

    // 3-5. Poll loop
    for time.Since(start) < timeout {
        time.Sleep(100 * time.Millisecond)
        // ... check Target.getTargets for new pages ...
    }
}
```

### Why Polling Instead of Events?

CDP has `Target.targetCreated` events that fire when a new tab opens. Using events would be faster (instant notification vs 100ms polling delay). However, event subscription requires:
- Background goroutine listening for events on the WebSocket
- Event demultiplexing (routing events to the right handler)
- Callback registration and cleanup

This is significant infrastructure. Polling at 100ms is simple, reliable, and the 100ms delay is imperceptible in test execution. The event-driven approach can be added later as an optimization.

### FAQ: What if two tabs open simultaneously?

The current implementation returns the first new tab it finds. If two tabs open at once, only one is returned from `WaitForNewPage`. The second tab would need to be found via `PageByURL` or `PageByIndex`. This matches Playwright's behavior.

---

## Page.Close and Lifecycle

### The Close Sequence

1. Check `p.closed` flag — if already closed, return nil (idempotent)
2. Send `Page.close` CDP command — tells Chrome to close the tab
3. Set `p.closed = true` — mark as closed locally
4. Call `p.browser.removePage(p)` — remove from browser's tracked pages

### removePage: O(n) Slice Removal

```go
func (b *Browser) removePage(page *Page) {
    b.pagesMu.Lock()
    defer b.pagesMu.Unlock()
    for i := 0; i < len(b.pages); i++ {
        if b.pages[i] == page {
            b.pages = append(b.pages[:i], b.pages[i+1:]...)
            return
        }
    }
}
```

**What does `append(b.pages[:i], b.pages[i+1:]...)` do?** It creates a new slice that contains everything before index `i` and everything after index `i` — effectively removing the element at `i`. This is Go's standard pattern for removing an element from a slice.

**Why pointer comparison (`==`) and not targetId comparison?** A pointer comparison is faster and unambiguous. Two different `*Page` objects could theoretically have the same targetId if one was closed and a new tab reused the ID. Pointer comparison is always correct.

---

## Pitfalls to Avoid

1. **Don't forget to close tabs** — Orphan tabs consume memory. Use `CloseAllPagesExcept` or close tabs in test cleanup.
2. **Page ordering is not guaranteed** — `PageByIndex(1)` may not return the tab you expect if tabs were opened/closed out of order. Use `PageByURL` for reliable identification.
3. **WaitForNewPage timeout** — Set a reasonable timeout (5-10s). If the action doesn't open a new tab, you'll wait the full timeout before getting an error.
4. **Tab close propagation** — When Chrome closes a tab externally (e.g., user action, crash), the `page.closed` flag may not be updated. Use `IsClosed()` as a hint, not a guarantee.
