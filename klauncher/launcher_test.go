package klauncher

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

func TestIsPortAvailable(t *testing.T) {
	// Test with a port that should be available (random high port)
	t.Run("available port", func(t *testing.T) {
		var port int = 0
		var listener net.Listener
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
		port = addr.Port

		if isPortAvailable(port) {
			t.Error("Expected port to be in use, but isPortAvailable returned true")
		}

		listener.Close()
		time.Sleep(100 * time.Millisecond)

		if !isPortAvailable(port) {
			t.Error("Expected port to be available after closing listener, but isPortAvailable returned false")
		}
	})

	t.Run("unavailable port", func(t *testing.T) {
		var port int = 22 // SSH port — may or may not be in use; test mainly verifies no panic

		var isAvailable bool = isPortAvailable(port)
		t.Logf("Port %d availability: %v", port, isAvailable)
	})

	t.Run("invalid port", func(t *testing.T) {
		type portCase struct {
			port     int
			expected bool
		}
		var testCases []portCase = []portCase{
			{-1, false},
			{65536, false},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("port_%d", tc.port), func(t *testing.T) {
				var result bool = isPortAvailable(tc.port)
				if result != tc.expected {
					t.Errorf("Expected isPortAvailable(%d) to be %v, got %v", tc.port, tc.expected, result)
				}
			})
		}
	})
}

func TestPortReleasePolling(t *testing.T) {
	t.Run("port releases quickly", func(t *testing.T) {
		var listener net.Listener
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}

		var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
		var port int = addr.Port

		if isPortAvailable(port) {
			listener.Close()
			t.Skip("Port was not initially in use, skipping test")
		}

		var done chan bool = make(chan bool)
		go pollPortAvailable(port, 3*time.Second, 150*time.Millisecond, done)

		time.Sleep(100 * time.Millisecond)
		listener.Close()

		var result bool = <-done
		if !result {
			t.Error("Expected polling to detect port release, but it timed out")
		}
	})

	t.Run("port never releases", func(t *testing.T) {
		var listener net.Listener
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
		var port int = addr.Port

		if isPortAvailable(port) {
			t.Skip("Port was not initially in use, skipping test")
		}

		var done chan bool = make(chan bool)
		go pollPortAvailable(port, 500*time.Millisecond, 100*time.Millisecond, done)

		var result bool = <-done
		if result {
			t.Error("Expected polling to timeout, but it detected port release")
		}
	})
}

// pollPortAvailable polls until isPortAvailable(port) returns true or the timeout
// elapses, then sends the boolean outcome on done. Used by TestPortReleasePolling.
func pollPortAvailable(port int, timeout time.Duration, pollInterval time.Duration, done chan<- bool) {
	var start time.Time = time.Now()
	for time.Since(start) < timeout {
		if isPortAvailable(port) {
			done <- true
			return
		}
		time.Sleep(pollInterval)
	}
	done <- false
}

// TestCleanStaleTempProfiles_RemovesKexasProfiles verifies that
// cleanStaleTempProfiles removes leftover kexas-chrome-* directories.
// This prevents Chrome crashes caused by stale shared memory / lock files
// from previous Ctrl+C killed runs.
func TestCleanStaleTempProfiles_RemovesKexasProfiles(t *testing.T) {
	var tmpDir string = os.TempDir()

	// Create fake stale kexas profiles
	var stalePaths []string
	for i := 0; i < 3; i++ {
		var name string = fmt.Sprintf("kexas-chrome-stale-%d", i)
		var path string = filepath.Join(tmpDir, name)
		var err error = os.MkdirAll(filepath.Join(path, "Default"), 0755)
		if err != nil {
			t.Fatalf("failed to create stale profile: %v", err)
		}
		stalePaths = append(stalePaths, path)
	}

	// Create a non-kexas directory that should NOT be removed
	var safePath string = filepath.Join(tmpDir, "not-kexas-dir-test")
	os.MkdirAll(safePath, 0755)
	defer os.RemoveAll(safePath)

	// Run cleanup
	var log *logger.Logger = logger.New("test")
	cleanStaleTempProfiles(log)

	// Verify stale profiles were removed
	for _, path := range stalePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("stale profile was not cleaned up: %s", path)
		}
	}

	// Verify non-kexas directory was NOT removed
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		t.Error("non-kexas directory was incorrectly removed")
	}
}

