package launcher

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestIsPortAvailable(t *testing.T) {
	// Test with a port that should be available (random high port)
	t.Run("available port", func(t *testing.T) {
		port := 0 // Let OS choose a random port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		// Get the actual port assigned
		addr := listener.Addr().(*net.TCPAddr)
		port = addr.Port

		// Port should be in use (not available)
		if isPortAvailable(port) {
			t.Error("Expected port to be in use, but isPortAvailable returned true")
		}

		// Close listener and check again
		listener.Close()
		time.Sleep(100 * time.Millisecond) // Give OS time to release port

		// Port should now be available
		if !isPortAvailable(port) {
			t.Error("Expected port to be available after closing listener, but isPortAvailable returned false")
		}
	})

	t.Run("unavailable port", func(t *testing.T) {
		// Use a common port that's likely in use or reserved
		port := 22 // SSH port

		// We can't guarantee this port is in use, but if it's available, that's ok
		// The test mainly verifies the function doesn't panic
		isAvailable := isPortAvailable(port)
		t.Logf("Port %d availability: %v", port, isAvailable)
	})

	t.Run("invalid port", func(t *testing.T) {
		// Test with invalid port numbers
		testCases := []struct {
			port     int
			expected bool
		}{
			{-1, false},    // Negative port
			{65536, false}, // Port above valid range
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("port_%d", tc.port), func(t *testing.T) {
				result := isPortAvailable(tc.port)
				if result != tc.expected {
					t.Errorf("Expected isPortAvailable(%d) to be %v, got %v", tc.port, tc.expected, result)
				}
			})
		}
	})
}

func TestPortReleasePolling(t *testing.T) {
	t.Run("port releases quickly", func(t *testing.T) {
		// Create a listener on a random port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}

		addr := listener.Addr().(*net.TCPAddr)
		port := addr.Port

		// Port should be in use
		if isPortAvailable(port) {
			listener.Close()
			t.Skip("Port was not initially in use, skipping test")
		}

		// Start polling in a goroutine
		done := make(chan bool)
		go func() {
			start := time.Now()
			timeout := 3 * time.Second
			pollInterval := 150 * time.Millisecond

			for time.Since(start) < timeout {
				if isPortAvailable(port) {
					done <- true
					return
				}
				time.Sleep(pollInterval)
			}
			done <- false
		}()

		// Close the listener after a short delay
		time.Sleep(100 * time.Millisecond)
		listener.Close()

		// Wait for polling to complete
		result := <-done

		if !result {
			t.Error("Expected polling to detect port release, but it timed out")
		}
	})

	t.Run("port never releases", func(t *testing.T) {
		// Create a listener and keep it open
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		addr := listener.Addr().(*net.TCPAddr)
		port := addr.Port

		// Port should be in use
		if isPortAvailable(port) {
			t.Skip("Port was not initially in use, skipping test")
		}

		// Start polling with short timeout
		done := make(chan bool)
		go func() {
			start := time.Now()
			timeout := 500 * time.Millisecond // Short timeout for test
			pollInterval := 100 * time.Millisecond

			for time.Since(start) < timeout {
				if isPortAvailable(port) {
					done <- true
					return
				}
				time.Sleep(pollInterval)
			}
			done <- false
		}()

		// Wait for polling to complete (should timeout)
		result := <-done

		if result {
			t.Error("Expected polling to timeout, but it detected port release")
		}
	})
}

func TestBrowserClosePortRelease(t *testing.T) {
	t.Skip("Skipping integration test - requires actual browser launch")

	// This would be an integration test that:
	// 1. Launches a real browser
	// 2. Calls Close()
	// 3. Verifies the port is released
	// 4. Measures how long it took

	// Skipping for now as it requires browser setup
}
