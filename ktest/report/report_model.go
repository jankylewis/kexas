// Package report provides HTML test report generation for ktest.
//
// It produces a single self-contained .html file with embedded CSS and JS,
// featuring dark/light theme toggle, test filtering, and summary statistics.
package report

import "time"

// TestStatus represents the outcome of a test execution.
type TestStatus string

const (
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
)

// TestStep represents a single step within a test execution.
// Inspired by Playwright's test.step() for structured debugging.
type TestStep struct {
	Title    string        `json:"title"`
	Status   string        `json:"status"` // "passed" or "failed"
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

// TestCaseResult holds the result of a single test execution.
type TestCaseResult struct {
	Name     string        `json:"name"`
	Status   TestStatus    `json:"status"`
	Duration time.Duration `json:"duration"`
	Filename string        `json:"filename"`
	WorkerID int           `json:"workerID"`
	ErrorMsg string        `json:"errorMsg,omitempty"`
	Steps    []TestStep    `json:"steps,omitempty"`
	Logs     []string      `json:"logs,omitempty"`
}

// TestSuiteResult groups test results by file or group name.
type TestSuiteResult struct {
	Name  string           `json:"name"`
	Tests []TestCaseResult `json:"tests"`
}

// TestReport is the top-level report data structure.
type TestReport struct {
	ProjectName string            `json:"projectName"`
	Timestamp   time.Time         `json:"timestamp"`
	Duration    time.Duration     `json:"duration"`
	Suites      []TestSuiteResult `json:"suites"`
	TotalTests  int               `json:"totalTests"`
	Passed      int               `json:"passed"`
	Failed      int               `json:"failed"`
	Skipped     int               `json:"skipped"`
	PassRate    float64           `json:"passRate"`
	Workers     int               `json:"workers"`
}

// ComputeStats calculates TotalTests, Passed, Failed, Skipped, and PassRate
// from the Suites data. Call this after all results have been added.
func (r *TestReport) ComputeStats() {
	r.TotalTests = 0
	r.Passed = 0
	r.Failed = 0
	r.Skipped = 0

	for _, suite := range r.Suites {
		for _, tc := range suite.Tests {
			r.TotalTests++
			switch tc.Status {
			case StatusPassed:
				r.Passed++
			case StatusFailed:
				r.Failed++
			case StatusSkipped:
				r.Skipped++
			}
		}
	}

	if r.TotalTests > 0 {
		r.PassRate = float64(r.Passed) / float64(r.TotalTests) * 100.0
	}
}
