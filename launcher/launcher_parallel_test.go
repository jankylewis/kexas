package launcher

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kexas-project/kexas/internal/logger"
)

// ========================================
// CRITICAL: Parallel launch fix tests
// ========================================
//
// These tests cover the fixes for Chrome parallel launch failures:
// 1. TOCTOU race in findFreePort() → fixed with --remote-debugging-port=0
// 2. File-based stdout polling lag → fixed with pipe-based stderr reading
// 3. No retry on transient failures → fixed with exponential backoff
// 4. Simultaneous worker startup → fixed with stagger delay

// TestExtractPortFromWSURL_ValidURL verifies that extractPortFromWSURL
// correctly parses the port from a standard Chrome WebSocket URL.
func TestExtractPortFromWSURL_ValidURL(t *testing.T) {
	var testCases []struct {
		url      string
		expected int
	} = []struct {
		url      string
		expected int
	}{
		{"ws://127.0.0.1:51234/devtools/browser/abc-123", 51234},
		{"ws://127.0.0.1:9222/devtools/browser/def-456", 9222},
		{"ws://localhost:42000/devtools/browser/ghi", 42000},
		{"ws://127.0.0.1:1/devtools/browser/x", 1},
		{"ws://127.0.0.1:65535/devtools/browser/y", 65535},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("port_%d", tc.expected), func(t *testing.T) {
			var port int = extractPortFromWSURL(tc.url)
			if port != tc.expected {
				t.Errorf("extractPortFromWSURL(%q) = %d, want %d", tc.url, port, tc.expected)
			}
		})
	}
}

// TestExtractPortFromWSURL_InvalidURL verifies that extractPortFromWSURL
// returns 0 for malformed or missing port URLs.
func TestExtractPortFromWSURL_InvalidURL(t *testing.T) {
	var testCases []string = []string{
		"",
		"not-a-url",
		"ws://127.0.0.1",      // no port, no path
		"ws://127.0.0.1/path", // no port
		"http://example.com",  // no port, wrong scheme
	}

	for _, url := range testCases {
		t.Run(url, func(t *testing.T) {
			var port int = extractPortFromWSURL(url)
			if port != 0 {
				t.Errorf("extractPortFromWSURL(%q) = %d, want 0", url, port)
			}
		})
	}
}

// TestBuildArgs_Port0_PassedDirectlyToChrome verifies that when Port=0,
// buildArgs passes --remote-debugging-port=0 to Chrome, letting Chrome
// atomically self-assign a free port. This is the CRITICAL fix for the
// TOCTOU race condition where findFreePort() released a port that another
// worker grabbed before Chrome could bind it.
func TestBuildArgs_Port0_PassedDirectlyToChrome(t *testing.T) {
	var opts *Options = &Options{Port: 0, Headless: true}
	var args []string
	var userDataDir string
	args, userDataDir = buildArgs(opts)
	defer os.RemoveAll(userDataDir)

	var foundPortArg bool = false
	for _, arg := range args {
		if arg == "--remote-debugging-port=0" {
			foundPortArg = true
			break
		}
	}

	if !foundPortArg {
		t.Error("CRITICAL: buildArgs with Port=0 must produce --remote-debugging-port=0 so Chrome self-assigns a free port atomically")
	}
}

// TestFindFreePort_TOCTOU_Race demonstrates the TOCTOU vulnerability in
// findFreePort(): the port is released between listener.Close() and Chrome
// binding it, allowing another caller to grab the same port.
func TestFindFreePort_TOCTOU_Race(t *testing.T) {
	var portCount int = 100
	var ports []int = make([]int, 0, portCount)
	var collisions int = 0

	for i := 0; i < portCount; i++ {
		var port int
		var err error
		port, err = findFreePort()
		if err != nil {
			t.Fatalf("findFreePort failed on iteration %d: %v", i, err)
		}
		for _, existing := range ports {
			if existing == port {
				collisions++
				break
			}
		}
		ports = append(ports, port)
	}

	t.Logf("findFreePort called %d times, collisions: %d", portCount, collisions)
	t.Log("NOTE: Even without collisions in this test, the TOCTOU window exists.")
	t.Log("The fix is --remote-debugging-port=0, which eliminates the window entirely.")
}

// TestFindFreePort_ConcurrentTOCTOU demonstrates the TOCTOU race under
// concurrent load — the exact scenario in parallel test execution.
func TestFindFreePort_ConcurrentTOCTOU(t *testing.T) {
	var concurrency int = 20
	var portChan chan int = make(chan int, concurrency)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var port int
			var err error
			port, err = findFreePort()
			if err != nil {
				t.Logf("findFreePort error: %v", err)
				portChan <- -1
				return
			}
			portChan <- port
		}()
	}
	wg.Wait()
	close(portChan)

	var seen map[int]int = make(map[int]int)
	for port := range portChan {
		if port > 0 {
			seen[port]++
		}
	}

	var collisions int = 0
	for port, count := range seen {
		if count > 1 {
			collisions++
			t.Logf("TOCTOU collision: port %d assigned to %d callers", port, count)
		}
	}

	t.Logf("Concurrent findFreePort: %d goroutines, %d unique ports, %d collisions",
		concurrency, len(seen), collisions,
	)
	t.Log("FIX: --remote-debugging-port=0 lets Chrome pick its own port atomically.")
}

// TestKillExistingChromeProcesses_NoProcesses verifies that
// killExistingChromeProcesses handles the case where no processes exist
// without errors or panics.
func TestKillExistingChromeProcesses_NoProcesses(t *testing.T) {
	var listener net.Listener
	var err error
	listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	var addr *net.TCPAddr = listener.Addr().(*net.TCPAddr)
	var port int = addr.Port
	listener.Close()
	time.Sleep(50 * time.Millisecond)

	killExistingChromeProcesses(port)

	if !isPortAvailable(port) {
		t.Errorf("port %d should be available when no processes exist", port)
	}
}

