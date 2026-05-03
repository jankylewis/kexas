package ktest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jankylewis/kexas/ktest/report"
)

// generateHTMLReport converts test results into an HTML report file.
func generateHTMLReport(results []testResult, config *Config) {
	var collector *report.Collector = report.NewCollector(config.ParallelSet)

	for _, r := range results {
		var status report.TestStatus = report.StatusPassed
		if !r.passed {
			status = report.StatusFailed
		}
		var reportSteps []report.TestStep = convertSteps(r.steps)
		// Convert absolute artifact paths to relative-to-reportDir so the rendered
		// HTML references work both when opened locally and when the test-results
		// directory is moved/zipped/served as a static site.
		collector.Add(report.TestCaseResult{
			Name:           r.name,
			Status:         status,
			Duration:       r.elapsed,
			Filename:       r.filename,
			WorkerID:       r.workerID,
			ErrorMsg:       r.errorMsg,
			Steps:          reportSteps,
			Logs:           r.logs,
			ScreenshotPath: relativizeArtifact(config.ReportDir, r.screenshotPath),
			VideoPath:      relativizeArtifact(config.ReportDir, r.videoPath),
		})
	}

	var cwd string
	cwd, _ = os.Getwd()
	var projectName string = report.ProjectNameFromDir(cwd)
	var htmlReport *report.TestReport = collector.BuildReport(projectName)
	var outputPath string
	var versionedPath string
	var err error
	outputPath, versionedPath, err = report.GenerateFiles(htmlReport, config.ReportDir)
	if err != nil {
		ktestLog.Error("failed to generate HTML report", "err", err)
		return
	}
	fmt.Printf("\n📊 HTML report: %s\n", outputPath)
	if versionedPath != "" {
		fmt.Printf("🗂  Archived copy: %s\n", versionedPath)
	}
}

// relativizeArtifact converts an absolute artifact path to a path relative to
// reportDir for embedding in the HTML report. Falls back to the original path on
// failure (e.g., paths on different volumes). Empty paths pass through.
func relativizeArtifact(reportDir, artifactPath string) string {
	if artifactPath == "" {
		return ""
	}
	var absReport string
	var absArtifact string
	var err error
	absReport, err = filepath.Abs(reportDir)
	if err != nil {
		return artifactPath
	}
	absArtifact, err = filepath.Abs(artifactPath)
	if err != nil {
		return artifactPath
	}
	var rel string
	rel, err = filepath.Rel(absReport, absArtifact)
	if err != nil {
		return artifactPath
	}
	return rel
}

// convertSteps converts internal testStep slices to report.TestStep slices.
func convertSteps(steps []testStep) []report.TestStep {
	if len(steps) == 0 {
		return nil
	}
	var result []report.TestStep = make([]report.TestStep, len(steps))
	for i, s := range steps {
		result[i] = report.TestStep{
			Title:    s.Title,
			Status:   s.Status,
			Duration: s.Duration,
			Error:    s.Error,
		}
	}
	return result
}

// printTestSummary prints the final test summary and exits with error code if needed.
func printTestSummary(results []testResult, filteredTests []NamedTest) {
	if len(filteredTests) == 1 {
		fmt.Printf("\nTest finished:\n")
	} else {
		fmt.Printf("\nAll %d tests finished:\n", len(filteredTests))
	}

	var hasFailures bool
	for _, result := range results {
		var displayName string = formatResultDisplayName(result)
		if result.passed {
			fmt.Printf("%s passed\n", displayName)
			continue
		}
		fmt.Printf("%s failed\n", displayName)
		hasFailures = true
	}

	if hasFailures {
		ktestLog.Error("test suite finished with failures")
		os.Exit(1)
	}
	ktestLog.Info("test suite finished", "total", len(filteredTests), "passed", len(filteredTests))
}
