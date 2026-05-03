// Package launcher handles finding, launching, and managing Chromium browser processes.
//
// It locates the Chromium binary on the system, starts it with the appropriate
// command-line flags for automation (headless mode, remote debugging port, etc.),
// and extracts the WebSocket debugger URL for CDP communication.
package klauncher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
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

	b.waitForPortRelease()
	b.cleanupTempProfile()
	b.cleanupLogFile()

	b.log.Info("browser closed")
	return nil
}

// waitForPortRelease polls until the debugging port is free or 3s elapses.
func (b *Browser) waitForPortRelease() {
	var portTimeout time.Duration = 3 * time.Second
	var pollInterval time.Duration = 150 * time.Millisecond
	var start time.Time = time.Now()

	b.log.Debug("waiting for port to be released", "port", b.port)

	for time.Since(start) < portTimeout {
		if isPortAvailable(b.port) {
			b.log.Debug("port released successfully", "port", b.port, "elapsed", time.Since(start))
			return
		}
		time.Sleep(pollInterval)
	}
	b.log.Warn("port still in use after timeout", "port", b.port, "timeout", portTimeout)
}

// cleanupTempProfile removes the per-launch temp profile directory.
func (b *Browser) cleanupTempProfile() {
	if b.userDataDir == "" {
		return
	}
	b.log.Debug("cleaning up temp profile", "path", b.userDataDir)
	var cleanupErr error = os.RemoveAll(b.userDataDir)
	if cleanupErr != nil {
		b.log.Warn("failed to clean up temp profile", "err", cleanupErr)
		return
	}
	b.log.Debug("temp profile cleaned up")
}

// cleanupLogFile removes the per-launch temp log file (if one was created).
func (b *Browser) cleanupLogFile() {
	if b.logFile == "" {
		return
	}
	var logErr error = os.Remove(b.logFile)
	if logErr != nil && !os.IsNotExist(logErr) {
		b.log.Debug("failed to clean up log file", "err", logErr)
	}
}
