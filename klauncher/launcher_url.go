package klauncher

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

// extractPortFromWSURL extracts the port number from a WebSocket URL.
// Example: "ws://127.0.0.1:51234/devtools/browser/abc" → 51234
func extractPortFromWSURL(wsURL string) int {
	var portPattern *regexp.Regexp = regexp.MustCompile(`:(\d+)/`)
	var match []string = portPattern.FindStringSubmatch(wsURL)
	if len(match) < 2 {
		return 0
	}
	var port int
	fmt.Sscanf(match[1], "%d", &port)
	return port
}

// pipeReadResult carries either the discovered DevTools URL or a read error
// out of the background scanner goroutine in extractDebuggerURLFromPipe.
type pipeReadResult struct {
	url string
	err error
}

// extractDebuggerURLFromPipe reads Chrome's stderr via a pipe in real-time.
// More reliable than file-based polling: no OS file-buffer lag, no seek/re-read
// overhead, no race between Chrome writing and our code reading.
func extractDebuggerURLFromPipe(pipe io.ReadCloser, timeout time.Duration, log *logger.Logger) (string, error) {
	var ch chan pipeReadResult = make(chan pipeReadResult, 1)
	go scanPipeForDebuggerURL(pipe, log, ch)
	select {
	case res := <-ch:
		return res.url, res.err
	case <-time.After(timeout):
		log.Error("timeout waiting for debugger URL from pipe")
		return "", ErrNoDebuggerURL
	}
}

// scanPipeForDebuggerURL drives the line-by-line scan, sending the first
// matching DevTools URL (or a fatal scan error) to ch.
func scanPipeForDebuggerURL(pipe io.ReadCloser, log *logger.Logger, ch chan<- pipeReadResult) {
	var scanner *bufio.Scanner = bufio.NewScanner(pipe)
	var pattern *regexp.Regexp = regexp.MustCompile(`ws://[^\s]+`)
	var lastLines []string
	for scanner.Scan() {
		var line string = scanner.Text()
		if len(lastLines) < 20 {
			lastLines = append(lastLines, line)
		}
		if res, done := classifyPipeLine(line, pattern, log); done {
			ch <- res
			return
		}
	}
	if scanner.Err() != nil {
		ch <- pipeReadResult{err: fmt.Errorf("pipe read error: %w", scanner.Err())}
		return
	}
	log.Error("chrome stderr closed without DevTools URL", "lastOutput", lastLines)
	ch <- pipeReadResult{err: ErrNoDebuggerURL}
}

// classifyPipeLine inspects one stderr line. Returns (result, done=true) when
// a terminal condition fires (URL found / port-bind failure); (_, false)
// otherwise. ERROR/FATAL lines are logged but don't stop the scan.
func classifyPipeLine(line string, pattern *regexp.Regexp, log *logger.Logger) (pipeReadResult, bool) {
	if strings.Contains(line, "DevTools listening on") {
		var match string = pattern.FindString(line)
		if match != "" {
			return pipeReadResult{url: match}, true
		}
	}
	if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
		log.Error("chrome error", "msg", line)
	}
	if strings.Contains(line, "bind") && strings.Contains(line, "address already in use") {
		return pipeReadResult{err: fmt.Errorf("chrome cannot bind to port - address already in use: %s", line)}, true
	}
	return pipeReadResult{}, false
}
