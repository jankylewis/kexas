package ktest

import (
	"fmt"
	"os"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/launcher"
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
// ========================================

// runRegisteredTests executes all tests registered via AlphaInit/Group system
func runRegisteredTests(tests []NamedTest) {
	// Create test runner
	var t KTestT = newKTestT()

	// Load configuration and apply it
	var config *Config = loadConfig()

	// Create screenshot directory if screenshot on fail is enabled
	if config.ScreenshotOnFail {
		os.MkdirAll(config.ScreenshotDir, 0755)
		fmt.Printf("📸 Screenshots enabled: %s\n", config.ScreenshotDir)
	}

	// Execute global BeforeAll hook
	ExecuteGlobalBeforeAll()

	// Run each test with its own browser (sequential execution)
	for _, test := range tests {
		var testName string = test.Name
		t.Run(testName, func(t KTestT) {
			t.Logf("🧪 ktest: running %s", testName)

			// Launch new browser for this test only
			fmt.Printf("🌐 ktest: launching browser for %s\n", testName)
			var opts *launcher.Options = launcher.DefaultOptions()
			opts.Headless = config.Headless
			if config.BrowserExecutable != "" {
				opts.ExecutablePath = config.BrowserExecutable
			}

			var browser *kexas.Browser
			var err error
			browser, err = kexas.Launch(opts)
			if err != nil {
				t.Errorf("❌ ktest: failed to launch browser: %v", err)
				return
			}
			defer browser.Close() // Guaranteed cleanup after each test

			// Get the first page (single tab only)
			var page *kexas.Page
			page, err = browser.FirstPage()
			if err != nil {
				t.Errorf("❌ ktest: failed to get first page: %v", err)
				return
			}
			defer page.Close()

			// Execute global BeforeEach hook
			ExecuteGlobalBeforeEach(page)

			// Recover from panics and take screenshot if needed
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("❌ ktest: test %s panicked: %v", testName, r)

					// Take screenshot on failure if enabled
					if config.ScreenshotOnFail {
						takeScreenshot(t, page, testName, config.ScreenshotDir)
					}
				}

				// Execute global AfterEach hook
				ExecuteGlobalAfterEach(page)
			}()

			// Call the registered test function
			test.Func(page, t)

			// Check if test failed and take screenshot
			if t.Failed() && config.ScreenshotOnFail {
				takeScreenshot(t, page, testName, config.ScreenshotDir)
			}

			t.Logf("✅ ktest: %s completed", testName)
		})
	}

	// Execute global AfterAll hook
	ExecuteGlobalAfterAll()

	// Print summary
	if t.Failed() {
		fmt.Println("\n❌ Some tests failed")
		os.Exit(1)
	} else {
		fmt.Printf("\n✅ All %d tests passed!\n", len(tests))
	}
}
