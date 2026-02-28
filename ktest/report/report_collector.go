package report

import (
	"path/filepath"
	"sync"
	"time"
)

// Collector accumulates test results during execution.
// Thread-safe: guarded by sync.Mutex for parallel test runs.
type Collector struct {
	mu        sync.Mutex
	results   []TestCaseResult
	startTime time.Time
	workers   int
}

// NewCollector creates a new result collector.
// Call this before test execution begins.
func NewCollector(workers int) *Collector {
	return &Collector{
		results:   make([]TestCaseResult, 0),
		startTime: time.Now(),
		workers:   workers,
	}
}

// Add records a single test result. Safe for concurrent use.
func (c *Collector) Add(result TestCaseResult) {
	c.mu.Lock()
	c.results = append(c.results, result)
	c.mu.Unlock()
}

// Results returns a copy of all collected results.
func (c *Collector) Results() []TestCaseResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	var copy []TestCaseResult = make([]TestCaseResult, len(c.results))
	for i, r := range c.results {
		copy[i] = r
	}
	return copy
}

// BuildReport assembles a TestReport from the collected results.
// Groups tests by filename into suites.
func (c *Collector) BuildReport(projectName string) *TestReport {
	c.mu.Lock()
	defer c.mu.Unlock()

	var suiteMap map[string]*TestSuiteResult = make(map[string]*TestSuiteResult)
	var suiteOrder []string

	for _, result := range c.results {
		var suiteName string = result.Filename
		if suiteName == "" {
			suiteName = "default"
		}

		if _, exists := suiteMap[suiteName]; !exists {
			suiteMap[suiteName] = &TestSuiteResult{
				Name:  suiteName,
				Tests: make([]TestCaseResult, 0),
			}
			suiteOrder = append(suiteOrder, suiteName)
		}
		suiteMap[suiteName].Tests = append(suiteMap[suiteName].Tests, result)
	}

	// Preserve insertion order
	var suites []TestSuiteResult
	for _, name := range suiteOrder {
		suites = append(suites, *suiteMap[name])
	}

	var report *TestReport = &TestReport{
		ProjectName: projectName,
		Timestamp:   c.startTime,
		Duration:    time.Since(c.startTime),
		Suites:      suites,
		Workers:     c.workers,
	}

	report.ComputeStats()
	return report
}

// ProjectNameFromDir extracts a project name from a directory path.
// Uses the last path component as the name.
func ProjectNameFromDir(dir string) string {
	var base string = filepath.Base(dir)
	if base == "." || base == "/" {
		return "kexas"
	}
	return base
}
