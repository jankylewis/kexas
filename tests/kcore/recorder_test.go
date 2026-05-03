//go:build integration

package kcore_test

import (
	"testing"

	"github.com/jankylewis/kexas"
	kexaserrors "github.com/jankylewis/kexas/errors"
)

// ============================================================
// Recorder Config Tests (pure unit tests, no browser needed)
// ============================================================

func TestRecorderConfig_Defaults(t *testing.T) {
	var config kexas.RecorderConfig = kexas.DefaultRecorderConfig()

	if config.Format != "jpeg" {
		t.Errorf("expected format 'jpeg', got '%s'", config.Format)
	}
	if config.Quality != 80 {
		t.Errorf("expected quality 80, got %d", config.Quality)
	}
	if config.MaxWidth != 1280 {
		t.Errorf("expected maxWidth 1280, got %d", config.MaxWidth)
	}
	if config.MaxHeight != 720 {
		t.Errorf("expected maxHeight 720, got %d", config.MaxHeight)
	}
	if config.EveryNthFrame != 1 {
		t.Errorf("expected everyNthFrame 1, got %d", config.EveryNthFrame)
	}
	if config.OutputDir != "./recordings" {
		t.Errorf("expected outputDir './recordings', got '%s'", config.OutputDir)
	}
	if config.OutputFilename != "recording" {
		t.Errorf("expected outputFilename 'recording', got '%s'", config.OutputFilename)
	}
}

func TestRecorderConfig_CustomValues(t *testing.T) {
	var config kexas.RecorderConfig = kexas.RecorderConfig{
		Format:         "png",
		Quality:        100,
		MaxWidth:       1920,
		MaxHeight:      1080,
		EveryNthFrame:  2,
		OutputDir:      "/tmp/videos",
		OutputFilename: "test_video",
	}

	if config.Format != "png" {
		t.Errorf("expected format 'png', got '%s'", config.Format)
	}
	if config.Quality != 100 {
		t.Errorf("expected quality 100, got %d", config.Quality)
	}
	if config.MaxWidth != 1920 {
		t.Errorf("expected maxWidth 1920, got %d", config.MaxWidth)
	}
	if config.MaxHeight != 1080 {
		t.Errorf("expected maxHeight 1080, got %d", config.MaxHeight)
	}
	if config.EveryNthFrame != 2 {
		t.Errorf("expected everyNthFrame 2, got %d", config.EveryNthFrame)
	}
}

// ============================================================
// Recorder Error Sentinel Tests
// ============================================================

func TestRecorderErrors_SentinelsDefined(t *testing.T) {
	if kexaserrors.ErrRecorderNotStarted == nil {
		t.Error("ErrRecorderNotStarted should not be nil")
	}
	if kexaserrors.ErrRecorderAlreadyStarted == nil {
		t.Error("ErrRecorderAlreadyStarted should not be nil")
	}
	if kexaserrors.ErrFfmpegNotFound == nil {
		t.Error("ErrFfmpegNotFound should not be nil")
	}
}

func TestRecorderErrors_Messages(t *testing.T) {
	if kexaserrors.ErrRecorderNotStarted.Error() != "recorder not started" {
		t.Errorf("unexpected error message: %s", kexaserrors.ErrRecorderNotStarted.Error())
	}
	if kexaserrors.ErrRecorderAlreadyStarted.Error() != "recorder already started" {
		t.Errorf("unexpected error message: %s", kexaserrors.ErrRecorderAlreadyStarted.Error())
	}
	if kexaserrors.ErrFfmpegNotFound.Error() != "ffmpeg not found" {
		t.Errorf("unexpected error message: %s", kexaserrors.ErrFfmpegNotFound.Error())
	}
}

// ============================================================
// Frame Tests (pure unit tests)
// ============================================================

func TestFrame_DefaultValues(t *testing.T) {
	var frame kexas.Frame

	if frame.Data != nil {
		t.Error("default frame data should be nil")
	}
	if frame.Index != 0 {
		t.Error("default frame index should be 0")
	}
	if !frame.Timestamp.IsZero() {
		t.Error("default frame timestamp should be zero")
	}
}

func TestFrame_WithData(t *testing.T) {
	var data []byte = []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	var frame kexas.Frame = kexas.Frame{
		Data:  data,
		Index: 5,
	}

	if len(frame.Data) != 4 {
		t.Errorf("expected data length 4, got %d", len(frame.Data))
	}
	if frame.Index != 5 {
		t.Errorf("expected index 5, got %d", frame.Index)
	}
}

// ============================================================
// Recorder Integration Tests (require real browser)
// ============================================================

func TestPage_StartRecording_DefaultConfig(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestPage_StartRecording_CustomConfig(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_Stop_BeforeStart(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_Stop_AfterStart(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_DoubleStart(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_DoubleStop_Idempotent(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_IsRecording_BeforeStart(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_IsRecording_AfterStart(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_IsRecording_AfterStop(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_FrameCount(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_SaveFrames_NoFrames(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_SaveFrames_CreatesFiles(t *testing.T) {
	t.Skip("requires real browser - integration test")
}

func TestRecorder_SaveVideo_WithFfmpeg(t *testing.T) {
	t.Skip("requires real browser + ffmpeg - integration test")
}

func TestRecorder_SaveVideo_WithoutFfmpeg(t *testing.T) {
	t.Skip("requires real browser - integration test")
}
