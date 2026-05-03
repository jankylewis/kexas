package klauncher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// ChromeForTestingAPIURL is the JSON API endpoint for Chrome for Testing.
const ChromeForTestingAPIURL string = "https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions-with-downloads.json"

// ChromeReleaseChannel specifies which Chrome release channel to use.
// Options: "Stable", "Beta", "Dev", "Canary"
// Default: "Stable"
const ChromeReleaseChannel string = "Stable"

// ChromeForTestingResponse represents the JSON API response structure.
type ChromeForTestingResponse struct {
	Channels map[string]ChromeChannelData `json:"channels"`
}

// ChromeChannelData represents a Chrome release channel.
type ChromeChannelData struct {
	Version   string                      `json:"version"`
	Revision  string                      `json:"revision"`
	Downloads map[string][]ChromeDownload `json:"downloads"`
}

// ChromeDownload represents a download URL for a specific platform.
type ChromeDownload struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

// getBrowserCacheDir returns the directory where browsers are cached.
func getBrowserCacheDir() (string, error) {
	var homeDir string = os.Getenv("HOME")
	if homeDir == "" {
		return "", fmt.Errorf("launcher: HOME environment variable not set")
	}

	var cacheDir string = filepath.Join(homeDir, ".kexas", "browsers")
	return cacheDir, nil
}

// getChromiumPath returns the path to the cached Chrome for Testing binary.
func getChromiumPath() (string, error) {
	var cacheDir string
	var err error
	cacheDir, err = getBrowserCacheDir()
	if err != nil {
		return "", err
	}

	var version string
	version, err = getChromeVersion()
	if err != nil {
		return "", err
	}

	var chromeDir string = filepath.Join(cacheDir, fmt.Sprintf("chrome-%s", version))

	var binaryPath string
	switch runtime.GOOS {
	case "darwin":
		// Chrome for Testing app bundle layout
		binaryPath = filepath.Join(chromeDir, "chrome-mac-arm64", "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing")
		if runtime.GOARCH != "arm64" {
			binaryPath = filepath.Join(chromeDir, "chrome-mac-x64", "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing")
		}
	case "linux":
		binaryPath = filepath.Join(chromeDir, "chrome-linux64", "chrome")
	case "windows":
		binaryPath = filepath.Join(chromeDir, "chrome-win64", "chrome.exe")
	default:
		return "", fmt.Errorf("launcher: unsupported platform: %s", runtime.GOOS)
	}

	return binaryPath, nil
}

// isChromiumInstalled checks if Chromium is already cached.
func isChromiumInstalled() (bool, error) {
	var path string
	var err error
	path, err = getChromiumPath()
	if err != nil {
		return false, err
	}

	var info os.FileInfo
	info, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// getChromeVersion fetches the Chrome for Testing version based on ChromeReleaseChannel.
func getChromeVersion() (string, error) {
	var resp *http.Response
	var err error
	resp, err = http.Get(ChromeForTestingAPIURL)
	if err != nil {
		return "", fmt.Errorf("launcher: failed to fetch Chrome version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("launcher: API returned status %s", resp.Status)
	}

	var apiResp ChromeForTestingResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return "", fmt.Errorf("launcher: failed to parse API response: %w", err)
	}

	var channel ChromeChannelData
	var ok bool
	channel, ok = apiResp.Channels[ChromeReleaseChannel]
	if !ok {
		return "", fmt.Errorf("launcher: %s channel not found in API response", ChromeReleaseChannel)
	}

	return channel.Version, nil
}
