package launcher

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/jankylewis/kexas/internal/logger"
)

// Launch starts a Chromium browser with the given options.
func Launch(ctx context.Context, opts *Options) (*Browser, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	var log *logger.Logger = logger.New("launcher")

	var execPath string
	var err error
	execPath, err = resolveChromiumPath(opts, log)
	if err != nil {
		return nil, err
	}
	log.Debug("found chromium", "path", execPath)
	logPortStrategy(opts, log)

	var cmd *exec.Cmd
	var userDataDir string
	var stderrPipe io.ReadCloser
	var cancel context.CancelFunc
	cmd, userDataDir, stderrPipe, cancel, err = prepareLaunchProcess(ctx, opts, execPath)
	if err != nil {
		return nil, err
	}

	log.Info("launching chromium", "headless", opts.Headless, "port", opts.Port)
	cleanOnce.Do(func() { cleanStaleTempProfiles(log) })

	err = ensurePortAvailable(opts.Port, log)
	if err != nil {
		cancel()
		return nil, err
	}

	var wsURL string
	wsURL, err = startProcessAndCaptureURL(cmd, stderrPipe, opts.Port, log)
	if err != nil {
		cancel()
		return nil, err
	}

	var actualPort int = resolveActualPort(opts.Port, wsURL, log)
	log.Info("chromium launched successfully", "wsURL", wsURL)

	return &Browser{
		cmd:         cmd,
		wsURL:       wsURL,
		log:         log,
		cancelFunc:  cancel,
		userDataDir: userDataDir,
		logFile:     "", // No log file — using pipe-based stderr capture
		port:        actualPort,
	}, nil
}

// resolveChromiumPath returns opts.ExecutablePath if set, otherwise auto-discovers Chromium.
func resolveChromiumPath(opts *Options, log *logger.Logger) (string, error) {
	if opts.ExecutablePath != "" {
		return opts.ExecutablePath, nil
	}
	var path string
	var err error
	path, err = findChromium()
	if err != nil {
		log.Error("chromium not found", "err", err)
		return "", err
	}
	return path, nil
}

// logPortStrategy logs whether Chrome will self-assign a port or use a fixed one.
//
// CRITICAL: When Port == 0, Chrome self-assigns and prints the URL on stderr.
// This eliminates the TOCTOU race where findFreePort() releases a port that
// another worker or OS process grabs before Chrome can bind it.
func logPortStrategy(opts *Options, log *logger.Logger) {
	if opts.Port == 0 {
		log.Info("using Chrome self-assigned port (--remote-debugging-port=0)")
	} else {
		log.Info("using fixed port", "port", opts.Port)
	}
}

// prepareLaunchProcess builds args, allocates a cancellable context, and creates the
// exec.Cmd with stderr piped — but does not start the process.
//
// Pipe-based stderr capture is used because file-based capture has buffering lag
// that caused 10s timeouts under parallel load.
func prepareLaunchProcess(ctx context.Context, opts *Options, execPath string) (*exec.Cmd, string, io.ReadCloser, context.CancelFunc, error) {
	var args []string
	var userDataDir string
	args, userDataDir = buildArgs(opts)

	var cmdCtx context.Context
	var cancel context.CancelFunc
	cmdCtx, cancel = context.WithCancel(ctx)
	var cmd *exec.Cmd = exec.CommandContext(cmdCtx, execPath, args...)

	var stderrPipe io.ReadCloser
	var err error
	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, "", nil, nil, fmt.Errorf("launcher: failed to create stderr pipe: %w", err)
	}
	cmd.Stdout = nil // Chrome outputs DevTools URL to stderr only
	return cmd, userDataDir, stderrPipe, cancel, nil
}

// ensurePortAvailable kills any existing Chrome on a fixed port and verifies it is free.
// No-op when port == 0 (Chrome self-assigns; conflicts are impossible).
func ensurePortAvailable(port int, log *logger.Logger) error {
	if port == 0 {
		return nil
	}
	log.Info("step 1: killing existing Chrome processes on port", "port", port)
	killExistingChromeProcesses(port)

	time.Sleep(500 * time.Millisecond)

	log.Info("step 2: checking if port is available", "port", port)
	if !isPortAvailable(port) {
		log.Error("port still in use after cleanup", "port", port)
		return fmt.Errorf("launcher: port %d is already in use - kill existing Chrome processes or wait for port release", port)
	}
	return nil
}

// startProcessAndCaptureURL starts cmd and reads the DevTools URL from stderrPipe.
// On URL-extraction failure, the Chrome process is killed before returning.
func startProcessAndCaptureURL(cmd *exec.Cmd, stderrPipe io.ReadCloser, port int, log *logger.Logger) (string, error) {
	log.Info("starting Chrome", "port", port)
	var err error = cmd.Start()
	if err != nil {
		log.Error("failed to start chromium", "err", err)
		return "", fmt.Errorf("%w: %v", ErrLaunchFailed, err)
	}
	log.Debug("chromium process started", "pid", cmd.Process.Pid)

	var wsURL string
	wsURL, err = extractDebuggerURLFromPipe(stderrPipe, 10*time.Second, log)
	if err != nil {
		cmd.Process.Kill()
		log.Error("failed to extract debugger URL", "err", err)
		return "", err
	}
	return wsURL, nil
}

// resolveActualPort returns the declared port if non-zero, otherwise extracts the
// Chrome-assigned port from the WebSocket URL (the port=0 self-assignment case).
func resolveActualPort(declaredPort int, wsURL string, log *logger.Logger) int {
	if declaredPort != 0 {
		return declaredPort
	}
	var actual int = extractPortFromWSURL(wsURL)
	log.Info("Chrome self-assigned port", "port", actual)
	return actual
}
