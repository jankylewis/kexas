// Package pages — POM wrappers for English Wikipedia.
package pages

import (
	"fmt"
	"net/url"

	"github.com/jankylewis/kexas"
)

const (
	// Vector skin (default modern UI). Both Vector 2010 and Vector 2022 use the
	// same input id; if Wikipedia ever migrates to a new skin, update here.
	mainSearchInputSelector  = "#searchInput"
	articleFirstHeadingSelector = "#firstHeading"
)

// EnglishWikipediaURL is the canonical English Wikipedia entry point.
const EnglishWikipediaURL string = "https://en.wikipedia.org/wiki/Main_Page"

// MainPage wraps the Wikipedia main page (and the search input it carries).
type MainPage struct {
	page *kexas.Page
}

// NewMainPage constructs a MainPage bound to the given page.
func NewMainPage(p *kexas.Page) *MainPage { return &MainPage{page: p} }

// Open navigates to the English Wikipedia main page.
func (mp *MainPage) Open() (*MainPage, error) {
	if err := mp.page.Navigate(EnglishWikipediaURL); err != nil {
		return nil, fmt.Errorf("wikipedia: navigate failed: %w", err)
	}
	return mp, nil
}

// HasSearchInput reports whether the search input is present on the page.
func (mp *MainPage) HasSearchInput() (bool, error) {
	_, err := mp.page.Find(mainSearchInputSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// SearchFor submits a query through Wikipedia's `Special:Search?...&go=Go`
// endpoint. Wikipedia 302-redirects to the matching article when the query
// has an exact title match (the same path the on-page search button takes
// when you click "Go" instead of "Search"). Returns an ArticlePage assuming
// the redirect landed on an article.
//
// Why not driver-level Fill + form-submit: Wikipedia ships at least three
// search-form layouts simultaneously (Vector 2010 / Vector 2022 / Minerva),
// each with different selectors for the form + submit button. Hitting the
// canonical Special:Search URL is robust across skins and across years.
func (mp *MainPage) SearchFor(query string) (*ArticlePage, error) {
	var searchURL string = "https://en.wikipedia.org/wiki/Special:Search?go=Go&search=" + url.QueryEscape(query)
	if err := mp.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("wikipedia: navigate to search failed: %w", err)
	}
	return NewArticlePage(mp.page), nil
}
