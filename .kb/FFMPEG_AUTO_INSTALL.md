# FFmpeg auto-install — 2026-05-02

Replaced the "GIF fallback when ffmpeg is missing" UX with Playwright-style auto-download. First run fetches a static ffmpeg into `~/.kexas/bin/`; subsequent runs reuse it. No user setup. Output is real `.mp4` (H.264 / yuv420p), not GIF.

## Why we did it

Pre-fix flow when ffmpeg wasn't on `PATH`:
1. `recorder.SaveVideo(...)` saw missing ffmpeg → fell back to pure-Go animated GIF.
2. User got a `.gif` instead of the `.mp4` they asked for.

Acceptable as a no-deps backstop, but it leaks a system-tooling concern into the test author's mental model. Playwright (and Puppeteer) avoid this by bundling/auto-fetching browsers — same idea applied to ffmpeg.

## What landed

### `kcore/recorder_ffmpeg_install.go` (new, ~225 lines)

Public entry: `EnsureFFmpeg(log *logger.Logger) (string, error)`.

Resolution order:
1. `~/.kexas/bin/ffmpeg` — the kexas-managed install (checked first so a cached download is reused before scanning PATH).
2. `exec.LookPath("ffmpeg")` — system install (lets users override the bundled version).
3. Download + extract into `~/.kexas/bin/` — happens once, cached forever.

Static-build sources:

| OS / arch | Source | Format |
|---|---|---|
| darwin (universal) | `evermeet.cx/ffmpeg/getrelease/zip` | `.zip` |
| linux/amd64 | `github.com/BtbN/FFmpeg-Builds` (latest) | `.tar.xz` |
| linux/arm64 | `github.com/BtbN/FFmpeg-Builds` (latest) | `.tar.xz` |
| windows/amd64 | `github.com/BtbN/FFmpeg-Builds` (latest) | `.zip` |

`.zip` extracted via stdlib `archive/zip`. `.tar.xz` shells out to system `tar -xf` (auto-detects xz on every modern Linux) — this avoids adding a pure-Go xz decoder dependency.

### `kcore/recorder_save.go` (modified)

Replaced `exec.LookPath("ffmpeg")` with `EnsureFFmpeg(r.log)`. GIF stays as final fallback for offline / unsupported platforms / encode failures — but no longer the default for "user just hasn't installed ffmpeg."

## Bugs surfaced during verification

End-to-end testing on `samplings/ghtests` exposed two latent recorder bugs that the GIF fallback had been silently hiding:

### Bug 1 — frame-extension mismatch

Symptom: ffmpeg ran cleanly, exited 0, but produced empty `.mp4` files.

Root cause: `RecorderConfig.Format` defaults to `"jpeg"`, so `frameExtension()` returned `"jpg"`. But the actual frame bytes come from `Page.Screenshot()` which **always** returns PNG. Result: `frame_0000.jpg` files containing PNG bytes. ffmpeg detected them as PNG via magic-byte sniffing, then x264 errored on the codec mismatch (silently in some flows, explicitly in others).

Fix: `Recorder.detectActualFrameExt()` sniffs the first frame's magic bytes (`isPNG` / `isJPEG` helpers already existed in `recorder_gif.go`) and returns the real extension. `writeFramesToTempDir` and `runFFmpegEncode` both take this `ext` as a parameter rather than reading config.

The same trap is documented for the GIF path in `recorder_gif.go` — config flag and actual bytes are not the same source of truth. `Page.Screenshot()` is hard-coded PNG; trust the bytes.

### Bug 2 — x264 odd-dimension rejection

Symptom: ffmpeg log:
```
[enc:libx264] Error while opening encoder - maybe incorrect parameters
such as bit_rate, rate, width or height.
```

Root cause: viewport was 1280x633 (odd height). `libx264 + yuv420p` requires both dimensions even — yuv420p subsamples chroma 2:1 in each direction, which only divides cleanly on even pixel counts.

Fix: added `-vf "pad=ceil(iw/2)*2:ceil(ih/2)*2"` to the ffmpeg command line. Pads up to the next even dimension (1280x633 → 1280x634) by adding a 1-pixel black row. Cheaper and lossless vs. cropping or scaling.

## Verified

End-to-end on a clean machine (no ffmpeg on PATH, no `~/.kexas/bin/`):

```
$ which ffmpeg
ffmpeg not found
$ ls ~/.kexas/bin/
ls: ...: No such file or directory

$ go test -run TestGitHub ./tests/
ok  github.com/jankylewis/ghtests/tests  52.409s   # download incl.

$ ls ~/.kexas/bin/ffmpeg
-rwxr-xr-x  80269896 May  2 22:28 /Users/.../.kexas/bin/ffmpeg

$ go test ./tests/   # cached
ok  github.com/jankylewis/ghtests/tests  7.229s

$ file test-results/videos/*.mp4
ISO Media, MP4 Base Media v1
$ ffmpeg -i ... | grep Stream
Video: h264 (High), yuv420p, 1280x634, 691 kb/s, 10 fps
```

Cross-checked against `samplings/hntests` and `samplings/wikitests` — both produce real H.264 MP4 with no regressions.

## Resolution order rationale

Why managed dir before PATH (and not the other way around)?

- **Reproducibility for CI / parallel `go test`** — the managed install is a known version we tested with. A user's system ffmpeg might be 4.x with libx264 disabled in a packager build.
- **Self-healing** — if a user runs `kexas` once and gets the bundled ffmpeg, future `go test` runs work without depending on PATH state. Especially important for parallel test packages where each spawns its own ffmpeg.
- **User can still override** — installing system ffmpeg into PATH is fine, but only takes effect the first time before the managed copy is cached. To hard-override, delete `~/.kexas/bin/ffmpeg`.

## Known limitations

1. **Concurrent installs race.** Multiple `go test` packages launching simultaneously may each start a download before any cache exists. Both succeed (each writes the binary, last writer wins), but it's wasted bandwidth. Live with it for v0.1; add a flock if it bites in practice.

2. **No version pinning.** URLs use `releases/latest` and `getrelease` (always-latest). Reproducibility is good enough for test artifacts but not for "exact bytes" comparisons. If we ever need that, switch to dated/tagged release URLs.

3. **macOS arch detection.** evermeet.cx ships a universal binary, so `darwin/amd64` and `darwin/arm64` use the same URL. Works today, may need a split if they ever stop shipping universal.

4. **No 32-bit / unusual arches.** `linux/386`, `linux/arm`, `windows/arm64` etc. fall through to "no auto-install URL" → GIF fallback. Acceptable: those aren't realistic test-runner platforms.

## Supersedes

`V010_DOGFOOD_FIXES.md` open item: "ffmpeg-missing degrades video to frames dir" — gone. Default is now real MP4.
