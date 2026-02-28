package ktest

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/kexas-project/kexas"
)

// ========================================
// KTEST GROUPS - TEST ORGANIZATION SYSTEM
// ========================================
//
// Group provides flexible test organization similar to TestNG test classes
// or Playwright test.describe() blocks. It supports unlimited nesting
// and group-level hooks for complete test organization.
//
// Usage:
//   ktest.Group("Amazon Authentication", func() {
//       ktest.Test("should login successfully", func(page *kexas.Page) {
//           // Test logic
//       }),
//   })
//
// This gives you the organization of TestNG with the flexibility of Playwright!
// ========================================

// TestPriority controls execution order when tests are sorted before dispatch.
// Lower numeric value = higher priority = runs first.
// Kexas uses alphabetical ordering: A runs before B, B before C, etc.
type TestPriority int

// prioritySet provides namespaced priority constants accessed via ktest.Priority.
// Alphabetical ordering: A (highest) through Z (lowest).
// Tests with Priority.A execute before Priority.B, and so on.
type prioritySet struct {
	A TestPriority
	B TestPriority
	C TestPriority
	D TestPriority
	E TestPriority
	F TestPriority
	G TestPriority
	H TestPriority
	I TestPriority
	J TestPriority
	K TestPriority
	L TestPriority
	M TestPriority
	N TestPriority
	O TestPriority
	P TestPriority
	Q TestPriority
	R TestPriority
	S TestPriority
	T TestPriority
	U TestPriority
	V TestPriority
	W TestPriority
	X TestPriority
	Y TestPriority
	Z TestPriority
}

// Priority provides alphabetically-ordered priority constants (A–Z).
// Usage: ktest.Test("Login", fn).WithPriority(ktest.Priority.A)
var Priority prioritySet = prioritySet{
	A: 0, B: 1, C: 2, D: 3, E: 4, F: 5, G: 6, H: 7, I: 8, J: 9,
	K: 10, L: 11, M: 12, N: 13, O: 14, P: 15, Q: 16, R: 17, S: 18, T: 19,
	U: 20, V: 21, W: 22, X: 23, Y: 24, Z: 25,
}

// Legacy aliases — deprecated, use ktest.Priority.A / .B / .C instead.
const (
	PriorityHigh   TestPriority = 0 // Deprecated: use Priority.A
	PriorityNormal TestPriority = 1 // Deprecated: use Priority.B
	PriorityLow    TestPriority = 2 // Deprecated: use Priority.C
)

// NamedTest represents a named test function
type NamedTest struct {
	Name     string
	Func     func(*kexas.Page, KTestT)
	Filename string
	Priority TestPriority
	Order    int // registration order for stable sort
}

// SortTestsByPriority sorts tests by priority (high first), then by registration order.
func SortTestsByPriority(tests []NamedTest) {
	sort.SliceStable(tests, func(i int, j int) bool {
		if tests[i].Priority != tests[j].Priority {
			return tests[i].Priority < tests[j].Priority
		}
		return tests[i].Order < tests[j].Order
	})
}

// Global test registration
var (
	registeredTests []NamedTest
	testMutex       sync.RWMutex
	globalTestOrder int
)

// registerNamedTest registers a test function globally and returns a pointer to it.
func registerNamedTest(name string, testFunc func(*kexas.Page, KTestT)) *NamedTest {
	testMutex.Lock()
	defer testMutex.Unlock()

	// Get filename from caller stack
	var callerFilename string
	for depth := 0; depth <= 10; depth++ {
		var _, file, _, ok = runtime.Caller(depth)
		if !ok {
			break
		}
		// Look for user's test file (not ktest/kexas internal files)
		if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
			callerFilename = file
			break
		}
	}

	var baseFilename string = "test"
	if callerFilename != "" {
		baseFilename = filepath.Base(callerFilename)
		baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	}

	globalTestOrder++
	registeredTests = append(registeredTests, NamedTest{
		Name:     name,
		Func:     testFunc,
		Filename: baseFilename,
		Priority: PriorityNormal,
		Order:    globalTestOrder,
	})
	var displayName string = fmt.Sprintf("<%s.%s>", baseFilename, extractShortName(name))
	fmt.Printf("✅ Registered test: %s\n", displayName)
	return &registeredTests[len(registeredTests)-1]
}

// GroupContext represents a test group with its configuration
type GroupContext struct {
	Name      string
	Parent    *GroupContext
	BeforeAll func()
	AfterAll  func()
	Tests     []NamedTest
	Children  []*GroupContext
	mu        sync.RWMutex
}

// Global group management
var (
	groupStack   []*GroupContext
	currentGroup *GroupContext
	rootGroups   map[string]*GroupContext
	groupCounter int
	groupMutex   sync.RWMutex
)

