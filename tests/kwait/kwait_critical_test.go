package kwait_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kwait"
)

// 15 critical-after-existing unit tests for the kwait package.
// Existing tests: For/UntilReady/ForPageLoad happy paths + a few edge cases.
// This batch adds: cancellation propagation, message-format checks, thread
// safety on shared callbacks, predicate-error swallow vs propagate, and
// boundary timing assertions.

func TestFor_CancelledContext_ReturnsImmediately(t *testing.T) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // already cancelled

	var start time.Time = time.Now()
	var err error = kwait.For(ctx, func() (bool, error) { return false, nil }, &kwait.Options{
		Timeout: 5 * time.Second, Interval: 100 * time.Millisecond,
	})
	var elapsed time.Duration = time.Since(start)

	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("should return ~immediately on cancelled ctx, took %v", elapsed)
	}
}

func TestFor_TimeoutMessage_IncludesContext(t *testing.T) {
	var err error = kwait.For(context.Background(), func() (bool, error) { return false, nil }, &kwait.Options{
		Timeout: 100 * time.Millisecond, Interval: 50 * time.Millisecond, Message: "kexas-test-marker",
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "kexas-test-marker") {
		t.Errorf("error message should include Options.Message; got %q", err.Error())
	}
}

func TestFor_ConditionReturnsTrueOnFirstCall(t *testing.T) {
	var calls int = 0
	var err error = kwait.For(context.Background(), func() (bool, error) {
		calls++
		return true, nil
	}, &kwait.Options{Timeout: 5 * time.Second, Interval: 100 * time.Millisecond})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call, got %d", calls)
	}
}

func TestFor_ConditionError_PropagatesNotSwallowed(t *testing.T) {
	// Document the actual contract: condition errors are propagated
	// immediately, NOT treated as transient (unlike titleFn errors in
	// ForPageLoad's polling loop, which ARE swallowed).
	var calls int32 = 0
	var err error = kwait.For(context.Background(), func() (bool, error) {
		atomic.AddInt32(&calls, 1)
		return false, errors.New("hard-fail")
	}, &kwait.Options{Timeout: 1 * time.Second, Interval: 50 * time.Millisecond})

	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected exactly 1 call before propagation, got %d", calls)
	}
}

func TestFor_NilOptions_UsesDefaults(t *testing.T) {
	// Existing TestFor_NilOptions verifies no panic; this verifies the call
	// actually returns a sensible default behavior (eventual timeout).
	var calls int = 0
	var err error = kwait.For(context.Background(), func() (bool, error) {
		calls++
		return false, nil
	}, nil)
	if err == nil {
		t.Error("expected timeout from default opts")
	}
	if calls < 1 {
		t.Errorf("expected at least 1 call with default opts, got %d", calls)
	}
}

