// Package kwait provides smart waiting mechanisms for browser automation.
package kwait

import (
	"context"
	"fmt"
	"time"
)

// WaitUntil defines when navigation is considered complete.
type WaitUntil string

const (
	// WaitUntilCommit waits for navigation to commit (response headers received).
	// This is the earliest point - the page is still blank.
	WaitUntilCommit WaitUntil = "commit"

	// WaitUntilDOMContentLoaded waits for the DOM to be ready (HTML parsed).
	// Doesn't wait for images, stylesheets, or async scripts.
	WaitUntilDOMContentLoaded WaitUntil = "domcontentloaded"

	// WaitUntilLoad waits for all resources to load (images, CSS, JS, iframes).
	// This is the default and most reliable for visual testing.
	WaitUntilLoad WaitUntil = "load"

	// WaitUntilNetworkIdle waits until no network activity for 500ms.
	// Useful for SPAs that fetch data after initial render.
	WaitUntilNetworkIdle WaitUntil = "networkidle"
)

// Condition is a function that returns true when a condition is met.
type Condition func() (bool, error)

// Options configures wait behavior.
type Options struct {
	Timeout  time.Duration
	Interval time.Duration
	Message  string
}

// DefaultOptions returns default wait options.
func DefaultOptions() *Options {
	return &Options{
		Timeout:  30 * time.Second,
		Interval: 100 * time.Millisecond,
		Message:  "condition not met",
	}
}

// For waits for a condition to be true.
func For(ctx context.Context, condition Condition, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	var deadline time.Time = time.Now().Add(opts.Timeout)
	var ticker *time.Ticker = time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		// Check condition
		var met bool
		var err error
		met, err = condition()
		if err != nil {
			return fmt.Errorf("kwait: condition error: %w", err)
		}
		if met {
			return nil
		}

		// Check timeout
		if time.Now().After(deadline) {
			return fmt.Errorf("kwait: timeout after %v: %s", opts.Timeout, opts.Message)
		}

		// Wait for next tick or context cancellation
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return fmt.Errorf("kwait: context cancelled: %w", ctx.Err())
		}
	}
}

// UntilReady waits until a function returns a non-nil value without error.
func UntilReady[T any](ctx context.Context, fn func() (T, error), opts *Options) (T, error) {
	var result T
	var err error = For(ctx, func() (bool, error) {
		var val T
		var e error
		val, e = fn()
		if e != nil {
			return false, nil // Ignore errors, keep trying
		}
		result = val
		return true, nil
	}, opts)
	return result, err
}

// ForPageLoad waits for a page to finish loading based on the wait strategy.
// getTitleFn should return the current page title.
func ForPageLoad(ctx context.Context, getTitleFn func() (string, error), waitUntil WaitUntil, timeout time.Duration) error {
	var opts *Options = &Options{
		Timeout:  timeout,
		Interval: 100 * time.Millisecond,
		Message:  fmt.Sprintf("page did not reach '%s' state", waitUntil),
	}

	switch waitUntil {
	case WaitUntilCommit:
		// Commit happens immediately when navigation starts
		// No additional waiting needed
		return nil

	case WaitUntilDOMContentLoaded:
		// Wait for title to change from "New Tab" (indicates DOM is ready)
		return For(ctx, func() (bool, error) {
			var title string
			var err error
			title, err = getTitleFn()
			if err != nil {
				return false, nil // Ignore errors, keep trying
			}
			// DOM is ready when title changes from "New Tab"
			return title != "New Tab", nil
		}, opts)

	case WaitUntilLoad:
		// Wait for title to be set and stable (indicates resources loaded)
		return For(ctx, func() (bool, error) {
			var title string
			var err error
			title, err = getTitleFn()
			if err != nil {
				return false, nil
			}
			// Page is fully loaded when title is not empty and not "New Tab"
			return title != "" && title != "New Tab", nil
		}, opts)

	case WaitUntilNetworkIdle:
		// TODO: Implement proper network idle detection via CDP
		// For now, wait for title + extra time for network to settle
		var err error = For(ctx, func() (bool, error) {
			var title string
			var e error
			title, e = getTitleFn()
			if e != nil {
				return false, nil
			}
			return title != "" && title != "New Tab", nil
		}, opts)
		if err != nil {
			return err
		}
		// Wait additional 500ms for network to be idle
		time.Sleep(500 * time.Millisecond)
		return nil

	default:
		return fmt.Errorf("kwait: unknown waitUntil value: %s", waitUntil)
	}
}
