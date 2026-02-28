//go:build integration

package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/launcher"
)

// launchTestBrowser launches a headless browser and returns the browser and first page.
// Returns nil, nil if browser cannot be launched (test should be skipped).
func launchTestBrowser(t *testing.T) (*kexas.Browser, *kexas.Page) {
	t.Helper()

	var opts *launcher.Options = launcher.DefaultOptions()
	opts.Headless = true

	var browser *kexas.Browser
	var err error
	browser, err = kexas.Launch(opts)
	if err != nil {
		t.Skipf("skipping: requires real browser - %v", err)
		return nil, nil
	}

	var page *kexas.Page
	page, err = browser.FirstPage()
	if err != nil {
		browser.Close()
		t.Fatalf("failed to get page: %v", err)
		return nil, nil
	}

	return browser, page
}

// setPageContent uses CDP Page.setDocumentContent to set HTML directly,
// avoiding data: URI navigation timeout issues.
func setPageContent(t *testing.T, page *kexas.Page, html string) {
	t.Helper()

	var err error
	err = page.SetContent(html)
	if err != nil {
		t.Fatalf("failed to set page content: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
}

// TestPageFindCSS_MultipleMatches_ReturnsError verifies that Find() returns error
// when a CSS selector matches more than 1 element.
// This is a critical safety check — ambiguous selectors must fail loudly.
func TestPageFindCSS_MultipleMatches_ReturnsError(t *testing.T) {
	var browser *kexas.Browser
	var page *kexas.Page
	browser, page = launchTestBrowser(t)
	if browser == nil {
		return
	}
	defer browser.Close()
	defer page.Close()

	// Set page content with 3 elements having the same class
	setPageContent(t, page, `<html><body>
		<div class="item">First</div>
		<div class="item">Second</div>
		<div class="item">Third</div>
	</body></html>`)

	// Find with selector that matches 3 elements - should return error
	var element *kexas.Element
	var err error
	element, err = page.Find("div.item")

	if err == nil {
		t.Fatal("expected error for CSS selector matching 3 elements, got nil")
		return
	}

	if element != nil {
		t.Error("expected nil element when multiple matches found")
		return
	}

	// Verify error message indicates multiple matches
	if !strings.Contains(err.Error(), "matched") {
		t.Errorf("expected error about multiple matches, got: %s", err.Error())
		return
	}

	if !strings.Contains(err.Error(), "3") {
		t.Errorf("expected error to mention 3 matches, got: %s", err.Error())
		return
	}

	t.Logf("correctly rejected ambiguous CSS selector: %s", err.Error())
}

// TestPageFindXPath_MultipleMatches_ReturnsError verifies that FindByXPath() returns error
// when an XPath expression matches more than 1 element.
// This is a critical safety check — ambiguous selectors must fail loudly.
func TestPageFindXPath_MultipleMatches_ReturnsError(t *testing.T) {
	var browser *kexas.Browser
	var page *kexas.Page
	browser, page = launchTestBrowser(t)
	if browser == nil {
		return
	}
	defer browser.Close()
	defer page.Close()

	// Set page content with 2 buttons having the same type
	setPageContent(t, page, `<html><body>
		<button type="submit">Submit 1</button>
		<button type="submit">Submit 2</button>
	</body></html>`)

	// FindByXPath with expression that matches 2 elements - should return error
	var element *kexas.Element
	var err error
	element, err = page.FindByXPath("//button[@type='submit']")

	if err == nil {
		t.Fatal("expected error for XPath matching 2 elements, got nil")
		return
	}

	if element != nil {
		t.Error("expected nil element when multiple matches found")
		return
	}

	// Verify error message indicates multiple matches
	if !strings.Contains(err.Error(), "matched") {
		t.Errorf("expected error about multiple matches, got: %s", err.Error())
		return
	}

	if !strings.Contains(err.Error(), "2") {
		t.Errorf("expected error to mention 2 matches, got: %s", err.Error())
		return
	}

	t.Logf("correctly rejected ambiguous XPath: %s", err.Error())
}
