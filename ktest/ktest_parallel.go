package ktest

import (
	"fmt"
	"sync"
	"time"

	"github.com/jankylewis/kexas"
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
	name           string
	passed         bool
	filename       string
	elapsed        time.Duration
	workerID       int
	errorMsg       string
	steps          []testStep
	logs           []string
	screenshotPath string // end-of-test screenshot, captured for every test
	videoPath      string // per-test mp4/frames recording, captured for every test
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
	SortTestsByPriority(tests)

	var workerCount int = parallelSet
	if workerCount > len(tests) {
		workerCount = len(tests)
	}

	ktestLog.Info("parallel execution",
		"workers", workerCount,
		"tests", len(tests),
	)
	fmt.Printf("⚡ Parallel mode: %d workers for %d tests\n", workerCount, len(tests))

	var testQueue chan NamedTest = make(chan NamedTest, len(tests))
	var resultsChan chan testResult = make(chan testResult, len(tests))

	for _, test := range tests {
		testQueue <- test
	}
	close(testQueue)

	// Stagger worker startup by 300ms each to reduce Chrome startup contention.
	// Without stagger, all workers race to launch Chrome simultaneously,
	// competing for OS resources (file handles, temp dirs, CPU).
	var wg sync.WaitGroup
	for workerID := 0; workerID < workerCount; workerID++ {
		wg.Add(1)
		var id int = workerID
		go func() {
			if id > 0 {
				time.Sleep(time.Duration(id) * 300 * time.Millisecond)
			}
			defer wg.Done()
			runWorker(id, parentT, testQueue, resultsChan, config)
		}()
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

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

	ktestLog.Info("worker starting test", "worker", workerID, "test", shortName)

	var childT *ktestT = &ktestT{
		name:   test.Name,
		failed: false,
		depth:  0,
	}

	var browser *kexas.Browser
	var page *kexas.Page
	var launchErr error
	browser, page, launchErr = launchIsolatedBrowser(config)
	if launchErr != nil {
		return makeLaunchFailureResult(test, childT, workerID, shortName, testStart, launchErr)
	}
	defer browser.Close()
	defer page.Close()

	ExecuteGlobalBeforeEach(page)
	runTestWithRecovery(childT, test, page, shortName, workerID, config)
	ExecuteGlobalAfterEach(page)

	logWorkerTestComplete(workerID, shortName, childT, testStart)

	if childT.Failed() {
		parentT.Errorf("❌ %s failed", shortName)
	}
	return buildResult(test, childT, testStart, workerID)
}

// makeLaunchFailureResult builds the error testResult for a worker that could not
// launch a browser. childT receives the error so the result reflects "FAIL".
func makeLaunchFailureResult(test NamedTest, childT *ktestT, workerID int, shortName string, testStart time.Time, launchErr error) testResult {
	var errMsg string = fmt.Sprintf("worker[%d] failed to launch browser for %s: %v",
		workerID, shortName, launchErr,
	)
	childT.Errorf("❌ %s", errMsg)
	var r testResult = buildResult(test, childT, testStart, workerID)
	r.errorMsg = errMsg
	return r
}

// logWorkerTestComplete emits the standard "worker finished test" log line with
// status (PASS/FAIL) derived from childT and elapsed time since testStart.
func logWorkerTestComplete(workerID int, shortName string, childT *ktestT, testStart time.Time) {
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
}
