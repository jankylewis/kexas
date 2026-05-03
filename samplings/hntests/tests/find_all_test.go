package hn_test

import (
	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/hntests/pages"
)

// TestNFindAllTitleLinesReturns30Elements — Page.FindAll on HN's `.titleline`
// (story title wrapper). The original strict-Find errors with "matched 30
// elements"; FindAll should return all 30 as iterable *Elements with text.
//
// Verifies the new Page.FindAll API end-to-end: CSS path, multi-element
// return, GetText on each wrapped element.
func (s *HNSuite) TestNFindAllTitleLinesReturns30Elements() {
	_, err := pages.NewFrontPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	titles, err := s.Page.FindAll(".titleline")
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(titles)).Named("FindAll(.titleline) count").Equals(30)

	// Every wrapped element should respond to GetText with a non-empty string.
	// This proves the wrapping is real (objectId-bound), not just a fake count.
	for i, el := range titles {
		text, textErr := el.GetText()
		kassert.ThatError(s.T(), textErr).IsNil()
		kassert.That(s.T(), len(text) > 0).Named("title text non-empty").IsTrue()
		_ = i
	}
}
