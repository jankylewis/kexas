package report

import (
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const baseReportFilename string = "report.html"

// GenerateFiles writes the main report.html and a timestamped copy for archiving.
// Returns the primary file path and the versioned archive path.
func GenerateFiles(report *TestReport, dir string) (string, string, error) {
	var basePath string = filepath.Join(dir, baseReportFilename)
	var err error = Generate(report, basePath)
	if err != nil {
		return "", "", err
	}

	var versionedPath string
	versionedPath, err = createVersionedCopy(basePath, dir, report.Timestamp)
	if err != nil {
		return basePath, "", err
	}

	return basePath, versionedPath, nil
}

// Generate writes a self-contained HTML report to outputPath.
// The file includes embedded CSS and JS — no external dependencies.
func Generate(report *TestReport, outputPath string) error {
	var dir string = filepath.Dir(outputPath)
	var err error = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("report: failed to create directory: %w", err)
	}

	var content string = buildHTML(report)
	err = os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("report: failed to write file: %w", err)
	}

	return nil
}

func createVersionedCopy(basePath string, dir string, timestamp time.Time) (string, error) {
	const maxAttempts int = 1000
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var filename string = versionedFilename(timestamp, attempt)
		var candidate string = filepath.Join(dir, filename)

		var _, statErr = os.Stat(candidate)
		if statErr == nil {
			continue
		}
		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("report: failed to check versioned file: %w", statErr)
		}

		var err error = copyFile(basePath, candidate)
		if err != nil {
			return "", err
		}
		return candidate, nil
	}
	return "", fmt.Errorf("report: could not allocate versioned filename after %d attempts", maxAttempts)
}

func versionedFilename(timestamp time.Time, attempt int) string {
	var stamp string = timestamp.Format("20060102-150405")
	if attempt == 0 {
		return fmt.Sprintf("report-%s.html", stamp)
	}
	return fmt.Sprintf("report-%s-%03d.html", stamp, attempt)
}

func copyFile(src string, dst string) error {
	var in *os.File
	var err error
	in, err = os.Open(src)
	if err != nil {
		return fmt.Errorf("report: failed to open source report: %w", err)
	}
	defer in.Close()

	var out *os.File
	out, err = os.Create(dst)
	if err != nil {
		return fmt.Errorf("report: failed to create versioned report: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("report: failed to copy report: %w", err)
	}
	if err = out.Close(); err != nil {
		return fmt.Errorf("report: failed to flush versioned report: %w", err)
	}
	return nil
}

// buildHTML assembles the complete HTML document from the report data.
func buildHTML(r *TestReport) string {
	var b strings.Builder
	b.Grow(16384)

	writeDocOpen(&b)
	writeHeader(&b, r)
	writeSummaryCards(&b, r)
	writeProgressBar(&b, r)
	writeInsights(&b, r)
	writeToolbar(&b, r)
	writeSuites(&b, r)
	writeFooter(&b, r)
	writeDocClose(&b)

	return b.String()
}

// writeDocOpen writes the HTML head section with embedded CSS.
func writeDocOpen(b *strings.Builder) {
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\" data-theme=\"dark\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	b.WriteString("<title>Kexas Test Report</title>\n")
	b.WriteString("<style>")
	b.WriteString(reportCSS)
	b.WriteString("</style>\n</head>\n<body>\n<div class=\"container\">\n")
}

// writeHeader writes the report title and theme toggle button.
func writeHeader(b *strings.Builder, r *TestReport) {
	b.WriteString("<button id=\"theme-toggle\" class=\"theme-toggle\" aria-label=\"Toggle theme\"></button>\n")
	b.WriteString("<div class=\"header\">\n")
	fmt.Fprintf(b, "<h1>%s</h1>\n", html.EscapeString(r.ProjectName))
	fmt.Fprintf(b, "<div class=\"subtitle\">%s</div>\n",
		r.Timestamp.Format("January 2, 2006 at 3:04 PM"))
	b.WriteString("</div>\n")
}

// writeSummaryCards renders the 4 stat cards (total, passed, failed, skipped).
func writeSummaryCards(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"summary\">\n")
	writeStatCard(b, "total", "Tests", r.TotalTests)
	writeStatCard(b, "passed", "Passed", r.Passed)
	writeStatCard(b, "failed", "Failed", r.Failed)
	writeStatCard(b, "skipped", "Skipped", r.Skipped)
	b.WriteString("</div>\n")
}

