# Multi-Tab Management — High-Level Overview

**Files:** `kexas/browser.go`, `kexas/page.go`
**CDP Domain:** Target, Page
**Status:** Implemented

---

## What It Does

Enables tests to open, track, switch between, and close multiple browser tabs within a single browser instance. Use cases include OAuth popups, multi-tab workflows, and `target="_blank"` link testing.

## API Surface

### Browser-Level

| Method | Description |
|--------|-------------|
| `browser.Pages()` | List all tracked pages (returns copy of slice) |
| `browser.PageCount()` | Number of tracked pages |
| `browser.PageByIndex(i)` | Get page by index (0-based) |
| `browser.PageByURL(pattern)` | Find page by URL substring match |
| `browser.WaitForNewPage(action, timeout)` | Execute action, wait for new tab to appear |
| `browser.CloseAllPagesExcept(page)` | Close all tabs except the specified one |

### Page-Level

| Method | Description |
|--------|-------------|
| `page.IsClosed()` | Whether this page has been closed |
| `page.BringToFront()` | Activate this tab (bring to foreground) |
| `page.Close()` | Close page, mark as closed, remove from browser tracking |

## Architecture

### Page Tracking

- Browser struct has `pages []*Page` slice protected by `sync.Mutex`
- `NewPage()` and `FirstPage()` append to `pages` after attaching
- `page.Close()` sets `closed=true` and calls `browser.removePage(page)`
- All page-list operations lock `pagesMu` for thread safety

### WaitForNewPage Pattern

1. Record current page count
2. Execute user-provided action (e.g., click a link)
3. Poll `Target.getTargets` every 100ms
4. When new target appears, find the unknown targetID
5. Attach to new target, add to pages list, return new `*Page`

## CDP Commands

- `Target.getTargets` — List all open targets
- `Target.createTarget` — Create new tab
- `Target.attachToTarget` — Attach CDP session to tab
- `Target.closeTarget` — Close a tab
- `Page.bringToFront` — Activate a tab

## Key Design Decisions

1. **Mutex-protected pages slice** — Thread-safe for parallel test workers
2. **Copy-on-read for Pages()** — Returns a copy to prevent concurrent modification
3. **PageByURL uses substring match** — Simple and sufficient for most cases; regex can be added later
4. **Polling-based WaitForNewPage** — CDP event subscription would be faster but requires more complex infrastructure; polling at 100ms is reliable enough
5. **Close is idempotent** — Calling Close() twice on a page is safe (returns nil on second call)
6. **removePage is O(n)** — Linear scan is fine for typical tab counts (2-10)
