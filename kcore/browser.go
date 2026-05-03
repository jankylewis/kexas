// Package kcore is the kexas browser-automation engine implementation. Re-exported from package kexas via type aliases and forwarding wrappers.
package kcore

import (
	"context"
	"fmt"
	"sync"

	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/internal/logger"
	"github.com/jankylewis/kexas/klauncher"
)


// Browser represents a browser instance with CDP connection.
type Browser struct {
	process *klauncher.Browser
	client  *cdp.Client
	log     *logger.Logger
	ctx     context.Context
	pages   []*Page
	pagesMu sync.Mutex
}

// Launch starts a new browser instance and connects via CDP.
func Launch(opts *klauncher.Options) (*Browser, error) {
	var ctx context.Context = context.Background()
	return LaunchWithContext(ctx, opts)
}

// LaunchWithContext starts a new browser with a custom context.
func LaunchWithContext(ctx context.Context, opts *klauncher.Options) (*Browser, error) {
	var log *logger.Logger = logger.New("browser")

	// Launch browser process
	log.Debug("launching browser process")
	var process *klauncher.Browser
	var err error
	process, err = klauncher.Launch(ctx, opts)
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
