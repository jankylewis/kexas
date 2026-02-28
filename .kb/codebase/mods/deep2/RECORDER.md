# Video Recording — Plain Language Guide

---

## What Is Video Recording in Testing?

Imagine a security camera for your website tests. Every time a test runs, it records a video of what the browser is doing — clicking buttons, loading pages, filling forms. If something goes wrong, you can watch the video and see exactly what happened.

## Why Is This Useful?

| Situation | How Video Helps |
|-----------|----------------|
| **Test fails in CI** | You can't see CI's screen, but you can watch the recording |
| **Flaky test** | A test that sometimes passes and sometimes fails — the video shows what's different |
| **Bug report** | Attach the video to show developers exactly what the user sees |
| **Demo** | Record a walk-through of a feature for stakeholders |

## How Does It Work?

1. **Start recording** — Tell kexas "start the camera"
2. **Run your test** — Click buttons, fill forms, navigate pages — the camera captures everything
3. **Stop recording** — Tell kexas "stop the camera"
4. **Save the video** — Kexas combines all the captured pictures into an MP4 video file

Think of it like a flipbook: kexas takes about 10 pictures per second, then "flips" through them quickly to create a video.

## What Do You Need?

- **kexas** — The testing framework (you already have this)
- **ffmpeg** (optional) — A free tool that combines pictures into video. If you don't have it installed, kexas saves the individual pictures instead

### What if I don't have ffmpeg?

That's fine! Kexas will save each frame as a separate image file (like `frame_0001.jpg`, `frame_0002.jpg`, etc.). You can still see what happened by looking through the images.

## Settings You Can Customize

| Setting | What It Controls | Default |
|---------|-----------------|---------|
| **Format** | Image quality type (JPEG or PNG) | JPEG |
| **Quality** | How detailed each frame is (0-100) | 80 |
| **Max Width** | Maximum frame width in pixels | 1280 |
| **Max Height** | Maximum frame height in pixels | 720 |
| **Output Directory** | Where videos are saved | `./recordings/` |

## Tips

- **Keep recordings short** — Each frame takes memory. A 30-second recording at 10fps = 300 pictures in memory
- **Use JPEG format** — It's smaller than PNG and good enough for test recordings
- **Record only on failure** — In CI, you often only want to save videos for tests that fail, to save storage space
