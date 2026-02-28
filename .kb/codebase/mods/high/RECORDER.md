# Video Recording — High-Level Overview

**File:** `kexas/recorder.go`
**CDP Domain:** Page (screencast)
**Status:** Implemented

---

## What It Does

Records browser activity during test execution as video files for debugging, documentation, and CI artifacts. Supports both raw frame export and ffmpeg-based MP4 encoding with graceful fallback.

## API Surface

| Method | Description |
|--------|-------------|
| `page.StartRecording(config...)` | Begin capturing frames, returns `*Recorder` |
| `recorder.Stop()` | End recording session |
| `recorder.IsRecording()` | Whether recorder is active |
| `recorder.FrameCount()` | Number of captured frames |
| `recorder.SaveFrames(dir)` | Write frames as individual image files |
| `recorder.SaveVideo(path)` | Assemble frames into MP4 via ffmpeg (falls back to frames) |
| `recorder.SaveScreencastFrame(data, sessionID)` | Manually inject a base64 frame |

## Architecture

### Frame Capture

Uses a polling-based approach: a background goroutine calls `page.Screenshot()` every ~100ms (~10fps). This is simpler and more portable than CDP event-based screencast, which requires async event subscription infrastructure.

### Video Assembly Pipeline

1. Frames stored in memory as `[]Frame` (Data []byte, Timestamp, Index)
2. `SaveFrames()` writes numbered files: `frame_0001.jpg`, `frame_0002.jpg`, ...
3. `SaveVideo()` writes frames to temp dir, runs ffmpeg, cleans up temp files
4. If ffmpeg not found: graceful fallback to `SaveFrames()`

### RecorderConfig

| Field | Default | Description |
|-------|---------|-------------|
| Format | "jpeg" | Frame format (jpeg or png) |
| Quality | 80 | JPEG quality (0-100) |
| MaxWidth | 1280 | Max frame width |
| MaxHeight | 720 | Max frame height |
| EveryNthFrame | 1 | Capture every Nth frame |
| OutputDir | "./recordings" | Output directory |
| OutputFilename | "recording" | Base filename |

## CDP Commands

- `Page.startScreencast` — Begin frame capture
- `Page.stopScreencast` — End frame capture
- `Page.screencastFrameAck` — Acknowledge frame receipt
- `Page.captureScreenshot` — Capture individual frames (polling fallback)

## Key Design Decisions

1. **Polling over events** — CDP screencast events require complex async infrastructure; polling at 100ms is reliable and simple
2. **ffmpeg as optional dependency** — Detected at runtime via `exec.LookPath`; graceful fallback to raw frames
3. **Mutex-protected frame buffer** — Thread-safe for concurrent frame capture and read
4. **Stop channel for goroutine cleanup** — `close(stopCh)` signals the frame collector to exit
5. **Double-start prevention** — Returns `ErrRecorderAlreadyStarted`
6. **Stop-before-start returns error** — Returns `ErrRecorderNotStarted`
7. **SaveVideo auto-stops** — If still recording when SaveVideo is called, stops first

## Pitfalls

- **Memory usage** — Frames accumulate in memory; long recordings at high resolution can use significant RAM
- **Headless rendering** — Headless Chrome may render slightly differently than headed mode
- **ffmpeg availability in CI** — Docker images need ffmpeg pre-installed; the CI plan includes this
