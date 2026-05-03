package ktest

import (
	"fmt"
	"strings"
	"time"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/klauncher"
)

// maxLaunchRetries is the number of times to retry a failed browser launch.
// Chrome can transiently fail under parallel load (resource contention, temp file races).
const maxLaunchRetries int = 5

// launchIsolatedBrowser creates a new browser+page pair for one test.
// Retries up to maxLaunchRetries times with exponential backoff on failure.
// The wrapped error preserves whether the failure was in browser launch or first-page.
func launchIsolatedBrowser(config *Config) (*kexas.Browser, *kexas.Page, error) {
	var opts *klauncher.Options = klauncher.DefaultOptions()
	opts.Headless = config.Headless
	if config.BrowserExecutable != "" {
		opts.ExecutablePath = config.BrowserExecutable
	}

	var lastErr error
	for attempt := 1; attempt <= maxLaunchRetries; attempt++ {
		var browser *kexas.Browser
		var page *kexas.Page
		var err error
		browser, page, err = attemptLaunch(opts)
		if err == nil {
			if attempt > 1 {
				ktestLog.Info("browser launch retry succeeded", "attempt", attempt, "totalRetries", attempt-1)
			}
			return browser, page, nil
		}
		lastErr = err
		if attempt < maxLaunchRetries {
			// Exponential backoff: 500ms, 1s, 2s, 4s
			var backoff time.Duration = time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
			ktestLog.Warn("browser launch failed, retrying",
				"attempt", attempt,
				"maxRetries", maxLaunchRetries,
				"backoff", backoff.String(),
				"err", err,
			)
			time.Sleep(backoff)
		}
	}
	return nil, nil, fmt.Errorf("browser launch (after %d attempts): %w", maxLaunchRetries, lastErr)
}

// attemptLaunch performs a single browser+page launch attempt with no retry.
// On FirstPage failure the browser is closed before returning to avoid leaking it.
func attemptLaunch(opts *klauncher.Options) (*kexas.Browser, *kexas.Page, error) {
	var browser *kexas.Browser
	var err error
	browser, err = kexas.Launch(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("browser launch: %w", err)
	}

	var page *kexas.Page
	page, err = browser.FirstPage()
	if err != nil {
		browser.Close()
		return nil, nil, fmt.Errorf("first page: %w", err)
	}
	return browser, page, nil
}

// runTestWithRecovery executes the test function with panic recovery and screenshot.
func runTestWithRecovery(
	childT *ktestT,
	test NamedTest,
	page *kexas.Page,
	shortName string,
	workerID int,
	config *Config,
) {
	defer func() {
		if r := recover(); r != nil {
			childT.Errorf("❌ worker[%d] test %s panicked: %v",
				workerID, shortName, r,
			)
			if config.ScreenshotOnFail {
				var screenshotName string = fmt.Sprintf("%s.%s", test.Filename, shortName)
				takeScreenshot(childT, page, screenshotName, config.ScreenshotDir)
			}
		}
	}()

	test.Func(page, childT)

	if childT.Failed() && config.ScreenshotOnFail {
		var screenshotName string = fmt.Sprintf("%s.%s", test.Filename, shortName)
		takeScreenshot(childT, page, screenshotName, config.ScreenshotDir)
	}
}

// buildResult constructs a testResult from a completed test.
func buildResult(test NamedTest, childT *ktestT, start time.Time, workerID int) testResult {
	var r testResult = testResult{
		name:     test.Name,
		passed:   !childT.Failed(),
		filename: test.Filename,
		elapsed:  time.Since(start),
		workerID: workerID,
	}
	if childT.Failed() && len(childT.errors) > 0 {
		r.errorMsg = childT.errors[len(childT.errors)-1]
	}
	r.steps = childT.steps
	r.logs = childT.logs
	return r
}

// extractShortName extracts the last segment of a dotted test name.
func extractShortName(fullName string) string {
	if idx := strings.LastIndex(fullName, "."); idx >= 0 {
		return fullName[idx+1:]
	}
	return fullName
}
