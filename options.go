package kexas

import "github.com/jankylewis/kexas/launcher"

// LaunchOptions configures browser launch behavior.
type LaunchOptions = launcher.Options

// DefaultLaunchOptions returns sensible default launch options.
func DefaultLaunchOptions() *LaunchOptions {
	return launcher.DefaultOptions()
}
