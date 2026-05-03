//go:build integration

package kcore_test

import (
	"errors"
	"testing"

	"github.com/jankylewis/kexas"
	kexaserrors "github.com/jankylewis/kexas/errors"
)

// ============================================================
// Cookie Struct Tests (pure unit tests, no browser needed)
// ============================================================

func TestCookie_DefaultValues(t *testing.T) {
	var cookie kexas.Cookie

	if cookie.Name != "" {
		t.Error("default cookie name should be empty")
	}
	if cookie.Value != "" {
		t.Error("default cookie value should be empty")
	}
	if cookie.Domain != "" {
		t.Error("default cookie domain should be empty")
	}
	if cookie.Path != "" {
		t.Error("default cookie path should be empty")
	}
	if cookie.HTTPOnly != false {
		t.Error("default cookie httpOnly should be false")
	}
	if cookie.Secure != false {
		t.Error("default cookie secure should be false")
	}
	if cookie.SameSite != "" {
		t.Error("default cookie sameSite should be empty")
	}
	if cookie.Expires != 0 {
		t.Error("default cookie expires should be 0")
	}
	if cookie.Size != 0 {
		t.Error("default cookie size should be 0")
	}
}

func TestCookie_WithAllFields(t *testing.T) {
	var cookie kexas.Cookie = kexas.Cookie{
		Name:     "session_id",
		Value:    "abc123",
		Domain:   ".example.com",
		Path:     "/",
		Expires:  1893456000,
		Size:     42,
		HTTPOnly: true,
		Secure:   true,
		SameSite: kexas.SameSiteLax,
	}

	if cookie.Name != "session_id" {
		t.Errorf("expected name 'session_id', got '%s'", cookie.Name)
	}
	if cookie.Value != "abc123" {
		t.Errorf("expected value 'abc123', got '%s'", cookie.Value)
	}
	if cookie.Domain != ".example.com" {
		t.Errorf("expected domain '.example.com', got '%s'", cookie.Domain)
	}
	if cookie.Path != "/" {
		t.Errorf("expected path '/', got '%s'", cookie.Path)
	}
	if cookie.HTTPOnly != true {
		t.Error("expected httpOnly true")
	}
	if cookie.Secure != true {
		t.Error("expected secure true")
	}
	if cookie.SameSite != "Lax" {
		t.Errorf("expected sameSite 'Lax', got '%s'", cookie.SameSite)
	}
}

func TestCookie_SameSiteConstants(t *testing.T) {
	if kexas.SameSiteStrict != "Strict" {
		t.Errorf("expected SameSiteStrict 'Strict', got '%s'", kexas.SameSiteStrict)
	}
	if kexas.SameSiteLax != "Lax" {
		t.Errorf("expected SameSiteLax 'Lax', got '%s'", kexas.SameSiteLax)
	}
	if kexas.SameSiteNone != "None" {
		t.Errorf("expected SameSiteNone 'None', got '%s'", kexas.SameSiteNone)
	}
}

// ============================================================
// Cookie Error Sentinel Tests
// ============================================================

func TestCookieErrors_SentinelsDefined(t *testing.T) {
	if kexaserrors.ErrCookieNameEmpty == nil {
		t.Error("ErrCookieNameEmpty should not be nil")
	}
	if kexaserrors.ErrCookieValueEmpty == nil {
		t.Error("ErrCookieValueEmpty should not be nil")
	}
	if kexaserrors.ErrCookieSetFailed == nil {
		t.Error("ErrCookieSetFailed should not be nil")
	}
	if kexaserrors.ErrCookieGetFailed == nil {
		t.Error("ErrCookieGetFailed should not be nil")
	}
	if kexaserrors.ErrCookieDeleteFailed == nil {
		t.Error("ErrCookieDeleteFailed should not be nil")
	}
	if kexaserrors.ErrCookieNotFound == nil {
		t.Error("ErrCookieNotFound should not be nil")
	}
}

func TestCookieErrors_Messages(t *testing.T) {
	if kexaserrors.ErrCookieNameEmpty.Error() != "cookie name cannot be empty" {
		t.Errorf("unexpected error message: %s", kexaserrors.ErrCookieNameEmpty.Error())
	}
	if kexaserrors.ErrCookieNotFound.Error() != "cookie not found" {
		t.Errorf("unexpected error message: %s", kexaserrors.ErrCookieNotFound.Error())
	}
}

func TestCookieErrors_IsComparison(t *testing.T) {
	// Ensure sentinel errors work with errors.Is
	var wrapped error = errors.New("cookie name cannot be empty")
	if wrapped == kexaserrors.ErrCookieNameEmpty {
		t.Error("different error instances should not be equal via ==")
	}
	if !errors.Is(kexaserrors.ErrCookieNameEmpty, kexaserrors.ErrCookieNameEmpty) {
		t.Error("errors.Is should match same sentinel")
	}
}

// ============================================================
// Cookie Integration Tests (require real browser)
// ============================================================

func TestPage_SetCookie_EmptyName(t *testing.T) {
	t.Skip("requires real browser - integration test")

	// var browser *kexas.Browser
	// browser, _ = kexas.Launch(nil)
	// defer browser.Close()
	// var page *kexas.Page
	// page, _ = browser.FirstPage()
	//
	// var cookie kexas.Cookie = kexas.Cookie{Name: "", Value: "test"}
	// var err error = page.SetCookie(cookie)
	// if !errors.Is(err, kexaserrors.ErrCookieNameEmpty) {
	// 	t.Errorf("expected ErrCookieNameEmpty, got %v", err)
	// }
}

func TestPage_SetCookie_ValidCookie(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_SetCookie_WithDomain(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_SetCookies_MultipleCookies(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookies_Empty(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookies_AfterSet(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookies_WithURLFilter(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_DeleteCookie_EmptyName(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_DeleteCookie_Existing(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_DeleteCookie_NonExisting(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_ClearCookies(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_SetAuthCookie_EmptyName(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_SetAuthCookie_Defaults(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_HasCookie_EmptyName(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_HasCookie_Exists(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_HasCookie_NotExists(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookieValue_EmptyName(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookieValue_Exists(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_GetCookieValue_NotExists(t *testing.T) {
	t.Skip("requires real browser - integration test")
}
