package klauncher

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

// findChromium locates the Chromium executable: tries the kexas-managed cache
// first, then auto-downloads, then falls back to system-installed browsers.
func findChromium() (string, error) {
	var log *logger.Logger = logger.New("launcher")
	if path, ok := tryCachedChromium(log); ok {
		return path, nil
	}
	if path, ok := tryDownloadChromium(log); ok {
		return path, nil
	}
	return findSystemChromium(log)
}

// tryCachedChromium returns the kexas-managed Chromium path if installed.
func tryCachedChromium(log *logger.Logger) (string, bool) {
	log.Debug("checking for cached chromium")
	var installed bool
	var err error
	installed, err = isChromiumInstalled()
	if err != nil {
		log.Warn("failed to check cached chromium", "err", err)
		return "", false
	}
	if !installed {
		return "", false
	}
	var cachedPath string
	cachedPath, err = getChromiumPath()
	if err != nil {
		return "", false
	}
	log.Info("using cached chromium", "path", cachedPath)
	return cachedPath, true
}

// tryDownloadChromium attempts a fresh kexas-managed download.
func tryDownloadChromium(log *logger.Logger) (string, bool) {
	log.Info("cached chromium not found, attempting download")
	var path string
	var err error
	path, err = downloadChromium()
	if err != nil {
		log.Warn("failed to download chromium, falling back to system browser", "err", err)
		return "", false
	}
	log.Info("chromium downloaded successfully", "path", path)
	return path, true
}

// findSystemChromium scans well-known macOS install paths for a usable browser.
func findSystemChromium(log *logger.Logger) (string, error) {
	log.Debug("searching for system-installed browsers")
	var paths []string = systemChromiumCandidatePaths()
	for _, path := range paths {
		var info os.FileInfo
		var err error
		info, err = os.Stat(path)
		if err == nil && !info.IsDir() {
			log.Info("found system browser", "path", path)
			return path, nil
		}
	}
	return "", ErrChromiumNotFound
}

// systemChromiumCandidatePaths returns macOS install locations to probe.
func systemChromiumCandidatePaths() []string {
	return []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		filepath.Join(os.Getenv("HOME"), "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
		filepath.Join(os.Getenv("HOME"), "Applications/Chromium.app/Contents/MacOS/Chromium"),
	}
}

// buildArgs constructs the command-line arguments for launching Chromium and
// the temporary user-data dir each invocation gets (fresh state per run).
func buildArgs(opts *Options) ([]string, string) {
	var userDataDir string = generateUniqueUserDataDir()
	var width, height int = resolveWindowSize(opts)
	var args []string = baseChromeArgs(opts.Port, userDataDir, width, height)
	if opts.Headless {
		args = append(args, "--headless=new")
	}
	args = append(args, opts.Args...)
	return args, userDataDir
}

// generateUniqueUserDataDir returns a fresh temp profile path per call. The
// UnixNano + 8 crypto-random bytes pattern guarantees uniqueness even when
// many goroutines race in the same nanosecond.
func generateUniqueUserDataDir() string {
	var randomBytes [8]byte
	cryptorand.Read(randomBytes[:])
	var suffix string = fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes[:]))
	return filepath.Join(os.TempDir(), fmt.Sprintf("kexas-chrome-%s", suffix))
}

// resolveWindowSize returns (width, height) defaulting to 1280x720 when opts
// don't specify one.
func resolveWindowSize(opts *Options) (int, int) {
	var width int = opts.WindowWidth
	if width <= 0 {
		width = 1280
	}
	var height int = opts.WindowHeight
	if height <= 0 {
		height = 720
	}
	return width, height
}

// baseChromeArgs returns the static set of Chromium flags kexas always passes.
// Headless and user-supplied flags are layered on top by buildArgs.
func baseChromeArgs(port int, userDataDir string, width, height int) []string {
	return []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		fmt.Sprintf("--user-data-dir=%s", userDataDir),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-breakpad",
		"--disable-client-side-phishing-detection",
		"--disable-component-extensions-with-background-pages",
		"--disable-default-apps",
		"--disable-dev-shm-usage",
		"--disable-extensions",
		"--disable-gpu",
		"--disable-features=TranslateUI",
		"--disable-hang-monitor",
		"--disable-ipc-flooding-protection",
		"--disable-popup-blocking",
		"--disable-prompt-on-repost",
		"--disable-renderer-backgrounding",
		"--disable-sync",
		"--force-color-profile=srgb",
		"--metrics-recording-only",
		"--no-service-autorun",
		"--password-store=basic",
		"--use-mock-keychain",
		"--disable-infobars",
		"--disable-crash-reporter",
		fmt.Sprintf("--window-size=%d,%d", width, height),
	}
}