// writeStatCard renders a single stat card.
func writeStatCard(b *strings.Builder, class string, label string, value int) {
	fmt.Fprintf(b, "<div class=\"stat-card %s\">\n", class)
	fmt.Fprintf(b, "<div class=\"stat-value\">%d</div>\n", value)
	fmt.Fprintf(b, "<div class=\"stat-label\">%s</div>\n", label)
	b.WriteString("</div>\n")
}

// writeProgressBar renders the pass/fail/skip progress bar with pass rate.
func writeProgressBar(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"progress-bar-container\">\n")
	b.WriteString("<div class=\"progress-bar-header\">\n")
	fmt.Fprintf(b, "<span class=\"pass-rate\">%.1f%% passed</span>\n", r.PassRate)
	b.WriteString("</div>\n")

	var passW float64 = safePercent(r.Passed, r.TotalTests)
	var failW float64 = safePercent(r.Failed, r.TotalTests)
	var skipW float64 = safePercent(r.Skipped, r.TotalTests)

	b.WriteString("<div class=\"progress-bar\">\n")
	fmt.Fprintf(b, "<div class=\"bar-pass\" style=\"width:%.2f%%\"></div>\n", passW)
	fmt.Fprintf(b, "<div class=\"bar-fail\" style=\"width:%.2f%%\"></div>\n", failW)
	fmt.Fprintf(b, "<div class=\"bar-skip\" style=\"width:%.2f%%\"></div>\n", skipW)
	b.WriteString("</div>\n</div>\n")
}

// writeInsights renders charts/trends similar to Allure-style dashboards.
func writeInsights(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"insights\">\n")
	writeOutcomeCard(b, r)
	writeDurationCard(b, r)
	b.WriteString("</div>\n")
}

func writeOutcomeCard(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"insight-card outcome\">\n")
	b.WriteString("<div class=\"insight-title\">Outcome mix</div>\n")
	b.WriteString("<div class=\"insight-subtitle\">Share of pass/fail/skip</div>\n")

	var passPct float64 = safePercent(r.Passed, r.TotalTests)
	var failPct float64 = safePercent(r.Failed, r.TotalTests)
	var skipPct float64 = safePercent(r.Skipped, r.TotalTests)
	var donutStyle string = buildDonutStyle(passPct, failPct, skipPct)

	b.WriteString("<div class=\"donut-wrapper\">\n")
	fmt.Fprintf(b, "<div class=\"donut\" style=\"%s\">\n", donutStyle)
	fmt.Fprintf(b, "<div class=\"donut-center\"><div class=\"donut-value\">%.0f%%</div><div class=\"donut-label\">pass</div></div>\n", passPct)
	b.WriteString("</div>\n")
	b.WriteString("<ul class=\"donut-legend\">\n")
	fmt.Fprintf(b, "<li><span class=\"legend-dot pass\"></span><span>Passed</span><strong>%d</strong><em>%.1f%%</em></li>\n", r.Passed, passPct)
	fmt.Fprintf(b, "<li><span class=\"legend-dot fail\"></span><span>Failed</span><strong>%d</strong><em>%.1f%%</em></li>\n", r.Failed, failPct)
	fmt.Fprintf(b, "<li><span class=\"legend-dot skip\"></span><span>Skipped</span><strong>%d</strong><em>%.1f%%</em></li>\n", r.Skipped, skipPct)
	b.WriteString("</ul>\n</div>\n</div>\n")
}

