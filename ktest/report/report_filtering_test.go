package report

import (
	"strings"
	"testing"
	"time"
)

// ========================================
// CRITICAL UNIT TESTS — HTML Structure for Filtering & Detail Toggling
// These tests guard against the tab-filtering bug where clicking
// test rows on Passed/Failed/Skipped tabs failed to expand details.
// Root cause: applyFilters() set inline display:none on .test-detail,
// which CSS could not override. The fix ensures inline styles are
// cleared for visible rows so CSS + click handlers can work.
// ========================================

// --- Test Row data-status / data-name attributes (filter prerequisites) ---

func TestWriteTestRow_DataStatusAttribute_Passed(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "Suite.TestOK",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, `data-status="passed"`) {
		t.Error("passed test row must have data-status=\"passed\" for JS filtering")
	}
}

func TestWriteTestRow_DataStatusAttribute_Failed(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "Suite.TestBad",
		Status:   StatusFailed,
		Duration: 1 * time.Second,
		ErrorMsg: "boom",
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, `data-status="failed"`) {
		t.Error("failed test row must have data-status=\"failed\" for JS filtering")
	}
}

func TestWriteTestRow_DataStatusAttribute_Skipped(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "Suite.TestSkip",
		Status:   StatusSkipped,
		Duration: 0,
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, `data-status="skipped"`) {
		t.Error("skipped test row must have data-status=\"skipped\" for JS filtering")
	}
}

func TestWriteTestRow_DataNameAttribute(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "Auth.TestLogin",
		Status:   StatusPassed,
		Duration: 500 * time.Millisecond,
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, `data-name="Auth.TestLogin"`) {
		t.Error("test row must have data-name attribute for JS search filtering")
	}
}

func TestWriteTestRow_DataNameAttribute_HTMLEscaping(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "Test<XSS>&Inject",
		Status:   StatusPassed,
		Duration: 100 * time.Millisecond,
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if strings.Contains(html, `data-name="Test<XSS>&Inject"`) {
		t.Error("data-name must HTML-escape special characters")
	}
	if !strings.Contains(html, `data-name="Test&lt;XSS&gt;&amp;Inject"`) {
		t.Error("data-name must contain escaped name")
	}
}

// --- Detail panel adjacency (sibling structure for JS nextElementSibling) ---

func TestWriteTestRow_DetailPanelIsImmediatelyAfterRow(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestAdjacentDetail",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
		Steps: []TestStep{
			{Title: "Step A", Status: "passed", Duration: 100 * time.Millisecond},
		},
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	// The JS uses row.nextElementSibling to find the detail panel.
	// The detail div MUST immediately follow the closing </div> of the test-row.
	var rowClose int = strings.Index(html, "</div>\n<div class=\"test-detail\">")
	if rowClose < 0 {
		t.Error("test-detail div must immediately follow the test-row closing tag (required for JS nextElementSibling)")
	}
}

func TestWriteTestRow_DetailPanelPresent_EvenWithoutSteps(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestEmpty",
		Status:   StatusPassed,
		Duration: 1 * time.Second,
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, "test-detail") {
		t.Error("every test row must have a test-detail panel for consistency")
	}
}

// --- has-detail class (click handler binding) ---

func TestWriteTestRow_HasDetailClass_Always(t *testing.T) {
	var cases []struct {
		name string
		tc   TestCaseResult
	} = []struct {
		name string
		tc   TestCaseResult
	}{
		{"passed no steps", TestCaseResult{Name: "T1", Status: StatusPassed, Duration: time.Second}},
		{"failed with error", TestCaseResult{Name: "T2", Status: StatusFailed, Duration: time.Second, ErrorMsg: "err"}},
		{"with steps", TestCaseResult{Name: "T3", Status: StatusPassed, Duration: time.Second, Steps: []TestStep{{Title: "s", Status: "passed", Duration: time.Millisecond}}}},
		{"with logs", TestCaseResult{Name: "T4", Status: StatusPassed, Duration: time.Second, Logs: []string{"log"}}},
		{"skipped", TestCaseResult{Name: "T5", Status: StatusSkipped, Duration: 0}},
	}

	for _, c := range cases {
		var b strings.Builder
		writeTestRow(&b, &c.tc, 1)
		if !strings.Contains(b.String(), "has-detail") {
			t.Errorf("[%s] all test rows must have has-detail class for JS click binding", c.name)
		}
	}
}

