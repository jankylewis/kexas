// Package kexas provides a high-performance browser automation library for Chromium.
//
// The browser-engine implementation lives in the internal `kcore` sub-package.
// Files at the root are thin re-export shims (type aliases + forwarding wrappers)
// that preserve the public API path: `kexas.Browser`, `kexas.Launch()`, etc.
package kexas

import (
	"context"

	"github.com/jankylewis/kexas/kcore"
)

// Browser is an alias for kcore.Browser — they are the same type.
type Browser = kcore.Browser

// Launch starts a new browser instance and connects via CDP.
func Launch(opts *LaunchOptions) (*Browser, error) {
	return kcore.Launch(opts)
}

// LaunchWithContext starts a new browser with a custom context.
func LaunchWithContext(ctx context.Context, opts *LaunchOptions) (*Browser, error) {
	return kcore.LaunchWithContext(ctx, opts)
}
