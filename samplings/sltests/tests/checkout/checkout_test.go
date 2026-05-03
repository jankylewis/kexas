package checkout_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// CheckoutSuite tests the 3-step checkout flow.
// BeforeEach: log in, add a known item, open the cart so each test starts on /cart.
type CheckoutSuite struct {
	ktest.Suite
}

func (s *CheckoutSuite) BeforeEach() {
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	inv := pages.NewInventoryPage(s.Page)
	if loaded, _ := inv.IsLoaded(); !loaded {
		login, _ := pages.NewLoginPage(s.Page).Open()
		_, _ = login.LoginAs(data.StandardUser())
	}
	_, _ = s.Page.Evaluate(`(function() { localStorage.removeItem("cart-contents"); return true; })()`)
	_ = s.Page.Navigate("https://www.saucedemo.com/inventory.html")
	_ = inv.AddItemToCart(data.Backpack())
	_ = inv.Header.OpenCart()
}

func (s *CheckoutSuite) TestACheckoutRequiresFirstName() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("", "Doe", "12345")
	_, _ = info.Continue()

	msg, _ := info.ErrorMessage()
	kassert.That(s.T(), msg).Contains("First Name is required")
}

func (s *CheckoutSuite) TestBCheckoutRequiresLastName() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "", "12345")
	_, _ = info.Continue()

	msg, _ := info.ErrorMessage()
	kassert.That(s.T(), msg).Contains("Last Name is required")
}

func (s *CheckoutSuite) TestCCheckoutRequiresPostalCode() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "")
	_, _ = info.Continue()

	msg, _ := info.ErrorMessage()
	kassert.That(s.T(), msg).Contains("Postal Code is required")
}

func (s *CheckoutSuite) TestDValidInfoAdvancesToOverview() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, err := info.Continue()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := overview.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *CheckoutSuite) TestEOverviewShowsSubtotalTaxTotal() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()

	subtotal, _ := overview.SubtotalLabel()
	kassert.That(s.T(), subtotal).Contains("$29.99") // Backpack price

	tax, _ := overview.TaxLabel()
	kassert.That(s.T(), tax).Contains("Tax")

	total, _ := overview.TotalLabel()
	kassert.That(s.T(), total).Contains("Total")
}

func (s *CheckoutSuite) TestFFinishOrderShowsCompletePage() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()
	complete, err := overview.Finish()
	kassert.ThatError(s.T(), err).IsNil()

	header, _ := complete.HeaderText()
	kassert.That(s.T(), header).Contains("Thank you for your order")
}

func (s *CheckoutSuite) TestGCancelFromInfoReturnsToCart() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	back, err := info.Cancel()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := back.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *CheckoutSuite) TestHBackToProductsAfterFinish() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()
	complete, _ := overview.Finish()
	inv, err := complete.BackToProducts()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := inv.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *CheckoutSuite) TestICompleteOrderEmptiesCart() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()
	complete, _ := overview.Finish()
	inv, _ := complete.BackToProducts()

	count, err := inv.Header.CartBadgeCount()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), count).Equals(0)
}

func (s *CheckoutSuite) TestJCancelFromOverviewReturnsToInventory() {
	cart := pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()
	inv, err := overview.Cancel()
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := inv.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *CheckoutSuite) TestKMultiItemCheckoutSubtotalSumsCorrectly() {
	// BeforeEach already added Backpack ($29.99) and opened cart.
	// Add Bike Light ($9.99) for a total of 2 items, expected subtotal $39.98.
	cart := pages.NewCartPage(s.Page)
	back, _ := cart.ContinueShopping()
	_ = back.AddItemToCart(data.BikeLight())
	_ = back.Header.OpenCart()

	cart = pages.NewCartPage(s.Page)
	info, _ := cart.Checkout()
	_ = info.FillInfo("Jane", "Doe", "12345")
	overview, _ := info.Continue()

	subtotal, _ := overview.SubtotalLabel()
	kassert.That(s.T(), subtotal).Contains("39.98")
}

func TestCheckout(t *testing.T) { ktest.Run(t, new(CheckoutSuite)) }
