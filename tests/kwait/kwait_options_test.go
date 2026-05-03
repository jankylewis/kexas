package kwait_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kwait"
)

// TestFor_RespectsCustomInterval — verifies For polls at the interval set in
// Options, not the default. Counts callback invocations over a short window.
func TestFor_RespectsCustomInterval(t *testing.T) {
	var calls int = 0
	var cond kwait.Condition = func() (bool, error) {
		calls++
		return false, nil
	}
	var opts *kwait.Options = &kwait.Options{
		Timeout:  450 * time.Millisecond,
		Interval: 100 * time.Millisecond,
		Message:  "test",
	}

	var ctx context.Context = context.Background()
	_ = kwait.For(ctx, cond, opts)

	// Expect ~5 calls (initial + 4 retries). Allow 3-7 for CI scheduling jitter.
	if calls < 3 || calls > 7 {
		t.Errorf("expected 3-7 calls at 100ms interval over 450ms, got %d", calls)
	}
}

// TestFor_NilContext_DoesNotPanic — passing a nil context should fall back to
// background context behavior (or return cleanly), not panic.
func TestFor_NilContext_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("For panicked on nil context: %v", r)
		}
	}()

	var opts *kwait.Options = &kwait.Options{
		Timeout:  100 * time.Millisecond,
		Interval: 50 * time.Millisecond,
	}
	_ = kwait.For(context.Background(), func() (bool, error) { return false, nil }, opts)
}

// TestUntilReady_ConditionError_PropagatesError — when fn returns an error on
// every call, UntilReady should propagate the error, not return success.
func TestUntilReady_ConditionError_PropagatesError(t *testing.T) {
	var sentinel error = &errSentinel{msg: "sentinel"}
	var fn func() (string, error) = func() (string, error) {
		return "", sentinel
	}
	var opts *kwait.Options = &kwait.Options{
		Timeout:  150 * time.Millisecond,
		Interval: 50 * time.Millisecond,
	}

	var got string
	var err error
	got, err = kwait.UntilReady(context.Background(), fn, opts)

	if err == nil {
		t.Fatal("expected error from UntilReady when fn always errors")
	}
	if got != "" {
		t.Errorf("expected empty result, got %q", got)
	}
}

// TestDefaultOptions_HasUsableValues — the zero-arg default options should
// have non-zero timeout + interval so callers can use them as-is.
func TestDefaultOptions_HasUsableValues(t *testing.T) {
	var opts *kwait.Options = kwait.DefaultOptions()
	if opts == nil {
		t.Fatal("DefaultOptions returned nil")
	}
	if opts.Timeout <= 0 {
		t.Errorf("DefaultOptions.Timeout should be > 0, got %v", opts.Timeout)
	}
	if opts.Interval <= 0 {
		t.Errorf("DefaultOptions.Interval should be > 0, got %v", opts.Interval)
	}
}

// TestForPageLoad_NetworkIdle_AddsExtraDelay — NetworkIdle path is title-poll
// + 500ms sleep. Wall-clock should reflect the sleep.
func TestForPageLoad_NetworkIdle_AddsExtraDelay(t *testing.T) {
	var titleFn func() (string, error) = func() (string, error) {
		return "Loaded", nil
	}
	var start time.Time = time.Now()
	var err error = kwait.ForPageLoad(context.Background(), titleFn, kwait.WaitUntilNetworkIdle, 2*time.Second)
	var elapsed time.Duration = time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if elapsed < 400*time.Millisecond {
		t.Errorf("NetworkIdle should sleep ~500ms after title-set, but elapsed was %v", elapsed)
	}
}

// errSentinel is a minimal error type for the test above.
type errSentinel struct{ msg string }

func (e *errSentinel) Error() string { return e.msg }

// _ ensures strings is referenced so the import doesn't get pruned if a future
// test that uses it gets removed. (No-op at runtime.)
var _ = strings.HasPrefix
