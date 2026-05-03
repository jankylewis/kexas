package ktest

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/internal/logger"
	"github.com/jankylewis/kexas/kassert"
)

// ========================================
// KTEST EXECUTION ENGINE FOR REGISTERED TESTS
// ========================================
//
// This file provides the execution engine for tests registered via
// the AlphaInit zero-boilerplate pattern. It connects the registration
// system to the actual test execution.
//
// Usage:
//   var _ = kexas.AlphaInit(
//       ktest.Test("MyTest", func(page *kexas.Page) {
//           page.Navigate("https://example.com")
//       }),
//   )
//
//   func main() {
//       ktest.AutoRun() // Executes registered tests
//   }

func formatTestDisplayName(test NamedTest) string {
	var filename string = test.Filename
	if filename == "" {
		filename = "test"
	}
	var shortName string = extractShortName(test.Name)
	return fmt.Sprintf("<%s.%s>", filename, shortName)
}

func formatResultDisplayName(result testResult) string {
	var filename string = result.filename
	if filename == "" {
		filename = "test"
	}
	var shortName string = extractShortName(result.name)
	return fmt.Sprintf("<%s.%s>", filename, shortName)
}

// ========================================

// ktestLog is the package-level logger for ktest lifecycle events.
var ktestLog *logger.Logger = logger.New("ktest")

// runRegisteredTests executes all tests registered via AlphaInit/Group system.
// Dispatches to sequential or parallel execution based on config.ParallelSet.
func runRegisteredTests(tests []NamedTest) {
	// Create test runner
	var t KTestT = newKTestT()

	// Reset assertion stats for this run
	kassert.ResetStats()

	// Filter tests by KEXAS_TEST_RUN environment variable
	var filteredTests []NamedTest = filterTests(tests)
	if len(filteredTests) == 0 {
		return
	}

	// Load configuration and apply it
	var config *Config = loadConfig()

	// Create screenshot directory if screenshot on fail is enabled
	if config.ScreenshotOnFail {
		os.MkdirAll(config.ScreenshotDir, 0755)
		fmt.Printf("📸 Screenshots enabled: %s\n", config.ScreenshotDir)
	}

	// Execute global BeforeAll hook (always on main goroutine)
	ktestLog.Info("executing global BeforeAll hook")
	ExecuteGlobalBeforeAll()

	// Dispatch: parallel or sequential
	var results []testResult
	if config.ParallelSet > 1 {
		results = runTestsParallelWorkers(t, filteredTests, config, config.ParallelSet)
	} else {
		results = runTestsSequentialRegistered(t, filteredTests, config)
	}

	// Execute global AfterAll hook (always on main goroutine)
	ktestLog.Info("executing global AfterAll hook")
	ExecuteGlobalAfterAll()

	// Print assertion stats
	kassert.PrintStats()

	// Generate HTML report
	generateHTMLReport(results, config)

	// Print summary and exit
	printTestSummary(results, filteredTests)
}

// filterTests applies the KEXAS_TEST_RUN environment variable filter.
func filterTests(tests []NamedTest) []NamedTest {
	var testFilter string = os.Getenv("KEXAS_TEST_RUN")
	if testFilter == "" {
		return tests
	}

	fmt.Printf("🔍 KEXAS_TEST_RUN filter: %s\n", testFilter)
	var filtered []NamedTest
	for _, test := range tests {
		if test.Name == testFilter || strings.Contains(test.Name, testFilter) {
			filtered = append(filtered, test)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("❌ No tests found matching filter: %s\n", testFilter)
		fmt.Printf("Available tests:\n")
		for _, test := range tests {
			fmt.Printf("  - %s\n", test.Name)
		}
		return nil
	}

	fmt.Printf("📋 Running %d filtered test(s):\n", len(filtered))
	for _, test := range filtered {
		fmt.Printf("  - %s\n", formatTestDisplayName(test))
	}
	return filtered
}

// runTestsSequentialRegistered runs tests one-by-one (parallelSet=1 path).
func runTestsSequentialRegistered(
	t KTestT,
	tests []NamedTest,
	config *Config,
) []testResult {
	SortTestsByPriority(tests)

	var results []testResult
	for _, test := range tests {
		var testName string = test.Name
		var shortName string = extractShortName(testName)
		var screenshotName string = fmt.Sprintf("%s.%s", test.Filename, shortName)
		var testPassed bool
		var testDuration time.Duration

		t.Run(testName, func(t KTestT) {
			testPassed, testDuration = executeSequentialTest(t, test, config, shortName, screenshotName)
		})

		results = append(results, testResult{
			name:     testName,
			passed:   testPassed,
			filename: test.Filename,
			elapsed:  testDuration,
			workerID: 0,
		})
	}
	return results
}

// executeSequentialTest runs one registered test in an isolated browser, taking screenshots
// on failure and invoking the global Before/After hooks. Duration is captured via deferred
// assignment so it includes browser teardown time, matching the original behavior.
func executeSequentialTest(t KTestT, test NamedTest, config *Config, shortName, screenshotName string) (testPassed bool, duration time.Duration) {
	var testStart time.Time = time.Now()
	testPassed = true
	defer func() { duration = time.Since(testStart) }()

	ktestLog.Info("starting test", "name", shortName)
	t.Logf("🧪 ktest: running %s", shortName)

	ktestLog.Info("launching browser", "test", shortName)
	fmt.Printf("🌐 ktest: launching browser for %s\n", shortName)
	var browser *kexas.Browser
	var page *kexas.Page
	var err error
	browser, page, err = launchIsolatedBrowser(config)
	if err != nil {
		t.Errorf("❌ ktest: failed to launch browser: %v", err)
		testPassed = false
		return
	}
	defer browser.Close()
	defer page.Close()
	ktestLog.Info("page created", "test", shortName)

	ktestLog.Debug("executing BeforeEach hook", "test", shortName)
	ExecuteGlobalBeforeEach(page)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("❌ ktest: test %s panicked: %v", test.Name, r)
			testPassed = false
			if config.ScreenshotOnFail {
				takeScreenshot(t, page, screenshotName, config.ScreenshotDir)
			}
		}
		ExecuteGlobalAfterEach(page)
	}()

	test.Func(page, t)

	if t.Failed() {
		testPassed = false
		if config.ScreenshotOnFail {
			takeScreenshot(t, page, screenshotName, config.ScreenshotDir)
		}
	}

	var elapsed time.Duration = time.Since(testStart)
	ktestLog.Info("completed test", "name", shortName, "elapsed", elapsed.String())
	t.Logf("✅ ktest: %s completed", shortName)
	return
}