// --- Combined steps + errors + logs rendering ---

func TestWriteTestRow_CombinedStepsErrorsLogs(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestFull",
		Status:   StatusFailed,
		Duration: 3 * time.Second,
		Steps: []TestStep{
			{Title: "Navigate", Status: "passed", Duration: 100 * time.Millisecond},
			{Title: "Click", Status: "failed", Duration: 50 * time.Millisecond, Error: "element not found"},
		},
		ErrorMsg: "test failed: element not found",
		Logs:     []string{"DEBUG: starting test", "ERROR: click failed"},
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	if !strings.Contains(html, "step-list") {
		t.Error("expected step-list in combined output")
	}
	if !strings.Contains(html, "Navigate") {
		t.Error("expected step title in combined output")
	}
	if !strings.Contains(html, "error-msg") {
		t.Error("expected error-msg in combined output")
	}
	if !strings.Contains(html, "log-section") {
		t.Error("expected log-section in combined output")
	}
	if !strings.Contains(html, "Console Output") {
		t.Error("expected Console Output header in combined output")
	}
	if !strings.Contains(html, "DEBUG: starting test") {
		t.Error("expected log line in combined output")
	}
}

func TestWriteTestRow_OrderIsStepsThenErrorThenLogs(t *testing.T) {
	var tc TestCaseResult = TestCaseResult{
		Name:     "TestOrder",
		Status:   StatusFailed,
		Duration: 1 * time.Second,
		Steps:    []TestStep{{Title: "S1", Status: "passed", Duration: time.Millisecond}},
		ErrorMsg: "assertion failed",
		Logs:     []string{"log line"},
	}
	var b strings.Builder
	writeTestRow(&b, &tc, 1)
	var html string = b.String()

	var stepIdx int = strings.Index(html, "step-list")
	var errorIdx int = strings.Index(html, "error-msg")
	var logIdx int = strings.Index(html, "log-section")

	if stepIdx < 0 || errorIdx < 0 || logIdx < 0 {
		t.Fatal("expected step-list, error-msg, and log-section all present")
	}
	if stepIdx > errorIdx {
		t.Error("steps must render before error")
	}
	if errorIdx > logIdx {
		t.Error("error must render before logs")
	}
}

// --- Toolbar filter buttons ---

func TestWriteToolbar_HasAllFilterButtons(t *testing.T) {
	var report *TestReport = &TestReport{TotalTests: 50}
	var b strings.Builder
	writeToolbar(&b, report)
	var html string = b.String()

	var expected []string = []string{
		`data-filter="all"`,
		`data-filter="passed"`,
		`data-filter="failed"`,
		`data-filter="skipped"`,
	}
	for _, e := range expected {
		if !strings.Contains(html, e) {
			t.Errorf("toolbar missing filter button: %s", e)
		}
	}
}

func TestWriteToolbar_AllIsActiveByDefault(t *testing.T) {
	var report *TestReport = &TestReport{TotalTests: 10}
	var b strings.Builder
	writeToolbar(&b, report)
	var html string = b.String()

	if !strings.Contains(html, `class="filter-btn active" data-filter="all"`) {
		t.Error("'All' filter button must be active by default")
	}
	// Other buttons must NOT be active
	if strings.Contains(html, `class="filter-btn active" data-filter="passed"`) {
		t.Error("'Passed' filter button must not be active by default")
	}
}

func TestWriteToolbar_SearchInputPresent(t *testing.T) {
	var report *TestReport = &TestReport{TotalTests: 100}
	var b strings.Builder
	writeToolbar(&b, report)
	var html string = b.String()

	if !strings.Contains(html, `id="search-input"`) {
		t.Error("toolbar must have search input with id 'search-input'")
	}
	if !strings.Contains(html, `Search 100 tests`) {
		t.Error("search placeholder must show test count")
	}
}

// --- JS script embedding ---

func TestReportJS_ContainsFilterLogic(t *testing.T) {
	if !strings.Contains(reportJS, "applyFilters") {
		t.Error("reportJS must contain applyFilters function")
	}
	if !strings.Contains(reportJS, "data-filter") {
		t.Error("reportJS must reference data-filter attribute")
	}
	if !strings.Contains(reportJS, "data-status") {
		t.Error("reportJS must reference data-status attribute for filtering")
	}
}

