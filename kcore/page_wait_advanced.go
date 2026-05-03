package kcore

import (
	"fmt"
	"strings"
	"time"

	"github.com/jankylewis/kexas/internal/cdp"
)

// Advanced wait strategies — companions to WaitForElement{,Visible,Clickable}
// and WaitForURLContains. Mirror Playwright (locator state: 'hidden' /
// 'detached'), Selenium (invisibilityOfElementLocated, titleContains,
// textToBePresentInElement, elementToBeAttribute), Cypress (should chains),
// and Rod (WaitInvisible / WaitStable).

// WaitForElementHidden polls until the selector matches an element that is
// not visible — display:none, visibility:hidden, opacity:0, zero-size, or
// removed from the DOM (latter case overlaps with WaitForElementDetached).
//
// Returns nil when the element becomes hidden, error on timeout.
func (p *Page) WaitForElementHidden(selector string, timeout time.Duration) error {
	if timeout < 100*time.Millisecond {
		return fmt.Errorf("kexas: timeout must be at least 100ms, got %v", timeout)
	}
	if p == nil {
		return fmt.Errorf("kexas: WaitForElementHidden called on nil page")
	}
	var jsExpr string = fmt.Sprintf(`(function() {
		const el = document.querySelector(%q);
		if (!el) return true;
		const cs = window.getComputedStyle(el);
		const rect = el.getBoundingClientRect();
		return cs.display === 'none' || cs.visibility === 'hidden' || cs.opacity === '0' || (rect.width === 0 && rect.height === 0);
	})()`, selector)
	return p.pollUntilJSTrue(jsExpr, timeout, fmt.Sprintf("element %s not hidden within %v", selector, timeout))
}

// WaitForElementDetached polls until the selector matches no element — the
// strict "removed from DOM" case. Distinct from WaitForElementHidden because
// a still-present-but-hidden element passes Hidden but fails Detached.
func (p *Page) WaitForElementDetached(selector string, timeout time.Duration) error {
	if timeout < 100*time.Millisecond {
		return fmt.Errorf("kexas: timeout must be at least 100ms, got %v", timeout)
	}
	if p == nil {
		return fmt.Errorf("kexas: WaitForElementDetached called on nil page")
	}
	var jsExpr string = fmt.Sprintf(`document.querySelector(%q) === null`, selector)
	return p.pollUntilJSTrue(jsExpr, timeout, fmt.Sprintf("element %s still in DOM within %v", selector, timeout))
}

// WaitForFunction polls a JavaScript expression until it returns a truthy
// value or timeout. The most general escape hatch — use when no built-in
// state matcher fits. Mirrors Playwright's `page.waitForFunction()`.
//
// Example: `WaitForFunction("window.myAppReady === true", 10*time.Second)`.
func (p *Page) WaitForFunction(jsExpression string, timeout time.Duration) error {
	if timeout < 100*time.Millisecond {
		return fmt.Errorf("kexas: timeout must be at least 100ms, got %v", timeout)
	}
	if p == nil {
		return fmt.Errorf("kexas: WaitForFunction called on nil page")
	}
	return p.pollUntilJSTrue(jsExpression, timeout, fmt.Sprintf("function did not become truthy within %v", timeout))
}

// WaitForTitleContains polls document.title until it contains substr.
// Mirrors Selenium's `ExpectedConditions.titleContains`.
func (p *Page) WaitForTitleContains(substr string, timeout time.Duration) error {
	if timeout < 100*time.Millisecond {
		return fmt.Errorf("kexas: timeout must be at least 100ms, got %v", timeout)
	}
	if p == nil {
		return fmt.Errorf("kexas: WaitForTitleContains called on nil page")
	}
	var start time.Time = time.Now()
	for time.Since(start) < timeout {
		var title string
		var err error
		title, err = p.Title()
		if err == nil && strings.Contains(title, substr) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("kexas: title did not contain %q within %v", substr, timeout)
}

// WaitForElementText polls until the selector's element exists AND its text
// content contains substr. Mirrors Selenium's
// `ExpectedConditions.textToBePresentInElement` + Playwright's
// `expect(locator).toContainText(...)` retry semantics.
func (p *Page) WaitForElementText(selector, substr string, timeout time.Duration) error {
	if timeout < 100*time.Millisecond {
		return fmt.Errorf("kexas: timeout must be at least 100ms, got %v", timeout)
	}
	if p == nil {
		return fmt.Errorf("kexas: WaitForElementText called on nil page")
	}
	var jsExpr string = fmt.Sprintf(`(function() {
		const el = document.querySelector(%q);
		if (!el) return false;
		return (el.textContent || '').indexOf(%q) !== -1;
	})()`, selector, substr)
	return p.pollUntilJSTrue(jsExpr, timeout, fmt.Sprintf("element %s text never contained %q within %v", selector, substr, timeout))
}

// pollUntilJSTrue is the shared poll body for the advanced waiters.
// Evaluates jsExpr at 100ms intervals, returns nil on truthy, formattedErr
// on timeout.
func (p *Page) pollUntilJSTrue(jsExpr string, timeout time.Duration, timeoutMsg string) error {
	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond
	for time.Since(start) < timeout {
		var result map[string]interface{}
		var err error
		result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
			"expression":    jsExpr,
			"returnByValue": true,
		})
		if err == nil && jsResultIsTruthy(result) {
			return nil
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("kexas: %s", timeoutMsg)
}

// jsResultIsTruthy unwraps a Runtime.evaluate response and returns true when
// the result represents JS truthy (true, non-zero number, non-empty string,
// non-null object). Used by every pollUntilJSTrue caller.
func jsResultIsTruthy(result map[string]interface{}) bool {
	resultObj, ok := result["result"].(map[string]interface{})
	if !ok {
		return false
	}
	value := resultObj["value"]
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return value != nil
	}
}