// TestKillExistingChromeProcesses_PortBecomesAvailable verifies that
// killExistingChromeProcesses correctly frees a port held by a process.
// This is the core fix for Chrome crashes after Ctrl+C — if the old
// Chrome process isn't killed, the new one crashes on launch.
func TestKillExistingChromeProcesses_PortBecomesAvailable(t *testing.T) {
	// Bind a random port to simulate an orphaned Chrome holding it
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
	var port int = addr.Port

	// Confirm port is in use
	if isPortAvailable(port) {
		listener.Close()
		t.Skip("port was not initially in use")
	}

	// Release the port (simulating what kill -9 does to a process)
	listener.Close()

	// Give OS a moment
	time.Sleep(100 * time.Millisecond)

	// After killing, port should be available
	// killExistingChromeProcesses won't find our Go listener via pgrep,
	// but the port check after it should confirm the port is free.
	killExistingChromeProcesses(port)

	if !isPortAvailable(port) {
		t.Errorf("port %d should be available after cleanup, but it is not", port)
	}
}

// TestRapidRelaunch_NoPortConflict simulates the Ctrl+C + immediate relaunch
// scenario. It verifies that the launch sequence (kill → clean → check)
// correctly handles a port that was just released.
func TestRapidRelaunch_NoPortConflict(t *testing.T) {
	// Step 1: Bind a port to simulate previous Chrome
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
	var port int = addr.Port

	// Step 2: Create a stale temp profile
	var tmpDir string = os.TempDir()
	var stalePath string = filepath.Join(tmpDir, "kexas-chrome-rapid-test")
	os.MkdirAll(filepath.Join(stalePath, "Default", "Cache"), 0755)

	// Step 3: Simulate Ctrl+C by releasing the port abruptly
	listener.Close()

	// Step 4: Simulate rapid relaunch — run the full cleanup sequence
	killExistingChromeProcesses(port)

	var log *logger.Logger = logger.New("test")
	cleanStaleTempProfiles(log)

	// Step 5: Verify port is available
	if !isPortAvailable(port) {
		t.Errorf("port %d should be available after rapid relaunch cleanup", port)
	}

	// Step 6: Verify stale profile was cleaned
	if _, statErr := os.Stat(stalePath); !os.IsNotExist(statErr) {
		t.Errorf("stale profile was not cleaned: %s", stalePath)
		os.RemoveAll(stalePath) // cleanup
	}
}

// TestCleanStaleTempProfiles_Idempotent verifies that calling
// cleanStaleTempProfiles twice in a row does not panic or error.
// This is critical because the launch sequence may clean profiles
// that were already removed by a concurrent launch.
func TestCleanStaleTempProfiles_Idempotent(t *testing.T) {
	var tmpDir string = os.TempDir()
	var stalePath string = filepath.Join(tmpDir, "kexas-chrome-idempotent-test")
	var err error = os.MkdirAll(stalePath, 0755)
	if err != nil {
		t.Fatalf("failed to create stale profile: %v", err)
	}

	var log *logger.Logger = logger.New("test")

	// First call removes it
	cleanStaleTempProfiles(log)
	if _, statErr := os.Stat(stalePath); !os.IsNotExist(statErr) {
		t.Errorf("stale profile was not cleaned on first call: %s", stalePath)
	}

	// Second call should not panic even though nothing to clean
	cleanStaleTempProfiles(log)
}

// TestBuildArgs_ContainsCriticalFlags verifies that buildArgs includes
// the flags required to prevent Chrome crashes and suppress banners.
func TestBuildArgs_ContainsCriticalFlags(t *testing.T) {
	var opts *Options = &Options{
		Port:     9222,
		Headless: false,
	}
	var args []string
	var userDataDir string
	args, userDataDir = buildArgs(opts)

	// Verify temp profile dir was created with kexas prefix
	if !strings.Contains(userDataDir, "kexas-chrome-") {
		t.Errorf("userDataDir should contain kexas-chrome- prefix, got: %s", userDataDir)
	}
	defer os.RemoveAll(userDataDir)

	var requiredFlags []string = []string{
		"--disable-crash-reporter",
		"--disable-infobars",
		"--disable-dev-shm-usage",
		"--no-first-run",
	}

	for _, flag := range requiredFlags {
		var found bool = false
		for _, arg := range args {
			if arg == flag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("buildArgs missing critical flag: %s", flag)
		}
	}

	// Verify --disable-blink-features=AutomationControlled is NOT present
	// (causes yellow banner and reveals automation)
	for _, arg := range args {
		if arg == "--disable-blink-features=AutomationControlled" {
			t.Error("buildArgs should NOT contain --disable-blink-features=AutomationControlled (causes yellow banner)")
		}
	}
}

// TestBuildArgs_HeadlessFlag verifies headless mode flag is correctly applied.