func TestReportJS_ClickHandlerManagesDetailDisplay(t *testing.T) {
	// Regression: the click handler must explicitly set display on detail panel
	// to avoid inline style conflicts with applyFilters().
	if !strings.Contains(reportJS, "row.classList.toggle('expanded')") {
		t.Error("reportJS click handler must toggle 'expanded' class")
	}
	if !strings.Contains(reportJS, "nextEl.style.display = 'block'") {
		t.Error("reportJS click handler must set display='block' on detail when expanding")
	}
	if !strings.Contains(reportJS, "nextEl.style.display = 'none'") {
		t.Error("reportJS click handler must set display='none' on detail when collapsing")
	}
}

func TestReportJS_FilterCollapsesHiddenExpandedRows(t *testing.T) {
	// Regression: when filtering hides a row that was expanded, it must
	// remove the 'expanded' class so the detail panel doesn't ghost-show.
	if !strings.Contains(reportJS, "row.classList.remove('expanded')") {
		t.Error("reportJS must collapse expanded rows when they become hidden by filter")
	}
}

func TestReportJS_FilterClearsInlineStyleForVisibleRows(t *testing.T) {
	// Regression: for visible non-expanded rows, the inline display style must
	// be cleared (set to '') so CSS can control visibility normally.
	// Without this, a previous display:'none' from a filter would persist.
	if !strings.Contains(reportJS, "nextEl.style.display = ''") {
		t.Error("reportJS must clear inline display style for visible non-expanded detail panels")
	}
}

func TestReportJS_FilterChecksSiblingIsTestDetail(t *testing.T) {
	// The JS must check nextElementSibling has 'test-detail' class before
	// manipulating it, to avoid modifying unrelated elements.
	if !strings.Contains(reportJS, "nextEl.classList.contains('test-detail')") {
		t.Error("reportJS must verify sibling is test-detail before changing display")
	}
}

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

// --- End-to-end HTML generation ---

func TestGenerateHTML_ContainsAllCriticalElements(t *testing.T) {
	var report *TestReport = &TestReport{
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
	report.ComputeStats()

	var html string = buildHTML(report)

	// Structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("must start with DOCTYPE")
	}
	if !strings.Contains(html, "<html") {
		t.Error("must contain html tag")
	}

	// Theme toggle
	if !strings.Contains(html, "theme-toggle") {
		t.Error("must contain theme toggle button")
	}

	// Toolbar
	if !strings.Contains(html, `data-filter="all"`) {
		t.Error("must contain All filter button")
	}
	if !strings.Contains(html, `data-filter="passed"`) {
		t.Error("must contain Passed filter button")
	}
	if !strings.Contains(html, `data-filter="failed"`) {
		t.Error("must contain Failed filter button")
	}
	if !strings.Contains(html, `data-filter="skipped"`) {
		t.Error("must contain Skipped filter button")
	}
	if !strings.Contains(html, `id="search-input"`) {
		t.Error("must contain search input")
	}

	// Test rows with data attributes
	if !strings.Contains(html, `data-status="passed"`) {
		t.Error("must have passed test row with data-status")
	}
	if !strings.Contains(html, `data-status="failed"`) {
		t.Error("must have failed test row with data-status")
	}
	if !strings.Contains(html, `data-name="Login.TestHappy"`) {
		t.Error("must have data-name on test rows")
	}

	// Steps
	if !strings.Contains(html, "step-list") {
		t.Error("must contain step-list")
	}
	if !strings.Contains(html, "Navigate") {
		t.Error("must contain step title")
	}

	// Error
	if !strings.Contains(html, "error-msg") {
		t.Error("must contain error-msg for failed test")
	}
	if !strings.Contains(html, "timeout") {
		t.Error("must contain error text")
	}

	// Logs
	if !strings.Contains(html, "log-section") {
		t.Error("must contain log-section")
	}
	if !strings.Contains(html, "page loaded") {
		t.Error("must contain log text")
	}

	// Worker tags (workers > 1)
	if !strings.Contains(html, "worker-tag") {
		t.Error("must contain worker tags when workers > 1")
	}

	// JS
	if !strings.Contains(html, "applyFilters") {
		t.Error("must embed JavaScript with applyFilters")
	}

	// CSS
	if !strings.Contains(html, "test-detail") {
		t.Error("must embed CSS with test-detail rules")
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
