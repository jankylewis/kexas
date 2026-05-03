package kexas

import "github.com/jankylewis/kexas/kcore"

// Recorder is an alias for kcore.Recorder — they are the same type.
type Recorder = kcore.Recorder

// RecorderConfig is an alias for kcore.RecorderConfig.
type RecorderConfig = kcore.RecorderConfig

// Frame is an alias for kcore.Frame — a single captured frame.
type Frame = kcore.Frame

// DefaultRecorderConfig returns sensible defaults for video recording.
func DefaultRecorderConfig() RecorderConfig {
	return kcore.DefaultRecorderConfig()
}
