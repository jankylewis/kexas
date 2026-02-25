# Chromium vs Chrome for Testing

## Important Discovery

The user correctly identified that Playwright changed what it downloads in version 1.57 (2024).

## Timeline

### Before Playwright 1.57 (Pre-2024)
- **Downloaded:** Chromium (open-source browser)
- **Source:** `playwright.azureedge.net/builds/chromium/{build}/`
- **Build numbers:** 1080, 1084, 1117, 1140, etc.
- **What it is:** Open-source Chromium without Google branding/services

### After Playwright 1.57 (2024+)
- **Downloads:** Chrome for Testing (Google's official testing browser)
- **Source:** Google's Chrome for Testing infrastructure
- **What it is:** Actual Chrome (branded) without auto-update
- **Exception:** ARM64 Linux still uses Chromium

## What is Chrome for Testing?

**Chrome for Testing** is Google's official browser for automation:

1. **Real Chrome** - Not Chromium, actual Chrome with Google branding
2. **No auto-update** - Version stays fixed (perfect for testing)
3. **Versioned binaries** - Every Chrome version available
4. **Official release** - Part of Chrome release process
5. **Matches user Chrome** - Same as what real users have

**Why Google created it:**
- Regular Chrome auto-updates (bad for testing reproducibility)
- Developers needed versioned, non-updating Chrome
- Previously had to use Chromium (which differs from Chrome)
- Chrome for Testing solves this: real Chrome without auto-update

## What Kexas Currently Does

**Current implementation:**
```go
const ChromiumVersion string = "1140"
// Downloads from: playwright.azureedge.net/builds/chromium/1140/chromium-mac-arm64.zip
```

**This downloads:** Chromium build 1140 (old Playwright infrastructure)

**Status:** Works perfectly, but uses pre-1.57 Playwright approach

## What Modern Playwright Does

**Playwright 1.57+ behavior:**
1. Downloads Chrome for Testing from Google's CDN
2. Uses Google's JSON API to find versions
3. Downloads from: `storage.googleapis.com/chrome-for-testing-public/...`
4. Gets actual Chrome (not Chromium)

**Example Playwright 1.57+ download:**
```
https://storage.googleapis.com/chrome-for-testing-public/131.0.6778.87/mac-arm64/chrome-mac-arm64.zip
```

## Key Differences

| Aspect | Chromium (old) | Chrome for Testing (new) |
|--------|---------------|-------------------------|
| **What** | Open-source Chromium | Branded Chrome |
| **Source** | Playwright Azure CDN | Google CDN |
| **Build numbers** | 1140, 1117, etc. | 131.0.6778.87, etc. |
| **Branding** | No Google branding | Google Chrome branding |
| **Services** | No Google services | Includes Google services |
| **Matches users** | Close, but not exact | Exact match to user Chrome |
| **Playwright version** | Pre-1.57 | 1.57+ |

## Should Kexas Switch?

### Option 1: Keep Chromium (Current)
**Pros:**
- ✅ Works perfectly right now
- ✅ Simple implementation
- ✅ Fast downloads from Azure Singapore
- ✅ No code changes needed

**Cons:**
- ❌ Uses older Playwright infrastructure
- ❌ Not matching modern Playwright behavior
- ❌ Chromium differs slightly from real Chrome
- ❌ May be deprecated in future

### Option 2: Switch to Chrome for Testing
**Pros:**
- ✅ Matches modern Playwright 1.57+ behavior
- ✅ Uses Google's official testing browser
- ✅ Exact match to real user Chrome
- ✅ Future-proof (Google's official solution)

**Cons:**
- ❌ More complex implementation (JSON API)
- ❌ Different download infrastructure
- ❌ Need to rewrite download logic
- ❌ May be slower (Google CDN vs Azure Singapore)

## Recommendation

**For now: Keep Chromium (Option 1)**

Reasons:
1. Current implementation works perfectly
2. Build 1140 is stable and tested
3. Azure Singapore provides fast downloads for Vietnam
4. Can switch to Chrome for Testing later if needed

**Future consideration:**
- Monitor if Playwright deprecates old Chromium builds
- Consider switching to Chrome for Testing in future release
- Would require rewriting download logic to use Google's JSON API

## Technical Details

### Current Kexas Download
```
URL: https://playwright.azureedge.net/builds/chromium/1140/chromium-mac-arm64.zip
Size: ~150MB compressed, ~430MB extracted
Contains: Chromium.app (full browser bundle)
Cache: ~/.kexas/browsers/chromium-1140/
```

### Modern Playwright Download (Chrome for Testing)
```
URL: https://storage.googleapis.com/chrome-for-testing-public/{version}/{platform}/chrome-{platform}.zip
Example: .../131.0.6778.87/mac-arm64/chrome-mac-arm64.zip
Contains: Google Chrome.app (full browser bundle)
Cache: ~/Library/Caches/ms-playwright/chrome-{version}/
```

## User's Question Answered

**Question:** "Playwright downloads Chrome binaries to local when a user's machine doesn't have Chrome locally?"

**Answer:** Yes, but it's more specific:

1. **Playwright 1.57+** downloads **Chrome for Testing** (not regular Chrome, not Chromium)
2. It downloads to cache folder: `~/Library/Caches/ms-playwright/`
3. This is separate from your system Chrome in `/Applications/`
4. It's a special non-auto-updating version of Chrome made for testing

**Kexas currently** downloads **Chromium** (older approach) to `~/.kexas/browsers/`

Both approaches work, but modern Playwright uses Chrome for Testing instead of Chromium.
