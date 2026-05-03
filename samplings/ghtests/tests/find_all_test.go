package gh_test

import (
	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/ghtests/pages"
)

// TestNFindAllOcticonReturnsManySVGs — CSS path, weird DOM (SVGs are not HTML
// elements but querySelectorAll handles them transparently). Original
// strict-Find on GitHub's `.octicon` errors with "matched 149 elements" —
// FindAll should return them all as iterable *Elements.
func (s *GitHubSuite) TestNFindAllOcticonReturnsManySVGs() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	icons, err := s.Page.FindAll(".octicon")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(icons) > 20).Named("FindAll(.octicon) count > 20").IsTrue()

	// Each SVG should respond to GetAttribute("class") — exercises the SVG +
	// objectId-dispatch path on every wrapped element.
	for _, icon := range icons {
		classAttr, attrErr := icon.GetAttribute("class")
		kassert.ThatError(s.T(), attrErr).IsNil()
		kassert.That(s.T(), classAttr).Contains("octicon")
	}
}

// TestOFindAllByXPathFindsAnchorsInsideNav — XPath path through FindAll.
// Targets every <a> inside any <nav> on the page; GitHub renders multiple
// navs (header, repo header, sidebar) so this is a robust multi-element
// signal regardless of UI redesigns.
func (s *GitHubSuite) TestOFindAllByXPathFindsAnchorsInsideNav() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	links, err := s.Page.FindAll(`//nav//a`)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(links) >= 5).Named("nav anchors >= 5").IsTrue()
}

// TestPFindAllOnNoMatchReturnsEmptySliceNotError — negative path. Selector
// matches zero elements; FindAll should return ([], nil), not an error.
func (s *GitHubSuite) TestPFindAllOnNoMatchReturnsEmptySliceNotError() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	missing, err := s.Page.FindAll(".kexas-totally-fake-class-xyz")
	kassert.ThatError(s.T(), err).Named("FindAll on no-match returns nil error").IsNil()
	kassert.That(s.T(), len(missing)).Named("empty slice on no-match").Equals(0)
}
