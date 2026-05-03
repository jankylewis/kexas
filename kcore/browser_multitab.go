package kcore

import (
	"fmt"
	"strings"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

// Pages returns all currently tracked pages.
func (b *Browser) Pages() []*Page {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	var result []*Page = make([]*Page, len(b.pages))
	copy(result, b.pages)
	return result
}

// PageCount returns the number of currently tracked pages.
func (b *Browser) PageCount() int {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()
	return len(b.pages)
}

// PageByIndex returns a page by its index in the tracked pages list.
func (b *Browser) PageByIndex(index int) (*Page, error) {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	if index < 0 || index >= len(b.pages) {
		return nil, fmt.Errorf("page index %d: %w", index, errors.ErrPageIndexOutOfRange)
	}

	return b.pages[index], nil
}

// PageByURL finds a page whose URL contains the given pattern (substring match).
func (b *Browser) PageByURL(urlPattern string) (*Page, error) {
	b.pagesMu.Lock()
	var pageCopy []*Page = make([]*Page, len(b.pages))
	copy(pageCopy, b.pages)
	b.pagesMu.Unlock()

	for i := 0; i < len(pageCopy); i++ {
		var pageURL string
		var err error
		pageURL, err = pageCopy[i].URL()
		if err != nil {
			continue
		}
		if strings.Contains(pageURL, urlPattern) {
			return pageCopy[i], nil
		}
	}

	return nil, fmt.Errorf("url pattern '%s': %w", urlPattern, errors.ErrNoPageMatchingURL)
}

// WaitForNewPage executes the given action and waits for a new page (tab) to appear.
// The action typically triggers a new tab (e.g., clicking a link with target="_blank").
// Returns the new page, or an error if no new page appears within the timeout.
func (b *Browser) WaitForNewPage(action func(), timeout time.Duration) (*Page, error) {
	b.log.Debug("waiting for new page", "timeout", timeout)

	b.pagesMu.Lock()
	var beforeCount int = len(b.pages)
	b.pagesMu.Unlock()

	action()

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		time.Sleep(pollInterval)

		var pageTargets []string
		var err error
		pageTargets, err = b.fetchPageTargetIDs()
		if err != nil {
			continue
		}
		if len(pageTargets) <= beforeCount {
			continue
		}

		var newPage *Page
		newPage, err = b.attachToFirstNewTarget(pageTargets)
		if err != nil {
			return nil, err
		}
		if newPage != nil {
			return newPage, nil
		}
	}

	return nil, fmt.Errorf("after %v: %w", timeout, errors.ErrWaitForNewPageTimeout)
}

// fetchPageTargetIDs returns the IDs of all targets of type "page" currently in the browser.
func (b *Browser) fetchPageTargetIDs() ([]string, error) {
	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, cdp.CmdTargetGetTargets, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var targetInfos []interface{}
	var ok bool
	targetInfos, ok = result["targetInfos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("targetInfos missing from CDP response")
	}

	var pageTargets []string
	for _, targetInfoRaw := range targetInfos {
		var targetInfo map[string]interface{}
		targetInfo, ok = targetInfoRaw.(map[string]interface{})
		if !ok {
			continue
		}
		var targetType string
		targetType, ok = targetInfo["type"].(string)
		if !ok || targetType != "page" {
			continue
		}
		var tid string
		tid, _ = targetInfo["targetId"].(string)
		pageTargets = append(pageTargets, tid)
	}
	return pageTargets, nil
}

// attachToFirstNewTarget attaches to the first target ID that is not already tracked
// by the browser. Returns (nil, nil) if every supplied ID is already known.
func (b *Browser) attachToFirstNewTarget(pageTargets []string) (*Page, error) {
	b.pagesMu.Lock()
	var knownTargets map[string]bool = make(map[string]bool)
	for _, p := range b.pages {
		knownTargets[p.targetID] = true
	}
	b.pagesMu.Unlock()

	for _, tid := range pageTargets {
		if knownTargets[tid] {
			continue
		}
		var page *Page
		var err error
		page, err = b.attachToPage(tid)
		if err != nil {
			return nil, fmt.Errorf("failed to attach to new page: %w", err)
		}

		b.pagesMu.Lock()
		b.pages = append(b.pages, page)
		b.pagesMu.Unlock()

		b.log.Info("new page detected", "targetId", tid)
		return page, nil
	}
	return nil, nil
}

// CloseAllPagesExcept closes all tracked pages except the specified one.
func (b *Browser) CloseAllPagesExcept(keep *Page) error {
	b.pagesMu.Lock()
	var pagesToClose []*Page = make([]*Page, 0)
	for _, p := range b.pages {
		if p != keep {
			pagesToClose = append(pagesToClose, p)
		}
	}
	b.pagesMu.Unlock()

	for _, p := range pagesToClose {
		var err error = p.Close()
		if err != nil {
			b.log.Error("failed to close page", "targetId", p.targetID, "error", err)
		}
	}

	// Update the pages list to only keep the one page
	b.pagesMu.Lock()
	if keep != nil {
		b.pages = []*Page{keep}
	} else {
		b.pages = nil
	}
	b.pagesMu.Unlock()

	return nil
}

// removePage removes a page from the tracked pages list.
func (b *Browser) removePage(page *Page) {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	for i := 0; i < len(b.pages); i++ {
		if b.pages[i] == page {
			b.pages = append(b.pages[:i], b.pages[i+1:]...)
			return
		}
	}
}
