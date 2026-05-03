# Samplings long-flows — 2026-05-02

Round 3 of samplings expansion: 3 multi-step "long" tests (~5s each) added to every project (sltests, wikitests, hntests, ghtests). 12 new tests; 1 real kexas bug found and fixed along the way.

## What got added

| Project | New tests | Theme |
|---|---|---|
| sltests | A. Login → add 3 → checkout → finish → back-to-products. B. Sort direction round-trip. C. Add-all-six → remove-all → cart empty. | Full lifecycle e2e + SortBy + AddItem/RemoveItem under repetition |
| wikitests | K. Search → click inline wikilink → assert title changed. L. Multi-section ScrollIntoView walk. M. Random→Random→Random chain (3 distinct titles). | Click-to-navigate, monotonic-scroll-position, redirect-following |
| hntests | K. Walk pagination 3 pages (count=30 each). L. Navigate front/newest/ask/show all render. M. ScrollIntoView .morelink → Click → URL contains p=2. | Full-reload navigation chain |
| ghtests | K. /issues → click first issue → URL has /issues/<id>. L. /pulls → click "Closed" → URL has is%3Aclosed. M. Multi-tab walk (Code, Issues, PRs, Actions) with header presence per tab. | SPA route changes + multi-nav under recorder pressure |

## Bug found and fixed

### `WaitForLoadState(DOMContentLoaded)` returns immediately on SPA navigation

**Surfaced by:** ghtests TestK + TestL.

**Symptom:** after `Element.Click()` on a GitHub issue/filter link, calling `WaitForLoadState(WaitUntilDOMContentLoaded, 15s)` returned without error, but `Page.URL()` immediately afterwards still showed the *previous* URL. The "wait" did nothing.

**Root cause:** `kwait.ForPageLoad` doesn't wait for `load` or `DOMContentLoaded` events. It polls the page title and returns the moment `title != "New Tab"` — which is satisfied the instant the *original* page loads, before any subsequent click. So the second-and-Nth waits return immediately on a tab whose title already exists. Code in `kwait/kwait.go:118-129`:

```go
case WaitUntilDOMContentLoaded:
    return For(ctx, func() (bool, error) {
        title, _ := getTitleFn()
        return title != "New Tab", nil  // ← satisfied forever after first load
    }, opts)
```

`WaitForLoad(timeout)` polls `document.readyState`, which has the same problem — readyState is "complete" the entire time after the initial load and doesn't reset on SPA navigation.

**Fix:** added `Page.WaitForURLContains(substr, timeout)` in `kcore/page_wait.go`. Polls `Page.URL()` directly until it contains the expected substring or times out. Reliable signal for any navigation regardless of mechanism (full reload, Turbo frame, React Router, history.pushState).

```go
func (p *Page) WaitForURLContains(substr string, timeout time.Duration) error
```

Two-line change in the failing tests: replace `WaitForLoadState(...)` with `WaitForURLContains(expected, timeout)`. Both pass.

**Why this surfaced now:** GitHub's `/issues` and `/pulls` filtering went Turbo-only sometime in 2024–25. HN still uses full reloads (its TestK pagination walk passed using the broken `WaitForLoadState` because the title actually does change between pages). Wikipedia + saucedemo also use full reloads. So the bug was latent until the gh tests exercised real SPA paths.

## Other findings

### `WaitForLoadState`/`WaitForLoad` are misnamed

The deeper issue is that **none of the existing wait-for-load primitives actually wait for load events.** They poll surrogate signals (title, readyState) that don't track SPA navigation at all. The names suggest event-based waiters; the implementation is title-polling.

Worth a future cleanup pass:
- Rename `WaitForLoadState` → `WaitForTitleSet` (or just delete in favour of WaitForURLContains + WaitForElementVisible).
- Rename `WaitForLoad` → `WaitForReadyStateComplete`.
- Add real CDP-event-based waiters using `Page.frameStoppedLoading` or `Page.lifecycleEvent` events for the case where you genuinely need a full-reload signal.

### `Click` on Turbo links works without explicit form-submit

Confirmed that `Element.Click()` on GitHub's Turbo-driven anchor links does fire the necessary event sequence to trigger the SPA route change. The bug was only on the wait side, not the click side. Good news: kexas doesn't need a separate "click and wait for SPA nav" combo method.

### Long flows surfaced no other kexas issues

Sort-direction round-trip, add-all-then-remove-all, multi-section scroll walks, hyperlink-chain navigation, random-to-random chains — all worked first try after the WaitForURLContains fix. No flakes across multiple runs.

## Test counts after this batch

| Project | Tests | Runtime |
|---|---|---|
| wikitests | 13 | ~15s |
| hntests | 13 | ~12s |
| ghtests | 13 | ~21s |
| sltests | 53 (50 prior + 3 flows) | ~95s across 6 packages |
| **total** | **92** | **~143s** |

All green, parallel + headless. Long-flow test runtimes range 0.6s (Wikipedia ScrollIntoView walk — tight DOM ops) to 9s (saucedemo full e2e checkout).

## Recommendations

1. **Prioritise the WaitForLoadState/WaitForLoad rename + deprecation.** Anyone migrating from Playwright will reach for `WaitForLoadState` first; current behavior is silently wrong for any SPA.
2. **Document `WaitForURLContains` in the README** as the recommended wait primitive after `Element.Click` triggers navigation.
3. **Consider event-based load detection.** CDP `Page.lifecycleEvent` is the right way; would allow proper `WaitForLoad`/`WaitForLoadState` semantics.
