package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/sltests/components"
	"github.com/jankylewis/sltests/data"
)

// Inventory page selectors.
const (
	inventoryListSelector       = `[data-test="inventory-list"]`
	inventoryItemSelector       = `[data-test="inventory-item"]`
	inventoryTitleSelector      = `[data-test="title"]`
	inventoryActiveSortSelector = `[data-test="active-option"]`
	inventorySortDropdown       = `[data-test="product-sort-container"]`
)

// SortOption values match saucedemo's <select> values.
type SortOption string

const (
	SortNameAZ     SortOption = "az"
	SortNameZA     SortOption = "za"
	SortPriceLowHi SortOption = "lohi"
	SortPriceHiLow SortOption = "hilo"
)

// InventoryPage wraps the post-login product listing.
type InventoryPage struct {
	page   *kexas.Page
	Header *components.Header
}

// NewInventoryPage constructs an InventoryPage bound to the given page.
func NewInventoryPage(p *kexas.Page) *InventoryPage {
	return &InventoryPage{page: p, Header: components.NewHeader(p)}
}

// IsLoaded checks for the inventory container by querying the title element.
// Returns false (no error) if not loaded — useful for post-login redirect checks.
func (ip *InventoryPage) IsLoaded() (bool, error) {
	_, err := ip.page.Find(inventoryTitleSelector)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Title returns the visible page title (typically "Products" on inventory).
func (ip *InventoryPage) Title() (string, error) {
	el, err := ip.page.Find(inventoryTitleSelector)
	if err != nil {
		return "", fmt.Errorf("inventory: title not found: %w", err)
	}
	return el.GetText()
}

// ItemCount returns the number of inventory items rendered on the page.
// Uses Page.Evaluate because Find() is strict-mode (errors on >1 match).
func (ip *InventoryPage) ItemCount() (int, error) {
	result, err := ip.page.Evaluate(
		fmt.Sprintf(`document.querySelectorAll('%s').length`, inventoryItemSelector),
	)
	if err != nil {
		return 0, fmt.Errorf("inventory: item count eval failed: %w", err)
	}
	return numericResult(result)
}

// AddItemToCart clicks "Add to cart" for the given product.
func (ip *InventoryPage) AddItemToCart(product data.Product) error {
	selector := fmt.Sprintf(`[data-test="add-to-cart-%s"]`, product.SelectorID)
	btn, err := ip.page.Find(selector)
	if err != nil {
		return fmt.Errorf("inventory: add-to-cart button for %q not found: %w", product.Name, err)
	}
	return btn.Click()
}

// RemoveItemFromCart clicks "Remove" for the given product (visible after adding).
func (ip *InventoryPage) RemoveItemFromCart(product data.Product) error {
	selector := fmt.Sprintf(`[data-test="remove-%s"]`, product.SelectorID)
	btn, err := ip.page.Find(selector)
	if err != nil {
		return fmt.Errorf("inventory: remove button for %q not found: %w", product.Name, err)
	}
	return btn.Click()
}

// SortBy selects a sort option from the dropdown via JS (kexas Element.Type does
// not support <select>; setting .value + dispatching change is the workaround).
func (ip *InventoryPage) SortBy(option SortOption) error {
	expr := fmt.Sprintf(`(() => {
		const sel = document.querySelector('%s');
		if (!sel) return false;
		sel.value = '%s';
		sel.dispatchEvent(new Event('change', { bubbles: true }));
		return true;
	})()`, inventorySortDropdown, option)
	result, err := ip.page.Evaluate(expr)
	if err != nil {
		return fmt.Errorf("inventory: SortBy(%q) eval failed: %w", option, err)
	}
	if !boolResult(result) {
		return fmt.Errorf("inventory: SortBy(%q) — sort dropdown not found", option)
	}
	return nil
}

// ActiveSortLabel returns the visible label of the active sort option (e.g., "Name (A to Z)").
func (ip *InventoryPage) ActiveSortLabel() (string, error) {
	el, err := ip.page.Find(inventoryActiveSortSelector)
	if err != nil {
		return "", fmt.Errorf("inventory: active-sort element not found: %w", err)
	}
	return el.GetText()
}

// ItemNames returns every visible inventory-item-name text, in display order.
// Uses Evaluate to bypass Find()'s strict-mode (>1 match errors).
func (ip *InventoryPage) ItemNames() ([]string, error) {
	expr := `Array.from(document.querySelectorAll('[data-test="inventory-item-name"]')).map(e => e.textContent)`
	result, err := ip.page.Evaluate(expr)
	if err != nil {
		return nil, fmt.Errorf("inventory: ItemNames eval failed: %w", err)
	}
	return stringSliceResult(result)
}

// ItemPrices returns every visible inventory-item-price text, in display order.
func (ip *InventoryPage) ItemPrices() ([]string, error) {
	expr := `Array.from(document.querySelectorAll('[data-test="inventory-item-price"]')).map(e => e.textContent)`
	result, err := ip.page.Evaluate(expr)
	if err != nil {
		return nil, fmt.Errorf("inventory: ItemPrices eval failed: %w", err)
	}
	return stringSliceResult(result)
}

// OpenItem clicks an inventory-item-name matching the given product, navigating
// to the item detail page. Uses Page.Evaluate because Find() is strict-mode and
// there are 6 inventory-item-name elements on the page.
func (ip *InventoryPage) OpenItem(product data.Product) (*ItemDetailPage, error) {
	expr := fmt.Sprintf(`(() => {
		const target = Array.from(document.querySelectorAll('[data-test="inventory-item-name"]'))
			.find(e => e.textContent === %q);
		if (!target) return false;
		target.click();
		return true;
	})()`, product.Name)
	result, err := ip.page.Evaluate(expr)
	if err != nil {
		return nil, fmt.Errorf("inventory: OpenItem(%q) eval failed: %w", product.Name, err)
	}
	if !boolResult(result) {
		return nil, fmt.Errorf("inventory: item %q not found on page", product.Name)
	}
	return NewItemDetailPage(ip.page), nil
}
