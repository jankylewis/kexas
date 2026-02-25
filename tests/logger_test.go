package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kexas-project/kexas/internal/logger"
)

func TestNew_DefaultValues(t *testing.T) {
	var log *logger.Logger = logger.New("test")

	if log.Component() != "test" {
		t.Errorf("expected component 'test', got '%s'", log.Component())
	}
	if log.MinLevel() != logger.LevelInfo {
		t.Errorf("expected default level Info, got %s", log.MinLevel())
	}
}

func TestLevelString_AllLevels(t *testing.T) {
	var tests []struct {
		level    logger.Level
		expected string
	} = []struct {
		level    logger.Level
		expected string
	}{
		{logger.LevelDebug, "DEBUG"},
		{logger.LevelInfo, "INFO"},
		{logger.LevelWarn, "WARN"},
		{logger.LevelError, "ERROR"},
		{logger.LevelSilent, "SILENT"},
		{logger.Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		var got string = tt.level.String()
		if got != tt.expected {
			t.Errorf("Level(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestInfo_OutputsCorrectFormat(t *testing.T) {
	var buf bytes.Buffer
	var log *logger.Logger = logger.New("cdp").WithOutput(&buf).WithColored(false)

	log.Info("connected successfully")

	var output string = buf.String()
	if !strings.Contains(output, "[INFO ]") {
		t.Errorf("expected '[INFO ]' in output, got: %s", output)
	}
	if !strings.Contains(output, "[cdp]") {
		t.Errorf("expected '[cdp]' in output, got: %s", output)
	}
	if !strings.Contains(output, "connected successfully") {
		t.Errorf("expected 'connected successfully' in output, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestDebug_FilteredWhenLevelIsInfo(t *testing.T) {
	var buf bytes.Buffer
	var log *logger.Logger = logger.New("test").WithOutput(&buf).WithColored(false)

	log.Debug("this should not appear")

	var output string = buf.String()
	if output != "" {
		t.Errorf("expected empty output for filtered debug, got: %s", output)
	}
}

func TestDebug_ShownWhenLevelIsDebug(t *testing.T) {
	var buf bytes.Buffer
	var log *logger.Logger = logger.New("test").
		WithOutput(&buf).
		WithColored(false).
		WithLevel(logger.LevelDebug)

	log.Debug("debug message visible")

	var output string = buf.String()
	if !strings.Contains(output, "debug message visible") {
		t.Errorf("expected debug message in output, got: %s", output)
	}
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected '[DEBUG]' in output, got: %s", output)
	}
}

func TestWithLevel_ChangesMinimumLevel(t *testing.T) {
	var log *logger.Logger = logger.New("test")
	var debugLog *logger.Logger = log.WithLevel(logger.LevelDebug)

	if log.MinLevel() != logger.LevelInfo {
		t.Error("original logger level should not change")
	}
	if debugLog.MinLevel() != logger.LevelDebug {
		t.Error("new logger should have debug level")
	}
}

func TestLogger_KeyValuePairs(t *testing.T) {
	var buf bytes.Buffer
	var log *logger.Logger = logger.New("page").WithOutput(&buf).WithColored(false)

	log.Info("navigated", "url", "https://example.com", "elapsed", "1.2s")

	var output string = buf.String()
	if !strings.Contains(output, "url=https://example.com") {
		t.Errorf("expected 'url=https://example.com' in output, got: %s", output)
	}
	if !strings.Contains(output, "elapsed=1.2s") {
		t.Errorf("expected 'elapsed=1.2s' in output, got: %s", output)
	}
}

func TestLogger_SilentLevel(t *testing.T) {
	var buf bytes.Buffer
	var log *logger.Logger = logger.New("test").
		WithOutput(&buf).
		WithColored(false).
		WithLevel(logger.LevelSilent)

	log.Debug("no")
	log.Info("no")
	log.Warn("no")
	log.Error("no")

	var output string = buf.String()
	if output != "" {
		t.Errorf("expected no output with Silent level, got: %s", output)
	}
}
