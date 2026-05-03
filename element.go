package kexas

import (
	"time"

	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/kcore"
)

// Element is an alias for kcore.Element — they are the same type.
// Methods (Click, WaitAndClick, Type, WaitAndType, Hover, WaitAndHover,
// GetText, GetAttribute, IsVisible, ScrollIntoView, Validate, ...) are
// inherited via the type alias.
type Element = kcore.Element

// NewElement creates a new Element bound to a Page (objectId resolved lazily).
func NewElement(page *Page, selector string, nodeID cdp.NodeID, timeout time.Duration) *Element {
	return kcore.NewElement(page, selector, nodeID, timeout)
}

// NewElementWithObject creates a new Element with an already-resolved CDP objectID.
func NewElementWithObject(page *Page, selector string, nodeID cdp.NodeID, objectID string, timeout time.Duration) *Element {
	return kcore.NewElementWithObject(page, selector, nodeID, objectID, timeout)
}
