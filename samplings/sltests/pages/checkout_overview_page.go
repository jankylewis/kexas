package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const (
	checkoutSummaryContainerSelector = `[data-test="checkout-summary-container"]`
	checkoutSubtotalSelector         = `[data-test="subtotal-label"]`
	checkoutTaxSelector              = `[data-test="tax-label"]`
	checkoutTotalSelector            = `[data-test="total-label"]`
	checkoutFinishSelector           = `[data-test="finish"]`
	checkoutOverviewCancelSelector   = `[data-test="cancel"]`
)

// CheckoutOverviewPage wraps the order summary (step 2 of checkout).
type CheckoutOverviewPage struct {
	page *kexas.Page
}

// NewCheckoutOverviewPage constructs a CheckoutOverviewPage.
func NewCheckoutOverviewPage(p *kexas.Page) *CheckoutOverviewPage {
	return &CheckoutOverviewPage{page: p}
}

// IsLoaded checks for the summary container.
func (cp *CheckoutOverviewPage) IsLoaded() (bool, error) {
	_, err := cp.page.Find(checkoutSummaryContainerSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// SubtotalLabel returns the rendered "Item total: $X.XX" text.
func (cp *CheckoutOverviewPage) SubtotalLabel() (string, error) {
	return cp.labelText(checkoutSubtotalSelector, "subtotal")
}

// TaxLabel returns the rendered "Tax: $X.XX" text.
func (cp *CheckoutOverviewPage) TaxLabel() (string, error) {
	return cp.labelText(checkoutTaxSelector, "tax")
}

// TotalLabel returns the rendered "Total: $X.XX" text.
func (cp *CheckoutOverviewPage) TotalLabel() (string, error) {
	return cp.labelText(checkoutTotalSelector, "total")
}

func (cp *CheckoutOverviewPage) labelText(selector, label string) (string, error) {
	el, err := cp.page.Find(selector)
	if err != nil {
		return "", fmt.Errorf("checkout-overview: %s label not found: %w", label, err)
	}
	return el.GetText()
}

// Finish clicks FINISH — submits the order, navigates to the complete page.
func (cp *CheckoutOverviewPage) Finish() (*CheckoutCompletePage, error) {
	btn, err := cp.page.Find(checkoutFinishSelector)
	if err != nil {
		return nil, fmt.Errorf("checkout-overview: finish button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewCheckoutCompletePage(cp.page), nil
}

// Cancel clicks CANCEL — returns to the inventory page.
func (cp *CheckoutOverviewPage) Cancel() (*InventoryPage, error) {
	btn, err := cp.page.Find(checkoutOverviewCancelSelector)
	if err != nil {
		return nil, fmt.Errorf("checkout-overview: cancel button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewInventoryPage(cp.page), nil
}
