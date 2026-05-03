package flows_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// FlowsSuite — long multi-step user-journey tests targeting ~5s wall-clock.
// Designed to surface state-handling bugs across the full saucedemo lifecycle:
// login → inventory → cart → checkout → confirmation → back-to-products.
//
// Distinct from auth/checkout/cart suites which test one stage in isolation.
type FlowsSuite struct {
	ktest.Suite
}

func (s *FlowsSuite) BeforeEach() {
	// Force-clear any prior session state so each long flow starts cold.
	_ = s.Page.Navigate("https://www.saucedemo.com/")
	_, _ = s.Page.Evaluate(`(function() { localStorage.clear(); return true; })()`)
}

// TestALoginAddThreeItemsAndCompleteCheckout — full happy-path lifecycle.
// login → add 3 items → cart shows 3 → checkout → fill info → continue →
// overview shows correct subtotal → finish → "Thank you" → back-to-products
// → cart empty. Touches every page in the saucedemo flow.
func (s *FlowsSuite) TestALoginAddThreeItemsAndCompleteCheckout() {
	login, err := pages.NewLoginPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	inv, err := login.LoginAs(data.StandardUser())
	kassert.ThatError(s.T(), err).IsNil()

	products := []data.Product{data.Backpack(), data.BikeLight(), data.BoltTShirt()}
	for _, p := range products {
		addErr := inv.AddItemToCart(p)
		kassert.ThatError(s.T(), addErr).Named("AddItemToCart " + p.Name).IsNil()
	}

	badge, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), badge).Named("badge after 3 adds").Equals(3)

	err = inv.Header.OpenCart()
	kassert.ThatError(s.T(), err).IsNil()

	cart := pages.NewCartPage(s.Page)
	count, err := cart.ItemCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Named("cart item count").Equals(3)

	info, err := cart.Checkout()
	kassert.ThatError(s.T(), err).IsNil()

	err = info.FillInfo("Jane", "Doe", "12345")
	kassert.ThatError(s.T(), err).IsNil()

	overview, err := info.Continue()
	kassert.ThatError(s.T(), err).IsNil()

	// Backpack $29.99 + BikeLight $9.99 + BoltTShirt $15.99 = $55.97.
	subtotal, err := overview.SubtotalLabel()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), subtotal).Named("3-item subtotal").Contains("55.97")

	complete, err := overview.Finish()
	kassert.ThatError(s.T(), err).IsNil()

	header, err := complete.HeaderText()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), header).Contains("Thank you for your order")

	invAfter, err := complete.BackToProducts()
	kassert.ThatError(s.T(), err).IsNil()

	emptyBadge, err := invAfter.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), emptyBadge).Named("cart empty after order").Equals(0)
}

// TestBSortDirectionRoundTripChangesFirstItem — login, capture default sort
// first item, switch to price-low-high (different first item), then
// price-high-low (different again), then back to name-A→Z (matches default).
// Tests that SortBy + ItemNames re-fetch state correctly across changes.
func (s *FlowsSuite) TestBSortDirectionRoundTripChangesFirstItem() {
	login, err := pages.NewLoginPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	inv, err := login.LoginAs(data.StandardUser())
	kassert.ThatError(s.T(), err).IsNil()

	defaultNames, err := inv.ItemNames()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), len(defaultNames) > 0).Named("default ItemNames non-empty").IsTrue()
	defaultFirst := defaultNames[0]

	err = inv.SortBy(pages.SortPriceLowHi)
	kassert.ThatError(s.T(), err).IsNil()
	lowFirst := firstName(s, inv)
	kassert.That(s.T(), lowFirst).Named("price-lo-hi first vs default").NotEquals(defaultFirst)

	err = inv.SortBy(pages.SortPriceHiLow)
	kassert.ThatError(s.T(), err).IsNil()
	hiFirst := firstName(s, inv)
	kassert.That(s.T(), hiFirst).Named("price-hi-lo first vs lo-hi").NotEquals(lowFirst)

	err = inv.SortBy(pages.SortNameAZ)
	kassert.ThatError(s.T(), err).IsNil()
	azFirst := firstName(s, inv)
	kassert.That(s.T(), azFirst).Named("name-A-Z restores default first").Equals(defaultFirst)
}

// firstName fetches inv.ItemNames()[0] with assertion-on-error. Extracted to
// keep TestB readable.
func firstName(s *FlowsSuite, inv *pages.InventoryPage) string {
	names, err := inv.ItemNames()
	kassert.ThatError(s.T(), err).IsNil()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

// TestCAddAllSixItemsThenRemoveAllLeavesCartEmpty — bulk add every inventory
// item, verify badge=6 and cart=6, then remove each one from the cart, verify
// cart=0. Stresses the AddItemToCart/RemoveItem code paths under sustained
// repetition + cart-page DOM reflows after each remove.
func (s *FlowsSuite) TestCAddAllSixItemsThenRemoveAllLeavesCartEmpty() {
	login, err := pages.NewLoginPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	inv, err := login.LoginAs(data.StandardUser())
	kassert.ThatError(s.T(), err).IsNil()

	all := data.AllProducts()
	for _, p := range all {
		addErr := inv.AddItemToCart(p)
		kassert.ThatError(s.T(), addErr).Named("AddItemToCart " + p.Name).IsNil()
	}

	badge, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), badge).Named("badge after add-all").Equals(len(all))

	err = inv.Header.OpenCart()
	kassert.ThatError(s.T(), err).IsNil()

	cart := pages.NewCartPage(s.Page)
	count, err := cart.ItemCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Named("cart count after add-all").Equals(len(all))

	for _, p := range all {
		removeErr := cart.RemoveItem(p)
		kassert.ThatError(s.T(), removeErr).Named("RemoveItem " + p.Name).IsNil()
	}

	finalCount, err := cart.ItemCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), finalCount).Named("cart empty after remove-all").Equals(0)
}

func TestFlows(t *testing.T) { ktest.Run(t, new(FlowsSuite)) }
