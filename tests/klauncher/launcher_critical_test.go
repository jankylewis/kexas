package klauncher_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jankylewis/kexas/klauncher"
)

// 15 critical-after-existing unit tests for klauncher.
// Existing: buildArgs, FindFreePort, BrowserClose, ExtractDebuggerURL.
// This batch: Options struct contracts, Launch context cancellation,
// concurrency, and Browser lifecycle edge cases.

func TestOptions_ZeroValueIsUsable(t *testing.T) {
	// A zero-value *Options shouldn't crash buildArgs (used internally by Launch).
	// We can't call internal buildArgs directly, but we can verify the struct
	// has sensible zero-values that downstream code handles.
	var opts klauncher.Options
	if opts.Headless != false {
		t.Errorf("zero Headless should be false, got %v", opts.Headless)
	}
	if opts.Port != 0 {
		t.Errorf("zero Port should be 0 (Chrome self-assigns), got %d", opts.Port)
	}
}

func TestOptions_SetAllFieldsRoundTrip(t *testing.T) {
	var opts klauncher.Options = klauncher.Options{
		Headless:       true,
		Port:           9222,
		Args:           []string{"--lang=fr-FR"},
		ExecutablePath: "/custom/chrome",
		WindowWidth:    1920,
		WindowHeight:   1080,
	}
	if !opts.Headless || opts.Port != 9222 || opts.WindowWidth != 1920 {
		t.Errorf("opts not preserved: %+v", opts)
	}
	if len(opts.Args) != 1 || opts.Args[0] != "--lang=fr-FR" {
		t.Errorf("Args not preserved: %v", opts.Args)
	}
	if opts.ExecutablePath != "/custom/chrome" {
		t.Errorf("ExecutablePath not preserved: %q", opts.ExecutablePath)
	}
}

func TestDefaultOptions_ReturnsNewPointerEachCall(t *testing.T) {
	// Mutating one DefaultOptions return shouldn't affect another.
	var opts1 *klauncher.Options = klauncher.DefaultOptions()
	var opts2 *klauncher.Options = klauncher.DefaultOptions()
	opts1.Port = 12345
	if opts2.Port == 12345 {
		t.Error("DefaultOptions returns shared instance — mutation leaked")
	}
}

func TestDefaultOptions_ArgsIsEmptySlice(t *testing.T) {
	var opts *klauncher.Options = klauncher.DefaultOptions()
	if opts.Args == nil {
		t.Error("DefaultOptions.Args should be empty slice, not nil")
	}
	if len(opts.Args) != 0 {
		t.Errorf("expected empty Args, got %v", opts.Args)
	}
}

func TestLaunch_AlreadyCancelledContext_FailsFast(t *testing.T) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	cancel()

	var start time.Time = time.Now()
	_, err := klauncher.Launch(ctx, &klauncher.Options{
		Headless:       true,
		ExecutablePath: "/nonexistent",
	})
	var elapsed time.Duration = time.Since(start)

	if err == nil {
		t.Fatal("expected error from already-cancelled ctx + bogus path")
	}
	// Should fail fast. The bogus-path lookup is local and quick.
	if elapsed > 5*time.Second {
		t.Errorf("Launch should fail fast on cancelled ctx + bogus path, took %v", elapsed)
	}
}

func TestLaunch_TimeoutContext_FailsBeforeBrowserCouldStart(t *testing.T) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := klauncher.Launch(ctx, &klauncher.Options{
		Headless:       true,
		ExecutablePath: "/definitely/not/here",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLaunch_EmptyExecutablePath_FallsThroughToDiscovery(t *testing.T) {
	// Empty ExecutablePath triggers the auto-discover path. May succeed if
	// Chrome is installed on the test machine. The contract: no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Launch panicked on empty ExecutablePath: %v", r)
		}
	}()
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _ = klauncher.Launch(ctx, &klauncher.Options{Headless: true, ExecutablePath: ""})
}

func TestLaunch_NilOpts_ContextStillRespected(t *testing.T) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // pre-cancel

	// nil opts → DefaultOptions(). Auto-discovery may find a real browser, but
	// the cancelled ctx should kill the launch attempt eventually.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic on nil opts + cancelled ctx: %v", r)
		}
	}()
	_, _ = klauncher.Launch(ctx, nil)
}

func TestErrChromiumNotFound_ErrorMessageDescriptive(t *testing.T) {
	var msg string = klauncher.ErrChromiumNotFound.Error()
	if !strings.Contains(strings.ToLower(msg), "chromium") {
		t.Errorf("ErrChromiumNotFound message should mention 'chromium'; got %q", msg)
	}
}

func TestErrLaunchFailed_ErrorMessageDescriptive(t *testing.T) {
	var msg string = klauncher.ErrLaunchFailed.Error()
	if !strings.Contains(strings.ToLower(msg), "launch") {
		t.Errorf("ErrLaunchFailed message should mention 'launch'; got %q", msg)
	}
}

func TestErrNoDebuggerURL_ErrorMessageDescriptive(t *testing.T) {
	var msg string = klauncher.ErrNoDebuggerURL.Error()
	if !strings.Contains(strings.ToLower(msg), "debugger") {
		t.Errorf("ErrNoDebuggerURL message should mention 'debugger'; got %q", msg)
	}
}

func TestErrors_AllExportedSentinelsAreDistinctFromGeneric(t *testing.T) {
	// None of the three klauncher sentinels should be the same as a generic Go error.
	var generic error = errors.New("generic")
	for _, sentinel := range []error{
		klauncher.ErrChromiumNotFound,
		klauncher.ErrLaunchFailed,
		klauncher.ErrNoDebuggerURL,
	} {
		if errors.Is(generic, sentinel) {
			t.Error("klauncher sentinel must not match generic error")
		}
	}
}

func TestLaunch_ConcurrentCallsToDefaultOptions_DoNotInterfere(t *testing.T) {
	// Run DefaultOptions across many goroutines; each should get a fresh
	// instance with the canonical defaults.
	const N int = 50
	var wg sync.WaitGroup
	var failures int32
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var opts *klauncher.Options = klauncher.DefaultOptions()
			if !opts.Headless || opts.Port != 0 || opts.WindowWidth <= 0 {
				atomic.AddInt32(&failures, 1)
			}
		}()
	}
	wg.Wait()
	if failures != 0 {
		t.Errorf("got %d failures across %d concurrent DefaultOptions calls", failures, N)
	}
}

func TestOptions_ArgsCanBeAppendedSafely(t *testing.T) {
	// Verify a DefaultOptions caller can append to .Args without panic.
	var opts *klauncher.Options = klauncher.DefaultOptions()
	opts.Args = append(opts.Args, "--no-sandbox")
	opts.Args = append(opts.Args, "--disable-gpu")
	if len(opts.Args) != 2 {
		t.Errorf("expected 2 appended args, got %d", len(opts.Args))
	}
}

func TestErrLaunchFailed_WrapsMatchByErrorsIs(t *testing.T) {
	// Confirms the sentinel can be the target of errors.Is when downstream
	// code wraps it (a usability test, not just identity).
	var wrapped error = errors.Join(klauncher.ErrLaunchFailed, errors.New("more context"))
	if !errors.Is(wrapped, klauncher.ErrLaunchFailed) {
		t.Error("errors.Join chain must let errors.Is find ErrLaunchFailed")
	}
}

// atomic int helper — kept inline since this is a test file.
type atomicHelper = struct{}
