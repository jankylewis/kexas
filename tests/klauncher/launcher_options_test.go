package klauncher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jankylewis/kexas/klauncher"
)

// TestLaunch_NilOpts_UsesDefaults — passing nil opts should fall back to
// DefaultOptions rather than panic. We can't actually launch a browser in
// unit tests, but we verify the nil-opts code path doesn't blow up before
// reaching browser-binary discovery.
func TestLaunch_NilOpts_UsesDefaults(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Launch panicked on nil opts: %v", r)
		}
	}()

	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// nil opts → DefaultOptions(). Will fail (no browser), but should NOT panic.
	_, err := klauncher.Launch(ctx, nil)
	if err == nil {
		t.Log("Launch unexpectedly succeeded — likely a test browser is on PATH; that's fine")
	}
	// Either error or success is acceptable here; only panic is a fail.
}

// TestLaunch_BogusExecutablePath_ReturnsError — explicit nonexistent path
// should error cleanly rather than spinning up some other browser.
func TestLaunch_BogusExecutablePath_ReturnsError(t *testing.T) {
	var opts *klauncher.Options = &klauncher.Options{
		Headless:       true,
		ExecutablePath: "/definitely/not/a/real/path/to/chromium-xyz123",
	}

	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := klauncher.Launch(ctx, opts)
	if err == nil {
		t.Fatal("expected Launch to fail with bogus executable path")
	}
}

// TestDefaultOptions_DefaultsAreUsable — nudge-test for Headless/Port/Args.
// Catches regressions where defaults silently change to something hostile.
func TestDefaultOptions_DefaultsAreUsable(t *testing.T) {
	var opts *klauncher.Options = klauncher.DefaultOptions()
	if opts == nil {
		t.Fatal("DefaultOptions returned nil")
	}
	if !opts.Headless {
		t.Error("expected Headless=true by default (CI-friendly)")
	}
	if opts.Port != 0 {
		t.Errorf("expected Port=0 (Chrome self-assigns) by default, got %d", opts.Port)
	}
	if opts.WindowWidth <= 0 || opts.WindowHeight <= 0 {
		t.Errorf("expected positive viewport, got %dx%d", opts.WindowWidth, opts.WindowHeight)
	}
}

// TestSentinelErrors_AreNonNil — confirms the package-level error sentinels
// are wired up. errors.Is() against them is the public contract.
func TestSentinelErrors_AreNonNil(t *testing.T) {
	if klauncher.ErrChromiumNotFound == nil {
		t.Error("ErrChromiumNotFound should not be nil")
	}
	if klauncher.ErrLaunchFailed == nil {
		t.Error("ErrLaunchFailed should not be nil")
	}
	if klauncher.ErrNoDebuggerURL == nil {
		t.Error("ErrNoDebuggerURL should not be nil")
	}
}

// TestSentinelErrors_DistinctIdentity — sentinels must not collapse to the
// same value (would break errors.Is dispatch).
func TestSentinelErrors_DistinctIdentity(t *testing.T) {
	if errors.Is(klauncher.ErrChromiumNotFound, klauncher.ErrLaunchFailed) {
		t.Error("ErrChromiumNotFound and ErrLaunchFailed must be distinct")
	}
	if errors.Is(klauncher.ErrLaunchFailed, klauncher.ErrNoDebuggerURL) {
		t.Error("ErrLaunchFailed and ErrNoDebuggerURL must be distinct")
	}
}
