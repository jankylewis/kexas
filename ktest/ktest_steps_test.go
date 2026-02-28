package ktest

import (
	"testing"
	"time"
)

// ========================================
// CRITICAL UNIT TESTS — ktest.Step() and Priority API
// ========================================

func TestStep_RecordsPassedStep(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestLogin"}

	Step(kt, "Navigate to page", func() {
		// simulate work
		time.Sleep(1 * time.Millisecond)
	})

	if len(kt.steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(kt.steps))
	}
	if kt.steps[0].Title != "Navigate to page" {
		t.Errorf("expected title 'Navigate to page', got %q", kt.steps[0].Title)
	}
	if kt.steps[0].Status != "passed" {
		t.Errorf("expected status 'passed', got %q", kt.steps[0].Status)
	}
	if kt.steps[0].Duration < 1*time.Millisecond {
		t.Errorf("expected duration >= 1ms, got %v", kt.steps[0].Duration)
	}
	if kt.steps[0].Error != "" {
		t.Errorf("expected no error, got %q", kt.steps[0].Error)
	}
}

func TestStep_RecordsMultipleSteps(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestCheckout"}

	Step(kt, "Step 1", func() {})
	Step(kt, "Step 2", func() {})
	Step(kt, "Step 3", func() {})

	if len(kt.steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(kt.steps))
	}
	if kt.steps[0].Title != "Step 1" {
		t.Errorf("step 0 title = %q, want 'Step 1'", kt.steps[0].Title)
	}
	if kt.steps[2].Title != "Step 3" {
		t.Errorf("step 2 title = %q, want 'Step 3'", kt.steps[2].Title)
	}
}

func TestStep_PanicMarksFailed(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestPanic"}

	// Step that panics should be marked failed and re-panic
	var didPanic bool = false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		Step(kt, "Panic step", func() {
			panic("something broke")
		})
	}()

	if !didPanic {
		t.Error("expected panic to be re-raised")
	}
	if len(kt.steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(kt.steps))
	}
	if kt.steps[0].Status != "failed" {
		t.Errorf("expected status 'failed', got %q", kt.steps[0].Status)
	}
	if kt.steps[0].Error != "something broke" {
		t.Errorf("expected error 'something broke', got %q", kt.steps[0].Error)
	}
}

func TestStep_NonKtestT_Fallback(t *testing.T) {
	// When using *testing.T wrapper, Step should just run the function
	var wrapper *testingTWrapper = &testingTWrapper{t: t}
	var ran bool = false

	Step(wrapper, "Fallback step", func() {
		ran = true
	})

	if !ran {
		t.Error("expected step function to execute in fallback mode")
	}
}

// ========================================
// PRIORITY API TESTS
// ========================================

func TestPriority_AlphabeticalOrdering(t *testing.T) {
	// Full A–Z: each letter must have a strictly increasing value
	var all []TestPriority = []TestPriority{
		Priority.A, Priority.B, Priority.C, Priority.D, Priority.E,
		Priority.F, Priority.G, Priority.H, Priority.I, Priority.J,
		Priority.K, Priority.L, Priority.M, Priority.N, Priority.O,
		Priority.P, Priority.Q, Priority.R, Priority.S, Priority.T,
		Priority.U, Priority.V, Priority.W, Priority.X, Priority.Y,
		Priority.Z,
	}
	if len(all) != 26 {
		t.Fatalf("expected 26 priorities, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i-1] >= all[i] {
			t.Errorf("Priority at index %d (%d) should be less than index %d (%d)",
				i-1, all[i-1], i, all[i])
		}
	}
	if Priority.A != 0 {
		t.Errorf("Priority.A should be 0, got %d", Priority.A)
	}
	if Priority.Z != 25 {
		t.Errorf("Priority.Z should be 25, got %d", Priority.Z)
	}
}

