package kcore

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jankylewis/kexas/internal/cdp"
)

// Returns the output directory path.
func (r *Recorder) SaveFrames(outputDir string) (string, error) {
	r.mu.Lock()
	var framesCopy []Frame = make([]Frame, len(r.frames))
	copy(framesCopy, r.frames)
	r.mu.Unlock()

	if len(framesCopy) == 0 {
		return "", fmt.Errorf("no frames to save")
	}

	var err error = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	var ext string = r.frameExtension()
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
	var err error = r.stopIfRecording()
	if err != nil {
		return "", err
	}

	var frameCount int = r.snapshotFrameCount()
	if frameCount == 0 {
		return "", fmt.Errorf("no frames to save as video")
	}

	var ffmpegPath string
	var ensureErr error
	ffmpegPath, ensureErr = EnsureFFmpeg(r.log)
	if ensureErr != nil {
		// Network down, unsupported platform, or extract failed: fall back to
		// pure-Go animated GIF so the test still produces a viewable artifact.
		r.log.Info("ffmpeg unavailable, falling back to animated GIF", "error", ensureErr)
		return r.SaveAnimatedGIF(outputPath)
	}

	var ext string = r.detectActualFrameExt()
	var tempDir string = outputPath + "_tmp_frames"
	err = r.writeFramesToTempDir(tempDir, ext)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}

	var output []byte
	output, err = r.runFFmpegEncode(ffmpegPath, tempDir, outputPath, ext)
	if err != nil {
		r.log.Error("ffmpeg failed, falling back to animated GIF", "error", err, "output", string(output))
		return r.SaveAnimatedGIF(outputPath)
	}

	r.log.Info("video saved", "path", outputPath, "frames", frameCount)
	return outputPath, nil
}

// frameExtension returns the file extension for frames based on the recorder config.
func (r *Recorder) frameExtension() string {
	if r.config.Format == "jpeg" {
		return "jpg"
	}
	return "png"
}

// stopIfRecording stops the recorder if it is still active. No-op if already stopped.
func (r *Recorder) stopIfRecording() error {
	r.mu.Lock()
	var stillRecording bool = r.recording
	r.mu.Unlock()
	if !stillRecording {
		return nil
	}
	return r.Stop()
}

// snapshotFrameCount returns the current frame count under the recorder mutex.
func (r *Recorder) snapshotFrameCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.frames)
}

// writeFramesToTempDir creates tempDir and writes each captured frame into it
// using the provided file extension. ext should match the frames' actual byte
// format (use detectActualFrameExt) so ffmpeg can decode them.
func (r *Recorder) writeFramesToTempDir(tempDir, ext string) error {
	var err error = os.MkdirAll(tempDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < len(r.frames); i++ {
		var filename string = fmt.Sprintf("frame_%04d.%s", i, ext)
		var filePath string = filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, r.frames[i].Data, 0644)
	}
	return nil
}

// detectActualFrameExt sniffs the first frame's magic bytes to choose between
// "png" and "jpg". The recorder config nominates a format ("jpeg"/"png") but
// the actual bytes come from Page.Screenshot which always returns PNG —
// trusting config.Format would mislabel files and break ffmpeg's decoder.
func (r *Recorder) detectActualFrameExt() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.frames) == 0 {
		return r.frameExtension()
	}
	var data []byte = r.frames[0].Data
	if isPNG(data) {
		return "png"
	}
	if isJPEG(data) {
		return "jpg"
	}
	return r.frameExtension()
}

// runFFmpegEncode invokes ffmpeg to assemble frames in tempDir into an MP4 at outputPath.
// ext is the input frame extension (must match what writeFramesToTempDir used).
// Returns ffmpeg's combined stdout/stderr output for diagnostic logging on failure.
func (r *Recorder) runFFmpegEncode(ffmpegPath, tempDir, outputPath, ext string) ([]byte, error) {
	var inputPattern string = filepath.Join(tempDir, fmt.Sprintf("frame_%%04d.%s", ext))
	// pad=ceil(iw/2)*2:ceil(ih/2)*2 rounds up to even dimensions — required by
	// libx264 + yuv420p, which can't encode frames with odd width or height.
	// Browser viewports often have odd heights (e.g., 1280x633), so without
	// this filter x264 errors out with "incorrect parameters such as ... width
	// or height".
	var cmd *exec.Cmd = exec.Command(ffmpegPath,
		"-y",
		"-framerate", "10",
		"-i", inputPattern,
		"-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "fast",
		outputPath,
	)
	return cmd.CombinedOutput()
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
