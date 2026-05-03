package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/sltests/data"
)

const (
	cartContainerSelector = `[data-test="cart-contents-container"]`
	cartItemSelector      = `[data-test="inventory-item"]` // same data-test as inventory page
	cartCheckoutSelector  = `[data-test="checkout"]`
	cartContinueSelector  = `[data-test="continue-shopping"]`
)

// CartPage wraps the saucedemo shopping-cart screen.
type CartPage struct {
	page *kexas.Page
}

// NewCartPage constructs a CartPage bound to the given page.
func NewCartPage(p *kexas.Page) *CartPage { return &CartPage{page: p} }

// IsLoaded checks for the cart container.
func (cp *CartPage) IsLoaded() (bool, error) {
	_, err := cp.page.Find(cartContainerSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// ItemCount returns the number of items currently in the cart.
func (cp *CartPage) ItemCount() (int, error) {
	expr := fmt.Sprintf(`document.querySelectorAll('%s').length`, cartItemSelector)
	result, err := cp.page.Evaluate(expr)
	if err != nil {
		return 0, fmt.Errorf("cart: ItemCount eval failed: %w", err)
	}
	return numericResult(result)
}

// ItemNames returns the names of every item currently in the cart.
func (cp *CartPage) ItemNames() ([]string, error) {
	expr := `Array.from(document.querySelectorAll('[data-test="inventory-item-name"]')).map(e => e.textContent)`
	result, err := cp.page.Evaluate(expr)
	if err != nil {
		return nil, fmt.Errorf("cart: ItemNames eval failed: %w", err)
	}
	return stringSliceResult(result)
}

// RemoveItem removes the given product from the cart.
func (cp *CartPage) RemoveItem(product data.Product) error {
	selector := fmt.Sprintf(`[data-test="remove-%s"]`, product.SelectorID)
	btn, err := cp.page.Find(selector)
	if err != nil {
		return fmt.Errorf("cart: remove button for %q not found: %w", product.Name, err)
	}
	return btn.Click()
}

// Checkout clicks the CHECKOUT button, advancing to the personal-info step.
func (cp *CartPage) Checkout() (*CheckoutInfoPage, error) {
	btn, err := cp.page.Find(cartCheckoutSelector)
	if err != nil {
		return nil, fmt.Errorf("cart: checkout button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewCheckoutInfoPage(cp.page), nil
}

// ContinueShopping clicks the CONTINUE SHOPPING button, returning to inventory.
func (cp *CartPage) ContinueShopping() (*InventoryPage, error) {
	btn, err := cp.page.Find(cartContinueSelector)
	if err != nil {
		return nil, fmt.Errorf("cart: continue-shopping button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewInventoryPage(cp.page), nil
}
