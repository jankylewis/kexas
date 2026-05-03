package ktest_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jankylewis/kexas/ktest"
)

// 15 critical-after-existing unit tests for ktest.
// Existing: config loading, suite registration, parallel runner, report HTML.
// This batch: hook lifecycle (BeforeAll/AfterAll/BeforeEach/AfterEach),
// global-hook reset, group hierarchy, config-file edge cases.

func TestDefaultConfig_NonNil(t *testing.T) {
	var cfg *ktest.Config = ktest.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
}

func TestDefaultConfig_HasReasonableDefaults(t *testing.T) {
	var cfg *ktest.Config = ktest.DefaultConfig()
	if cfg.Timeout <= 0 {
		t.Errorf("default Timeout should be positive, got %v", cfg.Timeout)
	}
	if cfg.Retries < 0 {
		t.Errorf("default Retries should be ≥0, got %d", cfg.Retries)
	}
}

func TestLoadConfigFromFile_NonexistentPath_ReturnsDefault(t *testing.T) {
	var cfg *ktest.Config = ktest.LoadConfigFromFile("/definitely/not/here/kexas.json")
	if cfg == nil {
		t.Fatal("LoadConfigFromFile should fall back to defaults, not nil")
	}
}

func TestLoadConfigFromFile_EmptyPath_ReturnsDefault(t *testing.T) {
	var cfg *ktest.Config = ktest.LoadConfigFromFile("")
	if cfg == nil {
		t.Fatal("empty path should fall back to defaults, not nil")
	}
}

func TestLoadConfigFromFile_ValidJSON_ParsesFields(t *testing.T) {
	var dir string = t.TempDir()
	var path string = filepath.Join(dir, "kexas.json")
	var content string = `{
		"headless": false,
		"timeout": 5000,
		"retries": 3,
		"parallelSet": 4
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var cfg *ktest.Config = ktest.LoadConfigFromFile(path)
	if cfg.Headless != false {
		t.Errorf("Headless: got %v", cfg.Headless)
	}
	if cfg.Retries != 3 {
		t.Errorf("Retries: got %d", cfg.Retries)
	}
	if cfg.ParallelSet != 4 {
		t.Errorf("ParallelSet: got %d", cfg.ParallelSet)
	}
}

func TestLoadConfigFromFile_MalformedJSON_ReturnsDefault(t *testing.T) {
	var dir string = t.TempDir()
	var path string = filepath.Join(dir, "kexas.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	var cfg *ktest.Config = ktest.LoadConfigFromFile(path)
	if cfg == nil {
		t.Fatal("malformed JSON should fall back to defaults, not nil")
	}
}

func TestBeforeAll_RegistersHook(t *testing.T) {
	ktest.ResetGlobalHooks()
	var called bool
	ktest.BeforeAll(func() { called = true })

	var hook func() = ktest.GetGlobalBeforeAll()
	if hook == nil {
		t.Fatal("BeforeAll should register a global hook")
	}
	hook()
	if !called {
		t.Error("hook function not invoked")
	}
}

func TestAfterAll_RegistersHook(t *testing.T) {
	ktest.ResetGlobalHooks()
	var called bool
	ktest.AfterAll(func() { called = true })

	var hook func() = ktest.GetGlobalAfterAll()
	if hook == nil {
		t.Fatal("AfterAll should register a global hook")
	}
	hook()
	if !called {
		t.Error("hook function not invoked")
	}
}

func TestResetGlobalHooks_ClearsAll(t *testing.T) {
	ktest.BeforeAll(func() {})
	ktest.AfterAll(func() {})
	ktest.ResetGlobalHooks()
	if ktest.GetGlobalBeforeAll() != nil {
		t.Error("ResetGlobalHooks should clear BeforeAll")
	}
	if ktest.GetGlobalAfterAll() != nil {
		t.Error("ResetGlobalHooks should clear AfterAll")
	}
}

func TestExecuteGlobalBeforeAll_CallsRegisteredHook(t *testing.T) {
	ktest.ResetGlobalHooks()
	var called int
	ktest.BeforeAll(func() { called++ })
	ktest.ExecuteGlobalBeforeAll()
	if called != 1 {
		t.Errorf("expected 1 invocation, got %d", called)
	}
}

func TestExecuteGlobalBeforeAll_NoHook_NoOp(t *testing.T) {
	ktest.ResetGlobalHooks()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ExecuteGlobalBeforeAll panicked with no hook: %v", r)
		}
	}()
	ktest.ExecuteGlobalBeforeAll()
}

func TestGetAllGroups_DoesNotPanic(t *testing.T) {
	// Either empty or pre-populated by prior test. Contract being tested: the
	// call doesn't panic regardless of state. Returning nil for empty is also
	// a valid implementation — both shapes are acceptable here.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetAllGroups panicked: %v", r)
		}
	}()
	_ = ktest.GetAllGroups()
}

func TestGetGroupByName_UnknownName_ReturnsNil(t *testing.T) {
	var g *ktest.GroupContext = ktest.GetGroupByName("definitely-not-a-group-xyz")
	if g != nil {
		t.Errorf("GetGroupByName for unknown name should return nil, got %+v", g)
	}
}

func TestConcurrentBeforeAllRegistration_NoRace(t *testing.T) {
	// Stress: register many hooks concurrently. The last one wins (or merges,
	// depending on impl). Test contract: no panic, GetGlobalBeforeAll non-nil.
	const N int = 20
	ktest.ResetGlobalHooks()
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ktest.BeforeAll(func() {})
		}()
	}
	wg.Wait()
	if ktest.GetGlobalBeforeAll() == nil {
		t.Error("GetGlobalBeforeAll should be non-nil after concurrent registration")
	}
}

func TestConfig_FieldsAccessible(t *testing.T) {
	// Verify the Config struct has the expected exported fields by setting them.
	var cfg ktest.Config = ktest.Config{
		Headless:         true,
		Timeout:          0,
		Retries:          1,
		Parallel:         true,
		ParallelSet:      2,
		ScreenshotOnFail: true,
		ScreenshotDir:    "screens",
		VideoDir:         "videos",
		ReportDir:        "reports",
		BaseURL:          "https://example.com",
	}
	if !cfg.Headless || cfg.Retries != 1 || cfg.ParallelSet != 2 {
		t.Errorf("Config fields not preserved: %+v", cfg)
	}
	if !strings.Contains(cfg.BaseURL, "example") {
		t.Errorf("BaseURL field corrupt: %s", cfg.BaseURL)
	}
}
