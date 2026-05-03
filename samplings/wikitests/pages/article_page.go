package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const randomArticleURL string = "https://en.wikipedia.org/wiki/Special:Random"

// ArticlePage wraps any Wikipedia article view (`/wiki/<title>`).
type ArticlePage struct {
	page *kexas.Page
}

// NewArticlePage constructs an ArticlePage bound to the given page.
func NewArticlePage(p *kexas.Page) *ArticlePage { return &ArticlePage{page: p} }

// Title returns the article's H1 (`#firstHeading`) text content.
func (ap *ArticlePage) Title() (string, error) {
	el, err := ap.page.Find(articleFirstHeadingSelector)
	if err != nil {
		return "", fmt.Errorf("article: H1 not found: %w", err)
	}
	return el.GetText()
}

// OpenRandom navigates to Special:Random — Wikipedia 302-redirects to a random
// article. Returns the ArticlePage bound to whatever article landed.
func OpenRandom(p *kexas.Page) (*ArticlePage, error) {
	if err := p.Navigate(randomArticleURL); err != nil {
		return nil, fmt.Errorf("article: random navigate failed: %w", err)
	}
	return NewArticlePage(p), nil
}
