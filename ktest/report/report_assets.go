package report

// reportCSS contains the embedded stylesheet for the HTML report.
// Aesthetic: Grafana/Prometheus panel style — dark, clean, information-dense.
const reportCSS string = `
/* ===== PALETTE — Grafana/Prometheus inspired ===== */
:root {
  --bg: #111217;
  --bg-raised: #181b23;
  --bg-card: #1e222a;
  --bg-card-hover: #252a34;
  --bg-input: #151820;
  --bg-panel: #0e1016;
  --border: #2a2e3a;
  --border-subtle: #22252f;
  --text: #d8dee9;
  --text-muted: #8892a4;
  --text-dim: #4c566a;
  --pass: #50fa7b;
  --pass-muted: #3fb950;
  --pass-bg: rgba(80,250,123,0.08);
  --pass-glow: rgba(80,250,123,0.25);
  --fail: #ff5555;
  --fail-muted: #f85149;
  --fail-bg: rgba(255,85,85,0.08);
  --fail-glow: rgba(255,85,85,0.20);
  --skip: #f1fa8c;
  --skip-muted: #d29922;
  --skip-bg: rgba(241,250,140,0.07);
  --accent: #8be9fd;
  --accent-alt: #bd93f9;
  --accent-bg: rgba(139,233,253,0.06);
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.4);
  --shadow: 0 4px 16px rgba(0,0,0,0.35);
  --shadow-lg: 0 8px 32px rgba(0,0,0,0.4);
  --radius: 8px;
  --radius-sm: 5px;
  --radius-xs: 3px;
  --font: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI',
           Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  --font-mono: 'JetBrains Mono', 'SF Mono', 'Fira Code', 'Cascadia Code',
               Consolas, monospace;
  --ease: cubic-bezier(0.4, 0, 0.2, 1);
  --transition: 0.2s var(--ease);
}

[data-theme="light"] {
  --bg: #f4f5f7;
  --bg-raised: #ebedf0;
  --bg-card: #ffffff;
  --bg-card-hover: #f7f8fa;
  --bg-input: #ffffff;
  --bg-panel: #f0f1f4;
  --border: #d4d7dd;
  --border-subtle: #e1e4ea;
  --text: #1a1d26;
  --text-muted: #5a6270;
  --text-dim: #a0a7b4;
  --pass: #1a7f37;
  --pass-muted: #1a7f37;
  --pass-bg: rgba(26,127,55,0.07);
  --pass-glow: rgba(26,127,55,0.15);
  --fail: #cf222e;
  --fail-muted: #cf222e;
  --fail-bg: rgba(207,34,46,0.06);
  --fail-glow: rgba(207,34,46,0.12);
  --skip: #9a6700;
  --skip-muted: #9a6700;
  --skip-bg: rgba(154,103,0,0.06);
  --accent: #0969da;
  --accent-alt: #8250df;
  --accent-bg: rgba(9,105,218,0.05);
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.06);
  --shadow: 0 2px 8px rgba(0,0,0,0.08);
  --shadow-lg: 0 4px 16px rgba(0,0,0,0.10);
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: var(--font);
  background: var(--bg);
  color: var(--text);
  line-height: 1.55;
  transition: background var(--transition), color var(--transition);
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.container {
  max-width: 1020px;
  margin: 0 auto;
  padding: 36px 28px;
}

/* ===== HEADER ===== */
.header {
  text-align: center;
  margin-bottom: 36px;
}
.header h1 {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.4px;
  margin-bottom: 2px;
  background: linear-gradient(135deg, var(--text), var(--accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.header .subtitle {
  color: var(--text-dim);
  font-size: 13px;
  font-weight: 400;
  letter-spacing: 0.2px;
}

.theme-toggle {
  position: fixed;
  top: 20px;
  right: 24px;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all var(--transition);
  z-index: 100;
  box-shadow: var(--shadow-sm);
}
.theme-toggle:hover {
  border-color: var(--accent);
  color: var(--accent);
  box-shadow: 0 0 12px rgba(139,233,253,0.15);
}

/* ===== SUMMARY CARDS — Grafana stat panels ===== */
.summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}
.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
  padding: 20px 16px;
  text-align: center;
  transition: all var(--transition);
  position: relative;
  overflow: hidden;
}
.stat-card::before {
  content: "";
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  opacity: 0;
  transition: opacity var(--transition);
}
.stat-card:hover {
  border-color: var(--border);
  box-shadow: var(--shadow);
  transform: translateY(-2px);
}
.stat-card:hover::before { opacity: 1; }
.stat-card.total::before { background: var(--accent); }
.stat-card.passed::before { background: var(--pass); }
.stat-card.failed::before { background: var(--fail); }
.stat-card.skipped::before { background: var(--skip-muted); }

.stat-card .stat-value {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -1.5px;
  line-height: 1;
}
.stat-card .stat-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  color: var(--text-dim);
  margin-top: 6px;
  font-weight: 500;
}
.stat-card.total .stat-value { color: var(--accent); }
.stat-card.passed .stat-value { color: var(--pass); }
.stat-card.failed .stat-value { color: var(--fail); }
.stat-card.skipped .stat-value { color: var(--skip-muted); }

/* ===== PROGRESS BAR — Grafana gauge style ===== */
.progress-bar-container {
  margin-bottom: 24px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
  padding: 18px 22px;
}
.progress-bar-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}
.progress-bar-header .pass-rate {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.3px;
}
.progress-bar-header .duration {
  font-size: 12px;
  color: var(--text-dim);
  font-family: var(--font-mono);
}
.progress-bar {
  height: 6px;
  border-radius: 3px;
  background: var(--bg-panel);
  overflow: hidden;
  display: flex;
}
.progress-bar .bar-pass {
  background: linear-gradient(90deg, var(--pass-muted), var(--pass));
  transition: width 0.8s var(--ease);
  box-shadow: 0 0 8px var(--pass-glow);
}
.progress-bar .bar-fail {
  background: linear-gradient(90deg, var(--fail-muted), var(--fail));
  transition: width 0.8s var(--ease);
}
.progress-bar .bar-skip {
  background: var(--skip-muted);
  transition: width 0.8s var(--ease);
}

/* ===== INSIGHTS — Prometheus/Grafana dashboard panels ===== */
.insights {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}
.insight-card {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
  padding: 20px 22px;
  box-shadow: var(--shadow-sm);
}
.insight-title {
  font-size: 14px;
  font-weight: 600;
  letter-spacing: -0.1px;
}
.insight-subtitle {
  font-size: 11px;
  color: var(--text-dim);
  margin-bottom: 16px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.donut-wrapper {
  display: flex;
  gap: 22px;
  align-items: center;
}
.donut {
  width: 140px;
  height: 140px;
  border-radius: 50%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 20px rgba(80,250,123,0.08);
  flex-shrink: 0;
}
.donut::after {
  content: "";
  position: absolute;
  width: 88px;
  height: 88px;
  border-radius: 50%;
  background: var(--bg-card);
}
.donut-center {
  position: absolute;
  text-align: center;
  z-index: 1;
}
.donut-value {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -1px;
}
.donut-label {
  font-size: 10px;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1.5px;
  font-weight: 500;
}

.donut-legend {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 13px;
  color: var(--text-muted);
}
.donut-legend li {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.donut-legend strong {
  margin-left: auto;
  color: var(--text);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.donut-legend em {
  font-style: normal;
  color: var(--text-dim);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}
.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
  display: inline-block;
}
.legend-dot.pass { background: var(--pass); box-shadow: 0 0 6px var(--pass-glow); }
.legend-dot.fail { background: var(--fail); box-shadow: 0 0 6px var(--fail-glow); }
.legend-dot.skip { background: var(--skip-muted); }

/* Duration bars — sparkline style */
.duration-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.duration-item {
  background: var(--bg-raised);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  transition: border-color var(--transition);
}
.duration-item:hover {
  border-color: var(--border);
}
.duration-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin-bottom: 6px;
  gap: 12px;
}
.duration-meta span:first-child {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 70%;
}
.duration-meta span:last-child {
  font-family: var(--font-mono);
  color: var(--text-dim);
  font-size: 11px;
  flex-shrink: 0;
}
.duration-bar {
  width: 100%;
  height: 4px;
  border-radius: 2px;
  background: var(--bg-panel);
  overflow: hidden;
}
.duration-bar-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--accent), var(--accent-alt));
  box-shadow: 0 0 8px rgba(139,233,253,0.15);
}
.insight-empty {
  color: var(--text-dim);
  font-size: 12px;
}

/* ===== TOOLBAR — search & filters ===== */
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 18px;
  align-items: center;
}
.search-input {
  flex: 1;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text);
  font-size: 13px;
  font-family: var(--font);
  outline: none;
  transition: all var(--transition);
}
.search-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(139,233,253,0.10);
}
.search-input::placeholder { color: var(--text-dim); }

.filter-btn {
  padding: 7px 14px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-card);
  color: var(--text-dim);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition);
  letter-spacing: 0.2px;
}
.filter-btn:hover {
  border-color: var(--accent);
  color: var(--text-muted);
}
.filter-btn.active {
  background: var(--accent-bg);
  border-color: var(--accent);
  color: var(--accent);
}

/* ===== SUITE — collapsible panels ===== */
.suite {
  margin-bottom: 12px;
}
.suite-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius) var(--radius) 0 0;
  cursor: pointer;
  user-select: none;
  transition: all var(--transition);
}
.suite-header:hover { background: var(--bg-card-hover); }
.suite-header .chevron {
  font-size: 10px;
  color: var(--text-dim);
  transition: transform 0.2s var(--ease);
}
.suite-header.collapsed .chevron { transform: rotate(-90deg); }
.suite-header .suite-name {
  font-weight: 600;
  font-size: 13px;
  flex: 1;
  letter-spacing: -0.1px;
}
.suite-header .suite-stats {
  font-size: 11px;
  color: var(--text-dim);
  font-family: var(--font-mono);
}

.suite-body {
  border: 1px solid var(--border-subtle);
  border-top: none;
  border-radius: 0 0 var(--radius) var(--radius);
  overflow: hidden;
}
.suite-body.hidden { display: none; }

/* ===== TEST ROW ===== */
.test-row {
  display: flex;
  align-items: center;
  padding: 9px 16px 9px 20px;
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition);
  gap: 10px;
}
.test-row:last-of-type { border-bottom: none; }
.test-row:hover { background: var(--bg-card-hover); }

.test-row .status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: box-shadow var(--transition);
}
.test-row .status-dot.passed {
  background: var(--pass);
  box-shadow: 0 0 6px var(--pass-glow);
}
.test-row .status-dot.failed {
  background: var(--fail);
  box-shadow: 0 0 6px var(--fail-glow);
}
.test-row .status-dot.skipped {
  background: var(--skip-muted);
}
.test-row:hover .status-dot.passed { box-shadow: 0 0 10px var(--pass-glow); }
.test-row:hover .status-dot.failed { box-shadow: 0 0 10px var(--fail-glow); }

.test-row .test-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
}
.test-row .test-name .group-prefix {
  color: var(--text-dim);
  font-weight: 400;
}

.test-row .badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.badge.passed { background: var(--pass-bg); color: var(--pass-muted); }
.badge.failed { background: var(--fail-bg); color: var(--fail-muted); }
.badge.skipped { background: var(--skip-bg); color: var(--skip-muted); }

.test-row .test-duration {
  font-size: 11px;
  color: var(--text-dim);
  font-family: var(--font-mono);
  min-width: 56px;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.test-row .worker-tag {
  font-size: 9px;
  color: var(--accent);
  font-family: var(--font-mono);
  background: var(--accent-bg);
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  font-weight: 600;
  letter-spacing: 0.3px;
}

/* ===== TEST DETAIL PANEL — Datadog log viewer style ===== */
.test-row.has-detail { cursor: pointer; }
.detail-toggle {
  font-size: 9px;
  color: var(--text-dim);
  transition: transform 0.2s var(--ease), color 0.2s;
}
.test-row:hover .detail-toggle { color: var(--text-muted); }
.test-row.expanded .detail-toggle {
  transform: rotate(90deg);
  color: var(--accent);
}
.test-detail {
  display: none;
  padding: 12px 20px 14px 38px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-panel);
  border-left: 2px solid var(--border);
  margin-left: 20px;
}
.test-row.expanded + .test-detail {
  display: block;
  animation: slideDown 0.2s var(--ease);
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-4px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* Steps — structured timeline */
.step-list { margin-bottom: 10px; }
.step-header {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  color: var(--text-dim);
  margin-bottom: 8px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}
.step-header::after {
  content: "";
  flex: 1;
  height: 1px;
  background: var(--border-subtle);
}
.step-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  font-size: 12px;
  border-left: 2px solid var(--border-subtle);
  padding-left: 12px;
  margin-left: 4px;
  transition: border-color var(--transition);
}
.step-row:hover {
  border-left-color: var(--text-dim);
}
.step-row.passed {
  border-left-color: rgba(80,250,123,0.25);
}
.step-row.failed {
  border-left-color: rgba(255,85,85,0.35);
}
.step-icon { font-size: 11px; width: 16px; text-align: center; flex-shrink: 0; }
.step-row.passed .step-icon { color: var(--pass-muted); }
.step-row.failed .step-icon { color: var(--fail); }
.step-title {
  flex: 1;
  color: var(--text);
}
.step-duration {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-dim);
  font-variant-numeric: tabular-nums;
}
.step-error {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--fail);
  padding: 4px 0 6px 30px;
  border-left: 2px solid rgba(255,85,85,0.25);
  margin-left: 4px;
}

/* Error message panel */
.error-msg {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--fail);
  background: var(--fail-bg);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--fail);
  white-space: pre-wrap;
  word-break: break-word;
  margin-top: 8px;
  line-height: 1.5;
}

/* ===== CONSOLE OUTPUT — Datadog log stream style ===== */
.log-section { margin-top: 12px; }
.log-output {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-muted);
  background: var(--bg);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  margin: 6px 0 0 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 280px;
  overflow-y: auto;
  line-height: 1.65;
  counter-reset: log-line;
}

/* ===== SCROLLBAR — Grafana-thin style ===== */
.log-output::-webkit-scrollbar { width: 5px; }
.log-output::-webkit-scrollbar-track { background: transparent; }
.log-output::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 3px;
}
.log-output::-webkit-scrollbar-thumb:hover { background: var(--text-dim); }

/* ===== FOOTER ===== */
.footer {
  text-align: center;
  margin-top: 48px;
  padding-top: 20px;
  border-top: 1px solid var(--border-subtle);
  color: var(--text-dim);
  font-size: 11px;
  letter-spacing: 0.2px;
}
.footer a {
  color: var(--accent);
  text-decoration: none;
  transition: color var(--transition);
}
.footer a:hover { color: var(--accent-alt); }

/* ===== RESPONSIVE ===== */
@media (max-width: 640px) {
  .summary { grid-template-columns: repeat(2, 1fr); }
  .toolbar { flex-wrap: wrap; }
  .container { padding: 20px 14px; }
  .test-detail { padding-left: 24px; margin-left: 10px; }
}
`

// reportJS is defined in report_assets_js.go
