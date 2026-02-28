package kexas

import (
	"fmt"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// Cookie represents a browser cookie with all standard fields.
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	Size     int     `json:"size"`
	HttpOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

// SameSite constants for cookie configuration.
const (
	SameSiteStrict = "Strict"
	SameSiteLax    = "Lax"
	SameSiteNone   = "None"
)

// GetCookies retrieves all cookies for the current page.
// If urls are provided, only cookies matching those URLs are returned.
func (p *Page) GetCookies(urls ...string) ([]Cookie, error) {
	p.log.Debug("getting cookies")

	var params map[string]interface{} = map[string]interface{}{}
	if len(urls) > 0 {
		params["urls"] = urls
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdNetworkGetCookies, params)
	if err != nil {
		p.log.Error("failed to get cookies", "error", err)
		return nil, fmt.Errorf("get cookies failed: %w", err)
	}

	var rawCookies []interface{}
	var ok bool
	rawCookies, ok = result["cookies"].([]interface{})
	if !ok {
		return []Cookie{}, nil
	}

	var cookies []Cookie = make([]Cookie, 0, len(rawCookies))
	for i := 0; i < len(rawCookies); i++ {
		var rawCookie map[string]interface{}
		rawCookie, ok = rawCookies[i].(map[string]interface{})
		if !ok {
			continue
		}
		var cookie Cookie = parseCookie(rawCookie)
		cookies = append(cookies, cookie)
	}

	p.log.Debug("cookies retrieved", "count", len(cookies))
	return cookies, nil
}

// SetCookie sets a single cookie on the page.
// Name and Value are required. Domain defaults to the current page's domain if empty.
func (p *Page) SetCookie(cookie Cookie) error {
	if cookie.Name == "" {
		return errors.ErrCookieNameEmpty
	}

	p.log.Debug("setting cookie", "name", cookie.Name)

	var params map[string]interface{} = map[string]interface{}{
		"name":  cookie.Name,
		"value": cookie.Value,
	}

	if cookie.Domain != "" {
		params["domain"] = cookie.Domain
	}
	if cookie.Path != "" {
		params["path"] = cookie.Path
	}
	if cookie.Expires > 0 {
		params["expires"] = cookie.Expires
	}
	if cookie.HttpOnly {
		params["httpOnly"] = true
	}
	if cookie.Secure {
		params["secure"] = true
	}
	if cookie.SameSite != "" {
		params["sameSite"] = cookie.SameSite
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdNetworkSetCookie, params)
	if err != nil {
		p.log.Error("failed to set cookie", "name", cookie.Name, "error", err)
		return fmt.Errorf("set cookie failed for '%s': %w", cookie.Name, err)
	}

	// Check if Chrome confirmed the cookie was set
	var success bool
	var ok bool
	success, ok = result["success"].(bool)
	if ok && !success {
		return fmt.Errorf("set cookie '%s': %w", cookie.Name, errors.ErrCookieSetFailed)
	}

	p.log.Info("cookie set", "name", cookie.Name)
	return nil
}

// SetCookies sets multiple cookies on the page.
func (p *Page) SetCookies(cookies []Cookie) error {
	for i := 0; i < len(cookies); i++ {
		var err error = p.SetCookie(cookies[i])
		if err != nil {
			return fmt.Errorf("set cookies failed at index %d: %w", i, err)
		}
	}
	return nil
}

// DeleteCookie deletes cookies matching the given name, domain, and path.
// Domain and path can be empty to match all cookies with the given name.
func (p *Page) DeleteCookie(name string, domain string, path string) error {
	if name == "" {
		return errors.ErrCookieNameEmpty
	}

	p.log.Debug("deleting cookie", "name", name)

	var params map[string]interface{} = map[string]interface{}{
		"name": name,
	}
	if domain != "" {
		params["domain"] = domain
	}
	if path != "" {
		params["path"] = path
	}

	var err error
	_, err = p.sendCommand(cdp.CmdNetworkDeleteCookies, params)
	if err != nil {
		p.log.Error("failed to delete cookie", "name", name, "error", err)
		return fmt.Errorf("delete cookie '%s' failed: %w", name, err)
	}

	p.log.Info("cookie deleted", "name", name)
	return nil
}

// ClearCookies deletes all cookies in the browser.
func (p *Page) ClearCookies() error {
	p.log.Debug("clearing all cookies")

	var err error
	_, err = p.sendCommand(cdp.CmdNetworkClearBrowserCookies, nil)
	if err != nil {
		p.log.Error("failed to clear cookies", "error", err)
		return fmt.Errorf("clear cookies failed: %w", err)
	}

	p.log.Info("all cookies cleared")
	return nil
}

// SetAuthCookie is a convenience method for setting an authentication cookie
// with sensible defaults (HttpOnly, Secure, SameSite=Lax, Path="/").
func (p *Page) SetAuthCookie(name string, value string, domain string) error {
	if name == "" {
		return errors.ErrCookieNameEmpty
	}

	var cookie Cookie = Cookie{
		Name:     name,
		Value:    value,
		Domain:   domain,
		Path:     "/",
		HttpOnly: true,
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
	cookie.HttpOnly, _ = raw["httpOnly"].(bool)
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
