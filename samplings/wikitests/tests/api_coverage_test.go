package wiki_test

import (
	"time"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/kassert"
)

// API-coverage tests (D-J) — pick kexas APIs the original A-C tests don't
// touch and exercise them against Wikipedia's real DOM. Goal is bug-discovery,
// not user-flow coverage. See `.kb/SAMPLINGS_API_COVERAGE.md` for the matrix.

const goLangArticleURL string = "https://en.wikipedia.org/wiki/Go_(programming_language)"

// TestDArticleScrollByMovesPageDown — Page.ScrollBy + Page.ScrollPosition.
func (s *WikiSuite) TestDArticleScrollByMovesPageDown() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	x0, y0, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y0).Named("initial scrollY").Equals(0)
	_ = x0

	err = s.Page.ScrollBy(0, 1500)
	kassert.ThatError(s.T(), err).IsNil()

	_, y1, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y1 > 0).Named("scrollY after ScrollBy").IsTrue()
}

// TestEScrollToTopReturnsToZero — Page.ScrollToTop after a deep scroll.
func (s *WikiSuite) TestEScrollToTopReturnsToZero() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.ScrollBy(0, 4000)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.ScrollToTop()
	kassert.ThatError(s.T(), err).IsNil()

	_, y, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y).Named("scrollY after ScrollToTop").Equals(0)
}

// TestFFindByXPathExtractsLeadParagraph — Page.FindByXPath against a
// non-trivial XPath that targets the first non-empty paragraph in the article
// body. Wikipedia's lead paragraph is generated server-side and stable.
func (s *WikiSuite) TestFFindByXPathExtractsLeadParagraph() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	// First non-empty <p> in document order. The outer `(...)[1]` is required
	// because kexas's FindByXPath enforces single-match (strict mode) — without
	// the positional wrapper, the inner `[1]` filters per-parent and returns
	// every "first paragraph in its parent" (8 matches on this article).
	el, err := s.Page.FindByXPath(`(//div[@id='mw-content-text']//p[normalize-space()])[1]`)
	kassert.ThatError(s.T(), err).IsNil()

	text, err := el.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(text) > 100).Named("lead paragraph length").IsTrue()
	kassert.That(s.T(), text).Contains("Go")
}

// TestGScrollIntoViewMovesViewportToHeading — Element.ScrollIntoView. Targets
// the "External links" h2 heading which sits near the bottom of any article.
func (s *WikiSuite) TestGScrollIntoViewMovesViewportToHeading() {
	err := s.Page.Navigate(goLangArticleURL)
	kassert.ThatError(s.T(), err).IsNil()

	// Wikipedia heading anchors use stable kebab/snake-case ids tied to the
	// section title text. "External_links" is present on virtually every article.
	heading, err := s.Page.Find("#External_links")
	kassert.ThatError(s.T(), err).IsNil()

	err = heading.ScrollIntoView()
	kassert.ThatError(s.T(), err).IsNil()

	_, y, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), y > 1000).Named("scrollY after ScrollIntoView on bottom heading").IsTrue()
}

// TestHCookieRoundtripPersistsAcrossNavigation — SetCookies + GetCookies +
// ClearCookies. Verifies CDP Network.setCookie + getAllCookies work end-to-end.
func (s *WikiSuite) TestHCookieRoundtripPersistsAcrossNavigation() {
	const cookieName string = "kexas_wiki_probe"
	const cookieValue string = "test_value_42"

	err := s.Page.SetCookies([]kexas.Cookie{{
		Name:   cookieName,
		Value:  cookieValue,
		Domain: ".wikipedia.org",
		Path:   "/",
	}})
	kassert.ThatError(s.T(), err).IsNil()

	// Re-navigate so the cookie should be sent on the request and visible after.
	err = s.Page.Navigate("https://en.wikipedia.org/wiki/Main_Page")
	kassert.ThatError(s.T(), err).IsNil()

	cookies, err := s.Page.GetCookies()
	kassert.ThatError(s.T(), err).IsNil()

	var found bool
	var foundValue string
	for _, c := range cookies {
		if c.Name == cookieName {
			found = true
			foundValue = c.Value
			break
		}
	}
	kassert.That(s.T(), found).Named("kexas_wiki_probe cookie present").IsTrue()
	kassert.That(s.T(), foundValue).Equals(cookieValue)

	err = s.Page.ClearCookies()
	kassert.ThatError(s.T(), err).IsNil()

	cookies, err = s.Page.GetCookies()
	kassert.ThatError(s.T(), err).IsNil()
	for _, c := range cookies {
		kassert.That(s.T(), c.Name).Named("no probe cookie after Clear").NotEquals(cookieName)
	}
}

// TestILocalStorageDirectAPIPersistsValue — Page.LocalStorage().Set/Get/Has/
// Remove. Avoids the Evaluate-based JS-string approach used in sltests so the
// typed Go API gets exercised end-to-end.
func (s *WikiSuite) TestILocalStorageDirectAPIPersistsValue() {
	const k string = "kexas_ls_probe"
	const v string = "kexas_ls_value"

	storage := s.Page.LocalStorage()
	err := storage.Set(k, v)
	kassert.ThatError(s.T(), err).IsNil()

	has, err := storage.Has(k)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), has).Named("LocalStorage.Has after Set").IsTrue()

	got, err := storage.Get(k)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), got).Equals(v)

	err = storage.Remove(k)
	kassert.ThatError(s.T(), err).IsNil()

	has, err = storage.Has(k)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), has).Named("LocalStorage.Has after Remove").IsFalse()
}

// TestJSetContentRendersCustomHTMLAndFindsElement — Page.SetContent. Replaces
// the document with synthetic HTML and queries it. Tests that SetContent
// returns control only after the new DOM is parsed and queryable.
func (s *WikiSuite) TestJSetContentRendersCustomHTMLAndFindsElement() {
	const html string = `<!doctype html>
<html><head><title>kexas synthetic page</title></head>
<body>
  <div id="kexas-probe" class="probe">kexas-marker-text</div>
  <input id="kexas-input" type="text" value="" />
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	el, err := s.Page.Find("#kexas-probe")
	kassert.ThatError(s.T(), err).IsNil()

	text, err := el.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), text).Equals("kexas-marker-text")

	title, err := s.Page.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), title).Equals("kexas synthetic page")

	// WaitForElementVisible should resolve immediately for an already-rendered
	// element. Tests the wait API on a synthetic DOM.
	visible, err := s.Page.WaitForElementVisible("#kexas-input", 2*time.Second)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), visible != nil).Named("WaitForElementVisible returned element").IsTrue()
}
