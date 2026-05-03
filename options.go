package kexas

import "github.com/jankylewis/kexas/klauncher"

// LaunchOptions configures browser launch behavior.
type LaunchOptions = klauncher.Options

// DefaultLaunchOptions returns sensible default launch options.
func DefaultLaunchOptions() *LaunchOptions {
	return klauncher.DefaultOptions()
}
