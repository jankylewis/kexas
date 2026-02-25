// Package kexas provides a high-performance browser automation library for Chromium.
//
// Kexas communicates with Chromium via the Chrome DevTools Protocol (CDP)
// over WebSocket, offering blazing-fast browser control with Go's concurrency model.
//
// Basic usage:
//
//	browser, err := kexas.Launch()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer browser.Close()
//
//	page, err := browser.NewPage()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = page.Navigate("https://example.com")
package kexas

// Version is the current version of the Kexas library.
const Version string = "0.1.0-dev"
