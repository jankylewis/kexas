package report

// reportJS contains the embedded JavaScript for the HTML report.
// Handles theme toggling, test filtering, search, suite collapsing,
// and test detail panel expand/collapse.
const reportJS string = `
(function() {
  // Theme toggle
  var themeBtn = document.getElementById('theme-toggle');
  var html = document.documentElement;
  var saved = localStorage.getItem('kexas-report-theme');
  if (saved) { html.setAttribute('data-theme', saved); }
  updateThemeIcon();

  themeBtn.addEventListener('click', function() {
    var current = html.getAttribute('data-theme') || 'dark';
    var next = current === 'dark' ? 'light' : 'dark';
    html.setAttribute('data-theme', next);
    localStorage.setItem('kexas-report-theme', next);
    updateThemeIcon();
  });

  function updateThemeIcon() {
    var theme = html.getAttribute('data-theme') || 'dark';
    themeBtn.textContent = theme === 'dark' ? '\u2600\uFE0F' : '\uD83C\uDF19';
  }

  // Suite collapse/expand
  var headers = document.querySelectorAll('.suite-header');
  headers.forEach(function(header) {
    header.addEventListener('click', function() {
      var body = header.nextElementSibling;
      header.classList.toggle('collapsed');
      body.classList.toggle('hidden');
    });
  });

  // Test detail toggle (steps/errors)
  document.querySelectorAll('.test-row.has-detail').forEach(function(row) {
    row.addEventListener('click', function() {
      row.classList.toggle('expanded');
      var nextEl = row.nextElementSibling;
      if (nextEl && nextEl.classList.contains('test-detail')) {
        if (row.classList.contains('expanded')) {
          nextEl.style.display = 'block';
        } else {
          nextEl.style.display = 'none';
        }
      }
    });
  });

  // Filter buttons
  var filterBtns = document.querySelectorAll('.filter-btn');
  filterBtns.forEach(function(btn) {
    btn.addEventListener('click', function() {
      filterBtns.forEach(function(b) { b.classList.remove('active'); });
      btn.classList.add('active');
      applyFilters();
    });
  });

  // Search
  var searchInput = document.getElementById('search-input');
  searchInput.addEventListener('input', function() { applyFilters(); });

  function applyFilters() {
    var activeBtn = document.querySelector('.filter-btn.active');
    var filter = activeBtn ? activeBtn.getAttribute('data-filter') : 'all';
    var query = searchInput.value.toLowerCase().trim();
    var rows = document.querySelectorAll('.test-row');

    rows.forEach(function(row) {
      var status = row.getAttribute('data-status');
      var name = (row.getAttribute('data-name') || '').toLowerCase();
      var matchFilter = (filter === 'all' || status === filter);
      var matchSearch = (query === '' || name.indexOf(query) !== -1);
      var visible = matchFilter && matchSearch;
      row.style.display = visible ? '' : 'none';

      // Collapse expanded rows that are now hidden
      if (!visible && row.classList.contains('expanded')) {
        row.classList.remove('expanded');
      }

      // Show/hide associated detail panel: let CSS handle visible+expanded,
      // but force-hide when the row itself is hidden.
      var nextEl = row.nextElementSibling;
      if (nextEl && nextEl.classList.contains('test-detail')) {
        if (!visible) {
          nextEl.style.display = 'none';
        } else if (row.classList.contains('expanded')) {
          nextEl.style.display = 'block';
        } else {
          nextEl.style.display = '';
        }
      }
    });

    // Hide suites with no visible tests
    var suites = document.querySelectorAll('.suite');
    suites.forEach(function(suite) {
      var visibleRows = suite.querySelectorAll('.test-row[style=""], .test-row:not([style])');
      var hasVisible = false;
      visibleRows.forEach(function(r) {
        if (r.style.display !== 'none') { hasVisible = true; }
      });
      suite.style.display = hasVisible ? '' : 'none';
    });
  }
})();
`
