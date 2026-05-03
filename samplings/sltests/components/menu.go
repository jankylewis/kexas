package components

import (
	"fmt"

	"github.com/jankylewis/kexas"
)

// Menu wraps the side menu opened by Header.OpenMenu(). Each link is identified
// by a stable id; the menu must be open before any link is clicked.
type Menu struct {
	page *kexas.Page
}

// NewMenu constructs a Menu bound to the given page.
func NewMenu(p *kexas.Page) *Menu { return &Menu{page: p} }

// Logout clicks the "Logout" link in the side menu.
func (m *Menu) Logout() error {
	return m.click("#logout_sidebar_link", "logout")
}

// ResetAppState clicks the "Reset App State" link, clearing cart + sort selection.
func (m *Menu) ResetAppState() error {
	return m.click("#reset_sidebar_link", "reset-app-state")
}

// AllItems clicks the "All Items" link, navigating back to the inventory page.
func (m *Menu) AllItems() error {
	return m.click("#inventory_sidebar_link", "all-items")
}

// About clicks the "About" link, navigating to saucelabs.com (external).
func (m *Menu) About() error {
	return m.click("#about_sidebar_link", "about")
}

// Close clicks the X to close the side menu.
func (m *Menu) Close() error {
	return m.click("#react-burger-cross-btn", "close-menu")
}

func (m *Menu) click(selector, label string) error {
	link, err := m.page.Find(selector)
	if err != nil {
		return fmt.Errorf("menu: %s link not found: %w", label, err)
	}
	return link.Click()
}