func TestPriority_LegacyAliasesMatch(t *testing.T) {
	if PriorityHigh != Priority.A {
		t.Errorf("PriorityHigh (%d) should equal Priority.A (%d)", PriorityHigh, Priority.A)
	}
	if PriorityNormal != Priority.B {
		t.Errorf("PriorityNormal (%d) should equal Priority.B (%d)", PriorityNormal, Priority.B)
	}
	if PriorityLow != Priority.C {
		t.Errorf("PriorityLow (%d) should equal Priority.C (%d)", PriorityLow, Priority.C)
	}
}

func TestSortTestsByPriority_OrdersCorrectly(t *testing.T) {
	var tests []NamedTest = []NamedTest{
		{Name: "TestC", Priority: Priority.C, Order: 1},
		{Name: "TestA", Priority: Priority.A, Order: 2},
		{Name: "TestE", Priority: Priority.E, Order: 3},
		{Name: "TestB", Priority: Priority.B, Order: 4},
	}

	SortTestsByPriority(tests)

	var expected []string = []string{"TestA", "TestB", "TestC", "TestE"}
	for i, test := range tests {
		if test.Name != expected[i] {
			t.Errorf("position %d: got %s, want %s", i, test.Name, expected[i])
		}
	}
}

func TestSortTestsByPriority_StableWithinSamePriority(t *testing.T) {
	var tests []NamedTest = []NamedTest{
		{Name: "Test3", Priority: Priority.B, Order: 3},
		{Name: "Test1", Priority: Priority.B, Order: 1},
		{Name: "Test2", Priority: Priority.B, Order: 2},
	}

	SortTestsByPriority(tests)

	// Same priority → sort by registration order
	var expected []string = []string{"Test1", "Test2", "Test3"}
	for i, test := range tests {
		if test.Name != expected[i] {
			t.Errorf("position %d: got %s, want %s", i, test.Name, expected[i])
		}
	}
}

func TestWithPriority_NilSafe(t *testing.T) {
	var ref *TestRef = (*TestRef)(nil)

	// Should not panic
	var result *TestRef = ref.WithPriority(Priority.A)
	if result != ref {
		t.Error("expected nil ref to return itself")
	}
}

func TestWithPriority_SetsPriority(t *testing.T) {
	var nt NamedTest = NamedTest{Name: "TestX", Priority: Priority.B}
	var ref *TestRef = &TestRef{test: &nt}

	ref.WithPriority(Priority.A)

	if nt.Priority != Priority.A {
		t.Errorf("expected priority A (%d), got %d", Priority.A, nt.Priority)
	}
}

// ========================================
// ERROR CAPTURE TESTS
// ========================================

func TestKtestT_Errorf_CapturesMessage(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestErr"}

	kt.Errorf("failed: %s", "timeout")

	if !kt.failed {
		t.Error("expected failed=true")
	}
	if len(kt.errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(kt.errors))
	}
	if kt.errors[0] != "failed: timeout" {
		t.Errorf("expected 'failed: timeout', got %q", kt.errors[0])
	}
}

func TestKtestT_Error_CapturesMessage(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestErr"}

	kt.Error("something went wrong")

	if len(kt.errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(kt.errors))
	}
	if kt.errors[0] != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", kt.errors[0])
	}
}

func TestKtestT_MultipleErrors_AllCaptured(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestMulti"}

	kt.Errorf("error 1")
	kt.Errorf("error 2")
	kt.Error("error 3")

	if len(kt.errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(kt.errors))
	}
}

func TestBuildResult_LastErrorCaptured(t *testing.T) {
	var kt *ktestT = &ktestT{name: "TestBuild"}
	kt.Errorf("first error")
	kt.Errorf("second error")

	Step(kt, "some step", func() {})

	var nt NamedTest = NamedTest{Name: "TestBuild", Filename: "test"}
	var result testResult = buildResult(nt, kt, time.Now().Add(-1*time.Second), 0)

	if result.errorMsg != "second error" {
		t.Errorf("expected last error 'second error', got %q", result.errorMsg)
	}
	if len(result.steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result.steps))
	}
	if result.steps[0].Title != "some step" {
		t.Errorf("expected step title 'some step', got %q", result.steps[0].Title)
	}
}
