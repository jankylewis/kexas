package ktest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds test execution configuration.
type Config struct {
	Headless          bool
	Timeout           time.Duration
	Retries           int
	Parallel          bool
	ParallelSet       int
	ScreenshotOnFail  bool
	ScreenshotDir     string
	VideoDir          string
	ReportDir         string
	SlowMo            time.Duration
	BaseURL           string
	BrowserExecutable string
	ViewportWidth     int
	ViewportHeight    int
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		Headless:         true,
		Timeout:          20 * time.Second,
		Retries:          0,
		Parallel:         false,
		ParallelSet:      1,
		ScreenshotOnFail: true,
		ScreenshotDir:    "./test-results/screenshots",
		VideoDir:         "./test-results/videos",
		ReportDir:        "./test-results",
		SlowMo:           0,
		BaseURL:          "",
		ViewportWidth:    1280,
		ViewportHeight:   720,
	}
}

// configJSON represents the JSON structure of kexas.config.json.
type configJSON struct {
	Headless          *bool   `json:"headless"`
	Timeout           *int    `json:"timeout"`
	Retries           *int    `json:"retries"`
	Parallel          *bool   `json:"parallel"`
	ParallelSet       *int    `json:"parallelSet"`
	ScreenshotOnFail  *bool   `json:"screenshotOnFail"`
	ScreenshotDir     *string `json:"screenshotDir"`
	VideoDir          *string `json:"videoDir"`
	ReportDir         *string `json:"reportDir"`
	SlowMo            *int    `json:"slowMo"`
	BaseURL           *string `json:"baseURL"`
	BrowserExecutable *string `json:"browserExecutable"`
	ViewportWidth     *int    `json:"viewportWidth"`
	ViewportHeight    *int    `json:"viewportHeight"`
}

// loadConfig walks upward from the current working directory looking for a
// kexas.config.json file (up to 10 parent levels). This lets `go test ./tests/auth/`
// find a config at the project root even when the test process's CWD is a deep
// sub-package. Falls back to DefaultConfig() when no config is found.
func loadConfig() *Config {
	var path string = findConfigUpward(".", "kexas.config.json", 10)
	if path == "" {
		return DefaultConfig()
	}
	return LoadConfigFromFile(path)
}

// findConfigUpward walks parent directories of startDir looking for filename.
// Returns the absolute path on first match, or "" if not found within maxLevels.
func findConfigUpward(startDir, filename string, maxLevels int) string {
	var current string
	var err error
	current, err = filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	for i := 0; i < maxLevels; i++ {
		var candidate string = filepath.Join(current, filename)
		_, err = os.Stat(candidate)
		if err == nil {
			return candidate
		}
		var parent string = filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
	return ""
}

// LoadConfigFromFile loads configuration from a specific file path.
// Returns DefaultConfig() if the file is missing or contains invalid JSON.
// Relative dir paths in the loaded config (ScreenshotDir, ReportDir, VideoDir)
// are anchored to the directory of the config file so subsequent path use is
// independent of the test process's CWD.
func LoadConfigFromFile(path string) *Config {
	var config *Config = DefaultConfig()

	var data []byte
	var err error
	data, err = os.ReadFile(path)
	if err != nil {
		return config
	}

	var jsonConfig configJSON
	err = json.Unmarshal(data, &jsonConfig)
	if err != nil {
		return config
	}

	applyJSONConfig(config, &jsonConfig)
	anchorConfigPaths(config, filepath.Dir(path))
	return config
}

// anchorConfigPaths converts relative ScreenshotDir/ReportDir/VideoDir entries
// to absolute paths rooted at baseDir. Already-absolute and empty paths pass through.
func anchorConfigPaths(config *Config, baseDir string) {
	config.ScreenshotDir = anchorOne(baseDir, config.ScreenshotDir)
	config.ReportDir = anchorOne(baseDir, config.ReportDir)
	config.VideoDir = anchorOne(baseDir, config.VideoDir)
}

func anchorOne(baseDir, dir string) string {
	if dir == "" || filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(baseDir, dir)
}

// applyJSONConfig overlays each non-nil field from jsonConfig onto config.
// Only fields explicitly present in the JSON file override the defaults.
func applyJSONConfig(config *Config, jsonConfig *configJSON) {
	if jsonConfig.Headless != nil {
		config.Headless = *jsonConfig.Headless
	}
	if jsonConfig.Timeout != nil {
		config.Timeout = time.Duration(*jsonConfig.Timeout) * time.Millisecond
	}
	if jsonConfig.Retries != nil {
		config.Retries = *jsonConfig.Retries
	}
	if jsonConfig.Parallel != nil {
		config.Parallel = *jsonConfig.Parallel
	}
	if jsonConfig.ParallelSet != nil && *jsonConfig.ParallelSet >= 1 {
		config.ParallelSet = *jsonConfig.ParallelSet
		if config.ParallelSet > 1 {
			config.Parallel = true
		}
	}
	if jsonConfig.ScreenshotOnFail != nil {
		config.ScreenshotOnFail = *jsonConfig.ScreenshotOnFail
	}
	if jsonConfig.ScreenshotDir != nil {
		config.ScreenshotDir = *jsonConfig.ScreenshotDir
	}
	if jsonConfig.VideoDir != nil {
		config.VideoDir = *jsonConfig.VideoDir
	}
	if jsonConfig.ReportDir != nil {
		config.ReportDir = *jsonConfig.ReportDir
	}
	if jsonConfig.SlowMo != nil {
		config.SlowMo = time.Duration(*jsonConfig.SlowMo) * time.Millisecond
	}
	if jsonConfig.BaseURL != nil {
		config.BaseURL = *jsonConfig.BaseURL
	}
	if jsonConfig.BrowserExecutable != nil {
		config.BrowserExecutable = *jsonConfig.BrowserExecutable
	}
	if jsonConfig.ViewportWidth != nil && *jsonConfig.ViewportWidth > 0 {
		config.ViewportWidth = *jsonConfig.ViewportWidth
	}
	if jsonConfig.ViewportHeight != nil && *jsonConfig.ViewportHeight > 0 {
		config.ViewportHeight = *jsonConfig.ViewportHeight
	}
}
