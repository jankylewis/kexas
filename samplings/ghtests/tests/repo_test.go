package gh_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/ghtests/pages"
)

// GitHubSuite tests GitHub's anonymous public-repo browsing surface.
type GitHubSuite struct {
	ktest.Suite
}

func (s *GitHubSuite) TestAGolangGoRepoLoadsWithCorrectTitle() {
	repo, err := pages.NewRepoPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	title, err := s.Page.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), title).Contains("golang/go")

	hasName, _ := repo.HasRepoNameHeader()
	kassert.That(s.T(), hasName).Named("repo name header present").IsTrue()
}

func (s *GitHubSuite) TestBRepoHasStarCounter() {
	repo, _ := pages.NewRepoPage(s.Page).OpenGoLangGo()

	stars, err := repo.StarCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(stars) > 0).Named("star counter non-empty").IsTrue()
}

func (s *GitHubSuite) TestCIssuesPageOpensWithIssuesList() {
	issues, err := pages.NewIssuesPage(s.Page).OpenGoLangGo()
	kassert.ThatError(s.T(), err).IsNil()

	count, err := issues.IssueCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count > 0).Named("at least one issue row visible").IsTrue()
}

func TestGitHub(t *testing.T) { ktest.Run(t, new(GitHubSuite)) }
