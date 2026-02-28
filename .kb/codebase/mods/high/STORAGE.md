# Local/Session Storage — High-Level Overview

**File:** `kexas/storage.go`
**CDP Domain:** None (uses Runtime.evaluate via page.Evaluate)
**Status:** Implemented

---

## What It Does

Provides type-safe access to browser localStorage and sessionStorage, enabling tests to:
- Pre-set application state (tokens, feature flags, preferences)
- Read storage values for assertion
- Clear storage between tests for isolation
- Bulk set/get operations

## API Surface

| Method | Description |
|--------|-------------|
| `page.LocalStorage()` | Returns `*Storage` bound to localStorage |
| `page.SessionStorage()` | Returns `*Storage` bound to sessionStorage |
| `storage.Set(key, value)` | Store a key-value pair |
| `storage.Get(key)` | Retrieve value by key (empty string if missing) |
| `storage.Remove(key)` | Delete a key |
| `storage.Clear()` | Remove all items |
| `storage.GetAll()` | Get all items as `map[string]string` |
| `storage.Length()` | Get number of stored items |
| `storage.Has(key)` | Check if key exists |
| `storage.SetMany(items)` | Bulk set from map in one JS call |

## Technical Approach

Uses JavaScript evaluation (`page.Evaluate`) rather than the CDP `DOMStorage` domain because:
1. `DOMStorage` requires a `securityOrigin` parameter that's complex to obtain
2. JS `localStorage.setItem()` is simpler and more reliable
3. Matches exactly what application code does

## Key Design Decisions

1. **Storage struct holds storageType field** — "localStorage" or "sessionStorage" — same struct, different JS target
2. **escapeJSString** — All keys and values are escaped before injection into JS string literals (handles quotes, newlines, backslashes)
3. **GetAll uses JSON.stringify** — Iterates all keys in JS, builds object, returns JSON string, parsed in Go via `json.Unmarshal`
4. **SetMany batches into single Evaluate** — Builds one IIFE with multiple `setItem` calls for performance
5. **Numbers returned as float64** — CDP returns all numbers as float64; `Length()` converts to int

## Pitfalls

- **Storage access fails on `about:blank`** — Must navigate to a real URL first
- **5MB browser storage limit** — Large values will hit browser limits
- **Session vs Local isolation** — sessionStorage is per-tab, localStorage is shared across tabs
