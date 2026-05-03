package report

import (
	"strings"
	"testing"
	"time"
)

// --- End-to-end HTML generation ---

func TestGenerateHTML_ContainsAllCriticalElements(t *testing.T) {
	var report *TestReport = criticalElementsReport()
	report.ComputeStats()

	var html string = buildHTML(report)
	assertCriticalHTMLElements(t, html)
}

// criticalElementsReport builds the two-test, two-worker fixture used to verify
// the HTML report includes every cosmetic + structural element listed below.
func criticalElementsReport() *TestReport {
	return &TestReport{
		ProjectName: "myproject",
		Timestamp:   time.Now(),
		Workers:     2,
		Suites: []TestSuiteResult{
			{
				Name: "login_tests",
				Tests: []TestCaseResult{
					{
						Name:     "Login.TestHappy",
						Status:   StatusPassed,
						Duration: 2 * time.Second,
						WorkerID: 0,
						Steps:    []TestStep{{Title: "Navigate", Status: "passed", Duration: 100 * time.Millisecond}},
						Logs:     []string{"page loaded"},
					},
					{
						Name:     "Login.TestBad",
						Status:   StatusFailed,
						Duration: 5 * time.Second,
						WorkerID: 1,
						ErrorMsg: "timeout",
						Steps:    []TestStep{{Title: "Click", Status: "failed", Duration: 50 * time.Millisecond, Error: "not found"}},
						Logs:     []string{"retrying..."},
					},
				},
			},
		},
	}
}

// assertCriticalHTMLElements checks that every required substring appears in html.
// Each missing substring is reported as a separate t.Error so a single run surfaces
// all gaps, not just the first one.
func assertCriticalHTMLElements(t *testing.T, html string) {
	type check struct {
		substring string
		message   string
	}
	var checks []check = []check{
		{"<!DOCTYPE html>", "must start with DOCTYPE"},
		{"<html", "must contain html tag"},
		{"theme-toggle", "must contain theme toggle button"},
		{`data-filter="all"`, "must contain All filter button"},
		{`data-filter="passed"`, "must contain Passed filter button"},
		{`data-filter="failed"`, "must contain Failed filter button"},
		{`data-filter="skipped"`, "must contain Skipped filter button"},
		{`id="search-input"`, "must contain search input"},
		{`data-status="passed"`, "must have passed test row with data-status"},
		{`data-status="failed"`, "must have failed test row with data-status"},
		{`data-name="Login.TestHappy"`, "must have data-name on test rows"},
		{"step-list", "must contain step-list"},
		{"Navigate", "must contain step title"},
		{"error-msg", "must contain error-msg for failed test"},
		{"timeout", "must contain error text"},
		{"log-section", "must contain log-section"},
		{"page loaded", "must contain log text"},
		{"worker-tag", "must contain worker tags when workers > 1"},
		{"applyFilters", "must embed JavaScript with applyFilters"},
		{"test-detail", "must embed CSS with test-detail rules"},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.substring) {
			t.Error(c.message)
		}
	}
}

func TestGenerateHTML_NoWorkerTags_SingleWorker(t *testing.T) {
	var report *TestReport = &TestReport{
		ProjectName: "solo",
		Timestamp:   time.Now(),
		Workers:     1,
		Suites: []TestSuiteResult{
			{
				Name:  "tests",
				Tests: []TestCaseResult{{Name: "T1", Status: StatusPassed, Duration: time.Second, WorkerID: 0}},
			},
		},
	}
	report.ComputeStats()

	var html string = buildHTML(report)
	// Check for actual worker tag element, not CSS class definition
	if strings.Contains(html, `class="worker-tag">W`) {
		t.Error("should not render worker tag elements when workers=1")
	}
}

// --- Collector preserves logs ---

func TestCollector_LogsPreservedInReport(t *testing.T) {
	var c *Collector = NewCollector(1)
	c.Add(TestCaseResult{
		Name:     "TestWithLogs",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		Filename: "auth",
		Logs:     []string{"line 1", "line 2", "line 3"},
	})

	var report *TestReport = c.BuildReport("proj")
	var tc TestCaseResult = report.Suites[0].Tests[0]

	if len(tc.Logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(tc.Logs))
	}
	if tc.Logs[0] != "line 1" {
		t.Errorf("expected first log 'line 1', got '%s'", tc.Logs[0])
	}
}

func TestCollector_StepsAndLogsPreservedTogether(t *testing.T) {
	var c *Collector = NewCollector(2)
	c.Add(TestCaseResult{
		Name:     "TestCombo",
		Status:   StatusFailed,
		Duration: 2 * time.Second,
		Filename: "cart",
		Steps:    []TestStep{{Title: "Click", Status: "failed", Duration: time.Millisecond, Error: "not found"}},
		ErrorMsg: "assertion failed",
		Logs:     []string{"debug: clicking button"},
	})

	var report *TestReport = c.BuildReport("proj")
	var tc TestCaseResult = report.Suites[0].Tests[0]

	if len(tc.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tc.Steps))
	}
	if tc.ErrorMsg != "assertion failed" {
		t.Errorf("expected error preserved, got '%s'", tc.ErrorMsg)
	}
	if len(tc.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(tc.Logs))
	}
}

// --- Collector groups by filename ---

func TestCollector_GroupsByFilename(t *testing.T) {
	var c *Collector = NewCollector(1)
	c.Add(TestCaseResult{Name: "T1", Status: StatusPassed, Filename: "auth"})
	c.Add(TestCaseResult{Name: "T2", Status: StatusPassed, Filename: "cart"})
	c.Add(TestCaseResult{Name: "T3", Status: StatusFailed, Filename: "auth", ErrorMsg: "err"})

	var report *TestReport = c.BuildReport("proj")

	if len(report.Suites) != 2 {
		t.Fatalf("expected 2 suites, got %d", len(report.Suites))
	}
	if report.Suites[0].Name != "auth" {
		t.Errorf("expected first suite 'auth', got '%s'", report.Suites[0].Name)
	}
	if len(report.Suites[0].Tests) != 2 {
		t.Errorf("expected 2 tests in auth suite, got %d", len(report.Suites[0].Tests))
	}
	if len(report.Suites[1].Tests) != 1 {
		t.Errorf("expected 1 test in cart suite, got %d", len(report.Suites[1].Tests))
	}
}

func TestCollector_EmptyFilenameUsesDefault(t *testing.T) {
	var c *Collector = NewCollector(1)
	c.Add(TestCaseResult{Name: "T1", Status: StatusPassed, Filename: ""})

	var report *TestReport = c.BuildReport("proj")
	if report.Suites[0].Name != "default" {
		t.Errorf("expected default suite name, got '%s'", report.Suites[0].Name)
	}
}

// --- topSlowTests ---
