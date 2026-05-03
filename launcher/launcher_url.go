package launcher

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

// extractDebuggerURLFromPipe reads Chrome's stderr via a pipe in real-time.
// This is significantly more reliable than file-based polling because:
//  1. No OS file buffer lag — lines arrive immediately via the pipe.
//  2. No seek/re-read overhead — each line is read exactly once.
//  3. No race between Chrome writing and our code reading.
func extractDebuggerURLFromPipe(pipe io.ReadCloser, timeout time.Duration, log *logger.Logger) (string, error) {
	type result struct {
		url string
		err error
	}
	var ch chan result = make(chan result, 1)

	go func() {
		var scanner *bufio.Scanner = bufio.NewScanner(pipe)
		var pattern *regexp.Regexp = regexp.MustCompile(`ws://[^\s]+`)
		var lastLines []string

		for scanner.Scan() {
			var line string = scanner.Text()

			if len(lastLines) < 20 {
				lastLines = append(lastLines, line)
			}

			if strings.Contains(line, "DevTools listening on") {
				var match string = pattern.FindString(line)
				if match != "" {
					ch <- result{url: match}
					return
				}
			}
			if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
				log.Error("chrome error", "msg", line)
			}
			if strings.Contains(line, "bind") && strings.Contains(line, "address already in use") {
				ch <- result{err: fmt.Errorf("chrome cannot bind to port - address already in use: %s", line)}
				return
			}
		}

		if scanner.Err() != nil {
			ch <- result{err: fmt.Errorf("pipe read error: %w", scanner.Err())}
		} else {
			log.Error("chrome stderr closed without DevTools URL", "lastOutput", lastLines)
			ch <- result{err: ErrNoDebuggerURL}
		}
	}()

	var res result
	select {
	case res = <-ch:
		return res.url, res.err
	case <-time.After(timeout):
		log.Error("timeout waiting for debugger URL from pipe")
		return "", ErrNoDebuggerURL
	}
}
