package gh_test

import (
	"strings"
	"time"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/kwait"
	"github.com/jankylewis/ghtests/pages"
)

// Long-flow tests — multi-step chains targeting ~5s wall-clock each. GitHub
// loads heavy assets (~2-3s per nav), so these tests double as a soak for the
// recorder's CDP screencast under sustained activity.

// TestKBrowseFromIssuesListToFirstIssueDetail — from /issues, click into the
// first issue, assert the detail URL pattern. Exercises Click + navigation +
// URL-after-redirect on a real PR-style detail page.
func (s *GitHubSuite) TestKBrowseFromIssuesListToFirstIssueDetail() {
	_, err := pages.NewIssuesPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	// `data-hovercard-type="issue"` marks each issue-row link in GitHub's
	// current React rewrites. Take the first via XPath.
	issueLink, err := s.Page.FindByXPath("(//a[@data-hovercard-type='issue'])[1]")
	kassert.ThatError(s.T(), err).IsNil()

	href, err := issueLink.GetAttribute("href")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), href).Named("issue href").Contains("/issues/")

	err = issueLink.Click()
	kassert.ThatError(s.T(), err).IsNil()

	// GitHub uses Turbo-style SPA nav for issue listings — no full reload, so
	// title/readyState waiters return immediately. Poll the URL bar instead.
	err = s.Page.WaitForURLContains("/issues/", 15*time.Second)
	kassert.ThatError(s.T(), err).IsNil()

	currentURL, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), currentURL).Contains("/issues/")
	// Detail URL has a numeric ID — sanity-check the path doesn't end at /issues.
	kassert.That(s.T(), strings.HasSuffix(currentURL, "/issues")).Named("not on listing").IsFalse()
}

// TestLPullsPageRendersOpenAndClosedTabState — navigate to /pulls, find the
// "Closed" filter link, click → wait → assert URL has is%3Aclosed (the
// URL-encoded `is:closed` query GitHub uses).
func (s *GitHubSuite) TestLPullsPageRendersOpenAndClosedTabState() {
	err := s.Page.Navigate("https://github.com/golang/go/pulls")
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForLoadState(kwait.WaitUntilDOMContentLoaded, 15*time.Second)
	kassert.ThatError(s.T(), err).IsNil()

	// The "Closed" filter is an <a> with href containing is%3Aclosed. Several
	// elements may match (sticky nav + sidebar), so positional XPath.
	closedLink, err := s.Page.FindByXPath("(//a[contains(@href, 'is%3Aclosed')])[1]")
	kassert.ThatError(s.T(), err).IsNil()

	err = closedLink.Click()
	kassert.ThatError(s.T(), err).IsNil()

	// Turbo-style SPA nav — see TestK comment on WaitForURLContains.
	err = s.Page.WaitForURLContains("is%3Aclosed", 15*time.Second)
	kassert.ThatError(s.T(), err).IsNil()

	currentURL, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), currentURL).Contains("is%3Aclosed")
}

// TestMVisitMultipleRepoTabsAllRespond — Code → Issues → Pull requests →
// Actions. Each navigation assertion verifies fresh state. Multi-nav under
// continuous recorder pressure.
func (s *GitHubSuite) TestMVisitMultipleRepoTabsAllRespond() {
	tabs := []struct {
		path     string
		expected string
	}{
		{"https://github.com/golang/go", "github.com/golang/go"},
		{"https://github.com/golang/go/issues", "/issues"},
		{"https://github.com/golang/go/pulls", "/pulls"},
		{"https://github.com/golang/go/actions", "/actions"},
	}

	for _, tab := range tabs {
		err := s.Page.Navigate(tab.path)
		kassert.ThatError(s.T(), err).Named("Navigate " + tab.path).IsNil()

		err = s.Page.WaitForLoadState(kwait.WaitUntilDOMContentLoaded, 15*time.Second)
		kassert.ThatError(s.T(), err).Named("WaitForLoadState " + tab.path).IsNil()

		currentURL, urlErr := s.Page.URL()
		kassert.ThatError(s.T(), urlErr).IsNil()
		kassert.That(s.T(), currentURL).Named("URL after nav " + tab.path).Contains(tab.expected)

		// Each tab keeps the repo header — sanity-check the schema.org name marker
		// is still present (and findable) after the route change.
		_, findErr := s.Page.FindByXPath(`(//*[@itemprop='name'])[1]`)
		kassert.ThatError(s.T(), findErr).Named("repo header on " + tab.expected).IsNil()
	}
}
