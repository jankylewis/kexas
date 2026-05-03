# Kexas Roadmap

Future features that are committed but not yet implemented. README's "🚧 Coming Soon" lists the headlines; this file holds the implementation plans.

Last updated: 2026-05-02.

## 1. Video recording integrated with ktest reporting

**Status:** Manual `Recorder` API exists; not integrated into the test framework or HTML report.

### What works today

- `Page.StartRecording(configs ...RecorderConfig) (*Recorder, error)` — public API in `recorder.go`. Users opt into recording manually inside test code.
- `Recorder.SaveVideo(outputPath)` — assembles captured frames into MP4 via `ffmpeg` if it's on PATH; falls back to writing individual JPEG/PNG frames if not.
- `RecorderConfig` — format (jpeg/png), quality, max width/height, every-Nth-frame, output directory.
- Tests: `tests/kcore/recorder_test.go` (integration) exercises the manual API.

### What's missing

| Gap | Where |
|---|---|
| ktest does not auto-start a recording per test | `ktest/ktest_parallel.go runTestWithRecovery` |
| Recordings are not saved under `config.VideoDir` automatically | `kexas.config.json videoDir` field is parsed but unused |
| HTML report does not embed videos | `ktest/report/report_html.go` renders screenshots-on-fail only |
| Frame capture is screenshot polling at ~10fps, not CDP screencast events | `recorder.go collectFrames` |
| `Page.startScreencast` CDP call is fired but its events are never subscribed | `recorder.go start()` — effectively dead code today |

### Implementation plan (~2–3 focused sessions)

**Phase 1 — Real CDP screencast event subscription**

Replace screenshot polling with subscription to `Page.screencastFrame` CDP events. Each event delivers a base64-encoded frame; ack via `Page.screencastFrameAck`. The destination handler `Recorder.SaveScreencastFrame(data, sessionID)` already exists at `recorder.go:320` waiting to be wired up. Side benefit: no more dead `startScreencast` call in `recorder.go start()`.

**Phase 2 — ktest auto-attach per test**

In `ktest/ktest_parallel.go runTestWithRecovery` (and `runTestsSequentialRegistered` in `ktest_runner.go`):

1. Before `test.Func(page, t)`, call `page.StartRecording(...)` with `RecorderConfig` derived from `config.VideoDir` and the test name.
2. After the test body returns, call `recorder.SaveVideo(...)` with a per-test output path (mirror screenshot naming: `<filename>.<TestName>.mp4`).
3. On panic / failure, save anyway (the existing `recover()` block + `if t.Failed()` handles this for screenshots — extend the same pattern for video).
4. On pass with `config.VideoOnFail = true` (new config flag), discard the recording instead of saving (saves disk + speeds up clean runs).

**Phase 3 — HTML report video embed**

In `ktest/report/report_html.go`:

1. Extend `TestCaseResult` model with `VideoPath string` field (empty if no video).
2. `Collector.Add(...)` populates `VideoPath` based on whether the file exists in the videos directory.
3. Report template renders `<video controls width="640" src="..."></video>` for any test row with a `VideoPath`. Match the screenshot-embed visual treatment.

### Effort estimate

| Phase | Effort | Dependencies |
|---|---|---|
| 1 | 1 session — CDP event subscription is well-scoped; existing `SaveScreencastFrame` reduces work | None |
| 2 | 0.5 session — straightforward wrapping in two ktest functions | Phase 1 |
| 3 | 1 session — model field + collector + template + per-test naming | Phase 2 |

Total: ~2.5 sessions of focused work.

### Order with other work

Not blocking. Should land after:

- The `kcore/` structural refactor (current focus).
- The 230-line / 50-line compliance pass on existing source.

Reasons: (a) the kcore/ move shifts file paths the report's screenshot-handling references, so doing report work after that move is cleaner; (b) recorder.go is one of the files being moved, so changes during the kcore refactor and changes here are best done sequentially, not concurrently.

---

## 2. Network interception + HAR capture

**Status:** Not started. Mentioned in README's "🚧 Coming Soon".

### Sketch (to be detailed when work begins)

- Subscribe to `Network.requestWillBeSent` / `Network.responseReceived` / `Network.loadingFinished` CDP events (the `Network` agent is already enabled lazily by `internal/agent/`).
- Maintain a per-page request/response log with timing.
- Serialize to HAR 1.2 format (industry standard, viewable in Chrome DevTools and many tools).
- Optionally, gate-block requests by URL pattern (`Fetch.requestPaused` for true interception).

### Open questions for that session

- HAR file location — under `config.HarDir` or alongside videos under `test-results/`?
- Auto-attach per test like the video plan, or opt-in via `page.StartHARCapture()`?

---

## 3. PDF export

**Status:** Not started. Mentioned in README's "🚧 Coming Soon".

### Sketch

- `Page.captureScreenshot` already supports full-page screenshots. Adding `Page.PrintPDF(opts)` would call CDP `Page.printToPDF` and return the PDF bytes.
- `PDFOptions`: orientation, paper size, margins, header/footer templates, print background, scale.
- Tiny method add — call CDP, base64-decode, return bytes. Probably ≤30 lines of code + tests.

---

## 4. Windows / Linux release builds

**Status:** Not started. macOS-only today.

### What's needed

- The launcher already supports `mac-arm64`, `mac-x64`, `linux64`, `win64`, `win32` for downloading Chrome for Testing — the binary-download path is portable.
- The `findChromium` system-fallback paths in `launcher/launcher.go` only cover macOS today. Need to add Linux paths (`/usr/bin/google-chrome`, `/usr/bin/chromium`, etc.) and Windows paths (`C:\Program Files\Google\Chrome\Application\chrome.exe`).
- The `killExistingChromeProcesses` helper in `launcher_cleanup.go` uses `lsof` and `pgrep` — both Unix-only. Windows would need `tasklist` / `taskkill`. Linux already supports `lsof`/`pgrep`.
- CI matrix for cross-platform testing.

### Effort estimate

~1 session for Linux support (low-risk extension of macOS code). 1-2 sessions for Windows (more divergence; needs `tasklist`/`taskkill` shim + path conventions).

---

## How to use this file

- New future features get a section. Keep them brief until they're imminent — then expand the implementation plan.
- When work starts, the section moves into a session SUMMARY's "What we did", not deleted from here — instead, mark `**Status: shipped 2026-MM-DD**` and link to the SUMMARY that completed it.
- Order is not priority order. See README's "🚧 Coming Soon" for the project's stated priority headline.
