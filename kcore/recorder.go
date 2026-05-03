package kcore

import (
	"fmt"
	"sync"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/internal/logger"
)

// RecorderConfig holds configuration for video recording.
type RecorderConfig struct {
	Format         string // "jpeg" or "png"
	Quality        int    // 0-100, only for jpeg
	MaxWidth       int    // Max frame width
	MaxHeight      int    // Max frame height
	EveryNthFrame  int    // Capture every Nth frame (1 = all frames)
	OutputDir      string // Directory for output files
	OutputFilename string // Base filename (without extension)
}

// DefaultRecorderConfig returns sensible defaults for recording.
func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		Format:         "jpeg",
		Quality:        80,
		MaxWidth:       1280,
		MaxHeight:      720,
		EveryNthFrame:  1,
		OutputDir:      "./recordings",
		OutputFilename: "recording",
	}
}

// Frame represents a single captured frame.
type Frame struct {
	Data      []byte
	Timestamp time.Time
	Index     int
}

// Recorder captures screenshots from a page to assemble into video.
type Recorder struct {
	page      *Page
	config    RecorderConfig
	frames    []Frame
	recording bool
	stopCh    chan struct{}
	mu        sync.Mutex
	log       *logger.Logger
}

// StartRecording begins capturing screenshots from the page at regular intervals.
// Uses CDP Page.startScreencast for efficient frame capture.
func (p *Page) StartRecording(configs ...RecorderConfig) (*Recorder, error) {
	var config RecorderConfig
	if len(configs) > 0 {
		config = configs[0]
	} else {
		config = DefaultRecorderConfig()
	}

	var recorder *Recorder = &Recorder{
		page:   p,
		config: config,
		frames: make([]Frame, 0),
		stopCh: make(chan struct{}),
		log:    logger.New("recorder"),
	}

	var err error = recorder.start()
	if err != nil {
		return nil, err
	}

	return recorder, nil
}

// start begins the screencast capture.
func (r *Recorder) start() error {
	r.mu.Lock()
	if r.recording {
		r.mu.Unlock()
		return errors.ErrRecorderAlreadyStarted
	}
	r.recording = true
	r.mu.Unlock()

	r.log.Debug("starting recording", "format", r.config.Format, "quality", r.config.Quality)

	// Start CDP screencast
	var params map[string]interface{} = map[string]interface{}{
		"format":    r.config.Format,
		"quality":   r.config.Quality,
		"maxWidth":  r.config.MaxWidth,
		"maxHeight": r.config.MaxHeight,
	}
	if r.config.EveryNthFrame > 0 {
		params["everyNthFrame"] = r.config.EveryNthFrame
	}

	var err error
	_, err = r.page.sendCommand(cdp.CmdPageStartScreencast, params)
	if err != nil {
		r.mu.Lock()
		r.recording = false
		r.mu.Unlock()
		return fmt.Errorf("start screencast failed: %w", err)
	}

	// Start background frame collector using polling
	go r.collectFrames()

	r.log.Info("recording started")
	return nil
}

// collectFrames periodically captures screenshots as frames.
// This is a polling-based fallback since CDP event subscription requires
// more complex infrastructure. Each frame is a Page.captureScreenshot call.
func (r *Recorder) collectFrames() {
	var frameIndex int = 0
	var interval time.Duration = 100 * time.Millisecond // ~10 fps

	for {
		select {
		case <-r.stopCh:
			return
		default:
			var screenshot []byte
			var err error
			screenshot, err = r.page.Screenshot()
			if err != nil {
				r.log.Debug("frame capture failed", "error", err)
				time.Sleep(interval)
				continue
			}

			r.mu.Lock()
			r.frames = append(r.frames, Frame{
				Data:      screenshot,
				Timestamp: time.Now(),
				Index:     frameIndex,
			})
			frameIndex++
			r.mu.Unlock()

			time.Sleep(interval)
		}
	}
}

// Stop ends the recording session.
func (r *Recorder) Stop() error {
	r.mu.Lock()
	if !r.recording {
		r.mu.Unlock()
		return errors.ErrRecorderNotStarted
	}
	r.recording = false
	r.mu.Unlock()

	// Signal the frame collector to stop
	close(r.stopCh)

	// Stop CDP screencast
	_, _ = r.page.sendCommand(cdp.CmdPageStopScreencast, nil)

	r.log.Info("recording stopped", "frames", len(r.frames))
	return nil
}

// IsRecording returns whether the recorder is currently active.
func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

// FrameCount returns the number of frames captured so far.
func (r *Recorder) FrameCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.frames)
}

