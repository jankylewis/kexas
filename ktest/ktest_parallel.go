package ktest

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/launcher"
)

// ========================================
// KTEST PARALLEL EXECUTION ENGINE
// ========================================
//
// Provides worker-pool based parallel test execution.
// Each worker gets its own browser instance for full isolation.
//
// Design:
//   - Tests are sorted by priority, then fed into a channel.
//   - N worker goroutines consume from the channel.
//   - Each worker: launch browser → get page → run test → close browser.
//   - Results flow back through a results channel.
//   - Main goroutine collects results after all workers finish.
// ========================================

// testResult holds the outcome of a single test execution.
type testResult struct {
	name     string
	passed   bool
	filename string
	elapsed  time.Duration
	workerID int
	errorMsg string
	steps    []testStep
	logs     []string
}

// testStep represents a single step within a test execution.
// Inspired by Playwright's test.step() for structured debugging.
type testStep struct {
	Title    string
	Status   string // "passed", "failed"
	Duration time.Duration
	Error    string
}

// runTestsParallelWorkers dispatches tests to N worker goroutines.
// Each worker launches its own browser for complete isolation.
func runTestsParallelWorkers(
	parentT KTestT,
	tests []NamedTest,
	config *Config,
	parallelSet int,
) []testResult {
	// Sort tests by priority before dispatch
	SortTestsByPriority(tests)

	// Cap workers to test count — no idle workers
	var workerCount int = parallelSet
	if workerCount > len(tests) {
		workerCount = len(tests)
	}

	ktestLog.Info("parallel execution",
		"workers", workerCount,
		"tests", len(tests),
	)
	fmt.Printf("⚡ Parallel mode: %d workers for %d tests\n", workerCount, len(tests))

	// Channels
	var testQueue chan NamedTest = make(chan NamedTest, len(tests))
	var resultsChan chan testResult = make(chan testResult, len(tests))

	// Feed all tests into queue, then close
	for _, test := range tests {
		testQueue <- test
	}
	close(testQueue)

	// Launch workers with stagger delay to reduce Chrome startup contention.
	// Without stagger, all workers race to launch Chrome simultaneously,
	// competing for OS resources (file handles, temp dirs, CPU).
	var wg sync.WaitGroup
	for workerID := 0; workerID < workerCount; workerID++ {
		wg.Add(1)
		var id int = workerID
		go func() {
			// Stagger: worker 0 starts immediately, worker 1 after 300ms, etc.
			if id > 0 {
				time.Sleep(time.Duration(id) * 300 * time.Millisecond)
			}
			defer wg.Done()
			runWorker(id, parentT, testQueue, resultsChan, config)
		}()
	}

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []testResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// runWorker is a single worker goroutine that consumes tests from the queue.
// Each test gets its own browser instance for full isolation.
func runWorker(
	workerID int,
	parentT KTestT,
	testQueue <-chan NamedTest,
	resultsChan chan<- testResult,
	config *Config,
) {
	for test := range testQueue {
		var result testResult = executeSingleTest(workerID, parentT, test, config)
		resultsChan <- result
	}
}

// executeSingleTest runs one test in an isolated browser.
// Returns the result — never panics out of this function.
func executeSingleTest(
	workerID int,
	parentT KTestT,
	test NamedTest,
	config *Config,
) testResult {
	var testStart time.Time = time.Now()
	var shortName string = extractShortName(test.Name)

	ktestLog.Info("worker starting test",
		"worker", workerID,
		"test", shortName,
	)

	// Create isolated T for this test
	var childT *ktestT = &ktestT{
		name:   test.Name,
		failed: false,
		depth:  0,
	}

	// Launch isolated browser
	var browser *kexas.Browser
	var page *kexas.Page
	var launchErr error
	browser, page, launchErr = launchIsolatedBrowser(config)

	if launchErr != nil {
		var errMsg string = fmt.Sprintf("worker[%d] failed to launch browser for %s: %v",
			workerID, shortName, launchErr,
		)
		childT.Errorf("❌ %s", errMsg)
		var r testResult = buildResult(test, childT, testStart, workerID)
		r.errorMsg = errMsg
		return r
	}
	defer browser.Close()
	defer page.Close()

	// Execute BeforeEach hook
	ExecuteGlobalBeforeEach(page)

	// Run the test with panic recovery
	runTestWithRecovery(childT, test, page, shortName, workerID, config)

	// Execute AfterEach hook
	ExecuteGlobalAfterEach(page)

	var elapsed time.Duration = time.Since(testStart)
	var status string = "PASS"
	if childT.Failed() {
		status = "FAIL"
	}

	ktestLog.Info("worker finished test",
		"worker", workerID,
		"test", shortName,
		"status", status,
		"elapsed", elapsed.String(),
	)

	// Propagate failure to parent T
	if childT.Failed() {
		parentT.Errorf("❌ %s failed", shortName)
	}

	return buildResult(test, childT, testStart, workerID)
}

// maxLaunchRetries is the number of times to retry a failed browser launch.
// Chrome can transiently fail under parallel load (resource contention, temp file races).
const maxLaunchRetries int = 5

// launchIsolatedBrowser creates a new browser+page pair for one test.
// Retries up to maxLaunchRetries times with exponential backoff on failure.
func launchIsolatedBrowser(config *Config) (*kexas.Browser, *kexas.Page, error) {
	var opts *launcher.Options = launcher.DefaultOptions()
	opts.Headless = config.Headless
	if config.BrowserExecutable != "" {
		opts.ExecutablePath = config.BrowserExecutable
	}

	var lastErr error
	for attempt := 1; attempt <= maxLaunchRetries; attempt++ {
		var browser *kexas.Browser
		var err error
		browser, err = kexas.Launch(opts)
		if err != nil {
			lastErr = err
			if attempt < maxLaunchRetries {
				// Exponential backoff: 500ms, 1s, 2s
				var backoff time.Duration = time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
				ktestLog.Warn("browser launch failed, retrying",
					"attempt", attempt,
					"maxRetries", maxLaunchRetries,
					"backoff", backoff.String(),
					"err", err,
				)
				time.Sleep(backoff)
				continue
			}
			return nil, nil, fmt.Errorf("browser launch (after %d attempts): %w", maxLaunchRetries, lastErr)
		}

		var page *kexas.Page
		page, err = browser.FirstPage()
		if err != nil {
			browser.Close()
			lastErr = err
			if attempt < maxLaunchRetries {
				var backoff time.Duration = time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
				ktestLog.Warn("first page failed, retrying",
					"attempt", attempt,
					"backoff", backoff.String(),
					"err", err,
				)
				time.Sleep(backoff)
				continue
			}
			return nil, nil, fmt.Errorf("first page (after %d attempts): %w", maxLaunchRetries, lastErr)
		}

		if attempt > 1 {
			ktestLog.Info("browser launch retry succeeded",
				"attempt", attempt,
				"totalRetries", attempt-1,
			)
		}

		return browser, page, nil
	}

	return nil, nil, fmt.Errorf("browser launch exhausted %d retries: %w", maxLaunchRetries, lastErr)
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

	// Screenshot on failure
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
