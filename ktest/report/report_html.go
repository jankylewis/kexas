package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// createVersionedCopy stamps an archive copy of basePath with the report timestamp.
// On filename collisions an -NNN suffix is appended (incrementing until a free slot).
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

// versionedFilename returns "report-YYYYMMDD-HHMMSS[-NNN].html" for the given timestamp.
// attempt 0 omits the numeric suffix; attempts 1+ append "-001", "-002", etc.
func versionedFilename(timestamp time.Time, attempt int) string {
	var stamp string = timestamp.Format("20060102-150405")
	if attempt == 0 {
		return fmt.Sprintf("report-%s.html", stamp)
	}
	return fmt.Sprintf("report-%s-%03d.html", stamp, attempt)
}

// copyFile copies src to dst, returning a wrapped error on any I/O failure.
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
