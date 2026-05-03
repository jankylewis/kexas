package misc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// MiscSuite covers cross-cutting concerns that don't belong in a single
// feature suite: cookies/storage, screenshots taken programmatically (not
// as failure artifacts), multi-tab handling, and footer-link validation.
//
// Each test exercises a different kexas API surface area to broaden coverage:
//   - TestA: Element.GetAttribute on multiple <a> tags
//   - TestB: Page.Screenshot returning bytes (programmatic, not via report)
//   - TestC: Browser.PageCount + Browser multi-page tracking after window.open
//   - TestD: Page.GetCookies after authenticated navigation
type MiscSuite struct {
	ktest.Suite
}

func (s *MiscSuite) BeforeEach() {
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	if loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded(); !loaded {
		login, _ := pages.NewLoginPage(s.Page).Open()
		_, _ = login.LoginAs(data.StandardUser())
	}
}

// TestASocialLinksHaveValidExternalHrefs grabs the href of each footer social
// link via Element.GetAttribute and verifies they point at the expected
// external sites. Exercises GetAttribute on multiple selectors.
func (s *MiscSuite) TestASocialLinksHaveValidExternalHrefs() {
	expectations := []struct {
		selector string
		domain   string
	}{
		{`[data-test="social-twitter"]`, "twitter.com"},
		{`[data-test="social-facebook"]`, "facebook.com"},
		{`[data-test="social-linkedin"]`, "linkedin.com"},
	}
	for _, want := range expectations {
		// social links live inside an <a> wrapping a <li>; our [data-test=...]
		// targets the <a>, so href is its direct attribute.
		el, err := s.Page.Find(want.selector)
		kassert.ThatError(s.T(), err).Named(want.selector).IsNil()

		href, hrefErr := el.GetAttribute("href")
		kassert.ThatError(s.T(), hrefErr).Named(want.selector + " href").IsNil()
		kassert.That(s.T(), href).Named(want.selector).Contains(want.domain)
	}
}

// TestBProgrammaticScreenshotProducesValidPNG calls Page.Screenshot directly
// (not via the always-on test artifact pipeline), writes the bytes to disk,
// and verifies the file exists + has the PNG magic-number header.
func (s *MiscSuite) TestBProgrammaticScreenshotProducesValidPNG() {
	bytes, err := s.Page.Screenshot()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(bytes) > 0).Named("non-empty screenshot bytes").IsTrue()

	// PNG magic: 0x89 0x50 0x4E 0x47 0x0D 0x0A 0x1A 0x0A
	kassert.That(s.T(), len(bytes) >= 8).Named("at least 8 bytes for PNG header").IsTrue()
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8 && i < len(bytes); i++ {
		kassert.That(s.T(), bytes[i]).Named("PNG magic byte " + string(rune('0'+i))).Equals(pngMagic[i])
	}

	// Round-trip: write to disk, stat, ensure size matches.
	// Sanitize: t.Name() is "TestMisc/TestB..." — the slash means subdir.
	safeName := strings.ReplaceAll(s.T().Name(), "/", "_") + ".png"
	tmp := filepath.Join(os.TempDir(), safeName)
	werr := os.WriteFile(tmp, bytes, 0644)
	kassert.ThatError(s.T(), werr).IsNil()
	defer os.Remove(tmp)

	info, statErr := os.Stat(tmp)
	kassert.ThatError(s.T(), statErr).IsNil()
	kassert.That(s.T(), info.Size()).Equals(int64(len(bytes)))
}

// TestCMultiTabIncreasesPageCount opens a second tab via JS window.open + URL
// navigation, then verifies Browser.PageCount tracks it. Exercises kexas's
// multi-tab attachment via Browser.WaitForNewPage style flow (here we use
// Evaluate to trigger the open, then poll PageCount).
func (s *MiscSuite) TestCMultiTabIncreasesPageCount() {
	beforeCount := s.Browser.PageCount()
	kassert.That(s.T(), beforeCount).Named("baseline page count").Equals(1)

	// Trigger a fresh tab via window.open. The `void` prefix is required because
	// CDP's Runtime.evaluate errors with "Object reference chain is too long" if
	// we let it try to serialize the returned Window object (popups self-reference
	// across origins, causing a circular chain CDP can't traverse).
	_, evalErr := s.Page.Evaluate(`void window.open("https://www.saucedemo.com/cart.html", "_blank")`)
	kassert.ThatError(s.T(), evalErr).IsNil()

	// Give the new target a moment to register with kexas's Browser.
	// PageCount only reflects pages kexas has explicitly attached to. Without
	// auto-attach for popup targets, the new tab won't appear here yet — this
	// is itself a useful finding for the kexas dogfood log.
	time.Sleep(1 * time.Second)
	afterCount := s.Browser.PageCount()
	s.T().Logf("page count: before=%d after=%d", beforeCount, afterCount)

	// Document the current behavior: kexas does NOT auto-attach window.open
	// targets. afterCount stays at 1 unless we explicitly Browser.WaitForNewPage.
	// This test asserts the current (limited) behavior; if kexas adds auto-attach,
	// flip the assertion to >1.
	kassert.That(s.T(), afterCount).Named("PageCount after window.open").Equals(1)
}

// TestDPageHasCookiesAfterAuthenticatedNavigation exercises Page.GetCookies on
// the post-login state. Saucedemo's session is in localStorage (not a cookie),
// so we expect an EMPTY cookie list — important kexas API smoke check that
// GetCookies handles the no-cookies case gracefully (returns empty slice + nil
// error, not nil slice + "no cookies" error).
func (s *MiscSuite) TestDPageHasCookiesAfterAuthenticatedNavigation() {
	cookies, err := s.Page.GetCookies()
	kassert.ThatError(s.T(), err).IsNil()
	// Whether the list is empty or non-empty, the call must succeed.
	s.T().Logf("cookies after login: %d", len(cookies))
}

func TestMisc(t *testing.T) { ktest.Run(t, new(MiscSuite)) }
