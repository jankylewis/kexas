package report

import (
	"strings"
	"testing"
	"time"
)

// ========================================
// CRITICAL UNIT TESTS — Test Steps & Error Messages in Reports
// ========================================

func TestWriteTestRow_WithSteps_HasDetailClass(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestLogin",
		Status:   StatusPassed,
		Duration: 2 * time.Second,
		Filename: "signin_tests",
		Steps: []TestStep{
			{Title: "Navigate to login", Status: "passed", Duration: 500 * time.Millisecond},
			{Title: "Fill credentials", Status: "passed", Duration: 300 * time.Millisecond},
		},
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, "has-detail") {
		t.Error("expected has-detail class when steps are present")
	}
	if !strings.Contains(html, "detail-toggle") {
		t.Error("expected detail-toggle arrow when steps are present")
	}
	if !strings.Contains(html, "test-detail") {
		t.Error("expected test-detail panel when steps are present")
	}
}

func TestWriteTestRow_WithoutSteps_AlwaysExpandable(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestSimple",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		Filename: "basic_tests",
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, "has-detail") {
		t.Error("all test rows should have has-detail class for expandability")
	}
	if !strings.Contains(html, "test-detail") {
		t.Error("all test rows should have test-detail panel")
	}
}

func TestWriteTestRow_FailedWithError_ShowsErrorMsg(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestBroken",
		Status:   StatusFailed,
		Duration: 10200 * time.Millisecond,
		Filename: "cart_tests",
		ErrorMsg: "worker[0] failed to launch browser: could not find debugger URL",
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 3)
	var html string = b.String()

	if !strings.Contains(html, "has-detail") {
		t.Error("expected has-detail class when error message present")
	}
	if !strings.Contains(html, "error-msg") {
		t.Error("expected error-msg div for failed test")
	}
	if !strings.Contains(html, "could not find debugger URL") {
		t.Error("expected error message text in output")
	}
}

func TestWriteTestSteps_PassedSteps(t *testing.T) {
	var steps []TestStep = []TestStep{
		{Title: "Navigate to cart", Status: "passed", Duration: 200 * time.Millisecond},
		{Title: "Click checkout", Status: "passed", Duration: 150 * time.Millisecond},
	}
	var b strings.Builder
	writeTestSteps(&b, steps)
	var html string = b.String()

	if !strings.Contains(html, "step-list") {
		t.Error("expected step-list wrapper")
	}
	if !strings.Contains(html, "Navigate to cart") {
		t.Error("expected step title")
	}
	if !strings.Contains(html, "Click checkout") {
		t.Error("expected second step title")
	}
	if strings.Count(html, "&#10003;") != 2 {
		t.Errorf("expected 2 checkmarks, got %d", strings.Count(html, "&#10003;"))
	}
}

func TestWriteTestSteps_FailedStep_ShowsCrossAndError(t *testing.T) {
	var steps []TestStep = []TestStep{
		{Title: "Navigate", Status: "passed", Duration: 100 * time.Millisecond},
		{Title: "Click button", Status: "failed", Duration: 50 * time.Millisecond, Error: "element not found"},
	}
	var b strings.Builder
	writeTestSteps(&b, steps)
	var html string = b.String()

	if strings.Count(html, "&#10007;") != 1 {
		t.Error("expected exactly 1 cross mark for failed step")
	}
	if !strings.Contains(html, "step-error") {
		t.Error("expected step-error div for failed step")
	}
	if !strings.Contains(html, "element not found") {
		t.Error("expected error text in step-error")
	}
}

func TestWriteTestSteps_EmptySteps_NoOutput(t *testing.T) {
	var b strings.Builder
	writeTestSteps(&b, nil)
	if b.Len() != 0 {
		t.Errorf("expected no output for nil steps, got %d bytes", b.Len())
	}

	b.Reset()
	writeTestSteps(&b, []TestStep{})
	if b.Len() != 0 {
		t.Errorf("expected no output for empty steps, got %d bytes", b.Len())
	}
}

func TestWriteTestError_FailedWithMsg(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Status:   StatusFailed,
		ErrorMsg: "timeout after 10s",
	}
	var b strings.Builder
	writeTestError(&b, &tc)
	var html string = b.String()

	if !strings.Contains(html, "error-msg") {
		t.Error("expected error-msg div")
	}
	if !strings.Contains(html, "timeout after 10s") {
		t.Error("expected error text")
	}
}