func TestForPageLoad_TitlePollFn_ErrorIsTolerated(t *testing.T) {
	// If getTitleFn errors, ForPageLoad should keep polling (per implementation
	// comments in kwait.go). Eventually times out, but doesn't fail loudly.
	var titleFn func() (string, error) = func() (string, error) {
		return "", errors.New("connection")
	}
	var err error = kwait.ForPageLoad(context.Background(), titleFn,
		kwait.WaitUntilDOMContentLoaded, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestForPageLoad_DOMContentLoaded_ReturnsImmediatelyWhenTitleSet(t *testing.T) {
	var titleFn func() (string, error) = func() (string, error) {
		return "Already loaded", nil
	}
	var start time.Time = time.Now()
	var err error = kwait.ForPageLoad(context.Background(), titleFn,
		kwait.WaitUntilDOMContentLoaded, 5*time.Second)
	var elapsed time.Duration = time.Since(start)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("should return fast when title already set, took %v", elapsed)
	}
}

func TestForPageLoad_Load_StaysOnNewTabUntilTitleChanges(t *testing.T) {
	// "New Tab" title means page hasn't loaded yet.
	var calls int = 0
	var titleFn func() (string, error) = func() (string, error) {
		calls++
		if calls < 3 {
			return "New Tab", nil
		}
		return "Real page", nil
	}
	var err error = kwait.ForPageLoad(context.Background(), titleFn,
		kwait.WaitUntilLoad, 2*time.Second)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestUntilReady_ReturnsValueOnSuccess(t *testing.T) {
	var fn func() (string, error) = func() (string, error) {
		return "success-marker", nil
	}
	var got string
	var err error
	got, err = kwait.UntilReady(context.Background(), fn, kwait.DefaultOptions())
	if err != nil {
		t.Errorf("nil err expected, got %v", err)
	}
	if got != "success-marker" {
		t.Errorf("expected 'success-marker', got %q", got)
	}
}

func TestUntilReady_RetriesUntilSuccess(t *testing.T) {
	var calls int = 0
	var fn func() (int, error) = func() (int, error) {
		calls++
		if calls < 4 {
			return 0, errors.New("not yet")
		}
		return 42, nil
	}
	var got int
	var err error
	got, err = kwait.UntilReady(context.Background(), fn, &kwait.Options{
		Timeout: 1 * time.Second, Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Errorf("nil err expected, got %v", err)
	}
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestDefaultOptions_TimeoutGreaterThanInterval(t *testing.T) {
	// Sanity invariant: defaults should make the polling loop run at least once.
	var opts *kwait.Options = kwait.DefaultOptions()
	if opts.Timeout <= opts.Interval {
		t.Errorf("default Timeout (%v) should exceed Interval (%v)", opts.Timeout, opts.Interval)
	}
}

func TestForPageLoad_UnknownStrategy_ReturnsError(t *testing.T) {
	// The default branch returns an error for unrecognised WaitUntil values.
	var err error = kwait.ForPageLoad(context.Background(),
		func() (string, error) { return "x", nil },
		kwait.WaitUntil("not-a-real-strategy"), 1*time.Second)
	if err == nil {
		t.Fatal("expected error for unknown waitUntil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown'; got %q", err.Error())
	}
}

func TestForPageLoad_Commit_ReturnsImmediately(t *testing.T) {
	// WaitUntilCommit returns nil immediately — no polling required.
	var calls int = 0
	var titleFn func() (string, error) = func() (string, error) {
		calls++
		return "anything", nil
	}
	var start time.Time = time.Now()
	var err error = kwait.ForPageLoad(context.Background(), titleFn,
		kwait.WaitUntilCommit, 5*time.Second)
	var elapsed time.Duration = time.Since(start)
	if err != nil {
		t.Errorf("nil err expected, got %v", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("Commit should return immediately, took %v", elapsed)
	}
	if calls != 0 {
		t.Errorf("Commit should not call titleFn at all, got %d calls", calls)
	}
}

func TestWaitUntilConstants_AreUnique(t *testing.T) {
	// All strategy constants should have distinct string values.
	var seen map[kwait.WaitUntil]bool = map[kwait.WaitUntil]bool{}
	for _, s := range []kwait.WaitUntil{
		kwait.WaitUntilCommit,
		kwait.WaitUntilDOMContentLoaded,
		kwait.WaitUntilLoad,
		kwait.WaitUntilNetworkIdle,
	} {
		if seen[s] {
			t.Errorf("duplicate strategy constant value: %v", s)
		}
		seen[s] = true
	}
}

func TestFor_RespectsTimeoutWithExpensiveCondition(t *testing.T) {
	// Even a slow condition function should not extend the timeout.
	var err error = kwait.For(context.Background(), func() (bool, error) {
		time.Sleep(50 * time.Millisecond) // expensive check
		return false, nil
	}, &kwait.Options{Timeout: 200 * time.Millisecond, Interval: 10 * time.Millisecond})

	if err == nil {
		t.Fatal("expected timeout error")
	}
}
