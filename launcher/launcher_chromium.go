package launcher

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

// findChromium locates the Chromium executable on the system.
func findChromium() (string, error) {
	var log *logger.Logger = logger.New("launcher")

	log.Debug("checking for cached chromium")
	var installed bool
	var err error
	installed, err = isChromiumInstalled()
	if err != nil {
		log.Warn("failed to check cached chromium", "err", err)
	} else if installed {
		var cachedPath string
		cachedPath, err = getChromiumPath()
		if err == nil {
			log.Info("using cached chromium", "path", cachedPath)
			return cachedPath, nil
		}
	}

	log.Info("cached chromium not found, attempting download")
	var downloadedPath string
	downloadedPath, err = downloadChromium()
	if err == nil {
		log.Info("chromium downloaded successfully", "path", downloadedPath)
		return downloadedPath, nil
	}
	log.Warn("failed to download chromium, falling back to system browser", "err", err)

	log.Debug("searching for system-installed browsers")
	var paths []string = []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		filepath.Join(os.Getenv("HOME"), "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
		filepath.Join(os.Getenv("HOME"), "Applications/Chromium.app/Contents/MacOS/Chromium"),
	}

	for _, path := range paths {
		var info os.FileInfo
		info, err = os.Stat(path)
		if err == nil && !info.IsDir() {
			log.Info("found system browser", "path", path)
			return path, nil
		}
	}

	return "", ErrChromiumNotFound
}

// buildArgs constructs the command-line arguments for launching Chromium.
// Returns the args slice and the temporary user data directory path.
//
// Each test run gets a fresh browser with no history/cookies/state via a unique
// temp profile. UnixNano + 8 crypto-random bytes guarantees uniqueness even when
// multiple goroutines call buildArgs at the same nanosecond.
func buildArgs(opts *Options) ([]string, string) {
	var randomBytes [8]byte
	cryptorand.Read(randomBytes[:])
	var suffix string = fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes[:]))
	var userDataDir string = filepath.Join(os.TempDir(), fmt.Sprintf("kexas-chrome-%s", suffix))

	var width int = opts.WindowWidth
	if width <= 0 {
		width = 1280
	}
	var height int = opts.WindowHeight
	if height <= 0 {
		height = 720
	}

	var args []string = []string{
		fmt.Sprintf("--remote-debugging-port=%d", opts.Port),
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

	if opts.Headless {
		args = append(args, "--headless=new")
	}

	args = append(args, opts.Args...)

	return args, userDataDir
}
