package hn_test

import (
	"time"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/hntests/pages"
)

// API-coverage tests (D-J) — pick kexas APIs the original A-C tests don't
// touch and exercise them against HN's real DOM. HN is a fixture: the HTML
// has barely changed since 2007, so failures here are kexas regressions, not
// site-evolution noise.

// TestDVoteArrowHoverDoesNotError — Element.Hover against HN's tiny
// `.votearrow` upvote target (~10x10 px). Hover-on-microtarget exercises CDP's
// Input.dispatchMouseEvent placement math.
//
// HN renders ~29 `.votearrow` divs on the front page; kexas's Find is strict
// (multi-match → error), so we use FindByXPath with `(...)[1]` to get just
// the first one.
func (s *HNSuite) TestDVoteArrowHoverDoesNotError() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	arrow, err := s.Page.FindByXPath("(//div[@class='votearrow'])[1]")
	kassert.ThatError(s.T(), err).IsNil()

	err = arrow.Hover()
	kassert.ThatError(s.T(), err).Named("Hover on tiny vote arrow").IsNil()
}

// TestEFindByXPathGetsFirstStoryRank — Page.FindByXPath with a positional
// predicate. HN ranks each story with `<span class="rank">1.</span>`.
func (s *HNSuite) TestEFindByXPathGetsFirstStoryRank() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	// "(//span[@class='rank'])[1]" picks the first rank in document order.
	el, err := s.Page.FindByXPath("(//span[@class='rank'])[1]")
	kassert.ThatError(s.T(), err).IsNil()

	rank, err := el.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), rank).Equals("1.")
}

// TestFFrontPageScrollByVerifiesPosition — Page.ScrollBy + ScrollPosition on
// a content-tall page. HN's front page is roughly 2000-3000px tall.
func (s *HNSuite) TestFFrontPageScrollByVerifiesPosition() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.ScrollBy(0, 800)
	kassert.ThatError(s.T(), err).IsNil()

	_, y, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y > 0).Named("scrollY > 0 after ScrollBy(800)").IsTrue()
}

// TestGScrollIntoViewBringsMoreLinkToTop — Element.ScrollIntoView. HN's
// `.morelink` ("More" pagination link) sits at the bottom of the listing.
// After ScrollIntoView, scrollY should be substantially non-zero.
func (s *HNSuite) TestGScrollIntoViewBringsMoreLinkToTop() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	more, err := s.Page.Find(".morelink")
	kassert.ThatError(s.T(), err).IsNil()

	err = more.ScrollIntoView()
	kassert.ThatError(s.T(), err).IsNil()

	_, y, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y > 500).Named("scrollY after ScrollIntoView on .morelink").IsTrue()
}

// TestHURLAfterNewestNavigationMatches — Page.URL() returns the resolved URL
// after navigation. A trivial-looking API but easy to break with redirect
// handling or async state.
func (s *HNSuite) TestHURLAfterNewestNavigationMatches() {
	_, err := pages.NewFrontPage(s.Page).OpenNewest()
	kassert.ThatError(s.T(), err).IsNil()

	currentURL, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), currentURL).Contains("/newest")
}

// TestICustomCookieRoundtrip — SetCookies + GetCookies + ClearCookies.
// HN sets its own session cookies; we add a synthetic one and verify our
// adds are visible alongside HN's.
func (s *HNSuite) TestICustomCookieRoundtrip() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	const cookieName string = "kexas_hn_probe"
	const cookieValue string = "probe_value"
	err = s.Page.SetCookies([]kexas.Cookie{{
		Name:   cookieName,
		Value:  cookieValue,
		Domain: "news.ycombinator.com",
		Path:   "/",
	}})
	kassert.ThatError(s.T(), err).IsNil()

	cookies, err := s.Page.GetCookies()
	kassert.ThatError(s.T(), err).IsNil()

	var found bool
	for _, c := range cookies {
		if c.Name == cookieName && c.Value == cookieValue {
			found = true
			break
		}
	}
	kassert.That(s.T(), found).Named("kexas_hn_probe cookie roundtripped").IsTrue()

	err = s.Page.ClearCookies()
	kassert.ThatError(s.T(), err).IsNil()
}

// TestJSetContentRendersSyntheticHNAndWaitsForElement — Page.SetContent +
// WaitForElementVisible. SetContent replaces the live document; WaitForElement
// then has to query the *new* DOM, not stale references from before.
func (s *HNSuite) TestJSetContentRendersSyntheticHNAndWaitsForElement() {
	const html string = `<!doctype html>
<html><head><title>kexas-hn-fake</title></head>
<body>
  <table><tr class="athing"><td>
    <span class="rank">99.</span>
    <span class="titleline"><a id="fake-title">kexas synthetic story</a></span>
  </td></tr></table>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	visible, err := s.Page.WaitForElementVisible("#fake-title", 2*time.Second)
	kassert.ThatError(s.T(), err).IsNil()

	text, err := visible.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), text).Equals("kexas synthetic story")
}