func TestWriteTestError_PassedTest_NoOutput(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Status:   StatusPassed,
		ErrorMsg: "should not appear",
	}
	var b strings.Builder
	writeTestError(&b, &tc)
	if b.Len() != 0 {
		t.Error("should not render error for passed test")
	}
}

func TestWriteTestError_FailedNoMsg_NoOutput(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Status:   StatusFailed,
		ErrorMsg: "",
	}
	var b strings.Builder
	writeTestError(&b, &tc)
	if b.Len() != 0 {
		t.Error("should not render error when ErrorMsg is empty")
	}
}

func TestBuildReport_StepsPreserved(t *testing.T) {
	var c *Collector = NewCollector(1)
	c.Add(TestCaseResult{
		Name:     "TestWithSteps",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		Filename: "auth",
		Steps: []TestStep{
			{Title: "Step A", Status: "passed", Duration: 100 * time.Millisecond},
			{Title: "Step B", Status: "passed", Duration: 200 * time.Millisecond},
		},
	})

	var report *TestReport = c.BuildReport("test-project")

	if len(report.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(report.Suites))
	}
	if len(report.Suites[0].Tests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(report.Suites[0].Tests))
	}
	var tc TestCaseResult = report.Suites[0].Tests[0]
	if len(tc.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(tc.Steps))
	}
	if tc.Steps[0].Title != "Step A" {
		t.Errorf("expected step title 'Step A', got '%s'", tc.Steps[0].Title)
	}
}

func TestBuildReport_ErrorMsgPreserved(t *testing.T) {
	var c *Collector = NewCollector(3)
	c.Add(TestCaseResult{
		Name:     "TestBroken",
		Status:   StatusFailed,
		Duration: 10200 * time.Millisecond,
		Filename: "cart",
		ErrorMsg: "launch failed: debugger URL timeout",
	})

	var report *TestReport = c.BuildReport("test-project")
	var tc TestCaseResult = report.Suites[0].Tests[0]

	if tc.ErrorMsg != "launch failed: debugger URL timeout" {
		t.Errorf("expected error msg preserved, got '%s'", tc.ErrorMsg)
	}
}

func TestGenerateHTML_StepsRendered(t *testing.T) {
	var report *TestReport = &TestReport{
		ProjectName: "test",
		Timestamp:   time.Now(),
		Workers:     1,
		Suites: []TestSuiteResult{
			{
				Name: "auth",
				Tests: []TestCaseResult{
					{
						Name:     "TestLogin",
						Status:   StatusPassed,
						Duration: 2 * time.Second,
						Steps: []TestStep{
							{Title: "Open page", Status: "passed", Duration: 100 * time.Millisecond},
						},
					},
				},
			},
		},
	}
	report.ComputeStats()

	var html string = buildHTML(report)
	if !strings.Contains(html, "step-list") {
		t.Error("expected step-list in generated HTML")
	}
	if !strings.Contains(html, "Open page") {
		t.Error("expected step title in generated HTML")
	}
}

func TestWriteTestRow_WorkerTag_OnlyWhenMultipleWorkers(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestX",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		WorkerID: 2,
	}

	// Single worker: no tag
	var b1 strings.Builder
	writeTestRow(&b1, &tc, 1)
	if strings.Contains(b1.String(), "worker-tag") {
		t.Error("should not show worker tag with 1 worker")
	}

	// Multiple workers: show tag
	var b2 strings.Builder
	writeTestRow(&b2, &tc, 3)
	if !strings.Contains(b2.String(), "worker-tag") {
		t.Error("should show worker tag with 3 workers")
	}
	if !strings.Contains(b2.String(), "W2") {
		t.Error("should show correct worker ID")
	}
}

func TestFormatDuration_VariousRanges(t *testing.T) {
	var cases []struct {
		d    time.Duration
		want string
	} = []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Microsecond, "500µs"},
		{250 * time.Millisecond, "250ms"},
		{2500 * time.Millisecond, "2.5s"},
		{10200 * time.Millisecond, "10.2s"},
		{90 * time.Second, "1m 30s"},
	}

	for _, tc := range cases {
		var got string = formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}
