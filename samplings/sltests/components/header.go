// Package components provides reusable UI components that appear across saucedemo pages.
package components

import (
	"fmt"
	"strconv"

	"github.com/jankylewis/kexas"
)

// Header wraps the top header bar present on every authenticated saucedemo page.
// Carries the burger menu trigger, page title, and shopping-cart icon + badge.
type Header struct {
	page *kexas.Page
}

// NewHeader constructs a Header bound to the given page.
func NewHeader(p *kexas.Page) *Header {
	return &Header{page: p}
}

// OpenMenu clicks the burger icon to open the side menu.
func (h *Header) OpenMenu() error {
	btn, err := h.page.Find("#react-burger-menu-btn")
	if err != nil {
		return fmt.Errorf("header: open-menu button not found: %w", err)
	}
	return btn.Click()
}

// OpenCart clicks the shopping-cart icon — navigates to the cart page.
func (h *Header) OpenCart() error {
	link, err := h.page.Find(`[data-test="shopping-cart-link"]`)
	if err != nil {
		return fmt.Errorf("header: cart link not found: %w", err)
	}
	return link.Click()
}

// CartBadgeCount returns the integer shown on the cart badge (number of items).
// Returns 0 when the badge is hidden (no items) — saucedemo removes the element entirely.
func (h *Header) CartBadgeCount() (int, error) {
	badge, err := h.page.Find(`[data-test="shopping-cart-badge"]`)
	if err != nil {
		return 0, nil // badge absent = empty cart
	}
	text, err := badge.GetText()
	if err != nil {
		return 0, fmt.Errorf("header: cart badge text read failed: %w", err)
	}
	count, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("header: cart badge text %q is not an integer: %w", text, err)
	}
	return count, nil
}
