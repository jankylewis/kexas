
package kwait_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kwait"
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
