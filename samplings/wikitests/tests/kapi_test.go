package wiki_test

import (
	"strings"

	"github.com/jankylewis/kexas/kapi"
	"github.com/jankylewis/kexas/kassert"
)

// API-side sampling tests — exercise kexas's HTTP client (kapi) against
// Wikipedia's REST API. This is the only samplings project using kapi today;
// it pairs nicely with the existing browser-side tests for cross-verification.
//
// kapi was completely unused across all 4 samplings projects before this file.

const wikipediaRESTBase string = "https://en.wikipedia.org/api/rest_v1"

// TestPAPI_GetGoArticleSummaryReturnsJSON — basic GET + JSON parse against the
// Wikipedia summary endpoint. Verifies kapi.Client can hit a real API and
// surface the response as a typed map.
func (s *WikiSuite) TestPAPI_GetGoArticleSummaryReturnsJSON() {
	client := kapi.NewClient(wikipediaRESTBase,
		kapi.WithHeader("User-Agent", "kexas-samplings/0.1 (https://github.com/jankylewis/kexas)"),
	)

	resp, err := client.Get("/page/summary/Go_(programming_language)")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), resp.IsOK()).Named("HTTP 2xx").IsTrue()

	body, jerr := resp.JSONMap()
	kassert.ThatError(s.T(), jerr).IsNil()

	title, _ := body["title"].(string)
	kassert.That(s.T(), title).Contains("Go")
	extract, _ := body["extract"].(string)
	kassert.That(s.T(), len(extract) > 100).Named("extract has substantial text").IsTrue()
}

// TestQAPI_NonExistentPageReturns404 — verify kapi reports 4xx via IsClientError.
// Wikipedia's REST API responds 404 for unknown titles.
func (s *WikiSuite) TestQAPI_NonExistentPageReturns404() {
	client := kapi.NewClient(wikipediaRESTBase,
		kapi.WithHeader("User-Agent", "kexas-samplings/0.1 (https://github.com/jankylewis/kexas)"),
	)

	resp, err := client.Get("/page/summary/Kexas_does_not_exist_xyz_zzzz9999")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), resp.IsClientError()).Named("HTTP 4xx for unknown title").IsTrue()
}

// TestRAPI_BrowserAndAPITitleAgree — cross-verify: fetch summary via REST API,
// load the page via browser, assert both report the same canonical title. Uses
// kapi alongside ktest's *kexas.Page — exercises the two halves of kexas in
// the same test.
func (s *WikiSuite) TestRAPI_BrowserAndAPITitleAgree() {
	const slug string = "Go_(programming_language)"
	client := kapi.NewClient(wikipediaRESTBase,
		kapi.WithHeader("User-Agent", "kexas-samplings/0.1 (https://github.com/jankylewis/kexas)"),
	)

	resp, err := client.Get("/page/summary/" + slug)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), resp.IsOK()).IsTrue()

	body, jerr := resp.JSONMap()
	kassert.ThatError(s.T(), jerr).IsNil()
	apiTitle, _ := body["title"].(string)

	err = s.Page.Navigate("https://en.wikipedia.org/wiki/" + slug)
	kassert.ThatError(s.T(), err).IsNil()

	heading, ferr := s.Page.Find("#firstHeading")
	kassert.ThatError(s.T(), ferr).IsNil()
	browserTitle, terr := heading.GetText()
	kassert.ThatError(s.T(), terr).IsNil()

	// Wikipedia normalises titles consistently: the API title and the H1 may
	// differ in parenthetical disambiguators. We assert the prefix matches.
	var apiHead string = apiTitle
	if idx := strings.Index(apiHead, " ("); idx > 0 {
		apiHead = apiHead[:idx]
	}
	kassert.That(s.T(), browserTitle).Contains(apiHead)
}
