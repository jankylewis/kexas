# Kexas dogfood — chronological batch log (2026-05-02)

Companion to [`SLTESTS_DOGFOOD_FINDINGS.md`](SLTESTS_DOGFOOD_FINDINGS.md). Each batch is a discrete chunk of fixes shipped on the same day, in order. Read top-to-bottom for the chronological picture, or jump to a specific batch via section heading.


## Update — 2026-05-02 PM batch (3 fixes shipped)

### Bug #3 → fixed: `Element.Fill(text)` added to kexas

`kcore/element_typing.go` now has `Fill()` alongside `Type()`. Implementation uses the React-aware HTMLInputElement.value setter + bubbling 'input' + 'change' events. Single CDP roundtrip (vs N×3 for Type). sltests `pages/login_page.go` and `pages/checkout_info_page.go` switched from the local `setReactInputValue` workaround to `field.Fill(value)`. The `pages/react_input.go` workaround file was deleted. All sltests still pass.

Fill also accepts `""` (clear-the-field) where `Type("")` errored with `ErrTextEmpty` — useful for "blank out then re-fill" flows.

### Config-loading scope → fixed: walk-up + path anchoring

`ktest/ktest_config.go` added `findConfigUpward(startDir, filename, maxLevels=10)` which walks parents from CWD looking for `kexas.config.json`. `LoadConfigFromFile` anchors `ScreenshotDir / ReportDir / VideoDir` to the directory of the config file when those paths are relative. Effect: running `go test ./tests/auth/` from a deeply nested package directory now finds the project-root config and writes screenshots/reports back to `<project-root>/test-results/`, not `<project-root>/tests/auth/test-results/`.

### HTML report wiring → fixed: `runTestLifecycle` now generates report.html

`ktest/ktest_impl.go::runTestsSequential / runTestsParallel / runSingleTest` refactored to return `[]testResult`. New `suiteFilename(suiteValue)` returns the suite's Go type name (e.g., "MenuSuite") for per-suite grouping. `ktest/ktest_main.go::runTestLifecycle` collects results and calls `generateHTMLReport(results, config)` after AfterAll. Effect: `ktest.Run(t, suite)` (the legacy + recommended testing-package API) now gets the same `report.html` AlphaInit-registered tests get.

### Newly visible gap — per-package report overwrite

Each Go test package runs as a separate process. Each `ktest.Run` call writes its own `report.html`, **overwriting** any earlier one. After `go test ./...` only the LAST package's report survives at `report.html`; the timestamped archives (`report-YYYYMMDD-HHMMSS.html`) preserve every run. Two fix paths:
- **(a) Per-suite subdirs** — `test-results/<SuiteName>/report.html`. Cheap.
- **(b) Cross-process aggregator** — JSON sidecar per suite, then merge into one `report.html`. Matches user expectation.

### Updated v0.1.0 priority list

1. ~~CDP connection-close hang~~ ✅ fixed
2. ~~`Element.Type` silent no-op on React~~ ✅ fixed via `Element.Fill`
3. ~~Config not found from sub-package CWDs~~ ✅ fixed via walk-up + anchoring
4. ~~No HTML report for legacy `ktest.Run` path~~ ✅ fixed
5. **Per-package `report.html` overwrite** — newly headline issue
6. **No `Page.FindAll` / `Locator`** — every multi-element check still goes through `Page.Evaluate`
7. **`Page.Evaluate` typed-result helpers** wanted
8. **`kassert.That(slice).Contains`** — slice membership; sltests uses manual loop
9. **`Element.SelectOption`** for native `<select>`
10. **`WaitForElementInteractive`** — partially mitigated by `Fill`

---

## Update — 2026-05-02 PM batch 2 (report UX + always-on artifacts)

### Three more fixes shipped

