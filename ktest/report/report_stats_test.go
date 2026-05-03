package report

import (
	"strings"
	"testing"
	"time"
)

// --- Suite rendering ---

func TestWriteSingleSuite_PassedCountAccurate(t *testing.T) {
	var suite TestSuiteResult = TestSuiteResult{
		Name: "auth_tests",
		Tests: []TestCaseResult{
			{Name: "T1", Status: StatusPassed, Duration: time.Second},
			{Name: "T2", Status: StatusFailed, Duration: time.Second, ErrorMsg: "err"},
			{Name: "T3", Status: StatusPassed, Duration: time.Second},
		},
	}
	var b strings.Builder
	writeSingleSuite(&b, &suite, 1)
	var html string = b.String()

	if !strings.Contains(html, "2 / 3 passed") {
		t.Error("suite header must show correct passed/total count")
	}
}

func TestWriteSingleSuite_ContainsSuiteStructure(t *testing.T) {
	var suite TestSuiteResult = TestSuiteResult{
		Name:  "cart_tests",
		Tests: []TestCaseResult{{Name: "T1", Status: StatusPassed, Duration: time.Second}},
	}
	var b strings.Builder
	writeSingleSuite(&b, &suite, 1)
	var html string = b.String()

	if !strings.Contains(html, `class="suite"`) {
		t.Error("expected suite wrapper div")
	}
	if !strings.Contains(html, `class="suite-header"`) {
		t.Error("expected suite-header div for JS collapse binding")
	}
	if !strings.Contains(html, `class="suite-body"`) {
		t.Error("expected suite-body div for collapse target")
	}
	if !strings.Contains(html, "cart_tests") {
		t.Error("expected suite name in output")
	}
}

// --- ComputeStats ---

func TestComputeStats_MixedStatuses(t *testing.T) {
	var r TestReport = TestReport{
		Suites: []TestSuiteResult{
			{
				Name: "s1",
				Tests: []TestCaseResult{
					{Status: StatusPassed},
					{Status: StatusPassed},
					{Status: StatusFailed},
					{Status: StatusSkipped},
				},
			},
			{
				Name: "s2",
				Tests: []TestCaseResult{
					{Status: StatusPassed},
					{Status: StatusFailed},
				},
			},
		},
	}
	r.ComputeStats()

	if r.TotalTests != 6 {
		t.Errorf("expected 6 total, got %d", r.TotalTests)
	}
	if r.Passed != 3 {
		t.Errorf("expected 3 passed, got %d", r.Passed)
	}
	if r.Failed != 2 {
		t.Errorf("expected 2 failed, got %d", r.Failed)
	}
	if r.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", r.Skipped)
	}
	if r.PassRate != 50.0 {
		t.Errorf("expected 50%% pass rate, got %.1f%%", r.PassRate)
	}
}

func TestComputeStats_Empty(t *testing.T) {
	var r TestReport = TestReport{}
	r.ComputeStats()

	if r.TotalTests != 0 {
		t.Errorf("expected 0 total, got %d", r.TotalTests)
	}
	if r.PassRate != 0 {
		t.Errorf("expected 0 pass rate, got %.1f", r.PassRate)
	}
}

func TestComputeStats_AllPassed(t *testing.T) {
	var r TestReport = TestReport{
		Suites: []TestSuiteResult{
			{Tests: []TestCaseResult{{Status: StatusPassed}, {Status: StatusPassed}}},
		},
	}
	r.ComputeStats()

	if r.PassRate != 100.0 {
		t.Errorf("expected 100%% pass rate, got %.1f%%", r.PassRate)
	}
	if r.Failed != 0 || r.Skipped != 0 {
		t.Error("no failed or skipped expected")
	}
}

// --- formatTestName ---

func TestFormatTestName_WithGroupPrefix(t *testing.T) {
	var result string = formatTestName("Inventory.TestSortDropdown")
	if !strings.Contains(result, "group-prefix") {
		t.Error("expected group-prefix span for dotted name")
	}
	if !strings.Contains(result, "Inventory.") {
		t.Error("expected 'Inventory.' as prefix")
	}
	if !strings.Contains(result, "TestSortDropdown") {
		t.Error("expected 'TestSortDropdown' as test name")
	}
}

func TestFormatTestName_NoGroup(t *testing.T) {
	var result string = formatTestName("TestSimple")
	if strings.Contains(result, "group-prefix") {
		t.Error("should not have group-prefix for name without dot")
	}
	if !strings.Contains(result, "TestSimple") {
		t.Error("expected plain test name")
	}
}

func TestFormatTestName_HTMLEscaping(t *testing.T) {
	var result string = formatTestName("Suite.<Test>")
	if strings.Contains(result, "<Test>") {
		t.Error("test name must be HTML-escaped")
	}
	if !strings.Contains(result, "&lt;Test&gt;") {
		t.Error("expected escaped test name")
	}
}
