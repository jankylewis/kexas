package wiki_test

import (
	"time"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/wikitests/pages"
)

// Long-flow tests — multi-step chains targeting ~5s wall-clock each. Designed
// to surface state-handling bugs (URL after multi-nav, scroll position after
// multi-find, title after random-redirect chain) that single-step API tests
// can't reach.

// TestKSearchHyperlinkChainNavigation — Search → article → click first inline
// hyperlink in the lead paragraph → wait for navigation → assert title changed.
// Exercises: Navigate, Find, FindByXPath, Element.Click + page navigation,
// WaitForElementVisible.
func (s *WikiSuite) TestKSearchHyperlinkChainNavigation() {
	main := pages.NewMainPage(s.Page)
	startArticle, err := main.SearchFor("Go (programming language)")
	kassert.ThatError(s.T(), err).IsNil()

	startTitle, err := startArticle.Title()
	kassert.ThatError(s.T(), err).IsNil()

	// First wikilink in the article body — articles always have inline links.
	link, err := s.Page.FindByXPath(`(//div[@id='mw-content-text']//p//a[starts-with(@href, '/wiki/')])[1]`)
	kassert.ThatError(s.T(), err).IsNil()

	href, err := link.GetAttribute("href")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), href).Named("inline link href").StartsWith("/wiki/")

	err = link.Click()
	kassert.ThatError(s.T(), err).IsNil()

	// Wikipedia uses traditional full-page navigation; the new H1 confirms the
	// destination article rendered.
	heading, err := s.Page.WaitForElementVisible("#firstHeading", 10*time.Second)
	kassert.ThatError(s.T(), err).IsNil()
	newTitle, err := heading.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), newTitle).Named("destination title").NotEquals(startTitle)
}

// TestLArticleScrollWalkAcrossMultipleSections — for each known section
// heading, Find + ScrollIntoView + ScrollPosition assert position increases
// monotonically. Tests Find + ScrollIntoView + ScrollPosition in a tight loop.
func (s *WikiSuite) TestLArticleScrollWalkAcrossMultipleSections() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	// Section ids are stable across Wikipedia's skin migrations because they're
	// derived from the heading text. Start near the top, end near the bottom.
	sections := []string{"History", "Design", "Examples", "External_links"}

	var lastY int = -1
	for _, sec := range sections {
		heading, findErr := s.Page.Find("#" + sec)
		kassert.ThatError(s.T(), findErr).Named("Find #" + sec).IsNil()

		scrollErr := heading.ScrollIntoView()
		kassert.ThatError(s.T(), scrollErr).Named("ScrollIntoView " + sec).IsNil()

		_, y, posErr := s.Page.ScrollPosition()
		kassert.ThatError(s.T(), posErr).IsNil()
		kassert.That(s.T(), y > lastY).Named(sec + " is below previous section").IsTrue()
		lastY = y
	}
}

// TestMRandomToRandomToRandomChain — open Special:Random three times, collect
// each title, assert all three differ. Tests redirect-following, fresh-page
// state on each navigation, and Title() consistency.
func (s *WikiSuite) TestMRandomToRandomToRandomChain() {
	titles := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		article, err := pages.OpenRandom(s.Page)
		kassert.ThatError(s.T(), err).IsNil()

		title, titleErr := article.Title()
		kassert.ThatError(s.T(), titleErr).IsNil()
		kassert.That(s.T(), len(title) > 0).Named("non-empty title").IsTrue()
		titles = append(titles, title)
	}

	// Birthday-paradox argument: 3 random pulls from 6.5M articles colliding is
	// vanishingly unlikely. If this ever fails, kexas is caching navigations.
	kassert.That(s.T(), titles[0]).Named("titles[0] vs [1]").NotEquals(titles[1])
	kassert.That(s.T(), titles[1]).Named("titles[1] vs [2]").NotEquals(titles[2])
	kassert.That(s.T(), titles[0]).Named("titles[0] vs [2]").NotEquals(titles[2])
}
