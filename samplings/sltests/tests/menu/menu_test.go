package menu_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/components"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// MenuSuite tests the burger side menu actions (Logout, Reset App State).
type MenuSuite struct {
	ktest.Suite
}

func (s *MenuSuite) BeforeEach() {
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	if loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded(); !loaded {
		login, _ := pages.NewLoginPage(s.Page).Open()
		_, _ = login.LoginAs(data.StandardUser())
	}
	_, _ = s.Page.Evaluate(`(function() { localStorage.removeItem("cart-contents"); return true; })()`)
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
}

func (s *MenuSuite) TestALogoutReturnsToLoginPage() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.Header.OpenMenu()

	menu := components.NewMenu(s.Page)
	_ = menu.Logout()

	// after logout, login form should be visible — locate username field
	_, err := s.Page.Find("#user-name")
	kassert.ThatError(s.T(), err).IsNil()
}

func (s *MenuSuite) TestBResetAppStateClearsCart() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())
	_ = inv.AddItemToCart(data.BikeLight())

	_ = inv.Header.OpenMenu()
	menu := components.NewMenu(s.Page)
	_ = menu.ResetAppState()
	_ = menu.Close()

	count, _ := inv.Header.CartBadgeCount()
	kassert.That(s.T(), count).Equals(0)
}

func (s *MenuSuite) TestCBurgerMenuOpensAndCloses() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.Header.OpenMenu()

	// Logout link should now be visible
	_, openErr := s.Page.Find("#logout_sidebar_link")
	kassert.ThatError(s.T(), openErr).IsNil()

	menu := components.NewMenu(s.Page)
	_ = menu.Close()

	// After close, the cart badge should still be reachable on the inventory page
	// (proves menu was dismissed and underlying page is interactive)
	_, badgeErr := s.Page.Find(`[data-test="shopping-cart-link"]`)
	kassert.ThatError(s.T(), badgeErr).IsNil()
}

func (s *MenuSuite) TestDLogoutClearsAuthenticatedSession() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.Header.OpenMenu()
	menu := components.NewMenu(s.Page)
	_ = menu.Logout()

	// After logout, navigating directly to a protected URL should NOT keep us
	// there — saucedemo bounces unauthenticated requests back to login.
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded()
	kassert.That(s.T(), loaded).Named("inventory should NOT be reachable after logout").IsFalse()
}

func (s *MenuSuite) TestEResetAppStateLeavesSessionIntact() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())

	_ = inv.Header.OpenMenu()
	menu := components.NewMenu(s.Page)
	_ = menu.ResetAppState()
	_ = menu.Close()

	// Cart cleared
	count, _ := inv.Header.CartBadgeCount()
	kassert.That(s.T(), count).Equals(0)

	// But still authenticated — inventory page is still loaded (no login redirect)
	loaded, _ := inv.IsLoaded()
	kassert.That(s.T(), loaded).Named("inventory should still load after reset").IsTrue()
}

// TestFAboutLinkLeavesSauceDemo clicks the About menu link which navigates the
// SAME tab to https://saucelabs.com/. Exercises Page.URL after a cross-domain
// navigation — proves kexas's CDP attachment survives even when the target
// origin changes.
func (s *MenuSuite) TestFAboutLinkLeavesSauceDemo() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.Header.OpenMenu()

	menu := components.NewMenu(s.Page)
	_ = menu.About()

	// Saucelabs.com replaces the saucedemo URL. Wait briefly for the navigation
	// + give the redirect time to settle (saucelabs has CDN/HSTS overhead).
	_ = s.Page.WaitForLoad(8 * time.Second)

	url, err := s.Page.URL()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), url).Named("post-About URL").NotContains("saucedemo")
}

func TestMenu(t *testing.T) { ktest.Run(t, new(MenuSuite)) }
