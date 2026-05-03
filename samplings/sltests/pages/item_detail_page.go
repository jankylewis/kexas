package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

const (
	// Item detail page selectors. Note: same data-test names as inventory tile
	// (the saucedemo single-item view re-uses the inventory item template), so
	// these only resolve to a single element on /inventory-item.html.
	itemDetailContainerSelector  = `[data-test="inventory-item"]`
	itemDetailNameSelector       = `[data-test="inventory-item-name"]`
	itemDetailDescSelector       = `[data-test="inventory-item-desc"]`
	itemDetailPriceSelector      = `[data-test="inventory-item-price"]`
	itemDetailBackButtonSelector = `[data-test="back-to-products"]`
)

// ItemDetailPage wraps the saucedemo single-item detail view (/inventory-item.html?id=N).
type ItemDetailPage struct {
	page *kexas.Page
}

// NewItemDetailPage constructs an ItemDetailPage bound to the given page.
func NewItemDetailPage(p *kexas.Page) *ItemDetailPage {
	return &ItemDetailPage{page: p}
}

// IsLoaded checks for the inventory-item container.
func (ip *ItemDetailPage) IsLoaded() (bool, error) {
	_, err := ip.page.Find(itemDetailContainerSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Name returns the visible product name.
func (ip *ItemDetailPage) Name() (string, error) {
	el, err := ip.page.Find(itemDetailNameSelector)
	if err != nil {
		return "", fmt.Errorf("item-detail: name not found: %w", err)
	}
	return el.GetText()
}

// Description returns the product description.
func (ip *ItemDetailPage) Description() (string, error) {
	el, err := ip.page.Find(itemDetailDescSelector)
	if err != nil {
		return "", fmt.Errorf("item-detail: description not found: %w", err)
	}
	return el.GetText()
}

// Price returns the product price.
func (ip *ItemDetailPage) Price() (string, error) {
	el, err := ip.page.Find(itemDetailPriceSelector)
	if err != nil {
		return "", fmt.Errorf("item-detail: price not found: %w", err)
	}
	return el.GetText()
}

// BackToProducts clicks the BACK TO PRODUCTS button — returns to inventory.
func (ip *ItemDetailPage) BackToProducts() (*InventoryPage, error) {
	btn, err := ip.page.Find(itemDetailBackButtonSelector)
	if err != nil {
		return nil, fmt.Errorf("item-detail: back button not found: %w", err)
	}
	if err := btn.Click(); err != nil {
		return nil, err
	}
	return NewInventoryPage(ip.page), nil
}
