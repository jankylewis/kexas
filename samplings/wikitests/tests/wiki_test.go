package wiki_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/wikitests/pages"
)

// WikiSuite exercises Wikipedia's search + article-navigation surface.
// Each test starts from the English main page (BeforeEach navigates).
type WikiSuite struct {
	ktest.Suite
}

func (s *WikiSuite) BeforeEach() {
	_, _ = pages.NewMainPage(s.Page).Open()
}

func (s *WikiSuite) TestAMainPageHasSearchInput() {
	main := pages.NewMainPage(s.Page)
	hasInput, err := main.HasSearchInput()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), hasInput).Named("Wikipedia search input present").IsTrue()
}

func (s *WikiSuite) TestBSearchForGoLanguageOpensCorrectArticle() {
	main := pages.NewMainPage(s.Page)
	article, err := main.SearchFor("Go (programming language)")
	kassert.ThatError(s.T(), err).IsNil()

	title, err := article.Title()
	kassert.ThatError(s.T(), err).IsNil()
	// Wikipedia normalises the title to "Go (programming language)" in H1.
	kassert.That(s.T(), title).Contains("Go")
	kassert.That(s.T(), title).Contains("programming language")
}

func (s *WikiSuite) TestCRandomArticleHasNonEmptyTitle() {
	article, err := pages.OpenRandom(s.Page)
	kassert.ThatError(s.T(), err).IsNil()

	title, err := article.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(title) > 0).Named("random article H1 non-empty").IsTrue()
}

func TestWiki(t *testing.T) { ktest.Run(t, new(WikiSuite)) }
