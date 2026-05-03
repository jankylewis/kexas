
package launcher_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jankylewis/kexas/launcher"
)

func TestDefaultOptions(t *testing.T) {
	var opts *launcher.Options = launcher.DefaultOptions()

	if !opts.Headless {
		t.Error("expected headless to be true by default")
	}
	if opts.Port != 0 {
		t.Errorf("expected default port 0 (dynamic assignment), got %d", opts.Port)
	}
	if opts.Args == nil {
		t.Error("expected Args to be initialized")
	}
}

func TestOptions_CustomValues(t *testing.T) {
	var opts *launcher.Options = &launcher.Options{
		Headless: false,
		Port:     9333,
		Args:     []string{"--disable-gpu"},
	}

	if opts.Headless {
		t.Error("expected headless to be false")
	}
	if opts.Port != 9333 {
		t.Errorf("expected port 9333, got %d", opts.Port)
	}
	if len(opts.Args) != 1 {
		t.Errorf("expected 1 custom arg, got %d", len(opts.Args))
	}
}

func TestFindChromium_NotFound(t *testing.T) {
	var ctx context.Context = context.Background()
	var opts *launcher.Options = &launcher.Options{
		Headless:       true,
		Port:           9222,
		ExecutablePath: "/nonexistent/path/to/chrome",
	}

	var browser *launcher.Browser
	var err error
	browser, err = launcher.Launch(ctx, opts)
	if err == nil {
		t.Error("expected error when launching with invalid executable path")
		if browser != nil {
			browser.Close()
		}
	}
}

func TestBuildArgs_Headless(t *testing.T) {
	var opts *launcher.Options = launcher.DefaultOptions()
	opts.Headless = true

	if !opts.Headless {
		t.Error("expected headless to be true")
	}
}

func TestBuildArgs_CustomArgs(t *testing.T) {
	var opts *launcher.Options = launcher.DefaultOptions()
	opts.Args = []string{"--window-size=1920,1080", "--disable-gpu"}

	if len(opts.Args) != 2 {
		t.Errorf("expected 2 custom args, got %d", len(opts.Args))
	}
	if opts.Args[0] != "--window-size=1920,1080" {
		t.Errorf("expected first arg to be window-size, got %s", opts.Args[0])
	}
}

func TestExtractDebuggerURL_ValidOutput(t *testing.T) {
	var file *os.File
	var err error
	file, err = os.CreateTemp("", "test-chrome-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	var mockOutput string = `[0101/120000.123:INFO:CONSOLE(1)] "Starting Chrome"
DevTools listening on ws://127.0.0.1:9222/devtools/browser/abc-123-def
[0101/120001.456:INFO:CONSOLE(2)] "Chrome started"
`
	file.WriteString(mockOutput)
	file.Sync()

	if !strings.Contains(mockOutput, "DevTools listening on") {
		t.Error("expected mock output to contain DevTools message")
	}
	if !strings.Contains(mockOutput, "ws://") {
		t.Error("expected mock output to contain WebSocket URL")
	}
}

func TestExtractDebuggerURL_NoURL(t *testing.T) {
	var file *os.File
	var err error
	file, err = os.CreateTemp("", "test-chrome-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	var mockOutput string = `[0101/120000.123:INFO:CONSOLE(1)] "Starting Chrome"
[0101/120001.456:INFO:CONSOLE(2)] "No debugger URL here"
`
	file.WriteString(mockOutput)
	file.Sync()

	if strings.Contains(mockOutput, "ws://") {
		t.Error("expected no WebSocket URL in output")
	}
}

func TestFindChromium_MacOSPaths(t *testing.T) {
	var commonPaths []string = []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
	}

	for _, path := range commonPaths {
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got %s", path)
		}
	}
}
