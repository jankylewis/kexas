package kcore

import (
	"fmt"

	"github.com/jankylewis/kexas/errors"
)

func (p *Page) SetAuthCookie(name string, value string, domain string) error {
	if name == "" {
		return errors.ErrCookieNameEmpty
	}

	var cookie Cookie = Cookie{
		Name:     name,
		Value:    value,
		Domain:   domain,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: SameSiteLax,
	}
	return p.SetCookie(cookie)
}

// HasCookie checks if a cookie with the given name exists.
func (p *Page) HasCookie(name string) (bool, error) {
	if name == "" {
		return false, errors.ErrCookieNameEmpty
	}

	var cookies []Cookie
	var err error
	cookies, err = p.GetCookies()
	if err != nil {
		return false, err
	}

	for i := 0; i < len(cookies); i++ {
		if cookies[i].Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetCookieValue returns the value of a specific cookie by name.
// Returns ErrCookieNotFound if the cookie does not exist.
func (p *Page) GetCookieValue(name string) (string, error) {
	if name == "" {
		return "", errors.ErrCookieNameEmpty
	}

	var cookies []Cookie
	var err error
	cookies, err = p.GetCookies()
	if err != nil {
		return "", err
	}

	for i := 0; i < len(cookies); i++ {
		if cookies[i].Name == name {
			return cookies[i].Value, nil
		}
	}

	return "", fmt.Errorf("cookie '%s': %w", name, errors.ErrCookieNotFound)
}

// parseCookie converts a raw CDP cookie map to a Cookie struct.
func parseCookie(raw map[string]interface{}) Cookie {
	var cookie Cookie

	var ok bool
	cookie.Name, _ = raw["name"].(string)
	cookie.Value, _ = raw["value"].(string)
	cookie.Domain, _ = raw["domain"].(string)
	cookie.Path, _ = raw["path"].(string)
	cookie.SameSite, _ = raw["sameSite"].(string)
	cookie.HTTPOnly, _ = raw["httpOnly"].(bool)
	cookie.Secure, _ = raw["secure"].(bool)

	var expiresFloat float64
	expiresFloat, ok = raw["expires"].(float64)
	if ok {
		cookie.Expires = expiresFloat
	}

	var sizeFloat float64
	sizeFloat, ok = raw["size"].(float64)
	if ok {
		cookie.Size = int(sizeFloat)
	}

	return cookie
}
