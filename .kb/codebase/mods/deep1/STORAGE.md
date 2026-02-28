# Local/Session Storage — Deep Technical Guide (Student Level)

**File:** `kexas/storage.go`

---

## What Is Web Storage?

Web Storage is a browser API that lets websites store key-value pairs locally. There are two types:

| Type | Scope | Lifetime | Size Limit |
|------|-------|----------|------------|
| **localStorage** | Shared across all tabs on the same origin | Permanent (survives browser restart) | ~5 MB per origin |
| **sessionStorage** | Per-tab (each tab has its own copy) | Until tab is closed | ~5 MB per origin |

**What is an "origin"?** A combination of protocol + domain + port. For example, `https://example.com:443` is one origin. `http://example.com:80` is a different origin. Storage is isolated per origin — one site can never read another site's storage.

### FAQ: When would a test use storage?

- **Set auth tokens** — Many SPAs store JWT tokens in localStorage instead of cookies
- **Set feature flags** — `localStorage.setItem("feature_dark_mode", "true")`
- **Set user preferences** — Skip onboarding flows by pre-setting "has_seen_tutorial"
- **Assert state** — Verify the app stored the correct data after an action

---

## Technical Approach: JavaScript Evaluation

We use `page.Evaluate()` to run JavaScript like `localStorage.setItem('key', 'value')` rather than the CDP `DOMStorage` domain. Why?

1. **Simpler** — `DOMStorage.setDOMStorageItem` requires a `storageId` with a `securityOrigin` that's hard to obtain
2. **Reliable** — JS storage API is the same API the application uses
3. **Playwright does the same** — Playwright also uses JS evaluation for storage

### The Storage Struct

```go
type Storage struct {
    page        *Page       // The page to execute JS on
    storageType StorageType // "localStorage" or "sessionStorage"
}
```

Both `page.LocalStorage()` and `page.SessionStorage()` return the same `*Storage` struct — the only difference is the `storageType` field, which determines which JS global to target.

---

## JavaScript String Escaping

### The Problem

When building JS expressions like `localStorage.setItem('key', 'value')`, if the key or value contains a single quote (`'`), the JS breaks:

```
localStorage.setItem('it's', 'value')  // SYNTAX ERROR!
```

### The Solution: escapeJSString

The `escapeJSString` function replaces dangerous characters:

| Character | Escaped To | Why |
|-----------|------------|-----|
| `\` | `\\` | Backslash is the JS escape character itself |
| `'` | `\'` | Single quote would end the string literal |
| `"` | `\"` | Double quote could break nested strings |
| newline | `\n` | Literal newlines break JS string literals |
| carriage return | `\r` | Same reason |
| tab | `\t` | Consistency |

**Why `strings.NewReplacer`?** It's Go's efficient multi-string replacer. It builds an internal trie for O(n) replacement in a single pass.

---

## GetAll: The IIFE Pattern

`GetAll()` needs to iterate all keys and return them as a Go map. It builds a JavaScript IIFE (Immediately Invoked Function Expression):

```javascript
(function() {
    var result = {};
    for (var i = 0; i < localStorage.length; i++) {
        var key = localStorage.key(i);
        result[key] = localStorage.getItem(key);
    }
    return JSON.stringify(result);
})()
```

**What is an IIFE?** A function that's defined and immediately called: `(function() { ... })()`. It creates a local scope so variables don't leak into the page's global scope. The result is a JSON string that Go parses with `json.Unmarshal`.

### FAQ: Why JSON.stringify instead of returning the object directly?

CDP's `Runtime.evaluate` with `returnByValue: true` can return objects, but complex objects with non-string values can cause serialization issues. Returning a JSON string and parsing it in Go is more reliable and explicit.

---

## SetMany: Batch Performance

Instead of making N separate `Evaluate` calls (each requiring a WebSocket round-trip), `SetMany` builds a single JS script:

```javascript
(function() {
    localStorage.setItem('key1', 'val1');
    localStorage.setItem('key2', 'val2');
    localStorage.setItem('key3', 'val3');
    return true;
})()
```

One WebSocket round-trip instead of N. For 10 items, that's ~10x faster.

**Implementation detail:** Uses `strings.Builder` for efficient string concatenation in Go. `strings.Builder` uses an internal `[]byte` buffer and avoids the O(n²) string concatenation problem.

---

## CDP Number Types

CDP returns all numbers as `float64` in Go (because JSON has no integer type). The `Length()` method must convert:

```go
var length float64
length, ok = result.(float64)  // CDP returns 5.0, not 5
return int(length), nil         // Convert to int
```

### FAQ: Can this lose precision?

For storage length (small integers), no. `float64` can exactly represent all integers up to 2^53. Storage length will never be that large.

---

## Pitfalls to Avoid

1. **Cannot access storage on `about:blank`** — Browser blocks storage API on special URLs. Navigate to a real page first.
2. **5MB limit** — Browser throws `QuotaExceededError` if you exceed ~5MB. Tests rarely hit this, but be aware.
3. **sessionStorage is per-tab** — If you set something in tab 1's sessionStorage, tab 2 won't see it. localStorage IS shared across tabs.
4. **Clear only affects the target type** — `page.LocalStorage().Clear()` does NOT clear sessionStorage, and vice versa.
5. **Map iteration order in Go is random** — `SetMany` iterates a `map[string]string`, so the order of `setItem` calls is non-deterministic. This is fine because `setItem` is idempotent per key.
