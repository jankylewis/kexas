package ktest

import (
	"fmt"
	"os"
	"testing"
)

// ktestT is our own implementation of KTestT for the Main() API.
type ktestT struct {
	name     string
	failed   bool
	depth    int
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
	fmt.Println(args...)
}

func (t *ktestT) Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (t *ktestT) Error(args ...interface{}) {
	fmt.Println(args...)
	t.failed = true
}

func (t *ktestT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	t.failed = true
}

func (t *ktestT) Fatal(args ...interface{}) {
	fmt.Println(args...)
	t.failed = true
	os.Exit(1)
}

func (t *ktestT) Fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	t.failed = true
	os.Exit(1)
}

func (t *ktestT) Failed() bool {
	return t.failed
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
	
	defer func() {
		if childT.failed {
			t.failed = true
		}
	}()
	
	f(childT)
	return !childT.failed
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
