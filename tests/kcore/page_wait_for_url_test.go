package kcore_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

// Critical unit tests for Page.WaitForURLContains — the SPA-friendly URL
// poller added after the WaitForLoadState/title-poll bug surfaced during
// ghtests dogfood. Mirror style of TestPage_WaitForElement_*.

// TestPage_WaitForURLContains_ZeroTimeout — timeout < 100ms is rejected.
// 100ms is the floor (not 1s as for WaitForElement) because URL polls are
// cheap and short flows may legitimately want sub-second timeouts.
func TestPage_WaitForURLContains_ZeroTimeout(t *testing.T) {
	var page *kexas.Page

	var err error
	err = page.WaitForURLContains("/foo", 0)

	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
	// Format from errors.TimeoutInvalidFormat: "timeout must be at least 1 second, got 0s"
	if !strings.Contains(err.Error(), "timeout must be at least") {
		t.Errorf("expected timeout-validation error, got '%s'", err.Error())
	}
}

// TestPage_WaitForURLContains_NegativeTimeout — same rejection path.
func TestPage_WaitForURLContains_NegativeTimeout(t *testing.T) {
	var page *kexas.Page

	var err error
	err = page.WaitForURLContains("/foo", -5*time.Second)

	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
	if !strings.Contains(err.Error(), "timeout must be at least") {
		t.Errorf("expected timeout-validation error, got '%s'", err.Error())
	}
}

// TestPage_WaitForURLContains_BelowFloor — exactly 99ms is below the 100ms
// floor; should be rejected.
func TestPage_WaitForURLContains_BelowFloor(t *testing.T) {
	var page *kexas.Page

	var err error
	err = page.WaitForURLContains("/foo", 99*time.Millisecond)

	if err == nil {
		t.Fatal("expected error for sub-100ms timeout")
	}
	if !strings.Contains(err.Error(), "timeout must be at least") {
		t.Errorf("expected timeout-validation error, got '%s'", err.Error())
	}
}

// TestPage_WaitForURLContains_NilPage — calling on a nil *Page returns the
// "URL did not contain X" error format. The nil-page guard short-circuits
// without panicking — added after the test caught the missing guard.
func TestPage_WaitForURLContains_NilPage(t *testing.T) {
	var page *kexas.Page

	var err error
	err = page.WaitForURLContains("/expected", 200*time.Millisecond)

	if err == nil {
		t.Fatal("expected error for nil page")
	}
	if !strings.Contains(err.Error(), "URL did not contain") {
		t.Errorf("expected URL-not-contained error, got '%s'", err.Error())
	}
	if !strings.Contains(err.Error(), "/expected") {
		t.Errorf("error should name the missed substring, got '%s'", err.Error())
	}
}

// TestPage_WaitForURLContains_ReturnsBeforeTimeoutOnNilPage — wall-clock
// confirms the nil-page guard short-circuits immediately rather than burning
// the full timeout polling.
func TestPage_WaitForURLContains_ReturnsBeforeTimeoutOnNilPage(t *testing.T) {
	var page *kexas.Page
	var requestedTimeout time.Duration = 5 * time.Second

	var start time.Time = time.Now()
	_ = page.WaitForURLContains("/anything", requestedTimeout)
	var elapsed time.Duration = time.Since(start)

	// The guard returns immediately; allow generous slack for slow CI but
	// still well under the 5s timeout.
	if elapsed > 1*time.Second {
		t.Errorf("nil-page guard should short-circuit, took %v", elapsed)
	}
}
