package hn_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/hntests/pages"
)

// HNSuite exercises HN's front page + alternate listing pages.
type HNSuite struct {
	ktest.Suite
}

func (s *HNSuite) TestAFrontPageLoads() {
	front, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	title, err := s.Page.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), title).Contains("Hacker News")
	_ = front
}

func (s *HNSuite) TestBFrontPageHas30Stories() {
	front, _ := pages.NewFrontPage(s.Page).Open()

	count, err := front.StoryCount()
	kassert.ThatError(s.T(), err).IsNil()
	// HN's front page has rendered exactly 30 stories per page since 2007.
	kassert.That(s.T(), count).Equals(30)
}

func (s *HNSuite) TestCNewestPageHasDifferentStories() {
	front, _ := pages.NewFrontPage(s.Page).Open()
	frontTitle, _ := front.FirstStoryTitle()

	newest, _ := pages.NewFrontPage(s.Page).OpenNewest()
	newestTitle, _ := newest.FirstStoryTitle()

	kassert.That(s.T(), len(frontTitle) > 0).Named("front first title non-empty").IsTrue()
	kassert.That(s.T(), len(newestTitle) > 0).Named("newest first title non-empty").IsTrue()
	kassert.That(s.T(), frontTitle).Named("front title").NotEquals(newestTitle)
}

func TestHN(t *testing.T) { ktest.Run(t, new(HNSuite)) }
