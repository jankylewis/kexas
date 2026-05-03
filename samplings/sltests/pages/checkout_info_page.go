package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const (
	checkoutInfoContainerSelector = `[data-test="checkout-info-container"]`
	checkoutFirstNameSelector     = "#first-name"
	checkoutLastNameSelector      = "#last-name"
	checkoutPostalSelector        = "#postal-code"
	checkoutContinueSelector      = `[data-test="continue"]`
	checkoutCancelSelector        = `[data-test="cancel"]`
	checkoutErrorSelector         = `[data-test="error"]`
)

// CheckoutInfoPage wraps the personal-info step (step 1 of checkout).
type CheckoutInfoPage struct {
	page *kexas.Page
}

// NewCheckoutInfoPage constructs a CheckoutInfoPage bound to the given page.
func NewCheckoutInfoPage(p *kexas.Page) *CheckoutInfoPage {
	return &CheckoutInfoPage{page: p}
}

// IsLoaded checks for the checkout-info container.
func (cp *CheckoutInfoPage) IsLoaded() (bool, error) {
	_, err := cp.page.Find(checkoutInfoContainerSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// FillInfo populates first name, last name, and postal code in one call.
// Uses kexas.Element.Fill (React-aware setter, fast + reliable on controlled inputs).
// Empty values are skipped — saucedemo expects to see the "required" error.
func (cp *CheckoutInfoPage) FillInfo(firstName, lastName, postalCode string) error {
	if err := cp.fillField(checkoutFirstNameSelector, firstName, "first name"); err != nil {
		return err
	}
	if err := cp.fillField(checkoutLastNameSelector, lastName, "last name"); err != nil {
		return err
	}
	return cp.fillField(checkoutPostalSelector, postalCode, "postal code")
}

func (cp *CheckoutInfoPage) fillField(selector, value, label string) error {
	if value == "" {
		return nil // intentionally leave empty so the form's required-field error fires
	}
	field, err := cp.page.Find(selector)
	if err != nil {
		return fmt.Errorf("checkout-info: %s field not found: %w", label, err)
	}
	if err := field.Fill(value); err != nil {
		return fmt.Errorf("checkout-info: %s field fill failed: %w", label, err)
	}
	return nil
}

// Continue clicks CONTINUE — proceeds to checkout overview when info is valid.
func (cp *CheckoutInfoPage) Continue() (*CheckoutOverviewPage, error) {
	btn, err := cp.page.Find(checkoutContinueSelector)
	if err != nil {
		return nil, fmt.Errorf("checkout-info: continue button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewCheckoutOverviewPage(cp.page), nil
}

// Cancel clicks CANCEL — returns to the cart.
func (cp *CheckoutInfoPage) Cancel() (*CartPage, error) {
	btn, err := cp.page.Find(checkoutCancelSelector)
	if err != nil {
		return nil, fmt.Errorf("checkout-info: cancel button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewCartPage(cp.page), nil
}

// ErrorMessage returns the visible error banner text, or "" if none.
func (cp *CheckoutInfoPage) ErrorMessage() (string, error) {
	el, err := cp.page.Find(checkoutErrorSelector)
	if err != nil {
		return "", nil
	}
	return el.GetText()
}
