package klauncher

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

// cleanOnce ensures cleanStaleTempProfiles runs exactly once per process,
// preventing a race where one worker's cleanup deletes another worker's
// active temp profile directory.
var cleanOnce sync.Once

// killExistingChromeProcesses terminates any existing Chrome processes using the debugging port.
// Also kills orphaned Chrome for Testing processes left behind by Ctrl+C.
func killExistingChromeProcesses(port int) {
	var log *logger.Logger = logger.New("launcher")
	var killed bool

	if killProcessesUsingPort(port, log) {
		killed = true
	}
	if killOrphanedChromeForTesting(log) {
		killed = true
	}

	if !killed {
		log.Info("no existing Chrome processes found on port", "port", port)
		return
	}

	waitForPortRelease(port, 5*time.Second, log)
}

// killProcessesUsingPort kills every PID that lsof reports as bound to port.
// Returns true if any kill succeeded.
func killProcessesUsingPort(port int, log *logger.Logger) bool {
	var cmd *exec.Cmd = exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	var output []byte
	var err error
	output, err = cmd.Output()
	if err != nil || len(output) == 0 {
		return false
	}
	var pids []string = strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Info("found processes using port", "port", port, "count", len(pids))

	var anyKilled bool
	for _, pid := range pids {
		if pid == "" {
			continue
		}
		log.Warn("killing process on port", "port", port, "pid", pid)
		var killErr error = exec.Command("kill", "-9", pid).Run()
		if killErr != nil {
			log.Error("failed to kill process", "pid", pid, "err", killErr)
			continue
		}
		anyKilled = true
	}
	return anyKilled
}

// killOrphanedChromeForTesting kills "Google Chrome for Testing" processes that may
// have been left behind by Ctrl+C-ing the Go process without Chrome's cleanup running.
func killOrphanedChromeForTesting(log *logger.Logger) bool {
	var cmd *exec.Cmd = exec.Command("pgrep", "-f", "Google Chrome for Testing")
	var output []byte
	var err error
	output, err = cmd.Output()
	if err != nil || len(output) == 0 {
		return false
	}
	var pids []string = strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Warn("found orphaned Chrome for Testing processes", "count", len(pids))

	var anyKilled bool
	for _, pid := range pids {
		if pid == "" {
			continue
		}
		log.Warn("killing orphaned Chrome process", "pid", pid)
		var killErr error = exec.Command("kill", "-9", pid).Run()
		if killErr != nil {
			log.Error("failed to kill orphaned process", "pid", pid, "err", killErr)
			continue
		}
		anyKilled = true
	}
	return anyKilled
}

// waitForPortRelease polls until isPortAvailable(port) returns true or timeout elapses.
func waitForPortRelease(port int, timeout time.Duration, log *logger.Logger) {
	var start time.Time = time.Now()
	for time.Since(start) < timeout {
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
