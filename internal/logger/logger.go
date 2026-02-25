// Package logger provides structured, leveled logging for Kexas.
//
// The logger supports Debug, Info, Warn, and Error levels with
// colored terminal output, component tagging, and timestamps.
//
// Usage:
//
//	log := logger.New("cdp")
//	log.Debug("sending command", "method", "Page.navigate")
//	log.Info("navigated successfully", "elapsed", "1.2s")
//	log.Error("navigation failed", "err", err)
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	// LevelDebug is for verbose development-time messages.
	LevelDebug Level = iota
	// LevelInfo is for general operational messages.
	LevelInfo
	// LevelWarn is for potentially harmful situations.
	LevelWarn
	// LevelError is for error conditions.
	LevelError
	// LevelSilent disables all logging.
	LevelSilent
)

// levelNames maps Level values to their string representations.
var levelNames map[Level]string = map[Level]string{
	LevelDebug:  "DEBUG",
	LevelInfo:   "INFO",
	LevelWarn:   "WARN",
	LevelError:  "ERROR",
	LevelSilent: "SILENT",
}

// String returns the human-readable name of the log level.
func (l Level) String() string {
	var name string
	var ok bool
	if name, ok = levelNames[l]; !ok {
		return "UNKNOWN"
	}
	return name
}

// ANSI color codes for terminal output.
const (
	colorReset  string = "\033[0m"
	colorCyan   string = "\033[36m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	colorRed    string = "\033[31m"
	colorGray   string = "\033[90m"
	colorBold   string = "\033[1m"
)

// levelColors maps Level values to their ANSI color codes.
var levelColors map[Level]string = map[Level]string{
	LevelDebug: colorCyan,
	LevelInfo:  colorGreen,
	LevelWarn:  colorYellow,
	LevelError: colorRed,
}

// Logger provides structured, leveled logging with component tagging.
type Logger struct {
	component string
	level     Level
	output    io.Writer
	colored   bool
	mu        sync.Mutex
}

// New creates a new Logger with the given component name.
// Defaults: level=LevelInfo, output=os.Stderr, colored=true.
func New(component string) *Logger {
	var l *Logger = &Logger{
		component: component,
		level:     LevelInfo,
		output:    os.Stderr,
		colored:   true,
	}
	return l
}

// WithLevel returns a new Logger with the specified minimum log level.
func (l *Logger) WithLevel(level Level) *Logger {
	var clone *Logger = l.clone()
	clone.level = level
	return clone
}

// WithOutput returns a new Logger that writes to the given writer.
func (l *Logger) WithOutput(w io.Writer) *Logger {
	var clone *Logger = l.clone()
	clone.output = w
	return clone
}

// WithColored returns a new Logger with colored output enabled or disabled.
func (l *Logger) WithColored(colored bool) *Logger {
	var clone *Logger = l.clone()
	clone.colored = colored
	return clone
}

// Component returns the logger's component name.
func (l *Logger) Component() string {
	return l.component
}

// Level returns the logger's minimum log level.
func (l *Logger) MinLevel() Level {
	return l.level
}

// Debug logs a message at Debug level with optional key-value pairs.
func (l *Logger) Debug(msg string, keyvals ...any) {
	l.log(LevelDebug, msg, keyvals...)
}

// Info logs a message at Info level with optional key-value pairs.
func (l *Logger) Info(msg string, keyvals ...any) {
	l.log(LevelInfo, msg, keyvals...)
}

// Warn logs a message at Warn level with optional key-value pairs.
func (l *Logger) Warn(msg string, keyvals ...any) {
	l.log(LevelWarn, msg, keyvals...)
}

// Error logs a message at Error level with optional key-value pairs.
func (l *Logger) Error(msg string, keyvals ...any) {
	l.log(LevelError, msg, keyvals...)
}

// log is the internal method that formats and writes the log line.
func (l *Logger) log(level Level, msg string, keyvals ...any) {
	if level < l.level {
		return
	}

	var timestamp string = time.Now().Format(time.RFC3339)
	var levelStr string = padRight(level.String(), 5)
	var kvStr string = formatKeyValues(keyvals...)

	var line string
	if l.colored {
		var color string = levelColors[level]
		line = fmt.Sprintf(
			"%s%s%s %s[%s]%s %s[%s]%s %s%s\n",
			colorGray, timestamp, colorReset,
			color, levelStr, colorReset,
			colorBold, l.component, colorReset,
			msg, kvStr,
		)
	} else {
		line = fmt.Sprintf(
			"%s [%s] [%s] %s%s\n",
			timestamp, levelStr, l.component, msg, kvStr,
		)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprint(l.output, line)
}

// clone creates a shallow copy of the logger.
func (l *Logger) clone() *Logger {
	return &Logger{
		component: l.component,
		level:     l.level,
		output:    l.output,
		colored:   l.colored,
	}
}

// formatKeyValues formats key-value pairs into a string.
// Pairs are formatted as " key=value key2=value2".
// If an odd number of args is given, the last value is set to "MISSING".
func formatKeyValues(keyvals ...any) string {
	if len(keyvals) == 0 {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(keyvals); i += 2 {
		b.WriteString(" ")
		var key string = fmt.Sprintf("%v", keyvals[i])
		if i+1 < len(keyvals) {
			b.WriteString(fmt.Sprintf("%s=%v", key, keyvals[i+1]))
		} else {
			b.WriteString(fmt.Sprintf("%s=MISSING", key))
		}
	}
	return b.String()
}

// padRight pads a string with spaces to the given width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
