package kexas

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
	"github.com/kexas-project/kexas/internal/logger"
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

// SaveFrames writes all captured frames as individual image files.
// Returns the output directory path.
func (r *Recorder) SaveFrames(outputDir string) (string, error) {
	r.mu.Lock()
	var framesCopy []Frame = make([]Frame, len(r.frames))
	copy(framesCopy, r.frames)
	r.mu.Unlock()

	if len(framesCopy) == 0 {
		return "", fmt.Errorf("no frames to save")
	}

	// Create output directory
	var err error = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	var ext string = "png"
	if r.config.Format == "jpeg" {
		ext = "jpg"
	}

	for i := 0; i < len(framesCopy); i++ {
		var filename string = fmt.Sprintf("frame_%04d.%s", i, ext)
		var filePath string = filepath.Join(outputDir, filename)

		err = os.WriteFile(filePath, framesCopy[i].Data, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write frame %d: %w", i, err)
		}
	}

	r.log.Info("frames saved", "count", len(framesCopy), "dir", outputDir)
	return outputDir, nil
}

// SaveVideo assembles captured frames into an MP4 video using ffmpeg.
// If ffmpeg is not available, falls back to saving individual frames.
// Returns the path to the output file.
func (r *Recorder) SaveVideo(outputPath string) (string, error) {
	// Stop recording if still active
	r.mu.Lock()
	var stillRecording bool = r.recording
	r.mu.Unlock()

	if stillRecording {
		var err error = r.Stop()
		if err != nil {
			return "", err
		}
	}

	r.mu.Lock()
	var frameCount int = len(r.frames)
	r.mu.Unlock()

	if frameCount == 0 {
		return "", fmt.Errorf("no frames to save as video")
	}

	// Check if ffmpeg is available
	var ffmpegPath string
	var lookErr error
	ffmpegPath, lookErr = exec.LookPath("ffmpeg")
	if lookErr != nil {
		r.log.Info("ffmpeg not found, falling back to saving frames")
		var framesDir string = outputPath + "_frames"
		return r.SaveFrames(framesDir)
	}

	// Create temporary directory for frames
	var tempDir string = outputPath + "_tmp_frames"
	var err error = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write frames to temp directory
	var ext string = "png"
	if r.config.Format == "jpeg" {
		ext = "jpg"
	}

	r.mu.Lock()
	for i := 0; i < len(r.frames); i++ {
		var filename string = fmt.Sprintf("frame_%04d.%s", i, ext)
		var filePath string = filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, r.frames[i].Data, 0644)
	}
	r.mu.Unlock()

	// Ensure output directory exists
	var outputDir string = filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	// Run ffmpeg to assemble video
	var inputPattern string = filepath.Join(tempDir, fmt.Sprintf("frame_%%04d.%s", ext))
	var cmd *exec.Cmd = exec.Command(ffmpegPath,
		"-y",
		"-framerate", "10",
		"-i", inputPattern,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "fast",
		outputPath,
	)

	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		r.log.Error("ffmpeg failed", "error", err, "output", string(output))
		// Fallback to saving frames
		var framesDir string = outputPath + "_frames"
		return r.SaveFrames(framesDir)
	}

	r.log.Info("video saved", "path", outputPath, "frames", frameCount)
	return outputPath, nil
}

// SaveScreencastFrame saves a single base64-encoded frame from CDP screencast event.
// This is used when CDP event subscription is available.
func (r *Recorder) SaveScreencastFrame(data string, sessionID int) error {
	var decoded []byte
	var err error
	decoded, err = base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode frame: %w", err)
	}

	r.mu.Lock()
	r.frames = append(r.frames, Frame{
		Data:      decoded,
		Timestamp: time.Now(),
		Index:     len(r.frames),
	})
	r.mu.Unlock()

	// Acknowledge the frame to Chrome
	_, _ = r.page.sendCommand(cdp.CmdPageScreencastFrameAck, map[string]interface{}{
		"sessionId": sessionID,
	})

	return nil
}
