
package ktest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas/ktest/report"
)

// TestGenerate_InsightsSection verifies donut and duration widgets render.
func TestGenerate_InsightsSection(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = insightsFixtureReport()
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	var mustContain []string = []string{
		"Outcome mix",
		"Slowest tests",
		"class=\"donut\"",
		"class=\"duration-list\"",
		"legend-dot pass",
		"legend-dot fail",
		"legend-dot skip",
	}
	for _, token := range mustContain {
		if !strings.Contains(html, token) {
			t.Fatalf("report missing insights token %q", token)
		}
	}

	var slowIdx int = strings.Index(html, "suite.TestSlow")
	var mediumIdx int = strings.Index(html, "suite.TestMedium")
	if slowIdx == -1 || mediumIdx == -1 {
		t.Fatalf("slow or medium test missing in duration list")
	}
	if slowIdx > mediumIdx {
		t.Fatalf("expected slowest test to appear before medium, got indices %d > %d", slowIdx, mediumIdx)
	}
}

// insightsFixtureReport builds the slow/medium/skipped 3-test fixture used by
// TestGenerate_InsightsSection to exercise donut + duration-list widgets.
func insightsFixtureReport() *report.TestReport {
	return &report.TestReport{
		ProjectName: "Insights",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "suite.TestSlow", Status: report.StatusFailed, Duration: 4 * time.Second},
					{Name: "suite.TestMedium", Status: report.StatusPassed, Duration: 2 * time.Second},
					{Name: "suite.TestSkipped", Status: report.StatusSkipped, Duration: 0},
				},
			},
		},
	}
}

// TestInsights_AllPassDonut verifies donut shows 100% pass when all tests pass.
func TestInsights_AllPassDonut(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "AllPass",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "TestA", Status: report.StatusPassed, Duration: 1 * time.Second},
					{Name: "TestB", Status: report.StatusPassed, Duration: 2 * time.Second},
				},
			},
		},
	}
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	// Donut center should show 100%
	if !strings.Contains(html, "100%") {
		t.Error("expected 100% pass in donut center")
	}
	// Legend should show 2 passed, 0 failed, 0 skipped
	if !strings.Contains(html, "<strong>2</strong>") {
		t.Error("expected passed count of 2 in legend")
	}
}

// TestInsights_AllFailDonut verifies donut shows 0% pass when all tests fail.
func TestInsights_AllFailDonut(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "AllFail",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "TestX", Status: report.StatusFailed, Duration: 1 * time.Second},
					{Name: "TestY", Status: report.StatusFailed, Duration: 500 * time.Millisecond},
				},
			},
		},
	}
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	// Donut center should show 0%
	if !strings.Contains(html, "0%") {
		t.Error("expected 0% pass in donut center when all tests fail")
	}
	// Legend should show 0 passed, 2 failed
	if !strings.Contains(html, "<strong>0</strong>") {
		t.Error("expected passed count of 0 in legend")
	}
}

// TestInsights_NoTests_EmptyState verifies duration card shows empty message.
func TestInsights_NoTests_EmptyState(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "Empty",
		Timestamp:   time.Now(),
		Suites:      []report.TestSuiteResult{},
	}
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	if !strings.Contains(html, "No tests executed") {
		t.Error("expected 'No tests executed' message in duration card when no tests exist")
	}
}

// TestInsights_SingleTest_DurationBar verifies bar width is 100% for a single test.
func TestInsights_SingleTest_DurationBar(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "Single",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "TestOnly", Status: report.StatusPassed, Duration: 3 * time.Second},
				},
			},
		},
	}
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	// Single test should have width 100%
	if !strings.Contains(html, "width:100.00%") {
		t.Error("expected single test duration bar to be 100% width")
	}
	if !strings.Contains(html, "TestOnly") {
		t.Error("expected test name 'TestOnly' in duration list")
	}
}

// TestInsights_ZeroDuration_NoPanic verifies no division-by-zero when all durations are 0.
func TestInsights_ZeroDuration_NoPanic(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "ZeroDur",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "TestZeroA", Status: report.StatusPassed, Duration: 0},
					{Name: "TestZeroB", Status: report.StatusPassed, Duration: 0},
				},
			},
		},
	}
	r.ComputeStats()

	// Should not panic on zero durations
	var err error = report.Generate(r, outputPath)
	if err != nil {
		t.Fatalf("Generate failed with zero durations: %v", err)
	}

	var data []byte
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	// Duration list should still render (with minimum bar widths)
	if !strings.Contains(html, "duration-list") {
		t.Error("expected duration-list to render even with zero durations")
	}
}

// TestInsights_DurationSortOrder verifies slowest tests appear first in the list.
func TestInsights_DurationSortOrder(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "SortOrder",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "TestFast", Status: report.StatusPassed, Duration: 100 * time.Millisecond},
					{Name: "TestSlowest", Status: report.StatusFailed, Duration: 10 * time.Second},
					{Name: "TestMid", Status: report.StatusPassed, Duration: 1 * time.Second},
				},
			},
		},
	}
	r.ComputeStats()

	if err := report.Generate(r, outputPath); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var html string = string(data)

	var slowIdx int = strings.Index(html, "TestSlowest")
	var midIdx int = strings.Index(html, "TestMid")
	var fastIdx int = strings.Index(html, "TestFast")

	if slowIdx == -1 || midIdx == -1 || fastIdx == -1 {
		t.Fatal("one or more test names missing from duration list")
	}
	if slowIdx > midIdx {
		t.Errorf("TestSlowest should appear before TestMid (indices: %d > %d)", slowIdx, midIdx)
	}
	if midIdx > fastIdx {
		t.Errorf("TestMid should appear before TestFast (indices: %d > %d)", midIdx, fastIdx)
	}
}
