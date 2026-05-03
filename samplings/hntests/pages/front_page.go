// Package pages — POM wrappers for Hacker News.
package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

// HN URLs.
const (
	HNFrontURL  string = "https://news.ycombinator.com/"
	HNNewestURL string = "https://news.ycombinator.com/newest"
)

// FrontPage wraps any HN listing page (front, newest, ask, show, etc.) since
// they all share the same DOM structure (table of `.titleline` rows).
type FrontPage struct {
	page *kexas.Page
}

// NewFrontPage constructs a FrontPage bound to the given page.
func NewFrontPage(p *kexas.Page) *FrontPage { return &FrontPage{page: p} }

// Open navigates to the standard front page.
func (fp *FrontPage) Open() (*FrontPage, error) {
	return fp.openURL(HNFrontURL)
}

// OpenNewest navigates to /newest.
func (fp *FrontPage) OpenNewest() (*FrontPage, error) {
	return fp.openURL(HNNewestURL)
}

func (fp *FrontPage) openURL(target string) (*FrontPage, error) {
	if err := fp.page.Navigate(target); err != nil {
		return nil, fmt.Errorf("hn: navigate %s failed: %w", target, err)
	}
	return fp, nil
}

// StoryCount returns the number of `.titleline` rows on the current page.
// Front + newest both render exactly 30 stories per page since launch (2007).
func (fp *FrontPage) StoryCount() (int, error) {
	expr := `document.querySelectorAll('.titleline').length`
	result, err := fp.page.Evaluate(expr)
	if err != nil {
		return 0, fmt.Errorf("hn: count eval failed: %w", err)
	}
	return numericFromEvaluate(result), nil
}

// FirstStoryTitle returns the text content of the first story's title link.
func (fp *FrontPage) FirstStoryTitle() (string, error) {
	expr := `(() => {
		const el = document.querySelector('.titleline > a');
		return el ? el.textContent : '';
	})()`
	result, err := fp.page.Evaluate(expr)
	if err != nil {
		return "", fmt.Errorf("hn: first title eval failed: %w", err)
	}
	return stringFromEvaluate(result), nil
}
