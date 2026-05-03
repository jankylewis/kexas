# v0.1.0 Dogfood Fixes — 2026-05-02

Six kexas fixes shipped from the sltests dogfood session. Full context: `SLTESTS_DOGFOOD_FINDINGS.md`.

| # | Fix | File(s) |
|---|---|---|
| 1 | `Element.Fill(text)` — Playwright-style React-aware setter; companion to `Type` | `kcore/element_typing.go` |
| 2 | Config-walking — `loadConfig` walks UP from CWD; relative dirs anchored to config dir | `ktest/ktest_config.go` |
| 3 | HTML report wired into `ktest.Run` path (was AlphaInit-only) | `ktest/ktest_impl.go`, `ktest/ktest_main.go` |
| 4 | Always-on per-test screenshot at config.ScreenshotDir/`<TestName>.png` | `ktest/ktest_impl.go` |
| 5 | Always-on per-test video at config.VideoDir/`<TestName>.mp4` (auto-installs ffmpeg on first use; pure-Go GIF as final offline fallback) | `ktest/ktest_impl.go`, `kcore/recorder_ffmpeg_install.go`, `kcore/recorder_save.go` |
| 6 | Report detail panel always shows `<img>` + `<video>` for every test (pass + fail) | `ktest/report/report_html_tests.go`, `report_assets.go` |

## Verified

`sltests/` consumer: 30/30 tests pass with `-p 1`. Single `sltests/test-results/` dir holds report.html, screenshots/, videos/. Detail panel meaningful for passed tests.

## Open

Per-package `report.html` overwrites; auto-step capture (so passed tests show step-by-step too).

## Follow-ups landed

- **ffmpeg auto-install** — see `FFMPEG_AUTO_INSTALL.md`. Replaces the original "frames dir if ffmpeg missing" UX with Playwright-style auto-download. Surfaced + fixed two latent recorder bugs (frame-extension mismatch, x264 odd-dimension rejection) along the way.
