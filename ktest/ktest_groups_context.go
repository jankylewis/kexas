package ktest

import (
	"fmt"
	"sync"

	"github.com/jankylewis/kexas"
)

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
		return
	}
	// This is a root group
	groupMutex.Lock()
	rootGroups[group.Name] = group
	groupMutex.Unlock()
}

// executeGroupFunction executes a group function with proper context management
func executeGroupFunction(group *GroupContext, groupFunc func()) {
	groupStack = append(groupStack, group)
	currentGroup = group

	groupFunc()

	popGroupFromStack()
}

// popGroupFromStack removes the current group from the stack
func popGroupFromStack() {
	groupStack = groupStack[:len(groupStack)-1]
	if len(groupStack) > 0 {
		currentGroup = groupStack[len(groupStack)-1]
		return
	}
	currentGroup = nil
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

	var group *GroupContext = createGroupContext(name)
	setGroupParent(group)
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
		return registerNamedTest(name, testFunc)
	}

	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	var fullName string = qualifiedTestName(name)
	var baseFilename string = resolveCallerBaseFilename()
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
		registerRootTestWithPriority(name, testFunc, priority)
		return
	}

	currentGroup.mu.Lock()
	defer currentGroup.mu.Unlock()

	var fullName string = qualifiedTestName(name)
	var baseFilename string = resolveCallerBaseFilename()
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

// registerRootTestWithPriority appends a priority-tagged test to the root-level
// registry (used when there is no active group context).
func registerRootTestWithPriority(name string, testFunc func(*kexas.Page, KTestT), priority TestPriority) {
	testMutex.Lock()
	defer testMutex.Unlock()

	var baseFilename string = resolveCallerBaseFilename()
	globalTestOrder++
	registeredTests = append(registeredTests, NamedTest{
		Name:     name,
		Func:     testFunc,
		Filename: baseFilename,
		Priority: priority,
		Order:    globalTestOrder,
	})
	fmt.Printf("✅ Registered root-level test: %s (priority=%d)\n", name, priority)
}

// qualifiedTestName prefixes name with the current group path (dot-separated) when
// inside a group, otherwise returns name unchanged.
func qualifiedTestName(name string) string {
	var fullName string = GetGroupPath()
	if fullName == "" {
		return name
	}
	return fullName + "." + name
}
