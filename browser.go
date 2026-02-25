// Package kexas provides high-performance browser automation for Chromium.
package kexas

import (
	"context"
	"fmt"

	"github.com/kexas-project/kexas/internal/cdp"
	"github.com/kexas-project/kexas/internal/logger"
	"github.com/kexas-project/kexas/launcher"
)

// Browser represents a browser instance with CDP connection.
type Browser struct {
	process *launcher.Browser
	client  *cdp.Client
	log     *logger.Logger
	ctx     context.Context
}

// Launch starts a new browser instance and connects via CDP.
func Launch(opts *LaunchOptions) (*Browser, error) {
	var ctx context.Context = context.Background()
	return LaunchWithContext(ctx, opts)
}

// LaunchWithContext starts a new browser with a custom context.
func LaunchWithContext(ctx context.Context, opts *LaunchOptions) (*Browser, error) {
	var log *logger.Logger = logger.New("browser")

	// Launch browser process
	log.Debug("launching browser process")
	var process *launcher.Browser
	var err error
	process, err = launcher.Launch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("kexas: failed to launch browser: %w", err)
	}

	// Connect to browser via CDP
	var wsURL string = process.WebSocketURL()
	log.Debug("connecting to CDP", "url", wsURL)

	var client *cdp.Client
	client, err = cdp.Connect(ctx, wsURL)
	if err != nil {
		process.Close()
		return nil, fmt.Errorf("kexas: failed to connect to CDP: %w", err)
	}

	log.Info("browser launched and connected")

	var browser *Browser = &Browser{
		process: process,
		client:  client,
		log:     log,
		ctx:     ctx,
	}

	return browser, nil
}

// NewPage creates a new browser page (tab).
func (b *Browser) NewPage() (*Page, error) {
	b.log.Debug("creating new page")

	// Create new target (tab)
	var params map[string]interface{} = map[string]interface{}{
		"url": "about:blank",
	}

	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.createTarget", params)
	if err != nil {
		b.log.Error("failed to create target", "err", err)
		return nil, fmt.Errorf("kexas: failed to create page: %w", err)
	}

	var targetID string
	var ok bool
	targetID, ok = result["targetId"].(string)
	if !ok {
		return nil, fmt.Errorf("kexas: invalid targetId in response")
	}

	b.log.Debug("page created", "targetId", targetID)

	return b.attachToPage(targetID)
}

// FirstPage gets the first existing page (the default tab).
func (b *Browser) FirstPage() (*Page, error) {
	b.log.Debug("getting first page")

	// Get all targets
	var params map[string]interface{} = map[string]interface{}{}
	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.getTargets", params)
	if err != nil {
		b.log.Error("failed to get targets", "err", err)
		return nil, fmt.Errorf("kexas: failed to get targets: %w", err)
	}

	var targetInfos []interface{}
	var ok bool
	targetInfos, ok = result["targetInfos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("kexas: invalid targetInfos in response")
	}

	// Find the first page target
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

		var targetID string
		targetID, ok = targetInfo["targetId"].(string)
		if !ok {
			continue
		}

		b.log.Debug("found first page", "targetId", targetID)
		return b.attachToPage(targetID)
	}

	return nil, fmt.Errorf("kexas: no page targets found")
}

// attachToPage attaches to a page target and returns a Page object.
func (b *Browser) attachToPage(targetID string) (*Page, error) {
	// Attach to the target to get a session
	var params map[string]interface{} = map[string]interface{}{
		"targetId": targetID,
		"flatten":  true,
	}

	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.attachToTarget", params)
	if err != nil {
		b.log.Error("failed to attach to target", "err", err)
		return nil, fmt.Errorf("kexas: failed to attach to page: %w", err)
	}

	var sessionID string
	var ok bool
	sessionID, ok = result["sessionId"].(string)
	if !ok {
		return nil, fmt.Errorf("kexas: invalid sessionId in response")
	}

	b.log.Info("attached to page", "sessionId", sessionID, "targetId", targetID)

	var page *Page = &Page{
		browser:   b,
		targetID:  targetID,
		sessionID: sessionID,
		log:       logger.New("page"),
		ctx:       b.ctx,
	}

	return page, nil
}

// Close closes the browser and cleans up resources.
func (b *Browser) Close() error {
	b.log.Debug("closing browser")

	var err error = b.client.Close()
	if err != nil {
		b.log.Error("failed to close CDP client", "err", err)
	}

	err = b.process.Close()
	if err != nil {
		b.log.Error("failed to close browser process", "err", err)
		return err
	}

	b.log.Info("browser closed")
	return nil
}
