package kexas

import "github.com/jankylewis/kexas/kcore"

// Page is an alias for kcore.Page — they are the same type.
// Methods on Page (Navigate, Title, URL, Screenshot, Evaluate, Find, FindAll,
// FindStrict, FindByXPath, WaitForElement, WaitForElementVisible,
// WaitForElementClickable, WaitForLoad, WaitForLoadState, ScrollToTop,
// ScrollToBottom, ScrollBy, ScrollPosition, BringToFront, Close, IsClosed,
// SetContent, GetCookies, SetCookie, DeleteCookies, ClearCookies,
// LocalStorage, SessionStorage, StartRecording, ...) are inherited via the
// type alias and accessible as `kexas.Page.MethodName(...)`.
type Page = kcore.Page
