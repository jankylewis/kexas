package report

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

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
