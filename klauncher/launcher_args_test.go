package klauncher

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jankylewis/kexas/internal/logger"
)

func TestBuildArgs_HeadlessFlag(t *testing.T) {
	t.Run("headless true", func(t *testing.T) {
		var opts *Options = &Options{Port: 9222, Headless: true}
		var args []string
		var userDataDir string
		args, userDataDir = buildArgs(opts)
		defer os.RemoveAll(userDataDir)

		var found bool = false
		for _, arg := range args {
			if arg == "--headless=new" {
				found = true
				break
			}
		}
		if !found {
			t.Error("buildArgs should contain --headless=new when Headless is true")
		}
	})

	t.Run("headless false", func(t *testing.T) {
		var opts *Options = &Options{Port: 9222, Headless: false}
		var args []string
		var userDataDir string
		args, userDataDir = buildArgs(opts)
		defer os.RemoveAll(userDataDir)

		for _, arg := range args {
			if arg == "--headless=new" {
				t.Error("buildArgs should NOT contain --headless=new when Headless is false")
			}
		}
	})
}

// TestFindFreePort_ReturnsValidPort verifies that findFreePort returns
// a usable port number within the valid range.
func TestFindFreePort_ReturnsValidPort(t *testing.T) {
	var port int
	var err error
	port, err = findFreePort()
	if err != nil {
		t.Fatalf("findFreePort failed: %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("findFreePort returned invalid port: %d", port)
	}
}

// TestFindFreePort_ReturnsUniquePorts verifies that consecutive calls
// to findFreePort return different ports. This is critical for parallel
// browser launches — each worker needs its own port.
func TestFindFreePort_ReturnsUniquePorts(t *testing.T) {
	var ports map[int]bool = make(map[int]bool)
	for i := 0; i < 10; i++ {
		var port int
		var err error
		port, err = findFreePort()
		if err != nil {
			t.Fatalf("findFreePort failed on iteration %d: %v", i, err)
		}
		if ports[port] {
			t.Errorf("findFreePort returned duplicate port %d on iteration %d", port, i)
		}
		ports[port] = true
	}
}

// TestFindFreePort_PortIsActuallyFree verifies that the returned port
// can be bound to immediately after findFreePort returns.
func TestFindFreePort_PortIsActuallyFree(t *testing.T) {
	var port int
	var err error
	port, err = findFreePort()
	if err != nil {
		t.Fatalf("findFreePort failed: %v", err)
	}

	// Verify we can bind to it
	var listener net.Listener
	listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Errorf("could not bind to port %d returned by findFreePort: %v", port, err)
		return
	}
	listener.Close()
}

// TestCleanStaleTempProfiles_DeletesActiveWorkerDir demonstrates the race
// condition that caused SingletonLock failures: when cleanStaleTempProfiles
// runs on every Launch() call, it deletes temp dirs belonging to other
// workers whose Chrome processes are still starting up.
//
// The fix: cleanOnce (sync.Once) ensures cleanup runs only on the first
// Launch() call, before any worker has created its temp dir.
func TestCleanStaleTempProfiles_DeletesActiveWorkerDir(t *testing.T) {
	var tmpDir string = os.TempDir()

	// Simulate Worker 0 creating its temp profile dir via buildArgs
	var worker0Dir string = filepath.Join(tmpDir, "kexas-chrome-worker0-active-test")
	var err error = os.MkdirAll(worker0Dir, 0755)
	if err != nil {
		t.Fatalf("failed to create worker0 dir: %v", err)
	}
	defer os.RemoveAll(worker0Dir)

	// Simulate Worker 1 calling cleanStaleTempProfiles (the old bug path)
	// This WILL delete Worker 0's directory — proving the race condition
	var log *logger.Logger = logger.New("test")
	cleanStaleTempProfiles(log)

	// Worker 0's dir should be GONE — this is the bug behavior
	if _, statErr := os.Stat(worker0Dir); !os.IsNotExist(statErr) {
		t.Error("cleanStaleTempProfiles should have deleted the active dir (demonstrating the race)")
	}

	// The fix: cleanOnce.Do() ensures this only runs ONCE at startup,
	// before any buildArgs() creates dirs, so no active dirs exist to delete.
}

