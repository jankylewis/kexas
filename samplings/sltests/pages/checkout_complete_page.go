package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const (
	checkoutCompleteContainerSelector = `[data-test="checkout-complete-container"]`
	checkoutCompleteHeaderSelector    = `[data-test="complete-header"]`
	checkoutCompleteTextSelector      = `[data-test="complete-text"]`
	checkoutBackToProductsSelector    = `[data-test="back-to-products"]`
)

// CheckoutCompletePage wraps the order-confirmation screen (final checkout step).
type CheckoutCompletePage struct {
	page *kexas.Page
}

// NewCheckoutCompletePage constructs a CheckoutCompletePage.
func NewCheckoutCompletePage(p *kexas.Page) *CheckoutCompletePage {
	return &CheckoutCompletePage{page: p}
}

// IsLoaded checks for the complete container.
func (cp *CheckoutCompletePage) IsLoaded() (bool, error) {
	_, err := cp.page.Find(checkoutCompleteContainerSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// HeaderText returns the "THANK YOU FOR YOUR ORDER" header text.
func (cp *CheckoutCompletePage) HeaderText() (string, error) {
	el, err := cp.page.Find(checkoutCompleteHeaderSelector)
	if err != nil {
		return "", fmt.Errorf("checkout-complete: header not found: %w", err)
	}
	return el.GetText()
}

// BodyText returns the descriptive paragraph below the header.
func (cp *CheckoutCompletePage) BodyText() (string, error) {
	el, err := cp.page.Find(checkoutCompleteTextSelector)
	if err != nil {
		return "", fmt.Errorf("checkout-complete: body text not found: %w", err)
	}
	return el.GetText()
}

// BackToProducts clicks BACK HOME — returns to the inventory page.
func (cp *CheckoutCompletePage) BackToProducts() (*InventoryPage, error) {
	btn, err := cp.page.Find(checkoutBackToProductsSelector)
	if err != nil {
		return nil, fmt.Errorf("checkout-complete: back-to-products button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewInventoryPage(cp.page), nil
}
