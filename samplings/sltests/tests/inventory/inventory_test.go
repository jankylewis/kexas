package inventory_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// InventorySuite tests the post-login product listing.
// BeforeEach logs in fresh + navigates to inventory so each test starts clean.
type InventorySuite struct {
	ktest.Suite
}

// BeforeEach ensures every test starts on a clean inventory page.
// Fast-path: navigate straight to /inventory.html (works if logged in via persisted localStorage).
// Slow-path: if not logged in, run the full login flow.
func (s *InventorySuite) BeforeEach() {
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	if loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded(); !loaded {
		login, _ := pages.NewLoginPage(s.Page).Open()
		_, _ = login.LoginAs(data.StandardUser())
	}
}

func (s *InventorySuite) TestAShowsAllSixProducts() {
	inv := pages.NewInventoryPage(s.Page)

	count, err := inv.ItemCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(6)
}

func (s *InventorySuite) TestBPageTitleIsProducts() {
	inv := pages.NewInventoryPage(s.Page)

	title, err := inv.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), title).Equals("Products")
}

func (s *InventorySuite) TestCSortByNameAToZ() {
	inv := pages.NewInventoryPage(s.Page)

	err := inv.SortBy(pages.SortNameAZ)
	kassert.ThatError(s.T(), err).IsNil()

	names, _ := inv.ItemNames()
	kassert.That(s.T(), len(names) > 0).IsTrue()
	kassert.That(s.T(), names[0]).Equals("Sauce Labs Backpack")
}

func (s *InventorySuite) TestDSortByNameZToA() {
	inv := pages.NewInventoryPage(s.Page)

	err := inv.SortBy(pages.SortNameZA)
	kassert.ThatError(s.T(), err).IsNil()

	names, _ := inv.ItemNames()
	kassert.That(s.T(), names[0]).Equals("Test.allTheThings() T-Shirt (Red)")
}

func (s *InventorySuite) TestESortByPriceLowToHigh() {
	inv := pages.NewInventoryPage(s.Page)

	err := inv.SortBy(pages.SortPriceLowHi)
	kassert.ThatError(s.T(), err).IsNil()

	prices, _ := inv.ItemPrices()
	kassert.That(s.T(), prices[0]).Equals("$7.99") // Onesie
}

func (s *InventorySuite) TestFSortByPriceHighToLow() {
	inv := pages.NewInventoryPage(s.Page)

	err := inv.SortBy(pages.SortPriceHiLow)
	kassert.ThatError(s.T(), err).IsNil()

	prices, _ := inv.ItemPrices()
	kassert.That(s.T(), prices[0]).Equals("$49.99") // Fleece Jacket
}

func (s *InventorySuite) TestGAddItemUpdatesCartBadge() {
	inv := pages.NewInventoryPage(s.Page)

	err := inv.AddItemToCart(data.Backpack())
	kassert.ThatError(s.T(), err).IsNil()

	count, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(1)
}

func (s *InventorySuite) TestHFooterContainsSocialLinks() {
	socials := []string{
		`[data-test="social-twitter"]`,
		`[data-test="social-facebook"]`,
		`[data-test="social-linkedin"]`,
	}
	for _, sel := range socials {
		_, err := s.Page.Find(sel)
		kassert.ThatError(s.T(), err).Named(sel).IsNil()
	}
}

func (s *InventorySuite) TestIItemDetailShowsCorrectName() {
	inv := pages.NewInventoryPage(s.Page)
	detail, err := inv.OpenItem(data.Backpack())
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := detail.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()

	name, _ := detail.Name()
	kassert.That(s.T(), name).Equals(data.Backpack().Name)
}

func (s *InventorySuite) TestJItemDetailShowsCorrectPrice() {
	inv := pages.NewInventoryPage(s.Page)
	detail, err := inv.OpenItem(data.Backpack())
	kassert.ThatError(s.T(), err).IsNil()

	price, _ := detail.Price()
	kassert.That(s.T(), price).Equals(data.Backpack().Price)
}

func (s *InventorySuite) TestKItemDetailHasNonEmptyDescription() {
	inv := pages.NewInventoryPage(s.Page)
	detail, _ := inv.OpenItem(data.BikeLight())

	desc, err := detail.Description()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(desc) > 0).Named("description non-empty").IsTrue()
}

func (s *InventorySuite) TestLBackToProductsFromItemDetail() {
	inv := pages.NewInventoryPage(s.Page)
	detail, _ := inv.OpenItem(data.Backpack())
	back, err := detail.BackToProducts()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := back.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

// TestMHoverOverShoppingCartIcon exercises Element.Hover (no error on a real
// hoverable element). Saucedemo doesn't surface a visual hover state we can
// assert on, so this is a smoke check for the Hover API itself.
func (s *InventorySuite) TestMHoverOverShoppingCartIcon() {
	icon, err := s.Page.Find(`[data-test="shopping-cart-link"]`)
	kassert.ThatError(s.T(), err).IsNil()

	hoverErr := icon.Hover()
	kassert.ThatError(s.T(), hoverErr).Named("hover").IsNil()
}

// TestNScrollToBottomBringsFooterIntoView exercises Page.ScrollToBottom +
// Element.IsVisible. Footer's social links should be off-screen on first paint
// (1280×720 viewport) and visible after scrolling to the bottom.
func (s *InventorySuite) TestNScrollToBottomBringsFooterIntoView() {
	err := s.Page.ScrollToBottom()
	kassert.ThatError(s.T(), err).IsNil()

	footer, err := s.Page.Find(`[data-test="footer-copy"]`)
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), footer.IsVisible()).Named("footer visible after scroll").IsTrue()
}

// TestOInventoryImagesArentBroken verifies all 6 product images render real
// product art — not the saucedemo "404 sad face" image that problem_user gets.
// Exercises Element.GetAttribute via JS evaluate of img src list.
func (s *InventorySuite) TestOInventoryImagesArentBroken() {
	expr := `Array.from(document.querySelectorAll('.inventory_item_img img')).map(img => img.src)`
	result, err := s.Page.Evaluate(expr)
	kassert.ThatError(s.T(), err).IsNil()

	// Coerce via the same helper sltests uses elsewhere.
	srcs, _ := stringSliceFromEvaluate(result)
	kassert.That(s.T(), len(srcs) > 0).Named("at least one img src").IsTrue()

	// Standard user must NOT see the broken-image substitute.
	for _, src := range srcs {
		hasBroken := false
		if len(src) >= 6 && src[len(src)-len("sl-404"):len(src)-len("sl-404")+len("sl-404")] == "sl-404" {
			hasBroken = true
		}
		kassert.That(s.T(), hasBroken).Named("img src not sl-404 — " + src).IsFalse()
	}
}

func TestInventory(t *testing.T) { ktest.Run(t, new(InventorySuite)) }

// stringSliceFromEvaluate is an inline copy of pages/eval_helpers.go::stringSliceResult
// because that helper is in package `pages` (private to sltests/pages).
func stringSliceFromEvaluate(raw interface{}) ([]string, error) {
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			return stringSliceFromEvaluate(v)
		}
	}
	if arr, ok := raw.([]interface{}); ok {
		out := make([]string, len(arr))
		for i, item := range arr {
			s, _ := item.(string)
			out[i] = s
		}
		return out, nil
	}
	return nil, nil
}
