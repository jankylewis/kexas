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
