package report

import (
	"strings"
	"testing"
)

// ========================================
// CRITICAL UNIT TESTS — Console Logs Rendering
// ========================================

func TestWriteTestLogs_RendersLogSection(t *testing.T) {
	var logs []string = []string{"step 1 executed", "found 6 items"}
	var b strings.Builder
	writeTestLogs(&b, logs)
	var html string = b.String()

	if !strings.Contains(html, "log-section") {
		t.Error("expected log-section wrapper")
	}
	if !strings.Contains(html, "Console Output") {
		t.Error("expected 'Console Output' header")
	}
	if !strings.Contains(html, "log-output") {
		t.Error("expected log-output pre block")
	}
	if !strings.Contains(html, "step 1 executed") {
		t.Error("expected first log line")
	}
	if !strings.Contains(html, "found 6 items") {
		t.Error("expected second log line")
	}
}

func TestWriteTestLogs_NilLogs_NoOutput(t *testing.T) {
	var b strings.Builder
	writeTestLogs(&b, nil)
	if b.Len() != 0 {
		t.Errorf("expected no output for nil logs, got %d bytes", b.Len())
	}
}

func TestWriteTestLogs_EmptySlice_NoOutput(t *testing.T) {
	var b strings.Builder
	writeTestLogs(&b, []string{})
	if b.Len() != 0 {
		t.Errorf("expected no output for empty logs, got %d bytes", b.Len())
	}
}

func TestWriteTestLogs_HTMLEscaping(t *testing.T) {
	var logs []string = []string{"value is <script>alert('xss')</script>"}
	var b strings.Builder
	writeTestLogs(&b, logs)
	var html string = b.String()

	if strings.Contains(html, "<script>") {
		t.Error("log output must HTML-escape script tags")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("expected escaped script tag in log output")
	}
}

func TestWriteTestLogs_PreservesNewlines(t *testing.T) {
	var logs []string = []string{"line1", "line2", "line3"}
	var b strings.Builder
	writeTestLogs(&b, logs)
	var html string = b.String()

	var count int = strings.Count(html, "\n")
	// Each log line adds a \n, plus structural newlines from divs/pre
	if count < 3 {
		t.Errorf("expected at least 3 newlines for 3 log lines, got %d", count)
	}
}

func TestWriteTestLogs_SingleLogLine(t *testing.T) {
	var logs []string = []string{"only one line"}
	var b strings.Builder
	writeTestLogs(&b, logs)
	var html string = b.String()

	if !strings.Contains(html, "only one line") {
		t.Error("expected single log line")
	}
	if !strings.Contains(html, "log-section") {
		t.Error("expected log-section even for a single line")
	}
}
