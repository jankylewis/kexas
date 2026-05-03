
package kwait_test

import (
	"context"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kwait"
)

// TestForPageLoad_Commit tests that commit strategy returns immediately.
func TestForPageLoad_Commit(t *testing.T) {
	var ctx context.Context = context.Background()
	var getTitleFn func() (string, error) = func() (string, error) {
		return "New Tab", nil
	}

	var startTime time.Time = time.Now()
	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntilCommit, 1*time.Second)
	var elapsed time.Duration = time.Since(startTime)

	if err != nil {
		t.Errorf("expected no error for commit, got: %v", err)
	}
	// Should return immediately (< 50ms)
	if elapsed > 50*time.Millisecond {
		t.Errorf("commit should return immediately, took %v", elapsed)
	}
}

// TestForPageLoad_DOMContentLoaded tests that domcontentloaded waits for title to change.
func TestForPageLoad_DOMContentLoaded(t *testing.T) {
	var ctx context.Context = context.Background()
	var callCount int = 0
	var getTitleFn func() (string, error) = func() (string, error) {
		callCount++
		if callCount >= 3 {
			return "Loaded Page", nil
		}
		return "New Tab", nil
	}

	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntilDOMContentLoaded, 2*time.Second)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
}

// TestForPageLoad_Load tests that load strategy waits for non-empty, non-NewTab title.
func TestForPageLoad_Load(t *testing.T) {
	var ctx context.Context = context.Background()
	var callCount int = 0
	var getTitleFn func() (string, error) = func() (string, error) {
		callCount++
		switch callCount {
		case 1:
			return "New Tab", nil
		case 2:
			return "", nil // Empty title
		default:
			return "Fully Loaded", nil
		}
	}

	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntilLoad, 2*time.Second)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
}

// TestForPageLoad_NetworkIdle tests that networkidle waits for title plus extra delay.
func TestForPageLoad_NetworkIdle(t *testing.T) {
	var ctx context.Context = context.Background()
	var callCount int = 0
	var getTitleFn func() (string, error) = func() (string, error) {
		callCount++
		if callCount >= 2 {
			return "Page Ready", nil
		}
		return "New Tab", nil
	}

	var startTime time.Time = time.Now()
	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntilNetworkIdle, 2*time.Second)
	var elapsed time.Duration = time.Since(startTime)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	// Should wait at least 500ms for network idle
	if elapsed < 500*time.Millisecond {
		t.Errorf("networkidle should wait at least 500ms, waited %v", elapsed)
	}
}

// TestForPageLoad_Timeout tests that page load times out when title never changes.
func TestForPageLoad_Timeout(t *testing.T) {
	var ctx context.Context = context.Background()
	var getTitleFn func() (string, error) = func() (string, error) {
		return "New Tab", nil // Never changes
	}

	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntilLoad, 500*time.Millisecond)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

// TestForPageLoad_UnknownStrategy tests that unknown wait strategy returns error.
func TestForPageLoad_UnknownStrategy(t *testing.T) {
	var ctx context.Context = context.Background()
	var getTitleFn func() (string, error) = func() (string, error) {
		return "Test", nil
	}

	var err error = kwait.ForPageLoad(ctx, getTitleFn, kwait.WaitUntil("invalid"), 1*time.Second)

	if err == nil {
		t.Error("expected error for unknown strategy, got nil")
	}
}

// TestKwaitDefaultOptions tests that DefaultOptions returns sensible defaults.
func TestKwaitDefaultOptions(t *testing.T) {
	var opts *kwait.Options = kwait.DefaultOptions()

	if opts.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", opts.Timeout)
	}
	if opts.Interval != 100*time.Millisecond {
		t.Errorf("expected default interval 100ms, got %v", opts.Interval)
	}
	if opts.Message == "" {
		t.Error("expected non-empty default message")
	}
}

// TestWaitUntilConstants tests that wait strategy constants are defined correctly.
func TestWaitUntilConstants(t *testing.T) {
	if kwait.WaitUntilCommit != "commit" {
		t.Errorf("expected 'commit', got %s", kwait.WaitUntilCommit)
	}
	if kwait.WaitUntilDOMContentLoaded != "domcontentloaded" {
		t.Errorf("expected 'domcontentloaded', got %s", kwait.WaitUntilDOMContentLoaded)
	}
	if kwait.WaitUntilLoad != "load" {
		t.Errorf("expected 'load', got %s", kwait.WaitUntilLoad)
	}
	if kwait.WaitUntilNetworkIdle != "networkidle" {
		t.Errorf("expected 'networkidle', got %s", kwait.WaitUntilNetworkIdle)
	}
}
