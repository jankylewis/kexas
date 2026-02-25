# Browser Documentation

This folder contains all documentation related to browser management, downloads, and configuration in Kexas.

## Quick Reference

### Getting Started
- **[CHROME_VERSION_CONTROL.md](CHROME_VERSION_CONTROL.md)** - How to control which Chrome version to download (Stable/Beta/Dev/Canary)

### Understanding Chrome for Testing
- **[CHROMIUM_VS_CHROME_FOR_TESTING.md](CHROMIUM_VS_CHROME_FOR_TESTING.md)** - Difference between Chromium and Chrome for Testing, why we switched

### Download Infrastructure
- **[CHROMIUM_DOWNLOAD.md](CHROMIUM_DOWNLOAD.md)** - Download strategy, CDN details, platform support

## Document Overview

### CHROME_VERSION_CONTROL.md
**Purpose:** Main guide for controlling Chrome versions

**Topics covered:**
- Available release channels (Stable, Beta, Dev, Canary)
- How to change channels
- Version caching behavior
- Pinning specific versions
- Troubleshooting

**When to read:** When you need to change Chrome version or understand version management

---

### CHROMIUM_VS_CHROME_FOR_TESTING.md
**Purpose:** Explains the migration from Chromium to Chrome for Testing

**Topics covered:**
- Timeline of changes (Playwright 1.57)
- What is Chrome for Testing
- Differences between Chromium and Chrome
- Why we switched
- Technical comparison

**When to read:** When you want to understand why Kexas uses Chrome for Testing instead of Chromium

---

### CHROMIUM_DOWNLOAD.md
**Purpose:** Technical details about browser download infrastructure

**Topics covered:**
- Current download status
- Download strategy (Chrome for Testing + System Chrome fallback)
- Why Playwright CDN
- Platform-specific details
- Testing instructions

**When to read:** When you need to understand how browser downloads work or troubleshoot download issues

---

## Chrome Release Channels Explained

### Stable (Default) ✅
- **Update frequency:** Every 4-6 weeks
- **Stability:** Most stable, production-ready
- **Use case:** Production testing, CI/CD
- **Current version:** 145.0.7632.117

### Beta
- **Update frequency:** Every 4-6 weeks (ahead of Stable)
- **Stability:** Generally stable, may have minor bugs
- **Use case:** Testing upcoming features before Stable release
- **Current version:** 146.0.7680.16

### Dev
- **Update frequency:** Weekly
- **Stability:** Less stable, more bugs
- **Use case:** Early access to new features
- **Current version:** 147.0.7695.0

### Canary 🐤
- **Update frequency:** Daily
- **Stability:** Least stable, experimental
- **Use case:** Bleeding edge testing, feature exploration
- **Current version:** 147.0.7701.0
- **Name origin:** "Canary in a coal mine" - early warning system for bugs

## Quick Tasks

### Change Chrome Version
1. Edit `launcher/downloader.go`
2. Change `ChromeReleaseChannel` constant
3. Remove cache: `rm -rf ~/.kexas/browsers/chrome-*`
4. Run tests: `go test -v ./tests`

See: [CHROME_VERSION_CONTROL.md](CHROME_VERSION_CONTROL.md)

### Check Current Version
```bash
ls -la ~/.kexas/browsers/
```

### Force Re-download
```bash
rm -rf ~/.kexas/browsers/chrome-*
go test -v ./tests
```

### Check Latest Available Versions
```bash
curl -s https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json | jq '.channels'
```

## Architecture

### Download Flow
```
1. Check cache (~/.kexas/browsers/chrome-{version}/)
   ├─ Exists? → Use cached version
   └─ Not exists? → Continue to download

2. Fetch version from API
   └─ https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json

3. Get download URL for platform
   ├─ mac-arm64
   ├─ mac-x64
   ├─ linux64
   ├─ win64
   └─ win32

4. Download .zip from Google CDN
   └─ https://storage.googleapis.com/chrome-for-testing-public/{version}/{platform}/chrome-{platform}.zip

5. Extract to cache directory
   └─ ~/.kexas/browsers/chrome-{version}/

6. Make binary executable (Unix)

7. Return path to binary
```

### Fallback Strategy
```
1. Try Chrome for Testing (downloaded)
   └─ ~/.kexas/browsers/chrome-{version}/

2. If download fails → Try System Chrome
   ├─ macOS: /Applications/Google Chrome.app/
   ├─ Linux: /usr/bin/google-chrome
   └─ Windows: C:\Program Files\Google\Chrome\
```

## File Locations

### Cache Directory
```
~/.kexas/browsers/
├── chrome-145.0.7632.117/    # Stable
├── chrome-146.0.7680.16/      # Beta (if downloaded)
└── chrome-147.0.7695.0/       # Dev (if downloaded)
```

### Binary Paths

**macOS ARM64:**
```
~/.kexas/browsers/chrome-{version}/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing
```

**macOS Intel:**
```
~/.kexas/browsers/chrome-{version}/chrome-mac-x64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing
```

**Linux:**
```
~/.kexas/browsers/chrome-{version}/chrome-linux64/chrome
```

**Windows:**
```
~/.kexas/browsers/chrome-{version}/chrome-win64/chrome.exe
```

## External Resources

- **Chrome for Testing Dashboard:** https://googlechromelabs.github.io/chrome-for-testing/
- **Chrome for Testing Blog Post:** https://developer.chrome.com/blog/chrome-for-testing/
- **JSON API Endpoint:** https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json
- **Playwright Release Notes (1.57):** https://playwright.dev/docs/release-notes#version-157

## Related Code

- **launcher/downloader.go** - Browser download implementation
- **launcher/launcher.go** - Browser launch logic
- **browser.go** - Browser control API
- **page.go** - Page automation API

## Glossary

- **Chrome for Testing** - Google's official Chrome build for automated testing (no auto-update)
- **Chromium** - Open-source browser project (Chrome is based on Chromium)
- **CDN** - Content Delivery Network (distributed file hosting)
- **Canary** - Daily experimental Chrome release (named after "canary in a coal mine")
- **Binary** - Executable program file (compiled machine code)
- **Release Channel** - Update track (Stable, Beta, Dev, Canary)
- **Revision** - Internal build number used by Chrome team
