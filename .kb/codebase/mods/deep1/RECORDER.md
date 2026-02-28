# Video Recording — Deep Technical Guide (Student Level)

**File:** `kexas/recorder.go`

---

## Why Record Tests?

When a test fails in CI, you can't see what happened. Was the button not visible? Did the page not load? Was there a popup blocking the element? A video recording answers all these questions instantly. It's also useful for:
- Creating demo videos for stakeholders
- Visual regression evidence
- Debugging flaky tests (tests that pass sometimes and fail other times)

---

## How CDP Screencast Works

Chrome has a built-in screen capture feature via the `Page.startScreencast` CDP command. When enabled, Chrome periodically captures the visible page and sends frames as base64-encoded images.

### The CDP Screencast Flow

```
Go Program                    Chrome
    |                            |
    |-- Page.startScreencast --> |  (start capturing)
    |                            |
    |<-- Page.screencastFrame -- |  (here's frame #1, base64 JPEG)
    |-- screencastFrameAck --->  |  (got it, send more)
    |                            |
    |<-- Page.screencastFrame -- |  (here's frame #2)
    |-- screencastFrameAck --->  |  (got it)
    |    ...                     |
    |                            |
    |-- Page.stopScreencast -->  |  (stop capturing)
```

**What is base64?** A way to encode binary data (like an image) as ASCII text. Every 3 bytes of binary become 4 ASCII characters. This lets you embed images in JSON (which only supports text). The overhead is ~33% larger than raw binary.

### FAQ: Why does Chrome need an "ack" for each frame?

**Flow control.** If Chrome sends frames faster than your program can process them, frames pile up in memory. The ack tells Chrome "I'm ready for the next frame." Without it, Chrome might send 60 frames per second and overwhelm a slow consumer.

---

## Our Implementation: Polling Fallback

The ideal approach uses CDP screencast events (described above). However, receiving CDP events requires **asynchronous event subscription** — a background goroutine that continuously reads from the WebSocket and routes events to handlers. This is complex infrastructure that kexas doesn't have yet.

### The Polling Approach

Instead, we use a simpler polling pattern:

```go
func (r *Recorder) collectFrames() {
    for {
        select {
        case <-r.stopCh:    // Stop signal received
            return
        default:
            screenshot, err := r.page.Screenshot()  // Capture one frame
            // ... store frame ...
            time.Sleep(100 * time.Millisecond)       // ~10 fps
        }
    }
}
```

Each iteration:
1. Check if stop was requested (via channel)
2. Call `page.Screenshot()` — a full CDP `Page.captureScreenshot` round-trip
3. Store the PNG bytes in the frame buffer
4. Sleep 100ms → approximately 10 frames per second

### FAQ: Why is polling slower than events?

Each `Screenshot()` call is a full WebSocket round-trip: send command → Chrome renders frame → Chrome encodes to PNG → Chrome base64-encodes → sends response → Go decodes base64. This takes ~50-100ms. With events, Chrome pushes frames proactively at its own pace, avoiding the request overhead.

### FAQ: When will event-based capture be added?

When the CDP client is enhanced to support event subscriptions (a background reader goroutine that demultiplexes incoming WebSocket messages into commands vs events).

---

## The Recorder Struct

```go
type Recorder struct {
    page      *Page            // The page being recorded
    config    RecorderConfig   // Format, quality, dimensions
    frames    []Frame          // Captured frames buffer
    recording bool             // Is currently recording?
    stopCh    chan struct{}     // Signal to stop the collector goroutine
    mu        sync.Mutex       // Protects frames and recording state
    log       *logger.Logger
}
```

### Key Fields Explained

**stopCh (chan struct{})** — An empty channel used purely as a signal. When you call `close(stopCh)`, all goroutines reading from it immediately receive a value (the zero value). This is Go's idiomatic pattern for broadcasting a "stop" signal to one or more goroutines.

**Why `chan struct{}` and not `chan bool`?** `struct{}` takes zero bytes of memory. We don't care about the value — only whether the channel is closed. This is a Go convention for signal-only channels.

**sync.Mutex** — Protects concurrent access to `frames` (the collector goroutine writes, `FrameCount()`/`SaveFrames()` read) and `recording` (Start/Stop/IsRecording).

---

## Video Assembly with ffmpeg

### What is ffmpeg?

ffmpeg is a command-line tool for video/audio processing. It can convert between formats, resize, compress, concatenate, and assemble individual images into video. It's the Swiss Army knife of multimedia.

### The Assembly Command

```bash
ffmpeg -y -framerate 10 -i frame_%04d.jpg -c:v libx264 -pix_fmt yuv420p -preset fast output.mp4
```

| Flag | Meaning |
|------|---------|
| `-y` | Overwrite output file without asking |
| `-framerate 10` | Input frames are at 10 fps |
| `-i frame_%04d.jpg` | Input pattern: frame_0001.jpg, frame_0002.jpg, ... |
| `-c:v libx264` | Use H.264 codec (universally supported) |
| `-pix_fmt yuv420p` | Pixel format for maximum compatibility |
| `-preset fast` | Encoding speed/quality tradeoff (fast = quicker, slightly larger file) |

### Graceful Fallback

```go
ffmpegPath, lookErr = exec.LookPath("ffmpeg")
if lookErr != nil {
    // ffmpeg not installed — save frames as individual images instead
    return r.SaveFrames(framesDir)
}
```

**What is exec.LookPath?** It searches the system's `PATH` environment variable for an executable named "ffmpeg". If found, returns its full path. If not found, returns an error. This is how we detect whether ffmpeg is available at runtime without requiring it as a hard dependency.

---

## Frame Storage: Memory vs Disk Tradeoff

Currently, frames are stored **in memory** as `[]Frame`. Each frame is a PNG image (~50-200KB). For a 30-second recording at 10fps:

```
300 frames × 100KB average = ~30MB RAM
```

For a 5-minute recording:
```
3000 frames × 100KB = ~300MB RAM
```

### Future Optimization

For long recordings, frames should be streamed directly to disk instead of accumulated in memory. This would change the architecture to:

1. Collector goroutine writes each frame to a numbered file immediately
2. `SaveVideo()` reads files from disk instead of memory
3. Memory usage stays constant regardless of recording length

---

## Pitfalls to Avoid

1. **Memory pressure** — Long recordings accumulate frames in RAM. Monitor memory usage for recordings >1 minute.
2. **ffmpeg in CI** — CI Docker images must include ffmpeg. The CI plan's Dockerfile installs it.
3. **Headless rendering differences** — Headless Chrome may render fonts and shadows slightly differently than headed mode. Screenshots/videos from CI may look slightly different from local.
4. **Recording during navigation** — `Screenshot()` may fail during page navigation (execution context destroyed). The collector ignores these errors and retries on the next cycle.
5. **Don't forget to Stop** — If you start recording and forget to stop, the collector goroutine runs forever (goroutine leak). Always defer `recorder.Stop()`.
