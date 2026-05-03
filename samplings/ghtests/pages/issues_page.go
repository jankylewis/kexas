package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

// IssuesPage wraps a GitHub repo's issues listing page (/<owner>/<repo>/issues).
type IssuesPage struct {
	page *kexas.Page
}

// NewIssuesPage constructs an IssuesPage bound to the given page.
func NewIssuesPage(p *kexas.Page) *IssuesPage { return &IssuesPage{page: p} }

// OpenGoLangGo navigates to /golang/go/issues.
func (ip *IssuesPage) OpenGoLangGo() (*IssuesPage, error) {
	if err := ip.page.Navigate(GoLangGoIssuesURL); err != nil {
		return nil, fmt.Errorf("github issues: navigate failed: %w", err)
	}
	return ip, nil
}

// IssueCount returns the number of issue-row elements rendered. GitHub's issue
// rows have evolved; this counts elements with data-test or aria-label markers
// likely to identify a row. Done via JS evaluate to avoid Find()'s strict-mode
// (multi-element flow).
func (ip *IssuesPage) IssueCount() (int, error) {
	// Try multiple selector candidates — GitHub HTML evolves; falling back
	// across known patterns keeps the test working across releases.
	expr := `(() => {
		const candidates = [
			'div[aria-label*="Issue"]',
			'a[data-hovercard-type="issue"]',
			'.js-issue-row',
		];
		for (const sel of candidates) {
			const n = document.querySelectorAll(sel).length;
			if (n > 0) return n;
		}
		return 0;
	})()`
	result, err := ip.page.Evaluate(expr)
	if err != nil {
		return 0, fmt.Errorf("github issues: count eval failed: %w", err)
	}
	return numericFromEvaluate(result), nil
}