// TestExtractDebuggerURLFromPipe_Success verifies that extractDebuggerURLFromPipe
// correctly reads a WebSocket URL from a simulated Chrome stderr pipe.
// This is the critical fix for file-based polling lag under parallel load.
func TestExtractDebuggerURLFromPipe_Success(t *testing.T) {
	var pr *io.PipeReader
	var pw *io.PipeWriter
	pr, pw = io.Pipe()

	var expectedURL string = "ws://127.0.0.1:44567/devtools/browser/test-uuid"

	// Simulate Chrome writing to stderr after a brief delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintln(pw, "[0228/110000.000:INFO] Some Chrome startup message")
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintf(pw, "DevTools listening on %s\n", expectedURL)
		pw.Close()
	}()

	var log *logger.Logger = logger.New("test")
	var wsURL string
	var err error
	wsURL, err = extractDebuggerURLFromPipe(pr, 5*time.Second, log)

	if err != nil {
		t.Fatalf("extractDebuggerURLFromPipe failed: %v", err)
	}
	if wsURL != expectedURL {
		t.Errorf("expected %q, got %q", expectedURL, wsURL)
	}
}

// TestExtractDebuggerURLFromPipe_Timeout verifies that extractDebuggerURLFromPipe
// correctly times out when Chrome never prints the DevTools URL.
func TestExtractDebuggerURLFromPipe_Timeout(t *testing.T) {
	var pr *io.PipeReader
	var pw *io.PipeWriter
	pr, pw = io.Pipe()

	// Simulate Chrome that prints noise but never the DevTools URL
	go func() {
		for i := 0; i < 5; i++ {
			fmt.Fprintf(pw, "Some Chrome log line %d\n", i)
			time.Sleep(50 * time.Millisecond)
		}
		// Never write DevTools URL — don't close pipe (simulates stuck Chrome)
	}()

	var log *logger.Logger = logger.New("test")
	var start time.Time = time.Now()
	var wsURL string
	var err error
	wsURL, err = extractDebuggerURLFromPipe(pr, 500*time.Millisecond, log)
	var elapsed time.Duration = time.Since(start)

	if err == nil {
		t.Fatalf("expected timeout error, got URL: %s", wsURL)
	}
	if err != ErrNoDebuggerURL {
		t.Errorf("expected ErrNoDebuggerURL, got: %v", err)
	}
	// Should timeout within reasonable bounds (500ms ± 200ms)
	if elapsed < 400*time.Millisecond || elapsed > 1*time.Second {
		t.Errorf("timeout took unexpected duration: %v", elapsed)
	}

	pw.Close() // Clean up
}

// TestExtractDebuggerURLFromPipe_PortBindError verifies that the pipe reader
// detects Chrome's "address already in use" error and returns it immediately
// instead of waiting for the full timeout.
func TestExtractDebuggerURLFromPipe_PortBindError(t *testing.T) {
	var pr *io.PipeReader
	var pw *io.PipeWriter
	pr, pw = io.Pipe()

	go func() {
		fmt.Fprintln(pw, "[ERROR] Failed to bind to port")
		fmt.Fprintln(pw, "bind: address already in use")
		pw.Close()
	}()

	var log *logger.Logger = logger.New("test")
	var start time.Time = time.Now()
	var wsURL string
	var err error
	wsURL, err = extractDebuggerURLFromPipe(pr, 5*time.Second, log)
	var elapsed time.Duration = time.Since(start)

	if err == nil {
		t.Fatalf("expected bind error, got URL: %s", wsURL)
	}
	if !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("expected 'address already in use' error, got: %v", err)
	}
	// Should fail fast, not wait for timeout
	if elapsed > 1*time.Second {
		t.Errorf("bind error detection took too long: %v (should be <1s)", elapsed)
	}
}

// TestExtractDebuggerURLFromPipe_PipeClosedEarly verifies that when Chrome
// crashes and the pipe closes without printing the URL, we get ErrNoDebuggerURL
// instead of hanging until timeout.
func TestExtractDebuggerURLFromPipe_PipeClosedEarly(t *testing.T) {
	var pr *io.PipeReader
	var pw *io.PipeWriter
	pr, pw = io.Pipe()

	// Simulate Chrome crashing immediately
	go func() {
		fmt.Fprintln(pw, "Chrome starting...")
		time.Sleep(20 * time.Millisecond)
		pw.Close() // Chrome crashes — pipe closes
	}()

	var log *logger.Logger = logger.New("test")
	var start time.Time = time.Now()
	var wsURL string
	var err error
	wsURL, err = extractDebuggerURLFromPipe(pr, 5*time.Second, log)
	var elapsed time.Duration = time.Since(start)

	if err == nil {
		t.Fatalf("expected error for crashed Chrome, got URL: %s", wsURL)
	}
	if err != ErrNoDebuggerURL {
		t.Errorf("expected ErrNoDebuggerURL, got: %v", err)
	}
	// Should detect closed pipe quickly, not wait 5s
	if elapsed > 1*time.Second {
		t.Errorf("pipe close detection took too long: %v (should be <1s)", elapsed)
	}
}

// TestDefaultOptions_PortIsZero verifies that DefaultOptions uses Port=0
// for automatic port assignment, which is required for parallel execution.
func TestDefaultOptions_PortIsZero(t *testing.T) {
	var opts *Options = DefaultOptions()
	if opts.Port != 0 {
		t.Errorf("DefaultOptions().Port should be 0 for dynamic assignment, got %d", opts.Port)
	}
}
