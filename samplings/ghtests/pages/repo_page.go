// Package pages — POM wrappers for GitHub public repository pages.
package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const (
	// Star counter is the small badge in the repo header showing the star total.
	starCounterSelector = "#repo-stars-counter-star"
	// itemprop="name" wraps the repo name in the header.
	repoNameSelector = `[itemprop="name"]`
)

// GoLangGoRepoURL is the canonical Go-language repo on GitHub — used as the
// stable reference target for these sample tests.
const GoLangGoRepoURL string = "https://github.com/golang/go"

// GoLangGoIssuesURL is the issues listing for the same repo.
const GoLangGoIssuesURL string = "https://github.com/golang/go/issues"

// RepoPage wraps a GitHub repository overview page (/<owner>/<repo>).
type RepoPage struct {
	page *kexas.Page
}

// NewRepoPage constructs a RepoPage bound to the given page.
func NewRepoPage(p *kexas.Page) *RepoPage { return &RepoPage{page: p} }

// OpenGoLangGo navigates to https://github.com/golang/go.
func (rp *RepoPage) OpenGoLangGo() (*RepoPage, error) {
	if err := rp.page.Navigate(GoLangGoRepoURL); err != nil {
		return nil, fmt.Errorf("github: navigate failed: %w", err)
	}
	return rp, nil
}

// StarCount returns the visible star-counter text (e.g., "121k").
func (rp *RepoPage) StarCount() (string, error) {
	el, err := rp.page.Find(starCounterSelector)
	if err != nil {
		return "", fmt.Errorf("github: star counter not found: %w", err)
	}
	return el.GetText()
}

// HasRepoNameHeader reports whether the repo-name header element is present.
func (rp *RepoPage) HasRepoNameHeader() (bool, error) {
	_, err := rp.page.Find(repoNameSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}
