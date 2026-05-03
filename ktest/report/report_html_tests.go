package report

import (
	"fmt"
	"html"
	"strings"
	"time"
)

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
	var passed int
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

	b.WriteString("<div class=\"test-detail\">\n")
	writeTestSteps(b, tc.Steps)
	writeTestError(b, tc)
	writeTestArtifacts(b, tc)
	writeTestLogs(b, tc.Logs)
	b.WriteString("</div>\n")
}

// writeTestArtifacts renders the per-test screenshot + video inside the detail
// panel — captured for every test (pass or fail) so the panel is meaningful for
// passed tests too. Renders nothing when both artifacts are absent.
func writeTestArtifacts(b *strings.Builder, tc *TestCaseResult) {
	if tc.ScreenshotPath == "" && tc.VideoPath == "" {
		return
	}
	b.WriteString("<div class=\"artifact-section\">\n")
	if tc.ScreenshotPath != "" {
		b.WriteString("<div class=\"step-header\">End-of-test screenshot</div>\n")
		fmt.Fprintf(b, "<a href=\"%s\" target=\"_blank\"><img class=\"artifact-screenshot\" src=\"%s\" alt=\"%s end-of-test screenshot\"></a>\n",
			html.EscapeString(tc.ScreenshotPath),
			html.EscapeString(tc.ScreenshotPath),
			html.EscapeString(tc.Name))
	}
	if tc.VideoPath != "" {
		b.WriteString("<div class=\"step-header\">Recording</div>\n")
		// Render based on file extension:
		//   .mp4 → <video controls> (real ffmpeg-encoded H.264)
		//   .gif → <img>            (kexas's pure-Go fallback; auto-loops in browser)
		//   _frames → fallback link  (legacy frames-dir path; kept for older results)
		switch {
		case strings.HasSuffix(tc.VideoPath, ".gif"):
			fmt.Fprintf(b, "<img class=\"artifact-video\" src=\"%s\" alt=\"%s recording\">\n",
				html.EscapeString(tc.VideoPath), html.EscapeString(tc.Name))
		case strings.HasSuffix(tc.VideoPath, "_frames"):
			fmt.Fprintf(b, "<p class=\"insight-empty\">Frame sequence: <a href=\"%s\">%s</a></p>\n",
				html.EscapeString(tc.VideoPath), html.EscapeString(tc.VideoPath))
		default:
			fmt.Fprintf(b, "<video class=\"artifact-video\" controls src=\"%s\"></video>\n",
				html.EscapeString(tc.VideoPath))
		}
	}
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

// formatDuration formats a duration into a human-friendly string.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
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
