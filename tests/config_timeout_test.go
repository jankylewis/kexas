
package tests

import (
	"testing"
	"time"

	"github.com/kexas-project/kexas/ktest"
)

func TestDefaultConfig_Timeout(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	// Test that default timeout is 20 seconds
	var expected time.Duration = 20 * time.Second
	if config.Timeout != expected {
		t.Errorf("expected timeout %v, got %v", expected, config.Timeout)
	}
}

func TestDefaultConfig_OtherFields(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	// Test other default values remain unchanged
	if !config.Headless {
		t.Error("expected Headless to be true")
	}

	if config.Retries != 0 {
		t.Error("expected Retries to be 0")
	}

	if config.Parallel {
		t.Error("expected Parallel to be false")
	}

	if !config.ScreenshotOnFail {
		t.Error("expected ScreenshotOnFail to be true")
	}

	var expectedScreenshotDir string = "./test-results/screenshots"
	if config.ScreenshotDir != expectedScreenshotDir {
		t.Errorf("expected ScreenshotDir %s, got %s", expectedScreenshotDir, config.ScreenshotDir)
	}

	var expectedVideoDir string = "./test-results/videos"
	if config.VideoDir != expectedVideoDir {
		t.Errorf("expected VideoDir %s, got %s", expectedVideoDir, config.VideoDir)
	}

	if config.SlowMo != 0 {
		t.Error("expected SlowMo to be 0")
	}

	if config.BaseURL != "" {
		t.Error("expected BaseURL to be empty")
	}

	if config.BrowserExecutable != "" {
		t.Error("expected BrowserExecutable to be empty")
	}
}

func TestDefaultConfig_Configuration(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	// Verify all fields are properly set
	if config == nil {
		t.Error("expected DefaultConfig to return non-nil config")
	}

	// Test that the config is a complete configuration object
	if config.Timeout == 0 {
		t.Error("expected Timeout to be set")
	}

	// Test that config contains all expected fields
	// This ensures the struct hasn't changed unexpectedly
	if config.ScreenshotDir == "" {
		t.Error("expected ScreenshotDir to be set")
	}

	if config.VideoDir == "" {
		t.Error("expected VideoDir to be set")
	}
}

func TestDefaultConfig_MultipleCalls(t *testing.T) {
	// Test that DefaultConfig returns consistent results
	var config1 *ktest.Config = ktest.DefaultConfig()
	var config2 *ktest.Config = ktest.DefaultConfig()

	if config1.Timeout != config2.Timeout {
		t.Error("expected DefaultConfig to return consistent timeout values")
	}

	if config1.Headless != config2.Headless {
		t.Error("expected DefaultConfig to return consistent Headless values")
	}

	if config1.ScreenshotOnFail != config2.ScreenshotOnFail {
		t.Error("expected DefaultConfig to return consistent ScreenshotOnFail values")
	}
}

func TestDefaultConfig_TimeoutValue(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	// Test that the timeout is exactly 20 seconds
	var expectedTimeout time.Duration = 20 * time.Second
	var actualTimeout time.Duration = config.Timeout

	if actualTimeout != expectedTimeout {
		t.Errorf("expected timeout to be exactly %v, got %v", expectedTimeout, actualTimeout)
	}

	// Test that it's not the old value (30 seconds)
	var oldTimeout time.Duration = 30 * time.Second
	if actualTimeout == oldTimeout {
		t.Error("expected timeout to be changed from 30s to 20s")
	}
}
