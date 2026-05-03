//go:build integration

package ktest_test

import (
	"os"
	"testing"

	"github.com/jankylewis/kexas/ktest"
)

// TestDefaultConfig tests the default configuration.
func TestDefaultConfig(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	if !config.Headless {
		t.Error("Expected headless to be true by default")
	}
	if config.Retries != 0 {
		t.Errorf("Expected retries to be 0, got %d", config.Retries)
	}
	if config.Parallel {
		t.Error("Expected parallel to be false by default")
	}
	if !config.ScreenshotOnFail {
		t.Error("Expected screenshot on fail to be true by default")
	}
	if config.ScreenshotDir != "./test-results/screenshots" {
		t.Errorf("Expected screenshot dir to be './test-results/screenshots', got %s", config.ScreenshotDir)
	}
	if config.ViewportWidth != 1280 {
		t.Errorf("Expected viewport width 1280, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 720 {
		t.Errorf("Expected viewport height 720, got %d", config.ViewportHeight)
	}
}

func TestLoadConfigFromFile_ViewportOverrides(t *testing.T) {
	var path string = writeTempConfig(t, `{"viewportWidth":1920,"viewportHeight":1080}`)
	var config *ktest.Config = ktest.LoadConfigFromFile(path)

	if config.ViewportWidth != 1920 {
		t.Errorf("expected viewport width 1920, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 1080 {
		t.Errorf("expected viewport height 1080, got %d", config.ViewportHeight)
	}
}

func TestLoadConfigFromFile_InvalidViewportFallsBack(t *testing.T) {
	var path string = writeTempConfig(t, `{"viewportWidth":-1,"viewportHeight":0}`)
	var config *ktest.Config = ktest.LoadConfigFromFile(path)

	if config.ViewportWidth != 1280 {
		t.Errorf("expected default viewport width 1280, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 720 {
		t.Errorf("expected default viewport height 720, got %d", config.ViewportHeight)
	}
}

// TestConfig_CustomValues tests custom configuration values.
func TestConfig_CustomValues(t *testing.T) {
	var config *ktest.Config = &ktest.Config{
		Headless:         false,
		Retries:          3,
		Parallel:         true,
		ScreenshotOnFail: false,
		ScreenshotDir:    "./custom-dir",
	}

	if config.Headless {
		t.Error("Expected headless to be false")
	}
	if config.Retries != 3 {
		t.Errorf("Expected retries to be 3, got %d", config.Retries)
	}
	if !config.Parallel {
		t.Error("Expected parallel to be true")
	}
	if config.ScreenshotOnFail {
		t.Error("Expected screenshot on fail to be false")
	}
}

// TestConfig_Timeout tests timeout configuration.
func TestConfig_Timeout(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	if config.Timeout == 0 {
		t.Error("Expected non-zero timeout")
	}
}

// TestConfig_BaseURL tests base URL configuration.
func TestConfig_BaseURL(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	config.BaseURL = "https://example.com"

	if config.BaseURL != "https://example.com" {
		t.Errorf("Expected base URL to be 'https://example.com', got %s", config.BaseURL)
	}
}

// TestLoadConfig tests configuration loading functionality.
func TestLoadConfig(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig() should not return nil")
	}
}

// TestConfigJSONStructure tests JSON configuration structure.
func TestConfigJSONStructure(t *testing.T) {
	// Test that config can be marshaled/unmarshaled
	var config *ktest.Config = ktest.DefaultConfig()

	// This would test JSON marshaling if we had that functionality
	// For now, just verify the structure is correct
	if config == nil {
		t.Error("Default config should not be nil")
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	var file *os.File
	var err error
	file, err = os.CreateTemp("", "kexas-config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err = file.WriteString(content); err != nil {
		file.Close()
		os.Remove(file.Name())
		t.Fatalf("failed to write temp config: %v", err)
	}
	file.Close()
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	return file.Name()
}
