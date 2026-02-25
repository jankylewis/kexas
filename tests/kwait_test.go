package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kexas-project/kexas/kwait"
)

// TestFor_ConditionMetImmediately tests that For returns immediately when condition is already true.
func TestFor_ConditionMetImmediately(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "test condition",
	}

	var callCount int = 0
	var condition kwait.Condition = func() (bool, error) {
		callCount++
		return true, nil
	}

	var err error = kwait.For(ctx, condition, opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected condition to be called once, got %d calls", callCount)
	}
}

// TestFor_ConditionMetAfterRetries tests that For waits and retries until condition becomes true.
func TestFor_ConditionMetAfterRetries(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  2 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "test condition",
	}

	var callCount int = 0
	var condition kwait.Condition = func() (bool, error) {
		callCount++
		// Return true on the 3rd call
		return callCount >= 3, nil
	}

	var startTime time.Time = time.Now()
	var err error = kwait.For(ctx, condition, opts)
	var elapsed time.Duration = time.Since(startTime)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
	// Should have waited at least 200ms (2 intervals)
	if elapsed < 200*time.Millisecond {
		t.Errorf("expected to wait at least 200ms, waited %v", elapsed)
	}
}

// TestFor_Timeout tests that For returns timeout error when condition never becomes true.
func TestFor_Timeout(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  500 * time.Millisecond,
		Interval: 100 * time.Millisecond,
		Message:  "test timeout",
	}

	var condition kwait.Condition = func() (bool, error) {
		return false, nil // Never true
	}

	var startTime time.Time = time.Now()
	var err error = kwait.For(ctx, condition, opts)
	var elapsed time.Duration = time.Since(startTime)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if elapsed < 500*time.Millisecond {
		t.Errorf("expected to wait at least 500ms, waited %v", elapsed)
	}
}

// TestFor_ConditionError tests that For returns error when condition function returns error.
func TestFor_ConditionError(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "test error",
	}

	var expectedErr error = errors.New("condition failed")
	var condition kwait.Condition = func() (bool, error) {
		return false, expectedErr
	}

	var err error = kwait.For(ctx, condition, opts)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got: %v", expectedErr, err)
	}
}

// TestFor_ContextCancellation tests that For respects context cancellation.
func TestFor_ContextCancellation(t *testing.T) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	var opts *kwait.Options = &kwait.Options{
		Timeout:  5 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "test cancellation",
	}

	var condition kwait.Condition = func() (bool, error) {
		return false, nil // Never true
	}

	// Cancel context after 200ms
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	var startTime time.Time = time.Now()
	var err error = kwait.For(ctx, condition, opts)
	var elapsed time.Duration = time.Since(startTime)

	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
	// Should return quickly after cancellation, not wait for full timeout
	if elapsed > 1*time.Second {
		t.Errorf("expected to return quickly after cancellation, took %v", elapsed)
	}
}

// TestFor_NilOptions tests that For uses default options when opts is nil.
func TestFor_NilOptions(t *testing.T) {
	var ctx context.Context = context.Background()

	var condition kwait.Condition = func() (bool, error) {
		return true, nil
	}

	var err error = kwait.For(ctx, condition, nil)
	if err != nil {
		t.Errorf("expected no error with nil options, got: %v", err)
	}
}

// TestUntilReady_Success tests that UntilReady returns value when function succeeds.
func TestUntilReady_Success(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "test ready",
	}

	var callCount int = 0
	var fn func() (string, error) = func() (string, error) {
		callCount++
		if callCount >= 2 {
			return "success", nil
		}
		return "", errors.New("not ready")
	}

	var result string
	var err error
	result, err = kwait.UntilReady(ctx, fn, opts)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got: %s", result)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls, got %d", callCount)
	}
}

// TestUntilReady_Timeout tests that UntilReady times out when function never succeeds.
func TestUntilReady_Timeout(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *kwait.Options = &kwait.Options{
		Timeout:  500 * time.Millisecond,
		Interval: 100 * time.Millisecond,
		Message:  "test timeout",
	}

	var fn func() (string, error) = func() (string, error) {
		return "", errors.New("always fails")
	}

	var result string
	var err error
	result, err = kwait.UntilReady(ctx, fn, opts)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if result != "" {
		t.Errorf("expected empty result on timeout, got: %s", result)
	}
}

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
