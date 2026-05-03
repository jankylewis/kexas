package wiki_test

import (
	"time"

	"github.com/jankylewis/kexas/kassert"
)

// Wait-strategy samplings — exercise the new advanced waiters
// (WaitForElementHidden, WaitForElementDetached, WaitForFunction,
// WaitForTitleContains, WaitForElementText) against a controlled DOM
// injected via Page.SetContent. Synthetic-page approach keeps the tests
// deterministic across Wikipedia's UI evolutions.

// TestSWaitForElementHidden_AfterJSToggle — inject a div, then run JS that
// hides it after 500ms. WaitForElementHidden should resolve in ~500ms.
func (s *WikiSuite) TestSWaitForElementHidden_AfterJSToggle() {
	const html string = `<!doctype html>
<html><body>
  <div id="hide-me" style="display:block">visible</div>
  <script>
    setTimeout(function() {
      document.getElementById('hide-me').style.display = 'none';
    }, 500);
  </script>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForElementHidden("#hide-me", 5*time.Second)
	kassert.ThatError(s.T(), err).Named("element should hide within 5s").IsNil()
}

// TestTWaitForElementDetached_AfterJSRemove — inject a div + a script that
// removes it from the DOM after 400ms. WaitForElementDetached should resolve.
func (s *WikiSuite) TestTWaitForElementDetached_AfterJSRemove() {
	const html string = `<!doctype html>
<html><body>
  <div id="remove-me">temp</div>
  <script>
    setTimeout(function() {
      document.getElementById('remove-me').remove();
    }, 400);
  </script>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForElementDetached("#remove-me", 5*time.Second)
	kassert.ThatError(s.T(), err).Named("element should detach within 5s").IsNil()
}

// TestUWaitForFunction_CustomJSCondition — inject a script that sets a
// global flag after 300ms. WaitForFunction polls for the flag.
func (s *WikiSuite) TestUWaitForFunction_CustomJSCondition() {
	const html string = `<!doctype html>
<html><body>
  <script>
    window.kexasReady = false;
    setTimeout(function() { window.kexasReady = true; }, 300);
  </script>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForFunction("window.kexasReady === true", 5*time.Second)
	kassert.ThatError(s.T(), err).Named("flag should be true within 5s").IsNil()
}

// TestVWaitForTitleContains_AfterJSChangeTitle — inject HTML with a script
// that renames document.title after 400ms. WaitForTitleContains polls until
// the new substring appears.
func (s *WikiSuite) TestVWaitForTitleContains_AfterJSChangeTitle() {
	const html string = `<!doctype html>
<html><head><title>initial-title</title></head>
<body>
  <script>
    setTimeout(function() { document.title = 'kexas-rebranded-title'; }, 400);
  </script>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForTitleContains("kexas-rebranded", 5*time.Second)
	kassert.ThatError(s.T(), err).Named("title should change within 5s").IsNil()
}

// TestWWaitForElementText_AfterJSAppend — inject a div whose text mutates
// over time. WaitForElementText polls until the expected substring appears.
func (s *WikiSuite) TestWWaitForElementText_AfterJSAppend() {
	const html string = `<!doctype html>
<html><body>
  <div id="text-target">loading...</div>
  <script>
    setTimeout(function() {
      document.getElementById('text-target').textContent = 'kexas-text-loaded';
    }, 400);
  </script>
</body></html>`

	err := s.Page.SetContent(html)
	kassert.ThatError(s.T(), err).IsNil()

	err = s.Page.WaitForElementText("#text-target", "kexas-text-loaded", 5*time.Second)
	kassert.ThatError(s.T(), err).Named("text should update within 5s").IsNil()
}
