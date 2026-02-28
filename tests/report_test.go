
package tests

import (
	"sync"
	"testing"

	"github.com/kexas-project/kexas/ktest/report"
)

// ========================================
// REPORT MODEL TESTS
// ========================================

// TestComputeStats_AllPassed verifies stats with 100% pass rate.
func TestComputeStats_AllPassed(t *testing.T) {
	var r *report.TestReport = &report.TestReport{
		Suites: []report.TestSuiteResult{
			{
				Name: "auth_tests",
				Tests: []report.TestCaseResult{
					{Name: "TestLogin", Status: report.StatusPassed},
					{Name: "TestLogout", Status: report.StatusPassed},
				},
			},
		},
	}
	r.ComputeStats()

	if r.TotalTests != 2 {
		t.Errorf("expected TotalTests=2, got %d", r.TotalTests)
	}
	if r.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", r.Passed)
	}
	if r.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", r.Failed)
	}
	if r.PassRate != 100.0 {
		t.Errorf("expected PassRate=100.0, got %.1f", r.PassRate)
	}
}

// TestComputeStats_MixedResults verifies stats with mixed pass/fail/skip.
func TestComputeStats_MixedResults(t *testing.T) {
	var r *report.TestReport = &report.TestReport{
		Suites: []report.TestSuiteResult{
			{
				Name: "suite1",
				Tests: []report.TestCaseResult{
					{Name: "T1", Status: report.StatusPassed},
					{Name: "T2", Status: report.StatusFailed},
					{Name: "T3", Status: report.StatusSkipped},
					{Name: "T4", Status: report.StatusPassed},
				},
			},
		},
	}
	r.ComputeStats()

	if r.TotalTests != 4 {
		t.Errorf("expected TotalTests=4, got %d", r.TotalTests)
	}
	if r.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", r.Passed)
	}
	if r.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", r.Failed)
	}
	if r.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d", r.Skipped)
	}
	if r.PassRate != 50.0 {
		t.Errorf("expected PassRate=50.0, got %.1f", r.PassRate)
	}
}

// TestComputeStats_Empty verifies stats with zero tests.
func TestComputeStats_Empty(t *testing.T) {
	var r *report.TestReport = &report.TestReport{}
	r.ComputeStats()

	if r.TotalTests != 0 {
		t.Errorf("expected TotalTests=0, got %d", r.TotalTests)
	}
	if r.PassRate != 0 {
		t.Errorf("expected PassRate=0, got %.1f", r.PassRate)
	}
}

// TestComputeStats_MultipleSuites verifies stats across multiple suites.
func TestComputeStats_MultipleSuites(t *testing.T) {
	var r *report.TestReport = &report.TestReport{
		Suites: []report.TestSuiteResult{
			{
				Name: "auth",
				Tests: []report.TestCaseResult{
					{Name: "T1", Status: report.StatusPassed},
				},
			},
			{
				Name: "cart",
				Tests: []report.TestCaseResult{
					{Name: "T2", Status: report.StatusFailed},
					{Name: "T3", Status: report.StatusPassed},
				},
			},
		},
	}
	r.ComputeStats()

	if r.TotalTests != 3 {
		t.Errorf("expected TotalTests=3, got %d", r.TotalTests)
	}
	if r.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", r.Passed)
	}
	if r.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", r.Failed)
	}
}

// ========================================
// COLLECTOR TESTS
// ========================================

// TestCollector_AddAndResults verifies basic add/retrieve flow.
func TestCollector_AddAndResults(t *testing.T) {
	var c *report.Collector = report.NewCollector(1)
	c.Add(report.TestCaseResult{Name: "TestA", Status: report.StatusPassed})
	c.Add(report.TestCaseResult{Name: "TestB", Status: report.StatusFailed})

	var results []report.TestCaseResult = c.Results()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "TestA" {
		t.Errorf("expected first result TestA, got %s", results[0].Name)
	}
	if results[1].Status != report.StatusFailed {
		t.Errorf("expected second result failed, got %s", results[1].Status)
	}
}

