# HTML Test Report System — Lessons Learned & Architecture

> Knowledge base entry for the ktest HTML report generator.
> Completed: 2026-02-27

---

## Architecture: Self-Contained Single HTML File

### Decision
The report is a **single `.html` file** with embedded CSS and JS — no external dependencies, no CDN links, no build step. Opens in any browser offline.

### Why NOT a multi-file approach
- External CSS/JS requires a web server or correct relative paths.
- CDN links break offline viewing.
- Multi-file output complicates the `reportDir` contract.
- Single file is trivially shareable (email, Slack, CI artifacts).

### Trade-off
- Larger file size (~15-20KB for typical reports). Acceptable for test reports.
- CSS/JS are Go string constants in `report_assets.go` — editing them requires recompiling.

---

## Package Structure

```
ktest/report/
├── report_model.go       # TestReport, TestCaseResult, TestSuiteResult, ComputeStats
├── report_collector.go   # Thread-safe Collector (mutex-guarded Add/Results/BuildReport)
├── report_html.go        # HTML generation (buildHTML, writeHeader, writeSuites, etc.)
├── report_assets.go      # CSS + JS as Go string constants
```

### Data Flow

```
ktest runner
  → report.NewCollector(workers)
  → collector.Add(result) for each test (thread-safe)
  → collector.BuildReport(projectName)
  → report.Generate(report, outputPath)
  → single .html file
```

---

## Design Decisions

### 1. Collector is Thread-Safe
The `Collector` uses `sync.Mutex` because parallel workers call `Add()` concurrently.
`Results()` returns a **copy** of the slice to prevent mutation leaks.

### 2. Suite Grouping by Filename
Tests are grouped into suites by their `Filename` field (the source file name).
This matches Playwright's convention of grouping by spec file.
Suite order is preserved by insertion order (first test seen from a file determines its position).

### 3. CSS Custom Properties for Theming
All colors are CSS variables on `:root` (dark) and `[data-theme="light"]` (light).
Theme toggle is a single `data-theme` attribute change on `<html>`.
Theme preference is persisted in `localStorage` under key `kexas-report-theme`.

### 4. HTML Escaping
All user-supplied strings (project name, test names, error messages) are escaped with `html.EscapeString()` to prevent XSS. This is verified by a dedicated unit test.

### 5. Worker Tags
Worker ID tags (`W0`, `W1`) are only rendered when `Workers > 1` (parallel mode).
This keeps sequential reports clean.

---

## Color Palette

| Token | Dark | Light |
|-------|------|-------|
| Background | `#0d1117` | `#f0f2f5` |
| Card | `#161b22` | `#ffffff` |
| Text | `#e6edf3` | `#1f2328` |
| Pass | `#3fb950` | `#1a7f37` |
| Fail | `#f85149` | `#cf222e` |
| Skip | `#d29922` | `#9a6700` |
| Accent | `#58a6ff` | `#0969da` |

Inspired by GitHub's dark/light themes for familiarity.

---

## Common Pitfalls & Mistakes to Avoid

| Pitfall | Explanation |
|---------|-------------|
| Testing for CSS class names in HTML | CSS class rules exist in the `<style>` block even when no elements use them. Test for actual rendered `<span class="...">` elements, not the class name string. |
| `fmt.Sprintf` in tight loops | Use `strings.Builder` with `Grow()` for HTML generation — much faster for large reports. |
| Forgetting `html.EscapeString` | Every user string must be escaped before embedding in HTML. Test names, project names, and error messages can contain `<`, `>`, `&`, `"`. |
| File size limit (500 lines) | CSS and JS are large string constants. Keep `report_assets.go` under 500 lines by being concise with CSS. If it grows, split CSS and JS into separate files. |
| Test file size limit (500 lines) | Split test files by concern: model+collector tests vs HTML generation tests. |
| `time.Duration` formatting | Use a custom `formatDuration()` that handles µs, ms, s, and m+s ranges. Don't use `Duration.String()` — it produces ugly output like `1.234567891s`. |
| Progress bar percentage rounding | Use `safePercent()` to handle division by zero when `TotalTests == 0`. |

---

## Testing Strategy

### 21 unit tests across two files:

**`tests/report_test.go`** — Model & Collector (10 tests):
- `ComputeStats` with all-pass, mixed, empty, multi-suite
- `Collector` add/retrieve, copy isolation, thread safety (`-race`), filename grouping, empty filename, worker count

**`tests/report_html_test.go`** — HTML Generation & Utilities (11 tests):
- File creation, HTML content verification, dark/light CSS, error messages
- Worker tags (parallel vs sequential), subdirectory creation
- `ProjectNameFromDir` normal/dot/root, HTML escaping (XSS prevention)

---

## Config Integration

```json
{
  "reportDir": "./test-results"
}
```

- Default: `./test-results`
- Output file: `<reportDir>/report.html`
- Directory is auto-created if it doesn't exist.

---

## Future Enhancements

- **Screenshot embedding**: Base64-encode failure screenshots into the report.
- **Retry history**: Show retry attempts for flaky tests.
- **Trend charts**: Compare pass rates across multiple runs.
- **Duration histogram**: Visualize test execution time distribution.
- **Export**: JSON/CSV export button in the report UI.

---

**Files**: `ktest/report/report_model.go`, `ktest/report/report_collector.go`, `ktest/report/report_html.go`, `ktest/report/report_assets.go`, `ktest/ktest_runner.go`
