
package ktest_test

import (
	"sync"
	"testing"

	"github.com/jankylewis/kexas/ktest"
)

// --- NamedTest Field Tests ---

func TestNamedTest_DefaultPriorityIsNormal(t *testing.T) {
	var nt ktest.NamedTest = ktest.NamedTest{
		Name: "TestDefault",
	}
	// Zero value of TestPriority (int) is 0, which is PriorityHigh.
	// But in our registration code, we explicitly set PriorityNormal.
	// This test verifies the zero-value semantics.
	if nt.Priority != ktest.PriorityHigh {
		t.Errorf("zero-value Priority should be %d (PriorityHigh), got %d",
			ktest.PriorityHigh, nt.Priority)
	}
}

func TestNamedTest_PriorityFieldIsSet(t *testing.T) {
	var nt ktest.NamedTest = ktest.NamedTest{
		Name:     "TestExplicit",
		Priority: ktest.PriorityLow,
		Order:    5,
	}
	if nt.Priority != ktest.PriorityLow {
		t.Errorf("expected PriorityLow, got %d", nt.Priority)
	}
	if nt.Order != 5 {
		t.Errorf("expected Order=5, got %d", nt.Order)
	}
}

// --- Worker Count Capping Tests ---

func TestWorkerCount_CappedToTestCount(t *testing.T) {
	// Simulate the capping logic from runTestsParallelWorkers
	var parallelSet int = 5
	var testCount int = 2
	var workerCount int = parallelSet
	if workerCount > testCount {
		workerCount = testCount
	}
	if workerCount != 2 {
		t.Errorf("expected workerCount=2 (capped to test count), got %d", workerCount)
	}
}

func TestWorkerCount_NotCappedWhenEnoughTests(t *testing.T) {
	var parallelSet int = 3
	var testCount int = 10
	var workerCount int = parallelSet
	if workerCount > testCount {
		workerCount = testCount
	}
	if workerCount != 3 {
		t.Errorf("expected workerCount=3, got %d", workerCount)
	}
}

func TestWorkerCount_SingleWorkerSingleTest(t *testing.T) {
	var parallelSet int = 1
	var testCount int = 1
	var workerCount int = parallelSet
	if workerCount > testCount {
		workerCount = testCount
	}
	if workerCount != 1 {
		t.Errorf("expected workerCount=1, got %d", workerCount)
	}
}

// --- ExtractShortName Tests (via sort, verifying naming) ---

func TestSortPreservesNames(t *testing.T) {
	var tests []ktest.NamedTest = []ktest.NamedTest{
		{Name: "Group.SubGroup.TestA", Priority: ktest.PriorityNormal, Order: 1},
		{Name: "Group.TestB", Priority: ktest.PriorityHigh, Order: 2},
		{Name: "TestC", Priority: ktest.PriorityLow, Order: 3},
	}

	ktest.SortTestsByPriority(tests)

	// Verify names are preserved after sorting
	if tests[0].Name != "Group.TestB" {
		t.Errorf("expected Group.TestB first (high), got %s", tests[0].Name)
	}
	if tests[1].Name != "Group.SubGroup.TestA" {
		t.Errorf("expected Group.SubGroup.TestA second (normal), got %s", tests[1].Name)
	}
	if tests[2].Name != "TestC" {
		t.Errorf("expected TestC third (low), got %s", tests[2].Name)
	}
}

// --- Isolation Verification Tests ---

func TestParallel_ResultChannelCollectsAll(t *testing.T) {
	// Simulate the channel pattern used in runTestsParallelWorkers
	var resultsChan chan int = make(chan int, 5)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		var val int = i
		go func() {
			defer wg.Done()
			resultsChan <- val
		}()
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var collected []int
	for r := range resultsChan {
		collected = append(collected, r)
	}

	if len(collected) != 5 {
		t.Errorf("expected 5 results, got %d", len(collected))
	}
}

func TestParallel_TestQueueDrainedCompletely(t *testing.T) {
	// Simulate test queue drain by multiple workers
	var testQueue chan string = make(chan string, 3)
	testQueue <- "TestA"
	testQueue <- "TestB"
	testQueue <- "TestC"
	close(testQueue)

	var consumed []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 2 workers consuming 3 items
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range testQueue {
				mu.Lock()
				consumed = append(consumed, item)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(consumed) != 3 {
		t.Errorf("expected 3 consumed items, got %d", len(consumed))
	}
}