// Initialize global group state
func init() {
	rootGroups = make(map[string]*GroupContext)
	groupStack = make([]*GroupContext, 0)
}

// createGroupContext creates a new group context with proper initialization
func createGroupContext(name string) *GroupContext {
	return &GroupContext{
		Name:     name,
		Tests:    make([]NamedTest, 0),
		Children: make([]*GroupContext, 0),
	}
}

// setGroupParent sets the parent relationship for a group
func setGroupParent(group *GroupContext) {
	if currentGroup != nil {
		group.Parent = currentGroup
		currentGroup.mu.Lock()
		currentGroup.Children = append(currentGroup.Children, group)
		currentGroup.mu.Unlock()
	} else {
		// This is a root group
		groupMutex.Lock()
		rootGroups[group.Name] = group
		groupMutex.Unlock()
	}
}

// executeGroupFunction executes a group function with proper context management
func executeGroupFunction(group *GroupContext, groupFunc func()) {
	// Push group onto stack
	groupStack = append(groupStack, group)
	currentGroup = group

	// Execute group function to register nested components
	groupFunc()

	// Pop group from stack
	popGroupFromStack()
}

// popGroupFromStack removes the current group from the stack
func popGroupFromStack() {
	groupStack = groupStack[:len(groupStack)-1]
	if len(groupStack) > 0 {
		currentGroup = groupStack[len(groupStack)-1]
	} else {
		currentGroup = nil
	}
}

// Group creates a new test group with the given name and configuration
func Group(name string, groupFunc func()) interface{} {
	if name == "" {
		fmt.Println("❌ ktest.Group: name cannot be empty")
		return nil
	}

	if groupFunc == nil {
		fmt.Printf("❌ ktest.Group: group function for '%s' cannot be nil\n", name)
		return nil
	}

	fmt.Printf("📁 Creating test group: %s\n", name)

	// Create new group context
	var group *GroupContext = createGroupContext(name)

	// Set parent relationship
	setGroupParent(group)

	// Execute group function with proper context management
	executeGroupFunction(group, groupFunc)

	fmt.Printf("✅ Group '%s' created with %d tests\n", name, len(group.Tests))
	return nil
}

// GetCurrentGroup returns the current active group context
func GetCurrentGroup() *GroupContext {
	groupMutex.RLock()
	defer groupMutex.RUnlock()
	return currentGroup
}

// GetGroupPath returns the full path of the current group (e.g., "Auth.Login")
func GetGroupPath() string {
	if currentGroup == nil {
		return ""
	}

	var path string = currentGroup.Name
	var parent *GroupContext = currentGroup.Parent
	for parent != nil {
		path = parent.Name + "." + path
		parent = parent.Parent
	}

	return path
}

// RegisterTestWithGroup registers a test with the current group context
// and returns a pointer to the registered NamedTest.
func RegisterTestWithGroup(name string, testFunc func(*kexas.Page, KTestT)) *NamedTest {
	if currentGroup == nil {
		// No current group, register as root-level test
		return registerNamedTest(name, testFunc)
	}

	// Register with current group
	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	// Create full test name with group path
	var fullName string = GetGroupPath()
	if fullName != "" {
		fullName = fullName + "." + name
	} else {
		fullName = name
	}

	// Get filename from caller stack
	var callerFilename string
	for depth := 0; depth <= 10; depth++ {
		var _, file, _, ok = runtime.Caller(depth)
		if !ok {
			break
		}
		// Look for user's test file (not ktest/kexas internal files)
		if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
			callerFilename = file
			break
		}
	}

	var baseFilename string = "test"
	if callerFilename != "" {
		baseFilename = filepath.Base(callerFilename)
		baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	}

	globalTestOrder++
	var namedTest NamedTest = NamedTest{
		Name:     fullName,
		Func:     testFunc,
		Filename: baseFilename,
		Priority: PriorityNormal,
		Order:    globalTestOrder,
	}

	currentGroup.Tests = append(currentGroup.Tests, namedTest)
	var displayName string = fmt.Sprintf("<%s.%s>", baseFilename, extractShortName(name))
	fmt.Printf("✅ Registered test: %s\n", displayName)
	return &currentGroup.Tests[len(currentGroup.Tests)-1]
}

