package kexas

import "github.com/jankylewis/kexas/kcore"

// Cookie is an alias for kcore.Cookie — they are the same type.
type Cookie = kcore.Cookie

// SameSite cookie attribute values.
const (
	SameSiteStrict = kcore.SameSiteStrict
	SameSiteLax    = kcore.SameSiteLax
	SameSiteNone   = kcore.SameSiteNone
)
