package ktest

import (
	"fmt"
)

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

	for _, group := range rootGroups {
		allGroups = append(allGroups, group)
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

	if group, exists := rootGroups[name]; exists {
		return group
	}

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

	if parent.Name == name {
		return parent
	}

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

	var indentStr string = ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	fmt.Printf("%s📁 %s (%d tests)\n", indentStr, group.Name, len(group.Tests))

	for _, test := range group.Tests {
		fmt.Printf("%s  🧪 %s\n", indentStr, test.Name)
	}

	for _, child := range group.Children {
		printGroup(child, indent+1)
	}
}
