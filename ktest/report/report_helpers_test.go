package report

import (
	"strings"
	"testing"
	"time"
)


func TestTopSlowTests_ReturnsTopN(t *testing.T) {
	var report *TestReport = &TestReport{
		Suites: []TestSuiteResult{
			{
				Tests: []TestCaseResult{
					{Name: "Fast", Duration: 100 * time.Millisecond},
					{Name: "Slow", Duration: 10 * time.Second},
					{Name: "Medium", Duration: 2 * time.Second},
					{Name: "Slower", Duration: 8 * time.Second},
				},
			},
		},
	}

	var top []TestCaseResult = topSlowTests(report, 2)
	if len(top) != 2 {
		t.Fatalf("expected 2 results, got %d", len(top))
	}
	if top[0].Name != "Slow" {
		t.Errorf("expected slowest first, got '%s'", top[0].Name)
	}
	if top[1].Name != "Slower" {
		t.Errorf("expected second slowest, got '%s'", top[1].Name)
	}
}

func TestTopSlowTests_EmptyReport(t *testing.T) {
	var report *TestReport = &TestReport{}
	var top []TestCaseResult = topSlowTests(report, 5)
	if top != nil {
		t.Errorf("expected nil for empty report, got %v", top)
	}
}

// --- buildDonutStyle ---

func TestBuildDonutStyle_Format(t *testing.T) {
	var style string = buildDonutStyle(80, 15, 5)
	if !strings.Contains(style, "conic-gradient") {
		t.Error("expected conic-gradient in donut style")
	}
	if !strings.Contains(style, "var(--pass)") {
		t.Error("expected --pass CSS variable")
	}
	if !strings.Contains(style, "var(--fail)") {
		t.Error("expected --fail CSS variable")
	}
}

func TestBuildDonutStyle_ZeroTotal(t *testing.T) {
	var style string = buildDonutStyle(0, 0, 0)
	if !strings.Contains(style, "conic-gradient") {
		t.Error("should still produce conic-gradient for zero values")
	}
}

// --- ProjectNameFromDir ---

func TestProjectNameFromDir_Normal(t *testing.T) {
	var name string = ProjectNameFromDir("/Users/dev/projects/myapp")
	if name != "myapp" {
		t.Errorf("expected 'myapp', got '%s'", name)
	}
}

func TestProjectNameFromDir_RootPath(t *testing.T) {
	var name string = ProjectNameFromDir("/")
	if name != "kexas" {
		t.Errorf("expected 'kexas' for root path, got '%s'", name)
	}
}

func TestProjectNameFromDir_DotPath(t *testing.T) {
	var name string = ProjectNameFromDir(".")
	if name != "kexas" {
		t.Errorf("expected 'kexas' for dot path, got '%s'", name)
	}
}
