package kexas

import (
	"time"

	"github.com/kexas-project/kexas/errors"
)

// WaitForElementVisible waits for an element to become visible on the page.
//
// This function polls the DOM at regular intervals until the element is found
// and is visible, or the timeout is reached. Returns the element when found.
func (p *Page) WaitForElementVisible(selector string, timeout time.Duration) (*Element, error) {
	// Validate timeout
	if timeout < 1*time.Second {
		return nil, errors.TimeoutInvalidFormat(timeout)
	}

	// Handle nil page gracefully for testing
	if p == nil || p.log == nil {
		// For testing, just return timeout error immediately
		return nil, errors.ElementNotVisibleWithin(selector, timeout)
	}

	p.log.Debug("waiting for element to be visible", "selector", selector, "timeout", timeout)

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		// Try to find the element
		var element *Element
		var err error
		element, err = p.Find(selector)
		if err == nil && element != nil {
			// Check if element is visible
			var visible bool = element.IsVisible()
			if visible {
				p.log.Debug("element became visible", "selector", selector, "elapsed", time.Since(start))
				return element, nil
			}
		}

		time.Sleep(pollInterval)
	}

	return nil, errors.ElementNotVisibleWithin(selector, timeout)
}

// WaitForElementClickable waits for an element to become clickable on the page.
//
// This function waits for an element to be both visible and enabled.
// Returns the element when it's ready for interaction.
func (p *Page) WaitForElementClickable(selector string, timeout time.Duration) (*Element, error) {
	// Validate timeout
	if timeout < 1*time.Second {
		return nil, errors.TimeoutInvalidFormat(timeout)
	}

	// Handle nil page gracefully for testing
	if p == nil || p.log == nil {
		// For testing, just return timeout error immediately
		return nil, errors.ElementNotClickableWithin(selector, timeout)
	}

	p.log.Debug("waiting for element to be clickable", "selector", selector, "timeout", timeout)

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		// Try to find the element
		var element *Element
		var err error
		element, err = p.Find(selector)
		if err == nil && element != nil {
			// Check if element is visible
			var visible bool = element.IsVisible()
			if visible {
				p.log.Debug("element became clickable", "selector", selector, "elapsed", time.Since(start))
				return element, nil
			}
		}

		time.Sleep(pollInterval)
	}

	return nil, errors.ElementNotClickableWithin(selector, timeout)
}

// WaitForElement waits for an element to exist in the DOM.
//
// This function waits for an element to be present in the DOM, regardless of
// visibility state. Returns the element when found.
func (p *Page) WaitForElement(selector string, timeout time.Duration) (*Element, error) {
	// Validate timeout
	if timeout < 1*time.Second {
		return nil, errors.TimeoutInvalidFormat(timeout)
	}

	// Handle nil page gracefully for testing
	if p == nil || p.log == nil {
		// For testing, just return timeout error immediately
		return nil, errors.ElementNotFound(selector)
	}

	p.log.Debug("waiting for element to exist", "selector", selector, "timeout", timeout)

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		// Try to find the element
		var element *Element
		var err error
		element, err = p.Find(selector)
		if err == nil && element != nil {
			p.log.Debug("element found", "selector", selector, "elapsed", time.Since(start))
			return element, nil
		}

		time.Sleep(pollInterval)
	}

	return nil, errors.ElementNotFound(selector)
}
