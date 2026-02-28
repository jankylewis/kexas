// Package launcher handles finding, launching, and managing Chromium browser processes.
//
// It locates the Chromium binary on the system, starts it with the appropriate
// command-line flags for automation (headless mode, remote debugging port, etc.),
// and extracts the WebSocket debugger URL for CDP communication.
package launcher

import (
	"bufio"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

// Options configures browser launch behavior.
type Options struct {
	Headless       bool
	Port           int
	Args           []string
	ExecutablePath string
	WindowWidth    int
	WindowHeight   int
}

// DefaultOptions returns sensible default launch options.
// Port defaults to 0, which means a free port is automatically assigned.
func DefaultOptions() *Options {
	return &Options{
		Headless:     true,
		Port:         0,
		Args:         []string{},
		WindowWidth:  1280,
		WindowHeight: 720,
	}
}

// Browser represents a launched browser process.
type Browser struct {
	cmd         *exec.Cmd
	wsURL       string
	log         *logger.Logger
	cancelFunc  context.CancelFunc
	userDataDir string // Temporary profile directory to clean up
	logFile     string // Temporary log file to clean up
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

	// CRITICAL: When Port == 0, pass it directly to Chrome as --remote-debugging-port=0.
	// Chrome will atomically pick its own free port and print it in stdout.
	// This eliminates the TOCTOU race where findFreePort() releases a port that
	// another worker or OS process grabs before Chrome can bind it.
	if opts.Port == 0 {
		log.Info("using Chrome self-assigned port (--remote-debugging-port=0)")
	} else {
		log.Info("using fixed port", "port", opts.Port)
	}

	// Build command-line arguments and get the temp profile directory
	var args []string
	var userDataDir string
	args, userDataDir = buildArgs(opts)

	// Create command with cancellable context
	var cmdCtx context.Context
	var cancel context.CancelFunc
	cmdCtx, cancel = context.WithCancel(ctx)

	var cmd *exec.Cmd = exec.CommandContext(cmdCtx, execPath, args...)

	// Use pipes to capture Chrome output in real-time.
	// Chrome prints "DevTools listening on ws://..." to stderr.
	// File-based capture has buffering lag that caused 10s timeouts under
	// parallel load. Pipe reading eliminates this entirely.
	var stderrPipe io.ReadCloser
	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("launcher: failed to create stderr pipe: %w", err)
	}
	cmd.Stdout = nil // Chrome outputs DevTools URL to stderr only

	log.Info("launching chromium", "headless", opts.Headless, "port", opts.Port)

	// Clean up stale temp profiles from previous crashed runs.
	// Uses sync.Once so this only runs on the FIRST Launch() call in the process.
	// Running it on every Launch() caused a race: worker A creates its temp dir,
	// then worker B's cleanup deletes it before Chrome can write SingletonLock.
	cleanOnce.Do(func() { cleanStaleTempProfiles(log) })

	// Only kill existing processes and wait if using a fixed (user-specified) port.
	// When port=0, Chrome self-assigns — no conflict possible, skip this step.
	if opts.Port != 0 {
		log.Info("step 1: killing existing Chrome processes on port", "port", opts.Port)
		killExistingChromeProcesses(opts.Port)

		// Settle delay for fixed port cleanup
		time.Sleep(500 * time.Millisecond)

		log.Info("step 2: checking if port is available", "port", opts.Port)
		if !isPortAvailable(opts.Port) {
			cancel()
			log.Error("port still in use after cleanup", "port", opts.Port)
			return nil, fmt.Errorf("launcher: port %d is already in use - kill existing Chrome processes or wait for port release", opts.Port)
		}
	}
	log.Info("starting Chrome", "port", opts.Port)

	// Start the browser process
	err = cmd.Start()
	if err != nil {
		cancel()
		log.Error("failed to start chromium", "err", err)
		return nil, fmt.Errorf("%w: %v", ErrLaunchFailed, err)
	}

	log.Debug("chromium process started", "pid", cmd.Process.Pid)

	// Extract WebSocket URL from stderr pipe in real-time
	var wsURL string
	wsURL, err = extractDebuggerURLFromPipe(stderrPipe, 10*time.Second, log)
	if err != nil {
		cancel()
		cmd.Process.Kill()
		log.Error("failed to extract debugger URL", "err", err)
		return nil, err
	}

	// Extract actual port from wsURL for cleanup (especially when port=0)
	var actualPort int = opts.Port
	if actualPort == 0 {
		actualPort = extractPortFromWSURL(wsURL)
		log.Info("Chrome self-assigned port", "port", actualPort)
	}

	log.Info("chromium launched successfully", "wsURL", wsURL)

	var browser *Browser = &Browser{
		cmd:         cmd,
		wsURL:       wsURL,
		log:         log,
		cancelFunc:  cancel,
		userDataDir: userDataDir,
		logFile:     "", // No log file — using pipe-based stderr capture
		port:        actualPort,
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
		} else {
			b.log.Debug("temp profile cleaned up")
		}
	}

	// Clean up temporary log file
	if b.logFile != "" {
		var logErr error = os.Remove(b.logFile)
		if logErr != nil && !os.IsNotExist(logErr) {
			b.log.Debug("failed to clean up log file", "err", logErr)
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
	// Combine UnixNano + 8 crypto-random bytes to guarantee uniqueness even
	// when multiple goroutines call buildArgs at the same nanosecond.
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

	// Add user-provided args
	args = append(args, opts.Args...)

	return args, userDataDir
}

// extractPortFromWSURL extracts the port number from a WebSocket URL.
// Example: "ws://127.0.0.1:51234/devtools/browser/abc" → 51234
func extractPortFromWSURL(wsURL string) int {
	// Pattern: ws://host:PORT/path
	var portPattern *regexp.Regexp = regexp.MustCompile(`:(\d+)/`)
	var match []string = portPattern.FindStringSubmatch(wsURL)
	if len(match) < 2 {
		return 0
	}
	var port int
	fmt.Sscanf(match[1], "%d", &port)
	return port
}

// extractDebuggerURLFromPipe reads Chrome's stderr via a pipe in real-time.
// This is significantly more reliable than file-based polling because:
// 1. No OS file buffer lag — lines arrive immediately via the pipe.
// 2. No seek/re-read overhead — each line is read exactly once.
// 3. No race between Chrome writing and our code reading.
func extractDebuggerURLFromPipe(pipe io.ReadCloser, timeout time.Duration, log *logger.Logger) (string, error) {
	type result struct {
		url string
		err error
	}
	var ch chan result = make(chan result, 1)

	go func() {
		var scanner *bufio.Scanner = bufio.NewScanner(pipe)
		var pattern *regexp.Regexp = regexp.MustCompile(`ws://[^\s]+`)
		var lastLines []string

		for scanner.Scan() {
			var line string = scanner.Text()

			// Keep last lines for diagnostics on timeout
			if len(lastLines) < 20 {
				lastLines = append(lastLines, line)
			}

			// Check for DevTools URL
			if strings.Contains(line, "DevTools listening on") {
				var match string = pattern.FindString(line)
				if match != "" {
					ch <- result{url: match}
					return
				}
			}
			// Log Chrome errors
			if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
				log.Error("chrome error", "msg", line)
			}
			// Detect port bind failure
			if strings.Contains(line, "bind") && strings.Contains(line, "address already in use") {
				ch <- result{err: fmt.Errorf("chrome cannot bind to port - address already in use: %s", line)}
				return
			}
		}

		// Scanner finished without finding URL (Chrome exited or pipe closed)
		if scanner.Err() != nil {
			ch <- result{err: fmt.Errorf("pipe read error: %w", scanner.Err())}
		} else {
			log.Error("chrome stderr closed without DevTools URL", "lastOutput", lastLines)
			ch <- result{err: ErrNoDebuggerURL}
		}
	}()

	select {
	case res := <-ch:
		return res.url, res.err
	case <-time.After(timeout):
		log.Error("timeout waiting for debugger URL from pipe")
		return "", ErrNoDebuggerURL
	}
}
