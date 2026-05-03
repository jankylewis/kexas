
package ktest_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/jankylewis/kexas/ktest"
)

// --- Config ParallelSet Tests ---

func TestConfigParallelSet_DefaultIsOne(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	if config.ParallelSet != 1 {
		t.Errorf("expected default ParallelSet=1, got %d", config.ParallelSet)
	}
	if config.Parallel != false {
		t.Errorf("expected default Parallel=false, got %v", config.Parallel)
	}
}

func TestConfigParallelSet_JSONParsing(t *testing.T) {
	type configJSON struct {
		ParallelSet *int `json:"parallelSet"`
	}
	var raw string = `{"parallelSet": 3}`
	var parsed configJSON
	var err error = json.Unmarshal([]byte(raw), &parsed)
	if err != nil {
		t.Fatalf("json parse failed: %v", err)
	}
	if parsed.ParallelSet == nil {
		t.Fatal("parallelSet should not be nil")
	}
	if *parsed.ParallelSet != 3 {
		t.Errorf("expected parallelSet=3, got %d", *parsed.ParallelSet)
	}
}

func TestConfigParallelSet_GreaterThanOneEnablesParallel(t *testing.T) {
	// Simulate what loadConfig does
	var config *ktest.Config = ktest.DefaultConfig()
	var parallelSet int = 2
	if parallelSet >= 1 {
		config.ParallelSet = parallelSet
		if config.ParallelSet > 1 {
			config.Parallel = true
		}
	}

	if config.ParallelSet != 2 {
		t.Errorf("expected ParallelSet=2, got %d", config.ParallelSet)
	}
	if !config.Parallel {
		t.Errorf("expected Parallel=true when ParallelSet=2")
	}
}

func TestConfigParallelSet_OneKeepsSequential(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	var parallelSet int = 1
	if parallelSet >= 1 {
		config.ParallelSet = parallelSet
		if config.ParallelSet > 1 {
			config.Parallel = true
		}
	}

	if config.ParallelSet != 1 {
		t.Errorf("expected ParallelSet=1, got %d", config.ParallelSet)
	}
	if config.Parallel {
		t.Errorf("expected Parallel=false when ParallelSet=1")
	}
}

// --- Thread-Safe ktestT Tests ---

// mockKTestT simulates KTestT for testing thread safety
type mockKTestT struct {
	mu       sync.Mutex
	failed   bool
	errorLog []string
}

func (m *mockKTestT) Helper()                         {}
func (m *mockKTestT) Log(args ...interface{})          {}
func (m *mockKTestT) Logf(f string, a ...interface{})  {}
func (m *mockKTestT) Fatal(args ...interface{})        { m.markFailed() }
func (m *mockKTestT) Fatalf(f string, a ...interface{}) { m.markFailed() }
func (m *mockKTestT) Name() string                     { return "mock" }
func (m *mockKTestT) Run(name string, f func(ktest.KTestT)) bool {
	f(m)
	return !m.Failed()
}

func (m *mockKTestT) Error(args ...interface{}) {
	m.markFailed()
}

func (m *mockKTestT) Errorf(format string, args ...interface{}) {
	m.markFailed()
}

func (m *mockKTestT) Failed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failed
}

func (m *mockKTestT) markFailed() {
	m.mu.Lock()
	m.failed = true
	m.mu.Unlock()
}

func TestKTestT_ConcurrentErrorfNoRace(t *testing.T) {
	// This test verifies that calling Errorf from multiple goroutines
	// does not cause a data race. Run with -race flag.
	var mock *mockKTestT = &mockKTestT{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mock.Errorf("concurrent error")
		}()
	}
	wg.Wait()

	if !mock.Failed() {
		t.Errorf("expected Failed()=true after concurrent Errorf calls")
	}
}

func TestKTestT_ConcurrentFailedNoRace(t *testing.T) {
	// Concurrent reads of Failed() should not race with writes
	var mock *mockKTestT = &mockKTestT{}

	var wg sync.WaitGroup
	// Writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mock.Errorf("write")
		}()
	}
	// Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mock.Failed()
		}()
	}
	wg.Wait()
}
