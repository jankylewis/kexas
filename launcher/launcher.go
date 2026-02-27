// Package launcher handles finding, launching, and managing Chromium browser processes.
//
// It locates the Chromium binary on the system, starts it with the appropriate
// command-line flags for automation (headless mode, remote debugging port, etc.),
// and extracts the WebSocket debugger URL for CDP communication.
package launcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kexas-project/kexas/internal/logger"
)

// Common launcher errors.
var (
	ErrChromiumNotFound error = errors.New("launcher: chromium binary not found")
	ErrLaunchFailed     error = errors.New("launcher: failed to launch browser")
	ErrNoDebuggerURL    error = errors.New("launcher: could not find debugger URL")
)

// killExistingChromeProcesses terminates any existing Chrome processes using the debugging port.
// Also kills orphaned Chrome for Testing processes left behind by Ctrl+C.
func killExistingChromeProcesses(port int) {
	var log *logger.Logger = logger.New("launcher")
	var killed bool = false

	// Method 1: Kill processes using the specific port via lsof
	var cmd *exec.Cmd = exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	var output []byte
	var err error
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		var pids []string = strings.Split(strings.TrimSpace(string(output)), "\n")
		log.Info("found processes using port", "port", port, "count", len(pids))
		for _, pid := range pids {
			if pid == "" {
				continue
			}
			log.Warn("killing process on port", "port", port, "pid", pid)
			var killCmd *exec.Cmd = exec.Command("kill", "-9", pid)
			var killErr error = killCmd.Run()
			if killErr != nil {
				log.Error("failed to kill process", "pid", pid, "err", killErr)
			} else {
				killed = true
			}
		}
	}

	// Method 2: Kill orphaned "Chrome for Testing" processes by name
	// These may exist after Ctrl+C kills the Go process but leaves Chrome alive
	var pgrepCmd *exec.Cmd = exec.Command("pgrep", "-f", "Google Chrome for Testing")
	var pgrepOut []byte
	var pgrepErr error
	pgrepOut, pgrepErr = pgrepCmd.Output()
	if pgrepErr == nil && len(pgrepOut) > 0 {
		var orphanPids []string = strings.Split(strings.TrimSpace(string(pgrepOut)), "\n")
		log.Warn("found orphaned Chrome for Testing processes", "count", len(orphanPids))
		for _, pid := range orphanPids {
			if pid == "" {
				continue
			}
			log.Warn("killing orphaned Chrome process", "pid", pid)
			var killCmd *exec.Cmd = exec.Command("kill", "-9", pid)
			var killErr error = killCmd.Run()
			if killErr != nil {
				log.Error("failed to kill orphaned process", "pid", pid, "err", killErr)
			} else {
				killed = true
			}
		}
	}

	if !killed {
		log.Info("no existing Chrome processes found on port", "port", port)
		return
	}

	// Poll for port release (up to 5 seconds) — must be long enough for OS to reclaim
	var start time.Time = time.Now()
	for time.Since(start) < 5*time.Second {
		if isPortAvailable(port) {
			log.Info("port released after killing processes", "port", port, "elapsed", time.Since(start))
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.Warn("port may still be in use after killing processes", "port", port)
}

func isPortAvailable(port int) bool {
	var addr string = fmt.Sprintf("127.0.0.1:%d", port)
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		// Port is in use
		return false
	}
	listener.Close()
	return true
}

// cleanStaleTempProfiles removes leftover kexas-chrome-* temp directories
// from previous crashed runs. These stale profiles can cause Chrome to crash
// on relaunch (mach_vm_read errors, shared memory conflicts).
func cleanStaleTempProfiles(log *logger.Logger) {
	var tmpDir string = os.TempDir()
	var entries []os.DirEntry
	var err error
	entries, err = os.ReadDir(tmpDir)
	if err != nil {
		return
	}

	var cleaned int = 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "kexas-chrome-") {
			continue
		}
		var fullPath string = filepath.Join(tmpDir, entry.Name())
		var removeErr error = os.RemoveAll(fullPath)
		if removeErr != nil {
			log.Debug("failed to clean stale profile", "path", fullPath, "err", removeErr)
		} else {
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Info("cleaned stale temp profiles", "count", cleaned)
	}
}