func writeDurationCard(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"insight-card duration\">\n")
	b.WriteString("<div class=\"insight-title\">Slowest tests</div>\n")
	b.WriteString("<div class=\"insight-subtitle\">Top offenders by runtime</div>\n")

	var top []TestCaseResult = topSlowTests(r, 5)
	if len(top) == 0 {
		b.WriteString("<p class=\"insight-empty\">No tests executed.</p>\n</div>\n")
		return
	}

	b.WriteString("<div class=\"duration-list\">\n")
	var maxDuration time.Duration = top[0].Duration
	if maxDuration < time.Millisecond {
		maxDuration = time.Millisecond
	}
	for _, tc := range top {
		var width float64 = (float64(tc.Duration) / float64(maxDuration)) * 100
		if width < 1 {
			width = 1
		}
		fmt.Fprintf(b, "<div class=\"duration-item\">\n")
		fmt.Fprintf(b, "<div class=\"duration-meta\"><span>%s</span><span>%s</span></div>\n",
			html.EscapeString(tc.Name), formatDuration(tc.Duration))
		fmt.Fprintf(b, "<div class=\"duration-bar\"><span class=\"duration-bar-fill\" style=\"width:%.2f%%\"></span></div>\n",
			width)
		b.WriteString("</div>\n")
	}
	b.WriteString("</div>\n</div>\n")
}

// writeToolbar renders the search input and filter buttons.
func writeToolbar(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"toolbar\">\n")
	fmt.Fprintf(b, "<input id=\"search-input\" class=\"search-input\" type=\"text\" placeholder=\"Search %d tests...\">\n", r.TotalTests)
	b.WriteString("<button class=\"filter-btn active\" data-filter=\"all\">All</button>\n")
	b.WriteString("<button class=\"filter-btn\" data-filter=\"passed\">Passed</button>\n")
	b.WriteString("<button class=\"filter-btn\" data-filter=\"failed\">Failed</button>\n")
	b.WriteString("<button class=\"filter-btn\" data-filter=\"skipped\">Skipped</button>\n")
	b.WriteString("</div>\n")
}

// writeSuites renders all test suites and their test rows.
func writeSuites(b *strings.Builder, r *TestReport) {
	for _, suite := range r.Suites {
		writeSingleSuite(b, &suite, r.Workers)
	}
}

// writeSingleSuite renders one suite with its header and test rows.
func writeSingleSuite(b *strings.Builder, suite *TestSuiteResult, workers int) {
	var passed int = 0
	var total int = len(suite.Tests)
	for _, tc := range suite.Tests {
		if tc.Status == StatusPassed {
			passed++
		}
	}

	b.WriteString("<div class=\"suite\">\n")
	fmt.Fprintf(b, "<div class=\"suite-header\">\n")
	b.WriteString("<span class=\"chevron\">&#9660;</span>\n")
	fmt.Fprintf(b, "<span class=\"suite-name\">%s</span>\n", html.EscapeString(suite.Name))
	fmt.Fprintf(b, "<span class=\"suite-stats\">%d / %d passed</span>\n", passed, total)
	b.WriteString("</div>\n")
	b.WriteString("<div class=\"suite-body\">\n")

	for _, tc := range suite.Tests {
		writeTestRow(b, &tc, workers)
	}

	b.WriteString("</div>\n</div>\n")
}

// writeTestRow renders a single test result row with expandable detail panel.
// Every test row is clickable — clicking expands to show steps, errors, and console output.
func writeTestRow(b *strings.Builder, tc *TestCaseResult, workers int) {
	var statusStr string = string(tc.Status)
	var displayName string = formatTestName(tc.Name)

	fmt.Fprintf(b, "<div class=\"test-row has-detail\" data-status=\"%s\" data-name=\"%s\">\n",
		statusStr, html.EscapeString(tc.Name))
	fmt.Fprintf(b, "<span class=\"status-dot %s\"></span>\n", statusStr)
	fmt.Fprintf(b, "<span class=\"test-name\">%s</span>\n", displayName)
	fmt.Fprintf(b, "<span class=\"badge %s\">%s</span>\n", statusStr, statusStr)
	fmt.Fprintf(b, "<span class=\"test-duration\">%s</span>\n", formatDuration(tc.Duration))

	if workers > 1 {
		fmt.Fprintf(b, "<span class=\"worker-tag\">W%d</span>\n", tc.WorkerID)
	}

	b.WriteString("<span class=\"detail-toggle\">&#9654;</span>\n")
	b.WriteString("</div>\n")

	// Expandable detail panel: steps → error → console output
	b.WriteString("<div class=\"test-detail\">\n")
	writeTestSteps(b, tc.Steps)
	writeTestError(b, tc)
	writeTestLogs(b, tc.Logs)
	b.WriteString("</div>\n")
}

