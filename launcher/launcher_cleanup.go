package launcher

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kexas-project/kexas/internal/logger"
)

// cleanOnce ensures cleanStaleTempProfiles runs exactly once per process,
// preventing a race where one worker's cleanup deletes another worker's
// active temp profile directory.
var cleanOnce sync.Once

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

	// Poll for port release (up to 5 seconds)
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

// isPortAvailable checks whether a TCP port can be bound on localhost.
func isPortAvailable(port int) bool {
	var addr string = fmt.Sprintf("127.0.0.1:%d", port)
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
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

// findFreePort asks the OS to assign an available TCP port.
// It binds to port 0, reads the assigned port, then releases it.
//
// WARNING: This function has a TOCTOU race — the port is free when checked
// but may be taken by another process before Chrome can bind it.
// For parallel launches, use --remote-debugging-port=0 instead (see Launch).
func findFreePort() (int, error) {
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("launcher: failed to find free port: %w", err)
	}
	var port int = listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}
