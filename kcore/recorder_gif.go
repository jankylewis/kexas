package kcore

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// SaveAnimatedGIF assembles captured frames into a single animated GIF using
// only the Go standard library — no system-level video tooling (ffmpeg etc.)
// required. The resulting file plays in any browser via an `<img>` tag, which
// is enough for CI/test-debug use cases where a real MP4 isn't worth a system
// dependency.
//
// outputPath should end in ".mp4" (the call site's intent); SaveAnimatedGIF
// rewrites the suffix to ".gif" before writing. Returns the actual file path.
//
// Tradeoffs vs. ffmpeg-encoded MP4:
//   - 256-color palette (Plan9 default) with Floyd-Steinberg dithering — some banding
//   - Larger file size for the same duration (no inter-frame delta compression)
//   - No browser seek bar (`<img>` autoplays the loop with no controls)
//   - Plays everywhere with zero install
func (r *Recorder) SaveAnimatedGIF(outputPath string) (string, error) {
	r.mu.Lock()
	var framesCopy []Frame = make([]Frame, len(r.frames))
	copy(framesCopy, r.frames)
	r.mu.Unlock()

	if len(framesCopy) == 0 {
		return "", fmt.Errorf("no frames to save as animated gif")
	}

	var g *gif.GIF = &gif.GIF{
		Image: make([]*image.Paletted, 0, len(framesCopy)),
		Delay: make([]int, 0, len(framesCopy)),
	}

	// 10fps matches the ffmpeg path (`-framerate 10`); 100ms / frame = 10 hundredths.
	const delayHundredths int = 10
	for i, frame := range framesCopy {
		var paletted *image.Paletted
		var convErr error
		paletted, convErr = decodeAndPaletteFrame(frame.Data)
		if convErr != nil {
			return "", fmt.Errorf("frame %d encode: %w", i, convErr)
		}
		g.Image = append(g.Image, paletted)
		g.Delay = append(g.Delay, delayHundredths)
	}

	var gifPath string = swapExtensionToGIF(outputPath)
	var err error = os.MkdirAll(filepath.Dir(gifPath), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create gif output dir: %w", err)
	}

	var f *os.File
	f, err = os.Create(gifPath)
	if err != nil {
		return "", fmt.Errorf("failed to create gif file: %w", err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return "", fmt.Errorf("gif.EncodeAll: %w", err)
	}

	r.log.Info("animated gif saved", "path", gifPath, "frames", len(framesCopy), "delay", delayHundredths)
	return gifPath, nil
}

// decodeAndPaletteFrame decodes a JPEG/PNG byte slice and quantises it to a
// 256-colour Plan9-palette image suitable for GIF encoding. Floyd-Steinberg
// dithering reduces the visible banding from palette quantisation.
//
// Detects format from magic bytes rather than trusting r.config.Format —
// kexas's Page.Screenshot always returns PNG regardless of the recorder's
// configured save-format, and frames sourced via SaveScreencastFrame can be
// either format depending on the CDP screencast configuration.
func decodeAndPaletteFrame(data []byte) (*image.Paletted, error) {
	var img image.Image
	var err error
	switch {
	case isPNG(data):
		img, err = png.Decode(bytes.NewReader(data))
	case isJPEG(data):
		img, err = jpeg.Decode(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unrecognised frame format (first bytes: % x)", firstBytes(data, 8))
	}
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	var bounds image.Rectangle = img.Bounds()
	var paletted *image.Paletted = image.NewPaletted(bounds, palette.Plan9)
	draw.FloydSteinberg.Draw(paletted, bounds, img, image.Point{})
	return paletted, nil
}

// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
func isPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A
}

// JPEG SOI marker: FF D8 FF
func isJPEG(data []byte) bool {
	if len(data) < 3 {
		return false
	}
	return data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
}

// firstBytes returns up to n leading bytes of data, for diagnostic logging.
func firstBytes(data []byte, n int) []byte {
	if len(data) < n {
		return data
	}
	return data[:n]
}

// swapExtensionToGIF returns the path with its extension replaced by ".gif".
// Strips any trailing ".mp4" so callers can pass the intended MP4 path and get
// the corresponding GIF path with the same stem.
func swapExtensionToGIF(p string) string {
	var trimmed string = strings.TrimSuffix(p, ".mp4")
	if !strings.HasSuffix(trimmed, ".gif") {
		trimmed = trimmed + ".gif"
	}
	return trimmed
}
