package hn_test

import (
	"strings"
	"time"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/kwait"
	"github.com/jankylewis/hntests/pages"
)

// Long-flow tests — multi-step chains targeting ~5s wall-clock each. HN's
// DOM is fixed since 2007, so these double as anti-flake regression coverage.

// TestKWalkPaginationAcrossThreePages — front → click "More" → wait → repeat.
// Verifies count=30 on each page. Exercises Element.Click that triggers full
// navigation (not a SPA route change), plus URL-after-nav state.
func (s *HNSuite) TestKWalkPaginationAcrossThreePages() {
	front, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	count, err := front.StoryCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Named("page 1 count").Equals(30)

	for page := 2; page <= 3; page++ {
		more, findErr := s.Page.Find(".morelink")
		kassert.ThatError(s.T(), findErr).Named("Find .morelink").IsNil()

		clickErr := more.Click()
		kassert.ThatError(s.T(), clickErr).Named("Click .morelink").IsNil()

		// HN serves the next page via a full reload — wait for DOMContentLoaded.
		waitErr := s.Page.WaitForLoadState(kwait.WaitUntilDOMContentLoaded, 10*time.Second)
		kassert.ThatError(s.T(), waitErr).IsNil()

		newCount, countErr := front.StoryCount()
		kassert.ThatError(s.T(), countErr).IsNil()
		kassert.That(s.T(), newCount).Named("page count after More").Equals(30)

		currentURL, urlErr := s.Page.URL()
		kassert.ThatError(s.T(), urlErr).IsNil()
		kassert.That(s.T(), currentURL).Contains("p=")
	}
}

// TestLNavigateAcrossSiteSectionsAllRender — front → newest → past → ask, each
// asserts >= 1 story. Tests fresh state per navigation and that the same POM
// works across listing variants (HN reuses the same DOM for all sections).
func (s *HNSuite) TestLNavigateAcrossSiteSectionsAllRender() {
	sections := []string{
		"https://news.ycombinator.com/",
		"https://news.ycombinator.com/newest",
		"https://news.ycombinator.com/ask",
		"https://news.ycombinator.com/show",
	}

	for _, url := range sections {
		err := s.Page.Navigate(url)
		kassert.ThatError(s.T(), err).Named("Navigate " + url).IsNil()

		// Reuse the FrontPage POM — HN renders all listings with .titleline rows.
		front := pages.NewFrontPage(s.Page)
		count, countErr := front.StoryCount()
		kassert.ThatError(s.T(), countErr).IsNil()
		kassert.That(s.T(), count > 0).Named("non-empty " + url).IsTrue()

		currentURL, urlErr := s.Page.URL()
		kassert.ThatError(s.T(), urlErr).IsNil()
		// Strip "https://news.ycombinator.com" prefix — URL() returns the
		// fully-qualified URL but Contains works on substrings.
		var pathOnly string = strings.TrimPrefix(url, "https://news.ycombinator.com")
		if pathOnly == "" {
			pathOnly = "/"
		}
		kassert.That(s.T(), currentURL).Contains(pathOnly)
	}
}

// TestMScrollIntoViewMoreLinkThenClickAdvancesPage — exercises the chain
// ScrollIntoView → Click → WaitForLoadState → URL assertion. The "More" link
// sits below the fold; ScrollIntoView is needed before Click can hit it on a
// short viewport.
func (s *HNSuite) TestMScrollIntoViewMoreLinkThenClickAdvancesPage() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	more, err := s.Page.Find(".morelink")
	kassert.ThatError(s.T(), err).IsNil()

	err = more.ScrollIntoView()
	kassert.ThatError(s.T(), err).IsNil()

	err = more.Click()
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForLoadState(kwait.WaitUntilDOMContentLoaded, 10*time.Second)
	kassert.ThatError(s.T(), err).IsNil()

	currentURL, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	// HN's pagination URL pattern is /news?p=2.
	kassert.That(s.T(), currentURL).Contains("p=2")
}