// RegisterTestWithGroupAndPriority registers a test with explicit priority.
func RegisterTestWithGroupAndPriority(name string, testFunc func(*kexas.Page, KTestT), priority TestPriority) {
	if currentGroup == nil {
		testMutex.Lock()
		defer testMutex.Unlock()

		var callerFilename string
		for depth := 0; depth <= 10; depth++ {
			var _, file, _, ok = runtime.Caller(depth)
			if !ok {
				break
			}
			if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
				callerFilename = file
				break
			}
		}

		var baseFilename string = "test"
		if callerFilename != "" {
			baseFilename = filepath.Base(callerFilename)
			baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
		}

		globalTestOrder++
		registeredTests = append(registeredTests, NamedTest{
			Name:     name,
			Func:     testFunc,
			Filename: baseFilename,
			Priority: priority,
			Order:    globalTestOrder,
		})
		fmt.Printf("✅ Registered root-level test: %s (priority=%d)\n", name, priority)
		return
	}

	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	var fullName string = GetGroupPath()
	if fullName != "" {
		fullName = fullName + "." + name
	} else {
		fullName = name
	}

	var callerFilename string
	for depth := 0; depth <= 10; depth++ {
		var _, file, _, ok = runtime.Caller(depth)
		if !ok {
			break
		}
		if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
			callerFilename = file
			break
		}
	}

	var baseFilename string = "test"
	if callerFilename != "" {
		baseFilename = filepath.Base(callerFilename)
		baseFilename = strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))
	}

	globalTestOrder++
	var namedTest NamedTest = NamedTest{
		Name:     fullName,
		Func:     testFunc,
		Filename: baseFilename,
		Priority: priority,
		Order:    globalTestOrder,
	}

	currentGroup.Tests = append(currentGroup.Tests, namedTest)
	fmt.Printf("✅ Registered test: %s (priority=%d)\n", fullName, priority)
}

// RegisterGroupHook registers a hook function with the current group
func RegisterGroupHook(hookType string, hookFunc interface{}) {
	if currentGroup == nil {
		fmt.Printf("❌ Cannot register %s hook - no active group\n", hookType)
		return
	}

	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	switch hookType {
	case "BeforeAll":
		if fn, ok := hookFunc.(func()); ok {
			currentGroup.BeforeAll = fn
			fmt.Printf("✅ Registered BeforeAll hook for group '%s'\n", currentGroup.Name)
		} else {
			fmt.Printf("❌ BeforeAll hook must be func()\n")
		}

	case "AfterAll":
		if fn, ok := hookFunc.(func()); ok {
			currentGroup.AfterAll = fn
			fmt.Printf("✅ Registered AfterAll hook for group '%s'\n", currentGroup.Name)
		} else {
			fmt.Printf("❌ AfterAll hook must be func()\n")
		}

	default:
		fmt.Printf("❌ Unknown hook type: %s\n", hookType)
	}
}

// GetAllGroups returns all registered groups (including nested)
func GetAllGroups() []*GroupContext {
	groupMutex.RLock()
	defer groupMutex.RUnlock()

	var allGroups []*GroupContext

	// Add root groups
	for _, group := range rootGroups {
		allGroups = append(allGroups, group)
		// Add nested groups recursively
		allGroups = append(allGroups, getNestedGroups(group)...)
	}

	return allGroups
}

// getNestedGroups recursively gets all nested groups
func getNestedGroups(parent *GroupContext) []*GroupContext {
	var nested []*GroupContext

	parent.mu.RLock()
	defer parent.mu.RUnlock()

	for _, child := range parent.Children {
		nested = append(nested, child)
		nested = append(nested, getNestedGroups(child)...)
	}

	return nested
}

// GetGroupByName finds a group by its full path name
func GetGroupByName(name string) *GroupContext {
	groupMutex.RLock()
	defer groupMutex.RUnlock()

	// Search in root groups first
	if group, exists := rootGroups[name]; exists {
		return group
	}

	// Search in all groups
	for _, group := range rootGroups {
		if found := findGroupByName(group, name); found != nil {
			return found
		}
	}

	return nil
}

// findGroupByName recursively searches for a group by name
func findGroupByName(parent *GroupContext, name string) *GroupContext {
	parent.mu.RLock()
	defer parent.mu.RUnlock()

	// Check if this group matches
	if parent.Name == name {
		return parent
	}

	// Check children
	for _, child := range parent.Children {
		if found := findGroupByName(child, name); found != nil {
			return found
		}
	}

	return nil
}

// PrintGroupHierarchy prints the complete group hierarchy for debugging
func PrintGroupHierarchy() {
	fmt.Println("🌳 Kexas Test Group Hierarchy:")
	fmt.Println("================================")

	groupMutex.RLock()
	defer groupMutex.RUnlock()

	for _, group := range rootGroups {
		printGroup(group, 0)
	}
}

// printGroup recursively prints a group and its children
func printGroup(group *GroupContext, indent int) {
	group.mu.RLock()
	defer group.mu.RUnlock()

	// Print group name
	var indentStr string = ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	fmt.Printf("%s📁 %s (%d tests)\n", indentStr, group.Name, len(group.Tests))

	// Print tests
	for _, test := range group.Tests {
		fmt.Printf("%s  🧪 %s\n", indentStr, test.Name)
	}

	// Print children
	for _, child := range group.Children {
		printGroup(child, indent+1)
	}
}
