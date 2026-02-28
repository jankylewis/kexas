package ktest

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// ktestT is our own implementation of KTestT for the Main() API.
// Thread-safe: all mutable state is guarded by mu.
type ktestT struct {
	name   string
	failed bool
	depth  int
	mu     sync.Mutex
	errors []string   // collected error messages for report
	steps  []testStep // collected test steps for report
	logs   []string   // collected console output for report
}

// newKTestT creates a new ktestT instance.
func newKTestT() *ktestT {
	return &ktestT{
		name:   "main",
		failed: false,
		depth:  0,
	}
}

func (t *ktestT) Helper() {
	t.depth++
}

func (t *ktestT) Log(args ...interface{}) {
	t.mu.Lock()
	var msg string = fmt.Sprint(args...)
	fmt.Println(msg)
	t.logs = append(t.logs, msg)
	t.mu.Unlock()
}

func (t *ktestT) Logf(format string, args ...interface{}) {
	t.mu.Lock()
	var msg string = fmt.Sprintf(format, args...)
	fmt.Println(msg)
	t.logs = append(t.logs, msg)
	t.mu.Unlock()
}

func (t *ktestT) Error(args ...interface{}) {
	t.mu.Lock()
	var msg string = fmt.Sprint(args...)
	fmt.Println(msg)
	t.failed = true
	t.errors = append(t.errors, msg)
	t.logs = append(t.logs, "ERROR: "+msg)
	t.mu.Unlock()
}

func (t *ktestT) Errorf(format string, args ...interface{}) {
	t.mu.Lock()
	var msg string = fmt.Sprintf(format, args...)
	fmt.Println(msg)
	t.failed = true
	t.errors = append(t.errors, msg)
	t.logs = append(t.logs, "ERROR: "+msg)
	t.mu.Unlock()
}

func (t *ktestT) Fatal(args ...interface{}) {
	t.mu.Lock()
	fmt.Println(args...)
	t.failed = true
	t.mu.Unlock()
	os.Exit(1)
}

func (t *ktestT) Fatalf(format string, args ...interface{}) {
	t.mu.Lock()
	fmt.Printf(format+"\n", args...)
	t.failed = true
	t.mu.Unlock()
	os.Exit(1)
}

func (t *ktestT) Failed() bool {
	t.mu.Lock()
	var f bool = t.failed
	t.mu.Unlock()
	return f
}

func (t *ktestT) Name() string {
	return t.name
}

func (t *ktestT) Run(name string, f func(KTestT)) bool {
	var childT *ktestT = &ktestT{
		name:   name,
		failed: false,
		depth:  t.depth + 1,
	}

	f(childT)

	if childT.Failed() {
		t.mu.Lock()
		t.failed = true
		t.mu.Unlock()
	}
	return !childT.Failed()
}

// testingTWrapper wraps *testing.T to implement KTestT.
type testingTWrapper struct {
	t *testing.T
}

func (w *testingTWrapper) Helper() {
	w.t.Helper()
}

func (w *testingTWrapper) Log(args ...interface{}) {
	w.t.Log(args...)
}

func (w *testingTWrapper) Logf(format string, args ...interface{}) {
	w.t.Logf(format, args...)
}

func (w *testingTWrapper) Error(args ...interface{}) {
	w.t.Error(args...)
}

func (w *testingTWrapper) Errorf(format string, args ...interface{}) {
	w.t.Errorf(format, args...)
}

func (w *testingTWrapper) Fatal(args ...interface{}) {
	w.t.Fatal(args...)
}

func (w *testingTWrapper) Fatalf(format string, args ...interface{}) {
	w.t.Fatalf(format, args...)
}

func (w *testingTWrapper) Failed() bool {
	return w.t.Failed()
}

func (w *testingTWrapper) Name() string {
	return w.t.Name()
}

func (w *testingTWrapper) Run(name string, f func(KTestT)) bool {
	return w.t.Run(name, func(t *testing.T) {
		var wrapper *testingTWrapper = &testingTWrapper{t: t}
		f(wrapper)
	})
}
