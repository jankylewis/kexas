package wiki_test

import (
	"github.com/jankylewis/kexas/kassert"
)

// TestNFindAllParagraphsInArticleViaXPath — XPath path through FindAll.
// Wikipedia's Go article has 8+ paragraphs in `mw-content-text`. The original
// strict-FindByXPath errored with "matched 8 elements"; FindAll should
// return them all in document order.
func (s *WikiSuite) TestNFindAllParagraphsInArticleViaXPath() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	// Lead-paragraph XPath that matched 8 elements during the strict-mode
	// finding earlier this week — perfect FindAll target.
	paragraphs, err := s.Page.FindAll(`//div[@id='mw-content-text']//p[normalize-space()][1]`)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(paragraphs) > 1).Named("multiple paragraphs returned").IsTrue()

	// First paragraph (document order) should mention "Go" — sanity-check
	// that ordering is preserved through Runtime.getProperties.
	firstText, err := paragraphs[0].GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), firstText).Contains("Go")
}

// TestOFindAllByCSSReturnsExternalLinksList — CSS path. Wikipedia article
// "External links" sections render as a <ul> of <li> rows; FindAll on the
// list-item selector should return all of them.
func (s *WikiSuite) TestOFindAllByCSSReturnsExternalLinksList() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	// `.mw-parser-output ul li` matches every list item across the article —
	// many sections use bullet lists. >= 5 is a safe lower bound.
	items, err := s.Page.FindAll(".mw-parser-output ul li")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(items) >= 5).Named("article has >= 5 <li> rows").IsTrue()
}