- **Always-on per-test screenshot.** End-of-test screenshot for every test (was failure-only). One PNG per test at `config.ScreenshotDir/<TestName>.png`. The "screenshotOnFail" config flag now reads as "screenshot enable/disable" — kept the name for backwards compat. Each test overwrites its own file (no timestamps in screenshot names) so reports always reference the latest run.
- **Always-on per-test video.** Each test wraps its execution in `page.StartRecording()` → `recorder.Stop()` → `recorder.SaveVideo("<config.VideoDir>/<TestName>.mp4")`. Recording covers BeforeEach + test + AfterEach + retries. Video starts fresh per test (state from previous test's recorder is gone because we create a new one).
- **Report detail panel now meaningful for passed tests.** `writeTestArtifacts` renders `<img class="artifact-screenshot">` + `<video class="artifact-video">` after steps and before logs in every test's detail panel. Click-to-expand is now useful for passed tests too — previously the panel was empty without explicit `ktest.Step()` calls.

### Bonus fix

**Donut legend column alignment.** The "Outcome mix" insight card had Pass/Fail/Skip count + percentage misaligned (varying label widths pushed numbers to different x-positions). Fixed by giving the label `<span>` `flex: 1` and the `<strong>` (count) + `<em>` (percentage) fixed widths with `text-align: right`.

### Lessons

- **Asymmetric code paths are dogfood bait.** kexas had TWO test-runner paths (`runRegisteredTests` for AlphaInit, `runTestLifecycle` for `ktest.Run`). Only the first generated HTML reports. Internal kexas tests didn't catch this because they exercise both paths but never check whether the report.html actually appears at the expected location. **Lesson:** any time a feature exists on path A but not path B, write an integration test that checks BOTH paths produce the artifact.
- **CWD-relative paths are a footgun for sub-package tests.** Go's `go test ./tests/auth/` runs in cwd=`tests/auth/`, not project root. Anything kexas reads/writes via `./<path>` resolves against the test's CWD. The fix: walk-up to find config, anchor relative paths in config to where it was found. **Lesson:** for any user-facing config field that points at a directory, anchor it to a known stable root (config-file dir, project root, or absolute), not the test's CWD.
- **`ktest.Suite` shares ONE Page across all tests in a suite.** Per-test artifacts (screenshot + video + state reset) must be set up + torn down explicitly per test. Recording in particular: must `page.StartRecording()` per test (returns a fresh `*Recorder`) and `Stop()` + `SaveVideo()` before the next test starts. If you started one recorder for the whole suite, you'd get one video covering all tests with no per-test boundaries.
- **kexas Recorder gracefully degrades on missing ffmpeg — but degrades into a directory.** `SaveVideo("foo.mp4")` returns `"foo.mp4_frames"` (a JPEG dir) when ffmpeg isn't on PATH. Embedding that path in `<video src=...>` fails silently in the browser. Workaround in `writeTestArtifacts`: detect the `_frames` suffix and render a fallback link. **Better long-term fix:** `SaveVideo` should return a richer struct like `{Path string, IsVideo bool, FallbackReason string}` so consumers don't have to suffix-sniff.
- **"Always-on" artifacts make pass/fail UX symmetric.** Tiny user-facing change with disproportionate impact. Detail panels for passed tests went from empty (visually broken) to meaningful (screenshot + recording). This is the kind of polish that makes a framework feel finished — and it's nearly free in implementation.
- **The HTML report's CSS layout assumes label widths are stable.** Donut legend rows used `display: flex` + `margin-left: auto` on the count, which works until labels differ in width (Passed/Failed/Skipped did). Fixed widths on count + percentage columns is the standard table-layout-in-flex idiom. Watch out for similar latent issues in any flex layout where a center column has variable width.

### Open after this batch

- **Auto-step capture.** Tests that don't call `ktest.Step()` show no steps in the report — even though kexas page actions (Find/Click/Type/Fill/Navigate) are all loggable events. Auto-emitting a step per kexas page action would make passed test details show "what the test actually did" without requiring opt-in code. Feature ask carried over.
- **ffmpeg dependency for video.** Currently silent fallback to frames dir + report shows "Video unavailable" link. Could either (a) bundle ffmpeg detection + install instructions in launcher logs, (b) implement a CDP-screencast-event-based recorder that doesn't need ffmpeg at all (already on `.kb/ROADMAP.md`), or (c) leave as-is and document the dependency.
- **Per-package report.html overwrite** still stands — `go test ./...` runs each package as a separate process, last one wins on `report.html`. Archives preserved.

### Files touched in PM Batch 2 (kexas)

| Path | Change |
|---|---|
| `kcore/element_typing.go` | `Element.Fill(text)` (PM Batch 1 — listed for completeness) |
| `ktest/ktest_parallel.go` | Added `screenshotPath` + `videoPath` to `testResult` struct |
| `ktest/ktest_impl.go` | `runSingleTest` now wraps test in recorder Start/Stop and end-of-test screenshot capture; new `startTestRecording`, `stopAndSaveTestVideo`, `captureEndOfTestScreenshot` helpers |
| `ktest/ktest_runner_report.go` | Threads ScreenshotPath/VideoPath through to `report.TestCaseResult`; new `relativizeArtifact` for relative-to-reportDir URI generation |
| `ktest/report/report_model.go` | `TestCaseResult` now has `ScreenshotPath` + `VideoPath` fields |
| `ktest/report/report_html_tests.go` | New `writeTestArtifacts` rendering `<img>` + `<video>` (or frames-dir fallback link) |
| `ktest/report/report_assets.go` | New `.artifact-screenshot` + `.artifact-video` CSS rules; donut-legend label flex + count/percentage fixed widths |

---

## Update — 2026-05-02 PM batch 3 (+10 tests, no new kexas bugs)

### Tests added (40 total now)

| Suite | New tests | Coverage added |
|---|---|---|
| `inventory` | 4 (TestI–L) | Item detail page — name / price / description / back-to-products |
| `cart` | 1 (TestF) | Add ALL six products → badge count = 6 |
| `checkout` | 3 (TestI–K) | Complete order empties cart; cancel from overview returns to inventory; 2-item subtotal sums correctly ($29.99 + $9.99 = $39.98) |
| `menu` | 2 (TestD–E) | Logout clears authenticated session (direct URL access bounces to login); Reset App State clears cart but DOES NOT log out |

### New POM additions

- `pages/item_detail_page.go` — `ItemDetailPage` POM with `IsLoaded / Name / Description / Price / BackToProducts`. Saucedemo's item detail re-uses the same `data-test` names as inventory tiles (`inventory-item-name / desc / price`); they only resolve to a single element on `/inventory-item.html` so kexas's strict-mode `Find()` works without contortions there.
- `InventoryPage.OpenItem(product)` — clicks an inventory item by Name to navigate to its detail page. Uses `Page.Evaluate` because `Find("[data-test='inventory-item-name']")` would error (6 matches, strict-mode).

### Findings & lessons from the +10 batch

- **Zero new kexas bugs.** All 40 tests passed first try with `-p 1`. The kexas API surface that the +10 tests exercised — `Element.Fill` (used by checkout flow), `Page.Evaluate` (item-name click + cart count), `Page.Navigate` (logout-bounce check), `Element.Click + GetText` (item detail) — is the same surface the first 30 tests used. No new code paths uncovered new bugs. **This is a positive signal:** dogfooding diminishing returns means kexas's core surface is stabilizing.
- **Strict-mode `Find()` reappears as a constraint** for any "click the Nth thing" / "find by text content" flow. `OpenItem(product)` has to go through `Page.Evaluate` to find-by-text and click. Reinforces the need for a `Locator`-style API (already in the v0.1.0 priority list as item #6).
- **Saucedemo session model:** "logout" tears down localStorage authentication so `/inventory.html` direct access bounces back to login. "Reset App State" clears `cart-contents` localStorage but leaves the auth token alone. Confirmed by TestD + TestE. Worth knowing for any future test that needs to verify isolation.
- **Test design — same-suite browser sharing kept simple:** the new tests didn't need to override `BeforeEach` because they all start from "logged-in standard_user on /inventory.html with empty cart" (the existing BeforeEach pattern). The CheckoutSuite's BeforeEach adds Backpack + opens cart, so TestK (multi-item subtotal) had to back out of the cart first via `ContinueShopping` before adding Bike Light. Slightly awkward — a per-test "starting state" helper would be cleaner, but acceptable for this suite size.
- **Direct URL navigation as a session probe** (TestD logout test) is a clean black-box check that doesn't depend on UI state. Saucedemo's React app reads `localStorage["session-username"]` on mount and redirects unauth'd visitors. Pattern worth reusing for any auth-test:
  ```go
  _ = s.Page.Navigate(protectedURL)
  loaded, _ := protectedPage.IsLoaded()
  kassert.That(s.T(), loaded).IsFalse()
  ```
- **Two-item subtotal test (TestK) caught nothing new** but is a useful smoke check — saucedemo's price math has been reliable for years; the test value is more "regression guard" than "bug finder". Worth keeping.
- **Slowdown for the menu suite** went from ~28s to ~50s after adding TestD + TestE. TestD does an extra navigate-to-protected-URL probe (~7s for Find retry on the missing inventory selector); TestE adds another menu open + click sequence. Cumulative effect of always-on screenshot + video per test is also ~1s/test. Worth noting that the per-test artifact overhead is real (not free) — tradeoff for test debugability.

### Updated v0.1.0 priority list (+10 batch unchanged the rankings)

The +10 batch confirmed the existing priority list. Order unchanged:

1. ~~CDP connection-close hang~~ ✅ fixed
2. ~~`Element.Type` silent no-op on React~~ ✅ fixed via `Element.Fill`
3. ~~Config not found from sub-package CWDs~~ ✅ fixed
4. ~~No HTML report for legacy `ktest.Run` path~~ ✅ fixed
5. ~~Per-test screenshot was failure-only~~ ✅ now always-on
6. ~~Per-test video missing entirely~~ ✅ now always-on (degrades to frames dir without ffmpeg)
7. ~~Donut legend column misalignment~~ ✅ fixed (CSS)
8. **Per-package `report.html` overwrite** — still the headline open issue
9. **Auto-step capture** — passed test details still need step-by-step "what the test did" without requiring opt-in `ktest.Step()` calls
10. **No `Page.FindAll` / `Locator`** — multi-element flows still go through `Page.Evaluate` (re-confirmed by `OpenItem`)
11. `Page.Evaluate` typed-result helpers
12. `kassert.That(slice).Contains`
13. `Element.SelectOption`
14. `WaitForElementInteractive`
15. ffmpeg dependency for video — could add CDP screencast-event-based recorder per `.kb/ROADMAP.md`

---

