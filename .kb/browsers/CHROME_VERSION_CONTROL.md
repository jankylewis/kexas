# Chrome Version Control

## How to Specify Chrome Version

Kexas downloads Chrome for Testing from Google's official CDN. You can control which version gets downloaded by changing the `ChromeReleaseChannel` constant in `launcher/downloader.go`.

## Available Channels

Chrome for Testing provides 4 release channels:

### 1. Stable (Default)
```go
const ChromeReleaseChannel string = "Stable"
```
- **Current version:** 145.0.7632.117
- **Best for:** Production use, stable testing
- **Update frequency:** Every 4-6 weeks
- **Recommended:** Yes, this is the default

### 2. Beta
```go
const ChromeReleaseChannel string = "Beta"
```
- **Current version:** 146.0.7680.16
- **Best for:** Testing upcoming features
- **Update frequency:** Every 4-6 weeks (ahead of Stable)
- **Stability:** Generally stable, but may have bugs

### 3. Dev
```go
const ChromeReleaseChannel string = "Dev"
```
- **Current version:** 147.0.7695.0
- **Best for:** Early testing of new features
- **Update frequency:** Weekly
- **Stability:** Less stable, more bugs

### 4. Canary
```go
const ChromeReleaseChannel string = "Canary"
```
- **Current version:** 147.0.7701.0
- **Best for:** Bleeding edge testing
- **Update frequency:** Daily
- **Stability:** Least stable, experimental features

## How to Change Channel

**Step 1:** Edit `launcher/downloader.go`

```go
// Change this line:
const ChromeReleaseChannel string = "Stable"

// To one of:
const ChromeReleaseChannel string = "Beta"
const ChromeReleaseChannel string = "Dev"
const ChromeReleaseChannel string = "Canary"
```

**Step 2:** Remove cached Chrome (optional, to force re-download)

```bash
rm -rf ~/.kexas/browsers/chrome-*
```

**Step 3:** Run your tests

```bash
go test -v ./tests
```

Kexas will automatically download the latest version from your chosen channel.

## How It Works

### Automatic Version Detection

Kexas fetches the latest version from Chrome for Testing's JSON API:

```
https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json
```

**API Response Structure:**
```json
{
  "channels": {
    "Stable": {
      "version": "145.0.7632.117",
      "revision": "1568190",
      "downloads": {
        "chrome": [
          {
            "platform": "mac-arm64",
            "url": "https://storage.googleapis.com/chrome-for-testing-public/145.0.7632.117/mac-arm64/chrome-mac-arm64.zip"
          }
        ]
      }
    },
    "Beta": { ... },
    "Dev": { ... },
    "Canary": { ... }
  }
}
```

### Download Process

1. **Fetch API:** Get latest version for chosen channel
2. **Select platform:** Determine correct platform (mac-arm64, mac-x64, linux64, win64, win32)
3. **Download:** Fetch `.zip` from Google CDN
4. **Extract:** Unzip to `~/.kexas/browsers/chrome-{version}/`
5. **Cache:** Subsequent runs use cached version

### Version Caching

Once downloaded, Chrome is cached by version:

```
~/.kexas/browsers/
├── chrome-145.0.7632.117/    # Stable
├── chrome-146.0.7680.16/      # Beta (if you switched)
└── chrome-147.0.7695.0/       # Dev (if you switched)
```

**Behavior:**
- If cached version exists, Kexas uses it (no download)
- If you switch channels, Kexas downloads the new version
- Old versions remain cached (manual cleanup needed)

## Pinning a Specific Version

Currently, Kexas always downloads the **latest** version from the chosen channel. If you need a specific version:

### Option 1: Manual Download (Temporary)

Download a specific version manually and place it in the cache:

```bash
# Example: Download Chrome 144.0.7626.0
VERSION="144.0.7626.0"
mkdir -p ~/.kexas/browsers/chrome-$VERSION
cd ~/.kexas/browsers/chrome-$VERSION

# Download for your platform
curl -O "https://storage.googleapis.com/chrome-for-testing-public/$VERSION/mac-arm64/chrome-mac-arm64.zip"
unzip chrome-mac-arm64.zip
rm chrome-mac-arm64.zip
```

Then modify `getChromeVersion()` to return your pinned version:

```go
func getChromeVersion() (string, error) {
    return "144.0.7626.0", nil  // Hardcoded version
}
```

### Option 2: Add Version Constant (Future Enhancement)

We could add a constant to pin a specific version:

```go
// If set, use this specific version instead of latest
const ChromeVersion string = ""  // Empty = use latest

func getChromeVersion() (string, error) {
    if ChromeVersion != "" {
        return ChromeVersion, nil
    }
    // Otherwise fetch latest from API...
}
```

This would require code changes to support.

## Checking Current Version

To see which Chrome version is currently cached:

```bash
ls -la ~/.kexas/browsers/
```

To see which version will be downloaded:

```bash
curl -s https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json | \
  jq '.channels.Stable.version'
```

Replace `Stable` with `Beta`, `Dev`, or `Canary` to check other channels.

## Download URLs

All Chrome for Testing downloads come from Google's CDN:

```
https://storage.googleapis.com/chrome-for-testing-public/{version}/{platform}/chrome-{platform}.zip
```

**Platforms:**
- `mac-arm64` - macOS Apple Silicon
- `mac-x64` - macOS Intel
- `linux64` - Linux x64
- `win64` - Windows 64-bit
- `win32` - Windows 32-bit

**Example URLs:**
```
# Stable 145.0.7632.117 for Mac ARM64
https://storage.googleapis.com/chrome-for-testing-public/145.0.7632.117/mac-arm64/chrome-mac-arm64.zip

# Beta 146.0.7680.16 for Linux
https://storage.googleapis.com/chrome-for-testing-public/146.0.7680.16/linux64/chrome-linux64.zip
```

## Comparison: Playwright vs Kexas

| Aspect | Playwright | Kexas |
|--------|-----------|-------|
| **Default channel** | Stable | Stable |
| **Version control** | Hardcoded in package | Dynamic from API |
| **Update method** | `npm update playwright` | Change constant + re-run |
| **Cache location** | `~/Library/Caches/ms-playwright/` | `~/.kexas/browsers/` |
| **Version pinning** | Via package version | Via channel selection |

## Best Practices

1. **Use Stable for production** - Most reliable, well-tested
2. **Use Beta for pre-release testing** - Test upcoming features before they hit Stable
3. **Clean old versions** - Manually remove old cached versions to save disk space
4. **Document your channel** - If using non-Stable, document why in your project

## Troubleshooting

### "Channel not found" error

Make sure you spelled the channel name correctly (case-sensitive):
- ✅ `"Stable"`, `"Beta"`, `"Dev"`, `"Canary"`
- ❌ `"stable"`, `"BETA"`, `"development"`

### Download fails

1. Check internet connection
2. Verify Google CDN is accessible
3. Check if version exists: visit the dashboard at https://googlechromelabs.github.io/chrome-for-testing/

### Wrong version downloaded

1. Remove cache: `rm -rf ~/.kexas/browsers/chrome-*`
2. Verify `ChromeReleaseChannel` constant is set correctly
3. Re-run tests to trigger fresh download

## Future Enhancements

Potential improvements for version control:

1. **Environment variable:** `KEXAS_CHROME_CHANNEL=Beta go test ./tests`
2. **Config file:** `.kexas/config.json` with channel setting
3. **Version pinning:** Specify exact version instead of channel
4. **Multiple versions:** Support testing against multiple Chrome versions
5. **Auto-cleanup:** Automatically remove old cached versions
