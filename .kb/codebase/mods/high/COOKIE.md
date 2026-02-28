# Cookie Management — High-Level Overview

**File:** `kexas/cookie.go`
**CDP Domain:** Network
**Status:** Implemented

---

## What It Does

Provides full cookie CRUD operations on browser pages, enabling tests to:
- Set authentication cookies to skip login flows
- Read cookies for assertion
- Delete specific cookies or clear all
- Use convenience helpers like `SetAuthCookie`, `HasCookie`, `GetCookieValue`

## API Surface

| Method | Description |
|--------|-------------|
| `page.GetCookies(urls...)` | Retrieve all cookies (optionally filtered by URL) |
| `page.SetCookie(cookie)` | Set a single cookie |
| `page.SetCookies(cookies)` | Set multiple cookies |
| `page.DeleteCookie(name, domain, path)` | Delete matching cookies |
| `page.ClearCookies()` | Delete ALL cookies |
| `page.SetAuthCookie(name, value, domain)` | Set auth cookie with secure defaults |
| `page.HasCookie(name)` | Check if cookie exists |
| `page.GetCookieValue(name)` | Get cookie value by name |

## CDP Commands

- `Network.getCookies` — Read cookies
- `Network.setCookie` — Set a cookie
- `Network.deleteCookies` — Delete cookies
- `Network.clearBrowserCookies` — Clear all

## Agent Manager Integration

Network agent is auto-enabled via `ensureAgentsForCommand()` in `page.go` when any cookie CDP command is sent. No manual `Network.enable` call needed.

## Key Design Decisions

1. **Cookie struct mirrors CDP fields** — Name, Value, Domain, Path, Expires, Size, HttpOnly, Secure, SameSite
2. **Name is required, Value is not** — Empty value is valid (used to clear cookies in some patterns)
3. **SetCookies loops over SetCookie** — CDP has no bulk set command
4. **SetAuthCookie uses secure defaults** — HttpOnly=true, Secure=true, SameSite=Lax, Path="/"
5. **parseCookie handles missing fields gracefully** — All fields default to zero values