// writeTestSteps renders the step list inside a test detail panel.
func writeTestSteps(b *strings.Builder, steps []TestStep) {
	if len(steps) == 0 {
		return
	}
	b.WriteString("<div class=\"step-list\">\n")
	b.WriteString("<div class=\"step-header\">Steps</div>\n")
	for _, step := range steps {
		var icon string = "&#10003;" // checkmark
		if step.Status == "failed" {
			icon = "&#10007;" // cross
		}
		fmt.Fprintf(b, "<div class=\"step-row %s\">\n", step.Status)
		fmt.Fprintf(b, "<span class=\"step-icon\">%s</span>\n", icon)
		fmt.Fprintf(b, "<span class=\"step-title\">%s</span>\n", html.EscapeString(step.Title))
		fmt.Fprintf(b, "<span class=\"step-duration\">%s</span>\n", formatDuration(step.Duration))
		b.WriteString("</div>\n")
		if step.Error != "" {
			fmt.Fprintf(b, "<div class=\"step-error\">%s</div>\n", html.EscapeString(step.Error))
		}
	}
	b.WriteString("</div>\n")
}

// writeTestError renders the error message inside a test detail panel.
func writeTestError(b *strings.Builder, tc *TestCaseResult) {
	if tc.Status != StatusFailed || tc.ErrorMsg == "" {
		return
	}
	fmt.Fprintf(b, "<div class=\"error-msg\">%s</div>\n", html.EscapeString(tc.ErrorMsg))
}

// writeTestLogs renders the per-test console output section.
func writeTestLogs(b *strings.Builder, logs []string) {
	if len(logs) == 0 {
		return
	}
	b.WriteString("<div class=\"log-section\">\n")
	b.WriteString("<div class=\"step-header\">Console Output</div>\n")
	b.WriteString("<pre class=\"log-output\">")
	for _, line := range logs {
		b.WriteString(html.EscapeString(line))
		b.WriteString("\n")
	}
	b.WriteString("</pre>\n")
	b.WriteString("</div>\n")
}

// writeFooter writes the report footer with branding.
func writeFooter(b *strings.Builder, r *TestReport) {
	b.WriteString("<div class=\"footer\">\n")
	fmt.Fprintf(b, "Generated by <a href=\"#\">Kexas</a> on %s\n",
		r.Timestamp.Format("2006-01-02 15:04:05"))
	b.WriteString("</div>\n")
}

// writeDocClose writes the closing tags and embedded JavaScript.
func writeDocClose(b *strings.Builder) {
	b.WriteString("</div>\n") // close container
	b.WriteString("<script>")
	b.WriteString(reportJS)
	b.WriteString("</script>\n")
	b.WriteString("</body>\n</html>\n")
}

// formatDuration formats a duration into a human-friendly string.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d\u00B5s", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	var mins int = int(d.Minutes())
	var secs int = int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", mins, secs)
}

// formatTestName extracts the short test name, prefixing group path in muted style.
func formatTestName(fullName string) string {
	var idx int = strings.LastIndex(fullName, ".")
	if idx < 0 {
		return html.EscapeString(fullName)
	}
	var prefix string = fullName[:idx+1]
	var name string = fullName[idx+1:]
	return fmt.Sprintf("<span class=\"group-prefix\">%s</span>%s",
		html.EscapeString(prefix), html.EscapeString(name))
}

// safePercent calculates percentage, returning 0 if total is 0.
func safePercent(part int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100.0
}

func buildDonutStyle(pass float64, fail float64, skip float64) string {
	var stopPass float64 = pass
	var stopFail float64 = pass + fail
	var stopTotal float64 = pass + fail + skip
	if stopTotal <= 0 {
		stopTotal = 100
	}
	return fmt.Sprintf(
		"background: conic-gradient(var(--pass) 0 %.2f%%, var(--fail) %.2f%% %.2f%%, var(--skip) %.2f%% %.2f%%);",
		stopPass, stopPass, stopFail, stopFail, stopTotal,
	)
}

func topSlowTests(r *TestReport, limit int) []TestCaseResult {
	var all []TestCaseResult
	for _, suite := range r.Suites {
		all = append(all, suite.Tests...)
	}
	if len(all) == 0 {
		return nil
	}
	sort.SliceStable(all, func(i int, j int) bool {
		return all[i].Duration > all[j].Duration
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}
