
package ktest_test

import (
	"testing"

	"github.com/jankylewis/kexas/ktest"
)

// ========================================

// --- Priority Sorting Tests ---

func TestSortTestsByPriority_HighBeforeNormal(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "NormalTest", Priority: ktest.PriorityNormal, Order: 1},
		{Name: "HighTest", Priority: ktest.PriorityHigh, Order: 2},
	}

	ktest.SortTestsByPriority(tests)

	if tests[0].Name != "HighTest" {
		t.Errorf("expected HighTest first, got %s", tests[0].Name)
	}
	if tests[1].Name != "NormalTest" {
		t.Errorf("expected NormalTest second, got %s", tests[1].Name)
	}
}

func TestSortTestsByPriority_AllThreeLevels(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "LowTest", Priority: ktest.PriorityLow, Order: 1},
		{Name: "HighTest", Priority: ktest.PriorityHigh, Order: 2},
		{Name: "NormalTest", Priority: ktest.PriorityNormal, Order: 3},
	}

	ktest.SortTestsByPriority(tests)

	if tests[0].Name != "HighTest" {
		t.Errorf("expected HighTest first, got %s", tests[0].Name)
	}
	if tests[1].Name != "NormalTest" {
		t.Errorf("expected NormalTest second, got %s", tests[1].Name)
	}
	if tests[2].Name != "LowTest" {
		t.Errorf("expected LowTest third, got %s", tests[2].Name)
	}
}

func TestSortTestsByPriority_SamePriorityPreservesOrder(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "TestA", Priority: ktest.PriorityNormal, Order: 1},
		{Name: "TestB", Priority: ktest.PriorityNormal, Order: 2},
		{Name: "TestC", Priority: ktest.PriorityNormal, Order: 3},
	}

	ktest.SortTestsByPriority(tests)

	if tests[0].Name != "TestA" {
		t.Errorf("expected TestA first, got %s", tests[0].Name)
	}
	if tests[1].Name != "TestB" {
		t.Errorf("expected TestB second, got %s", tests[1].Name)
	}
	if tests[2].Name != "TestC" {
		t.Errorf("expected TestC third, got %s", tests[2].Name)
	}
}

func TestSortTestsByPriority_EmptySlice(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{}
	ktest.SortTestsByPriority(tests)
	if len(tests) != 0 {
		t.Errorf("expected empty slice, got %d", len(tests))
	}
}

func TestSortTestsByPriority_SingleElement(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "OnlyTest", Priority: ktest.PriorityHigh, Order: 1},
	}
	ktest.SortTestsByPriority(tests)
	if tests[0].Name != "OnlyTest" {
		t.Errorf("expected OnlyTest, got %s", tests[0].Name)
	}
}

func TestSortTestsByPriority_MixedPriorityAndOrder(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "Normal2", Priority: ktest.PriorityNormal, Order: 4},
		{Name: "High1", Priority: ktest.PriorityHigh, Order: 3},
		{Name: "Normal1", Priority: ktest.PriorityNormal, Order: 1},
		{Name: "High2", Priority: ktest.PriorityHigh, Order: 5},
		{Name: "Low1", Priority: ktest.PriorityLow, Order: 2},
	}

	ktest.SortTestsByPriority(tests)

	// High priority first, ordered by Order
	if tests[0].Name != "High1" {
		t.Errorf("expected High1 first, got %s", tests[0].Name)
	}
	if tests[1].Name != "High2" {
		t.Errorf("expected High2 second, got %s", tests[1].Name)
	}
	// Normal priority next, ordered by Order
	if tests[2].Name != "Normal1" {
		t.Errorf("expected Normal1 third, got %s", tests[2].Name)
	}
	if tests[3].Name != "Normal2" {
		t.Errorf("expected Normal2 fourth, got %s", tests[3].Name)
	}
	// Low priority last
	if tests[4].Name != "Low1" {
		t.Errorf("expected Low1 fifth, got %s", tests[4].Name)
	}
}

// --- Priority Constants Tests ---

func TestPriorityConstants_Ordering(t *testing.T) {
	if ktest.PriorityHigh >= ktest.PriorityNormal {
		t.Errorf("PriorityHigh (%d) should be less than PriorityNormal (%d)",
			ktest.PriorityHigh, ktest.PriorityNormal)
	}
	if ktest.PriorityNormal >= ktest.PriorityLow {
		t.Errorf("PriorityNormal (%d) should be less than PriorityLow (%d)",
			ktest.PriorityNormal, ktest.PriorityLow)
	}
}
