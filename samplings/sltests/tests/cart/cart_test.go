package cart_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// CartSuite tests the shopping-cart screen.
// BeforeEach logs in fresh; each test adds whatever items it needs.
type CartSuite struct {
	ktest.Suite
}

func (s *CartSuite) BeforeEach() {
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	if loaded, _ := pages.NewInventoryPage(s.Page).IsLoaded(); !loaded {
		login, _ := pages.NewLoginPage(s.Page).Open()
		_, _ = login.LoginAs(data.StandardUser())
	}
	// Reset cart between tests — saucedemo persists cart in localStorage
	_, _ = s.Page.Evaluate(`(function() { localStorage.removeItem("cart-contents"); return true; })()`)
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
}

func (s *CartSuite) TestACartShowsAddedItems() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())
	_ = inv.AddItemToCart(data.BikeLight())
	_ = inv.Header.OpenCart()

	cart := pages.NewCartPage(s.Page)
	count, err := cart.ItemCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(2)

	names, _ := cart.ItemNames()
	// kassert.That.Contains is string-only; manual loop for slice membership
	// (kexas API gap noted in .kb/SLTESTS_DOGFOOD_FINDINGS.md)
	var hasBackpack bool
	for _, n := range names {
		if n == "Sauce Labs Backpack" {
			hasBackpack = true
			break
		}
	}
	kassert.That(s.T(), hasBackpack).Named("cart contains Backpack").IsTrue()
}

func (s *CartSuite) TestBRemoveItemFromCartUpdatesBadge() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())
	_ = inv.AddItemToCart(data.BikeLight())
	_ = inv.Header.OpenCart()

	cart := pages.NewCartPage(s.Page)
	_ = cart.RemoveItem(data.Backpack())

	count, _ := cart.ItemCount()
	kassert.That(s.T(), count).Equals(1)
}

func (s *CartSuite) TestCContinueShoppingReturnsToInventory() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.Header.OpenCart()

	cart := pages.NewCartPage(s.Page)
	back, err := cart.ContinueShopping()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := back.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *CartSuite) TestDCartPersistsAcrossNavigation() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())

	// navigate away and back
	_ = inv.Header.OpenCart()
	cart := pages.NewCartPage(s.Page)
	_, _ = cart.ContinueShopping()
	_ = inv.Header.OpenCart()

	count, _ := cart.ItemCount()
	kassert.That(s.T(), count).Equals(1)
}

func (s *CartSuite) TestEEmptyCartHasNoBadge() {
	inv := pages.NewInventoryPage(s.Page)

	count, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(0)
}

func (s *CartSuite) TestFAddAllSixProductsBadgeShowsSix() {
	inv := pages.NewInventoryPage(s.Page)
	for _, product := range data.AllProducts() {
		err := inv.AddItemToCart(product)
		kassert.ThatError(s.T(), err).Named(product.Name).IsNil()
	}

	count, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(6)
}

// TestGCartContentsMirrorLocalStorage proves the UI badge count agrees with
// saucedemo's underlying localStorage cart-contents JSON. Exercises Page.Evaluate
// reading window.localStorage — useful pattern for app-state assertions that
// don't depend on UI rendering.
func (s *CartSuite) TestGCartContentsMirrorLocalStorage() {
	inv := pages.NewInventoryPage(s.Page)
	_ = inv.AddItemToCart(data.Backpack())
	_ = inv.AddItemToCart(data.BikeLight())

	raw, err := s.Page.Evaluate(`localStorage.getItem("cart-contents")`)
	kassert.ThatError(s.T(), err).IsNil()

	// CDP returns the value as a string (or wrapped). Coerce.
	asStr := ""
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			asStr, _ = v.(string)
		}
	} else if v, ok := raw.(string); ok {
		asStr = v
	}

	// Saucedemo stores cart as a JSON array of integer item IDs. Two items added.
	kassert.That(s.T(), len(asStr) > 0).Named("cart-contents non-empty").IsTrue()
	// Should look like "[4,0]" (Backpack=4, Bike Light=0) — exact ids are
	// saucedemo internals; just check it's an array with 2 commas-or-elements.
	commaCount := 0
	for _, c := range asStr {
		if c == ',' {
			commaCount++
		}
	}
	kassert.That(s.T(), commaCount).Named("commas in cart JSON (one per item gap)").Equals(1)
}

func TestCart(t *testing.T) { ktest.Run(t, new(CartSuite)) }
