
package ktest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas/ktest/report"
)

// ========================================
// UTILITY TESTS
// ========================================

// TestProjectNameFromDir_Normal verifies extracting project name.
func TestProjectNameFromDir_Normal(t *testing.T) {
	var name string = report.ProjectNameFromDir("/Users/dev/projects/saucelab")
	if name != "saucelab" {
		t.Errorf("expected 'saucelab', got %q", name)
	}
}

// TestProjectNameFromDir_Dot verifies fallback for current directory.
func TestProjectNameFromDir_Dot(t *testing.T) {
	var name string = report.ProjectNameFromDir(".")
	if name != "kexas" {
		t.Errorf("expected 'kexas', got %q", name)
	}
}

// TestProjectNameFromDir_Root verifies fallback for root path.
func TestProjectNameFromDir_Root(t *testing.T) {
	var name string = report.ProjectNameFromDir("/")
	if name != "kexas" {
		t.Errorf("expected 'kexas', got %q", name)
	}
}

// TestHTMLEscaping verifies that test names with special chars are escaped.
func TestHTMLEscaping(t *testing.T) {
	var tmpDir string = t.TempDir()
	var outputPath string = filepath.Join(tmpDir, "report.html")

	var r *report.TestReport = &report.TestReport{
		ProjectName: "Escape<Test>&\"Project\"",
		Timestamp:   time.Now(),
		Suites: []report.TestSuiteResult{
			{
				Name: "suite",
				Tests: []report.TestCaseResult{
					{Name: "Test<XSS>", Status: report.StatusPassed},
				},
			},
		},
	}
	r.ComputeStats()

	var err error = report.Generate(r, outputPath)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var data []byte
	data, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var content string = string(data)

	// Raw < should NOT appear in project name or test name (must be escaped)
	if strings.Contains(content, "Escape<Test>") {
		t.Error("project name should be HTML-escaped, found unescaped < >")
	}
	if strings.Contains(content, "Test<XSS>") {
		t.Error("test name should be HTML-escaped, found unescaped < >")
	}
	// But the escaped versions should be present
	if !strings.Contains(content, "Escape&lt;Test&gt;") {
		t.Error("project name should contain HTML-escaped version")
	}
}
