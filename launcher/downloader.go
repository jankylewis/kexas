package launcher

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kexas-project/kexas/internal/logger"
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

	// Get version info
	var version string
	version, err = getChromeVersion()
	if err != nil {
		return "", err
	}

	var chromeDir string = filepath.Join(cacheDir, fmt.Sprintf("chrome-%s", version))

	// Platform-specific binary path (Chrome for Testing structure)
	var binaryPath string
	switch runtime.GOOS {
	case "darwin":
		// Chrome for Testing uses "Google Chrome for Testing.app"
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

// getChromeVersion fetches the Chrome for Testing version based on ChromeChannel.
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

// downloadChromium downloads and installs Chrome for Testing to the cache directory.
func downloadChromium() (string, error) {
	var log *logger.Logger = logger.New("downloader")

	// Check if already installed
	var installed bool
	var err error
	installed, err = isChromiumInstalled()
	if err != nil {
		return "", err
	}

	if installed {
		log.Debug("chrome for testing already installed")
		return getChromiumPath()
	}

	// Get latest version
	var version string
	version, err = getChromeVersion()
	if err != nil {
		return "", err
	}

	log.Info("downloading chrome for testing", "version", version)

	// Get download URL
	var downloadURL string
	downloadURL, err = getChromiumDownloadURL(version)
	if err != nil {
		return "", err
	}
	log.Debug("download URL", "url", downloadURL)

	// Create cache directory
	var cacheDir string
	cacheDir, err = getBrowserCacheDir()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", fmt.Errorf("launcher: failed to create cache dir: %w", err)
	}

	// Download to temp file
	var tempFile string = filepath.Join(cacheDir, "chrome-download.zip")
	log.Info("downloading to temp file", "path", tempFile)

	err = downloadFile(tempFile, downloadURL)
	if err != nil {
		return "", fmt.Errorf("launcher: download failed: %w", err)
	}
	defer os.Remove(tempFile)

	// Extract
	var chromeDir string = filepath.Join(cacheDir, fmt.Sprintf("chrome-%s", version))
	log.Info("extracting chrome for testing", "dest", chromeDir)

	err = unzip(tempFile, chromeDir)
	if err != nil {
		return "", fmt.Errorf("launcher: extraction failed: %w", err)
	}

	// Make binary executable on Unix
	if runtime.GOOS != "windows" {
		var binaryPath string
		binaryPath, err = getChromiumPath()
		if err != nil {
			return "", err
		}

		err = os.Chmod(binaryPath, 0755)
		if err != nil {
			return "", fmt.Errorf("launcher: failed to make binary executable: %w", err)
		}
	}

	log.Info("chrome for testing installed successfully")
	return getChromiumPath()
}

// InstallChromium is a public function to manually install Chrome for Testing.
func InstallChromium() (string, error) {
	return downloadChromium()
}

// getChromiumDownloadURL returns the download URL for the current platform.
// Fetches from Chrome for Testing JSON API.
func getChromiumDownloadURL(version string) (string, error) {
	var resp *http.Response
	var err error
	resp, err = http.Get(ChromeForTestingAPIURL)
	if err != nil {
		return "", fmt.Errorf("launcher: failed to fetch download URLs: %w", err)
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

	var channelData ChromeChannelData
	var ok bool
	channelData, ok = apiResp.Channels[ChromeReleaseChannel]
	if !ok {
		return "", fmt.Errorf("launcher: %s channel not found", ChromeReleaseChannel)
	}

	var chromeDownloads []ChromeDownload
	chromeDownloads, ok = channelData.Downloads["chrome"]
	if !ok {
		return "", fmt.Errorf("launcher: chrome downloads not found")
	}

	// Determine platform string
	var platform string
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			platform = "mac-arm64"
		} else {
			platform = "mac-x64"
		}
	case "linux":
		platform = "linux64"
	case "windows":
		if runtime.GOARCH == "amd64" {
			platform = "win64"
		} else {
			platform = "win32"
		}
	default:
		return "", fmt.Errorf("launcher: unsupported platform: %s", runtime.GOOS)
	}

	// Find matching download URL
	for _, download := range chromeDownloads {
		if download.Platform == platform {
			return download.URL, nil
		}
	}

	return "", fmt.Errorf("launcher: no download found for platform %s", platform)
}

// downloadFile downloads a file from a URL to a local path.
func downloadFile(filepath string, url string) error {
	var resp *http.Response
	var err error
	resp, err = http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	var out *os.File
	out, err = os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip extracts a zip file to a destination directory.
func unzip(src string, dest string) error {
	var r *zip.ReadCloser
	var err error
	r, err = zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		var fpath string = filepath.Join(dest, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		err = os.MkdirAll(filepath.Dir(fpath), 0755)
		if err != nil {
			return err
		}

		var outFile *os.File
		outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
