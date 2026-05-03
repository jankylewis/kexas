// Package pages contains Page Object Model wrappers for each saucedemo screen.
// Methods expose domain-level actions (LoginAs, AddItemToCart) — never raw Find/Click.
package pages

import (
	"fmt"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/sltests/data"
)

// Login page selectors — single source of truth.
const (
	loginUsernameSelector = "#user-name"
	loginPasswordSelector = "#password"
	loginButtonSelector   = "#login-button"
	loginErrorSelector    = `[data-test="error"]`
)

// LoginURL is the saucedemo entry point.
const LoginURL string = "https://www.saucedemo.com/"

// LoginPage wraps the saucedemo login screen.
type LoginPage struct {
	page *kexas.Page
}

// NewLoginPage constructs a LoginPage bound to the given page.
func NewLoginPage(p *kexas.Page) *LoginPage { return &LoginPage{page: p} }

// Open navigates to the login URL and returns the LoginPage for chaining.
func (lp *LoginPage) Open() (*LoginPage, error) {
	if err := lp.page.Navigate(LoginURL); err != nil {
		return nil, fmt.Errorf("login: navigate failed: %w", err)
	}
	return lp, nil
}

// FillUsername sets the username field via kexas's Element.Fill (React-aware setter).
func (lp *LoginPage) FillUsername(value string) error {
	field, err := lp.page.Find(loginUsernameSelector)
	if err != nil {
		return fmt.Errorf("login: username field not found: %w", err)
	}
	return field.Fill(value)
}

// FillPassword sets the password field via kexas's Element.Fill.
func (lp *LoginPage) FillPassword(value string) error {
	field, err := lp.page.Find(loginPasswordSelector)
	if err != nil {
		return fmt.Errorf("login: password field not found: %w", err)
	}
	return field.Fill(value)
}

// ClickLogin submits the login form by clicking the LOGIN button.
func (lp *LoginPage) ClickLogin() error {
	btn, err := lp.page.Find(loginButtonSelector)
	if err != nil {
		return fmt.Errorf("login: login button not found: %w", err)
	}
	return btn.Click()
}

// LoginAs fills username + password and submits — the standard happy-path login flow.
// Returns an InventoryPage assuming the login succeeded; the caller decides whether
// to assert the redirect actually happened.
func (lp *LoginPage) LoginAs(user data.User) (*InventoryPage, error) {
	if err := lp.FillUsername(user.Username); err != nil {
		return nil, err
	}
	if err := lp.FillPassword(user.Password); err != nil {
		return nil, err
	}
	if err := lp.ClickLogin(); err != nil {
		return nil, err
	}
	return NewInventoryPage(lp.page), nil
}

// ErrorMessage returns the visible error banner text, or "" if no error is shown.
func (lp *LoginPage) ErrorMessage() (string, error) {
	el, err := lp.page.Find(loginErrorSelector)
	if err != nil {
		return "", nil // no error banner present
	}
	return el.GetText()
}
