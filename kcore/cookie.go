package kcore

import (
	"fmt"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)


// Cookie represents a browser cookie with all standard fields.
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	Size     int     `json:"size"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

// SameSite constants for cookie configuration.
const (
	SameSiteStrict = "Strict"
	SameSiteLax    = "Lax"
	SameSiteNone   = "None"
)


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
	var params map[string]interface{} = buildSetCookieParams(cookie)
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdNetworkSetCookie, params)
	if err != nil {
		p.log.Error("failed to set cookie", "name", cookie.Name, "error", err)
		return fmt.Errorf("set cookie failed for '%s': %w", cookie.Name, err)
	}
	var success bool
	var ok bool
	success, ok = result["success"].(bool)
	if ok && !success {
		return fmt.Errorf("set cookie '%s': %w", cookie.Name, errors.ErrCookieSetFailed)
	}
	p.log.Info("cookie set", "name", cookie.Name)
	return nil
}

// buildSetCookieParams converts a Cookie struct into the CDP Network.setCookie
// params map, omitting empty/zero/false fields so CDP gets only what the
// caller intended.
func buildSetCookieParams(cookie Cookie) map[string]interface{} {
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
	if cookie.HTTPOnly {
		params["httpOnly"] = true
	}
	if cookie.Secure {
		params["secure"] = true
	}
	if cookie.SameSite != "" {
		params["sameSite"] = cookie.SameSite
	}
	return params
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
// with sensible defaults (HTTPOnly, Secure, SameSite=Lax, Path="/").
