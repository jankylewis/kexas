package ktest

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/internal/logger"
	"github.com/kexas-project/kexas/kassert"
	"github.com/kexas-project/kexas/ktest/report"
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
	// Sort by priority even in sequential mode
	SortTestsByPriority(tests)

	var results []testResult
	for _, test := range tests {
		var testName string = test.Name
		var testPassed bool = true
		var shortName string = extractShortName(testName)
		var screenshotName string = fmt.Sprintf("%s.%s", test.Filename, shortName)
		var testDuration time.Duration

		t.Run(testName, func(t KTestT) {
			var testStart time.Time = time.Now()
			defer func() {
				testDuration = time.Since(testStart)
			}()
			ktestLog.Info("starting test", "name", shortName)
			t.Logf("🧪 ktest: running %s", shortName)

			// Launch new browser for this test only
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

			// Execute global BeforeEach hook
			ktestLog.Debug("executing BeforeEach hook", "test", shortName)
			ExecuteGlobalBeforeEach(page)

			// Recover from panics and take screenshot if needed
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("❌ ktest: test %s panicked: %v", testName, r)
					testPassed = false
					if config.ScreenshotOnFail {
						takeScreenshot(t, page, screenshotName, config.ScreenshotDir)
					}
				}
				ExecuteGlobalAfterEach(page)
			}()

			// Call the registered test function
			test.Func(page, t)

			// Check if test failed and take screenshot
			if t.Failed() {
				testPassed = false
				if config.ScreenshotOnFail {
					takeScreenshot(t, page, screenshotName, config.ScreenshotDir)
				}
			}

			var elapsed time.Duration = time.Since(testStart)
			ktestLog.Info("completed test", "name", shortName, "elapsed", elapsed.String())
			t.Logf("✅ ktest: %s completed", shortName)
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

// generateHTMLReport converts test results into an HTML report file.
func generateHTMLReport(results []testResult, config *Config) {
	var collector *report.Collector = report.NewCollector(config.ParallelSet)

	for _, r := range results {
		var status report.TestStatus = report.StatusPassed
		if !r.passed {
			status = report.StatusFailed
		}
		var reportSteps []report.TestStep = convertSteps(r.steps)
		collector.Add(report.TestCaseResult{
			Name:     r.name,
			Status:   status,
			Duration: r.elapsed,
			Filename: r.filename,
			WorkerID: r.workerID,
			ErrorMsg: r.errorMsg,
			Steps:    reportSteps,
			Logs:     r.logs,
		})
	}

	var cwd string
	cwd, _ = os.Getwd()
	var projectName string = report.ProjectNameFromDir(cwd)
	var htmlReport *report.TestReport = collector.BuildReport(projectName)
	var outputPath string
	var versionedPath string
	var err error
	outputPath, versionedPath, err = report.GenerateFiles(htmlReport, config.ReportDir)
	if err != nil {
		ktestLog.Error("failed to generate HTML report", "err", err)
		return
	}
	fmt.Printf("\n📊 HTML report: %s\n", outputPath)
	if versionedPath != "" {
		fmt.Printf("🗂  Archived copy: %s\n", versionedPath)
	}
}

// convertSteps converts internal testStep slices to report.TestStep slices.
func convertSteps(steps []testStep) []report.TestStep {
	if len(steps) == 0 {
		return nil
	}
	var result []report.TestStep = make([]report.TestStep, len(steps))
	for i, s := range steps {
		result[i] = report.TestStep{
			Title:    s.Title,
			Status:   s.Status,
			Duration: s.Duration,
			Error:    s.Error,
		}
	}
	return result
}

// printTestSummary prints the final test summary and exits with error code if needed.
func printTestSummary(results []testResult, filteredTests []NamedTest) {
	if len(filteredTests) == 1 {
		fmt.Printf("\nTest finished:\n")
	} else {
		fmt.Printf("\nAll %d tests finished:\n", len(filteredTests))
	}

	var hasFailures bool = false
	for _, result := range results {
		var displayName string = formatResultDisplayName(result)
		if result.passed {
			fmt.Printf("%s passed\n", displayName)
		} else {
			fmt.Printf("%s failed\n", displayName)
			hasFailures = true
		}
	}

	if hasFailures {
		ktestLog.Error("test suite finished with failures")
		os.Exit(1)
	}
	ktestLog.Info("test suite finished", "total", len(filteredTests), "passed", len(filteredTests))
}