// Options configures browser launch behavior.
type Options struct {
	Headless       bool
	Port           int
	Args           []string
	ExecutablePath string
}

// DefaultOptions returns sensible default launch options.
func DefaultOptions() *Options {
	return &Options{
		Headless: true,
		Port:     9222,
		Args:     []string{},
	}
}

// Browser represents a launched browser process.
type Browser struct {
	cmd         *exec.Cmd
	wsURL       string
	log         *logger.Logger
	cancelFunc  context.CancelFunc
	userDataDir string // Temporary profile directory to clean up
	port        int    // Debugging port number
}

// Launch starts a Chromium browser with the given options.
func Launch(ctx context.Context, opts *Options) (*Browser, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	var log *logger.Logger = logger.New("launcher")

	// Find Chromium executable
	var execPath string
	var err error
	if opts.ExecutablePath != "" {
		execPath = opts.ExecutablePath
	} else {
		execPath, err = findChromium()
		if err != nil {
			log.Error("chromium not found", "err", err)
			return nil, err
		}
	}

	log.Debug("found chromium", "path", execPath)

	// Build command-line arguments and get the temp profile directory
	var args []string
	var userDataDir string
	args, userDataDir = buildArgs(opts)

	// Create command with cancellable context
	var cmdCtx context.Context
	var cancel context.CancelFunc
	cmdCtx, cancel = context.WithCancel(ctx)

	var cmd *exec.Cmd = exec.CommandContext(cmdCtx, execPath, args...)

	// Capture stdout to extract WebSocket URL
	var stdout *os.File
	stdout, err = os.CreateTemp("", "kexas-chrome-*.log")
	if err != nil {
		cancel()
		return nil, fmt.Errorf("launcher: failed to create temp file: %w", err)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	log.Info("launching chromium", "headless", opts.Headless, "port", opts.Port)

	// Kill any existing Chrome processes using the debugging port
	log.Info("step 1: killing existing Chrome processes on port", "port", opts.Port)
	killExistingChromeProcesses(opts.Port)

	// Clean up stale temp profiles from previous crashed runs
	cleanStaleTempProfiles(log)

	// Settle delay: let the OS fully release shared memory after killing processes
	// and removing stale profiles. Without this, Chrome may hit mach_vm_read errors.
	time.Sleep(500 * time.Millisecond)

	// Check if port is available before attempting to launch
	log.Info("step 2: checking if port is available", "port", opts.Port)
	if !isPortAvailable(opts.Port) {
		cancel()
		stdout.Close()
		log.Error("port still in use after cleanup", "port", opts.Port)
		return nil, fmt.Errorf("launcher: port %d is already in use - kill existing Chrome processes or wait for port release", opts.Port)
	}
	log.Info("step 3: port is available, starting Chrome", "port", opts.Port)

	// Start the browser process
	err = cmd.Start()
	if err != nil {
		cancel()
		stdout.Close()
		log.Error("failed to start chromium", "err", err)
		return nil, fmt.Errorf("%w: %v", ErrLaunchFailed, err)
	}

	log.Debug("chromium process started", "pid", cmd.Process.Pid)

	// Extract WebSocket URL from stdout
	var wsURL string
	wsURL, err = extractDebuggerURL(stdout, 10*time.Second, log)
	stdout.Close()
	if err != nil {
		cancel()
		cmd.Process.Kill()
		log.Error("failed to extract debugger URL", "err", err)
		return nil, err
	}

	log.Info("chromium launched successfully", "wsURL", wsURL)

	var browser *Browser = &Browser{
		cmd:         cmd,
		wsURL:       wsURL,
		log:         log,
		cancelFunc:  cancel,
		userDataDir: userDataDir,
		port:        opts.Port,
	}

	return browser, nil
}

// WebSocketURL returns the CDP WebSocket debugger URL.
func (b *Browser) WebSocketURL() string {
	return b.wsURL
}

// Close terminates the browser process and cleans up resources.
func (b *Browser) Close() error {
	b.log.Debug("closing browser")

	b.cancelFunc()

	var err error = b.cmd.Process.Kill()
	if err != nil {
		b.log.Error("failed to kill browser process", "err", err)
		return fmt.Errorf("launcher: failed to kill process: %w", err)
	}

	b.cmd.Wait()

	// Poll for port release with 3s timeout and 150ms intervals
	var portTimeout time.Duration = 3 * time.Second
	var pollInterval time.Duration = 150 * time.Millisecond
	var start time.Time = time.Now()

	b.log.Debug("waiting for port to be released", "port", b.port)

	for time.Since(start) < portTimeout {
		if isPortAvailable(b.port) {
			b.log.Debug("port released successfully", "port", b.port, "elapsed", time.Since(start))
			break
		}
		time.Sleep(pollInterval)
	}

	// If port is still in use after timeout, log a warning but continue
	if !isPortAvailable(b.port) {
		b.log.Warn("port still in use after timeout", "port", b.port, "timeout", portTimeout)
	}

	// Clean up temporary profile directory
	if b.userDataDir != "" {
		b.log.Debug("cleaning up temp profile", "path", b.userDataDir)
		var cleanupErr error = os.RemoveAll(b.userDataDir)
		if cleanupErr != nil {
			b.log.Warn("failed to clean up temp profile", "err", cleanupErr)
			// Don't return error, browser is already closed
		} else {
			b.log.Debug("temp profile cleaned up")
		}
	}

	b.log.Info("browser closed")
	return nil
}

// findChromium locates the Chromium executable on the system.
func findChromium() (string, error) {
	var log *logger.Logger = logger.New("launcher")

	// First, try to use cached Chromium
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

	// Try to download Chromium
	log.Info("cached chromium not found, attempting download")
	var downloadedPath string
	downloadedPath, err = downloadChromium()
	if err == nil {
		log.Info("chromium downloaded successfully", "path", downloadedPath)
		return downloadedPath, nil
	}
	log.Warn("failed to download chromium, falling back to system browser", "err", err)

	// Fall back to system-installed browsers
	log.Debug("searching for system-installed browsers")
	var paths []string

	// macOS paths
	paths = []string{
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
func buildArgs(opts *Options) ([]string, string) {
	// Use a temporary profile directory for complete isolation
	// Each test run gets a fresh browser with no history, cookies, or state
	var timestamp string = fmt.Sprintf("%d", time.Now().UnixNano())
	var userDataDir string = filepath.Join(os.TempDir(), fmt.Sprintf("kexas-chrome-%s", timestamp))

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
		"--window-size=1920,1080",
	}

	if opts.Headless {
		args = append(args, "--headless=new")
	}

	// Add user-provided args
	args = append(args, opts.Args...)

	return args, userDataDir
}

// extractDebuggerURL reads the browser stdout and extracts the WebSocket URL.
func extractDebuggerURL(file *os.File, timeout time.Duration, log *logger.Logger) (string, error) {
	var deadline time.Time = time.Now().Add(timeout)
	var pattern *regexp.Regexp = regexp.MustCompile(`ws://[^\s]+`)

	for time.Now().Before(deadline) {
		file.Seek(0, 0)
		var scanner *bufio.Scanner = bufio.NewScanner(file)

		for scanner.Scan() {
			var line string = scanner.Text()
			// Check for DevTools URL
			if strings.Contains(line, "DevTools listening on") {
				var match string = pattern.FindString(line)
				if match != "" {
					return match, nil
				}
			}
			// Log errors from Chrome
			if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
				log.Error("chrome error", "msg", line)
			}
			// Log if port binding failed
			if strings.Contains(line, "bind") && strings.Contains(line, "address already in use") {
				return "", fmt.Errorf("chrome cannot bind to port - address already in use: %s", line)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Timeout - read what we have for diagnostics
	file.Seek(0, 0)
	var scanner *bufio.Scanner = bufio.NewScanner(file)
	var lastLines []string
	for scanner.Scan() && len(lastLines) < 20 {
		lastLines = append(lastLines, scanner.Text())
	}
	log.Error("timeout waiting for debugger URL", "lastOutput", lastLines)

	return "", ErrNoDebuggerURL
}
