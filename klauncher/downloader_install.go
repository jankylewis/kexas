package klauncher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jankylewis/kexas/internal/logger"
)

// downloadChromium downloads and installs Chrome for Testing to the cache directory.
func downloadChromium() (string, error) {
	var log *logger.Logger = logger.New("downloader")

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

	var version string
	version, err = getChromeVersion()
	if err != nil {
		return "", err
	}
	log.Info("downloading chrome for testing", "version", version)

	var downloadURL string
	downloadURL, err = getChromiumDownloadURL(version)
	if err != nil {
		return "", err
	}
	log.Debug("download URL", "url", downloadURL)

	var cacheDir string
	cacheDir, err = ensureCacheDir()
	if err != nil {
		return "", err
	}

	err = downloadAndExtract(downloadURL, cacheDir, version, log)
	if err != nil {
		return "", err
	}

	err = makeBinaryExecutableOnUnix()
	if err != nil {
		return "", err
	}

	log.Info("chrome for testing installed successfully")
	return getChromiumPath()
}

// ensureCacheDir resolves the browser cache directory and creates it if missing.
func ensureCacheDir() (string, error) {
	var cacheDir string
	var err error
	cacheDir, err = getBrowserCacheDir()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", fmt.Errorf("launcher: failed to create cache dir: %w", err)
	}
	return cacheDir, nil
}

// downloadAndExtract fetches the Chrome zip into a temp file in cacheDir and unzips
// it into chrome-<version>/. The temp zip is removed when the function returns.
func downloadAndExtract(downloadURL, cacheDir, version string, log *logger.Logger) error {
	var tempFile string = filepath.Join(cacheDir, "chrome-download.zip")
	log.Info("downloading to temp file", "path", tempFile)

	var err error = downloadFile(tempFile, downloadURL)
	if err != nil {
		return fmt.Errorf("launcher: download failed: %w", err)
	}
	defer os.Remove(tempFile)

	var chromeDir string = filepath.Join(cacheDir, fmt.Sprintf("chrome-%s", version))
	log.Info("extracting chrome for testing", "dest", chromeDir)

	err = unzip(tempFile, chromeDir)
	if err != nil {
		return fmt.Errorf("launcher: extraction failed: %w", err)
	}
	return nil
}

// makeBinaryExecutableOnUnix chmods the Chromium binary to 0755 on non-Windows platforms.
// No-op on Windows where +x is not meaningful.
func makeBinaryExecutableOnUnix() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	var binaryPath string
	var err error
	binaryPath, err = getChromiumPath()
	if err != nil {
		return err
	}
	err = os.Chmod(binaryPath, 0755)
	if err != nil {
		return fmt.Errorf("launcher: failed to make binary executable: %w", err)
	}
	return nil
}

// InstallChromium is a public function to manually install Chrome for Testing.
func InstallChromium() (string, error) {
	return downloadChromium()
}