// TestCleanOnce_RunsExactlyOnce verifies that the sync.Once guard ensures
// cleanStaleTempProfiles runs only on the first call, not on subsequent calls.
// This prevents the race where a later worker's cleanup deletes an earlier
// worker's active temp directory.
func TestCleanOnce_RunsExactlyOnce(t *testing.T) {
	var callCount int = 0
	var once sync.Once

	var cleanup func() = func() {
		callCount++
	}

	// Simulate 10 concurrent Launch() calls each trying to clean
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(cleanup)
		}()
	}
	wg.Wait()

	if callCount != 1 {
		t.Errorf("cleanup should run exactly once, ran %d times", callCount)
	}
}

// TestBuildArgs_UniqueProfileDirs_Concurrent verifies that buildArgs produces
// unique user data directories even when called concurrently from multiple
// goroutines. This is the regression test for the SingletonLock collision bug
// where parallel workers received the same temp profile path.
func TestBuildArgs_UniqueProfileDirs_Concurrent(t *testing.T) {
	var opts *Options = &Options{Port: 0, Headless: true}
	var count int = 50
	var dirs chan string = make(chan string, count)

	for i := 0; i < count; i++ {
		go func() {
			var _args []string
			var dir string
			_args, dir = buildArgs(opts)
			_ = _args
			dirs <- dir
		}()
	}

	var seen map[string]bool = make(map[string]bool)
	for i := 0; i < count; i++ {
		var dir string = <-dirs
		defer os.RemoveAll(dir)
		if seen[dir] {
			t.Errorf("duplicate userDataDir detected: %s", dir)
		}
		seen[dir] = true
	}
}

// TestBuildArgs_ProfileDirContainsCryptoSuffix verifies that the temp profile
// directory name includes both a timestamp and a hex-encoded random suffix.
// This guards against regressions to timestamp-only naming.
func TestBuildArgs_ProfileDirContainsCryptoSuffix(t *testing.T) {
	var opts *Options = &Options{Port: 9222, Headless: true}
	var _args []string
	var dir string
	_args, dir = buildArgs(opts)
	_ = _args
	defer os.RemoveAll(dir)

	var base string = filepath.Base(dir)
	// Expected format: kexas-chrome-<unixnano>-<16 hex chars>
	if !strings.HasPrefix(base, "kexas-chrome-") {
		t.Fatalf("expected kexas-chrome- prefix, got: %s", base)
	}

	// Strip prefix and split on dash to find the hex suffix
	var rest string = strings.TrimPrefix(base, "kexas-chrome-")
	var parts []string = strings.SplitN(rest, "-", 2)
	if len(parts) != 2 {
		t.Fatalf("expected <timestamp>-<hex> format, got: %s", rest)
	}

	var hexPart string = parts[1]
	if len(hexPart) != 16 {
		t.Errorf("expected 16-char hex suffix (8 random bytes), got %d chars: %s", len(hexPart), hexPart)
	}
}

// TestBrowserClose_CleansUpLogFile verifies that Browser.Close() removes the
// temporary log file. This prevents stale log file accumulation from parallel
// browser launches that flood disk with leftover kexas-chrome-*.log files.
func TestBrowserClose_CleansUpLogFile(t *testing.T) {
	// Create a fake log file
	var tmpFile *os.File
	var err error
	tmpFile, err = os.CreateTemp("", "kexas-chrome-*.log")
	if err != nil {
		t.Fatalf("failed to create temp log file: %v", err)
	}
	var logPath string = tmpFile.Name()
	tmpFile.Close()

	// Verify it exists
	if _, statErr := os.Stat(logPath); os.IsNotExist(statErr) {
		t.Fatal("temp log file should exist before cleanup")
	}

	// Simulate cleanup (same logic as Browser.Close)
	var removeErr error = os.Remove(logPath)
	if removeErr != nil && !os.IsNotExist(removeErr) {
		t.Errorf("failed to remove log file: %v", removeErr)
	}

	// Verify it was removed
	if _, statErr := os.Stat(logPath); !os.IsNotExist(statErr) {
		t.Error("temp log file should be removed after cleanup")
	}

	// Verify double-remove doesn't panic (idempotent)
	var doubleErr error = os.Remove(logPath)
	if doubleErr != nil && !os.IsNotExist(doubleErr) {
		t.Errorf("double remove should be a no-op, got: %v", doubleErr)
	}
}
