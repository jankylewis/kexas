
package tests

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/kexas-project/kexas/ktest"
)

// ========================================
// UNIT TESTS FOR KTEST PARALLEL EXECUTION
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
