package gh_test

import (
	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/ghtests/pages"
)

// API-coverage tests (D-J) — pick kexas APIs the original A-C tests don't
// touch and exercise them against GitHub's real DOM. GitHub evolves quickly
// so selectors target stable schema.org/microformat hooks where possible.

// TestDFindByXPathExtractsRepoNameFromHeader — Page.FindByXPath against the
// `[itemprop="name"]` schema.org annotation that GitHub has carried forward
// across UI redesigns.
func (s *GitHubSuite) TestDFindByXPathExtractsRepoNameFromHeader() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	// `[itemprop="name"]` is the wrapper; its child `<a>` carries the repo name.
	el, err := s.Page.FindByXPath(`//*[@itemprop='name']/a`)
	kassert.ThatError(s.T(), err).IsNil()

	name, err := el.GetText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), name).Equals("go")
}

// TestEScrollIntoViewBringsHeaderBackToTop — Element.ScrollIntoView. After
// scrolling 4000px down, calling ScrollIntoView on the (sticky) repo-name
// header should snap the viewport back near the top.
func (s *GitHubSuite) TestEScrollIntoViewBringsHeaderBackToTop() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.ScrollBy(0, 4000)
	kassert.ThatError(s.T(), err).IsNil()

	_, yDeep, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), yDeep > 1000).Named("scrolled past header").IsTrue()

	header, err := s.Page.Find(`[itemprop="name"]`)
	kassert.ThatError(s.T(), err).IsNil()

	err = header.ScrollIntoView()
	kassert.ThatError(s.T(), err).IsNil()

	_, yAfter, err := s.Page.ScrollPosition()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), yAfter < yDeep).Named("ScrollIntoView reduced scrollY").IsTrue()
}

// TestFGetAttributeOnSVGOcticon — Element.GetAttribute on a non-HTML element.
// GitHub renders ~149 `.octicon` SVGs per repo page; we pick the first via
// XPath (kexas's Find is strict-single-match) to test SVG attribute access.
// SVG elements live in the SVG namespace but their `class` is read like HTML.
func (s *GitHubSuite) TestFGetAttributeOnSVGOcticon() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	octicon, err := s.Page.FindByXPath("(//*[name()='svg' and contains(@class, 'octicon')])[1]")
	kassert.ThatError(s.T(), err).IsNil()

	classAttr, err := octicon.GetAttribute("class")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), classAttr).Contains("octicon")
}

// TestGNonExistentSelectorReturnsErrorNotPanic — robustness check on the
// failure path. A selector that obviously matches nothing should error
// cleanly, not panic or hang.
func (s *GitHubSuite) TestGNonExistentSelectorReturnsErrorNotPanic() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	_, findErr := s.Page.Find(".kexas-totally-fake-selector-xyz123")
	kassert.That(s.T(), findErr != nil).Named("Find returns error on no-match").IsTrue()
}

// TestHURLAfterRepoNavigationMatches — Page.URL() returns the resolved URL.
// Easy to break with redirect handling or stale cached URL state.
func (s *GitHubSuite) TestHURLAfterRepoNavigationMatches() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	currentURL, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), currentURL).Contains("github.com/golang/go")
}

// TestICustomCookieRoundtrip — SetCookies + GetCookies + ClearCookies on a
// real https domain. GitHub sets its own session cookies; we add a synthetic
// one and verify it co-exists.
func (s *GitHubSuite) TestICustomCookieRoundtrip() {
	_, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	const cookieName string = "kexas_gh_probe"
	const cookieValue string = "probe_value"
	err = s.Page.SetCookies([]kexas.Cookie{{
		Name:   cookieName,
		Value:  cookieValue,
		Domain: ".github.com",
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
	kassert.That(s.T(), found).Named("kexas_gh_probe cookie roundtripped").IsTrue()

	err = s.Page.ClearCookies()
	kassert.ThatError(s.T(), err).IsNil()
}

// TestJSetContentWithFormFillSetsInputValue — Page.SetContent + Element.Fill.
// Renders a synthetic form, Fills the input, then reads the value attribute
// back via JS to verify Fill propagated to the controlled-input shape.
func (s *GitHubSuite) TestJSetContentWithFormFillSetsInputValue() {
	const html string = `<!doctype html>
<html><head><title>kexas-form</title></head>
<body>
  <form>
    <input id="kexas-input" type="text" value="" />
    <button id="kexas-button" type="button">Submit</button>
  </form>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	input, err := s.Page.Find("#kexas-input")
	kassert.ThatError(s.T(), err).IsNil()

	err = input.Fill("kexas-fill-marker")
	kassert.ThatError(s.T(), err).IsNil()

	// Read the .value via Evaluate — Fill should have set both the DOM `value`
	// property and dispatched input/change events. The `value` attribute on
	// the HTML may not reflect this (HTML attribute vs DOM property), so we
	// check the property explicitly.
	raw, err := s.Page.Evaluate(`document.getElementById('kexas-input').value`)
	kassert.ThatError(s.T(), err).IsNil()

	got, ok := stringValueOf(raw)
	kassert.That(s.T(), ok).Named("evaluate returned a string").IsTrue()
	kassert.That(s.T(), got).Equals("kexas-fill-marker")
}

// stringValueOf unwraps the typical CDP Runtime.evaluate return shape into a
// plain Go string. CDP returns either a bare value (when returnByValue) or a
// `{"value": ...}` map; kexas's Page.Evaluate flattens this for us, but the
// return type is still interface{}.
func stringValueOf(raw interface{}) (string, bool) {
	if s, ok := raw.(string); ok {
		return s, true
	}
	if m, ok := raw.(map[string]interface{}); ok {
		if v, ok := m["value"].(string); ok {
			return v, true
		}
	}
	return "", false
}