// TestCollector_ResultsReturnsCopy verifies that Results() returns a copy.
func TestCollector_ResultsReturnsCopy(t *testing.T) {
	var c *report.Collector = report.NewCollector(1)
	c.Add(report.TestCaseResult{Name: "TestA", Status: report.StatusPassed})

	var results1 []report.TestCaseResult = c.Results()
	results1[0].Name = "MUTATED"

	var results2 []report.TestCaseResult = c.Results()
	if results2[0].Name != "TestA" {
		t.Errorf("Results() should return a copy, but mutation leaked: got %s", results2[0].Name)
	}
}

// TestCollector_ThreadSafety verifies concurrent Add calls don't race.
func TestCollector_ThreadSafety(t *testing.T) {
	var c *report.Collector = report.NewCollector(4)
	var wg sync.WaitGroup
	var count int = 100

	for i := 0; i < count; i++ {
		wg.Add(1)
		var idx int = i
		go func() {
			defer wg.Done()
			c.Add(report.TestCaseResult{
				Name:     "Test" + string(rune('A'+idx%26)),
				Status:   report.StatusPassed,
				WorkerID: idx % 4,
			})
		}()
	}
	wg.Wait()

	var results []report.TestCaseResult = c.Results()
	if len(results) != count {
		t.Errorf("expected %d results, got %d", count, len(results))
	}
}

// TestCollector_BuildReport_GroupsByFilename verifies suite grouping.
func TestCollector_BuildReport_GroupsByFilename(t *testing.T) {
	var c *report.Collector = report.NewCollector(1)
	c.Add(report.TestCaseResult{Name: "T1", Filename: "auth_tests", Status: report.StatusPassed})
	c.Add(report.TestCaseResult{Name: "T2", Filename: "auth_tests", Status: report.StatusFailed})
	c.Add(report.TestCaseResult{Name: "T3", Filename: "cart_tests", Status: report.StatusPassed})

	var r *report.TestReport = c.BuildReport("myproject")
	if r.ProjectName != "myproject" {
		t.Errorf("expected ProjectName=myproject, got %s", r.ProjectName)
	}
	if len(r.Suites) != 2 {
		t.Fatalf("expected 2 suites, got %d", len(r.Suites))
	}
	if r.Suites[0].Name != "auth_tests" {
		t.Errorf("expected first suite auth_tests, got %s", r.Suites[0].Name)
	}
	if len(r.Suites[0].Tests) != 2 {
		t.Errorf("expected 2 tests in auth_tests, got %d", len(r.Suites[0].Tests))
	}
	if r.Suites[1].Name != "cart_tests" {
		t.Errorf("expected second suite cart_tests, got %s", r.Suites[1].Name)
	}
	if r.TotalTests != 3 {
		t.Errorf("expected TotalTests=3, got %d", r.TotalTests)
	}
}

// TestCollector_BuildReport_EmptyFilename verifies default suite name.
func TestCollector_BuildReport_EmptyFilename(t *testing.T) {
	var c *report.Collector = report.NewCollector(1)
	c.Add(report.TestCaseResult{Name: "T1", Filename: "", Status: report.StatusPassed})

	var r *report.TestReport = c.BuildReport("proj")
	if len(r.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(r.Suites))
	}
	if r.Suites[0].Name != "default" {
		t.Errorf("expected suite name 'default', got %s", r.Suites[0].Name)
	}
}

// TestCollector_BuildReport_Workers verifies worker count is recorded.
func TestCollector_BuildReport_Workers(t *testing.T) {
	var c *report.Collector = report.NewCollector(4)
	c.Add(report.TestCaseResult{Name: "T1", Status: report.StatusPassed})

	var r *report.TestReport = c.BuildReport("proj")
	if r.Workers != 4 {
		t.Errorf("expected Workers=4, got %d", r.Workers)
	}
}

// HTML generation, utility, and escaping tests are in report_html_test.go
