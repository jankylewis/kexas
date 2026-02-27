package ktest

import (
	"fmt"
	"path/filepath"
	"runtime"
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

// NamedTest represents a named test function
type NamedTest struct {
	Name     string
	Func     func(*kexas.Page, KTestT) // Add KTestT for kassert support
	Filename string                    // Source filename without extension
}

// Global test registration
var (
	registeredTests []NamedTest
	testMutex       sync.RWMutex
)

// registerNamedTest registers a test function globally
func registerNamedTest(name string, testFunc func(*kexas.Page, KTestT)) {
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

	registeredTests = append(registeredTests, NamedTest{
		Name:     name,
		Func:     testFunc,
		Filename: baseFilename,
	})
	fmt.Printf("✅ Registered root-level test: %s\n", name)
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

	path := currentGroup.Name
	parent := currentGroup.Parent
	for parent != nil {
		path = parent.Name + "." + path
		parent = parent.Parent
	}

	return path
}

// RegisterTestWithGroup registers a test with the current group context
func RegisterTestWithGroup(name string, testFunc func(*kexas.Page, KTestT)) {
	if currentGroup == nil {
		// No current group, register as root-level test
		registerNamedTest(name, testFunc)
		return
	}

	// Register with current group
	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	// Create full test name with group path
	fullName := GetGroupPath()
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

	namedTest := NamedTest{
		Name:     fullName,
		Func:     testFunc,
		Filename: baseFilename,
	}

	currentGroup.Tests = append(currentGroup.Tests, namedTest)
	fmt.Printf("✅ Registered test: %s\n", fullName)
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
	indentStr := ""
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
