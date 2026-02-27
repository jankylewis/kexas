package ktest

import (
	"fmt"

	"github.com/kexas-project/kexas"
)

// ========================================
// 🎯 KTEST TEST - TEST REGISTRATION SYSTEM
// ========================================
//
// Test provides test registration with support for both standalone tests
// and group-based tests. It integrates with the Group system for
// hierarchical test organization.
//
// Usage:
//   ktest.Test("should login successfully", func(page *kexas.Page) {
//       page.Navigate("https://amazon.com")
//       // Test logic
//   })
//
// This gives you the simplicity of Playwright with the organization of TestNG!
// ========================================

// Test registers a named test function with the current group context
// or as a root-level test if no group is active
func Test(name string, testFunc func(*kexas.Page, KTestT)) interface{} {
	if name == "" {
		fmt.Println("❌ ktest.Test: name cannot be empty")
		return nil
	}

	if testFunc == nil {
		fmt.Printf("❌ ktest.Test: test function for '%s' cannot be nil\n", name)
		return nil
	}

	// Register with current group if available, otherwise register globally
	RegisterTestWithGroup(name, testFunc)
	return nil
}

// GetRegisteredTests returns all registered tests (both group and root-level)
func GetRegisteredTests() []NamedTest {
	testMutex.RLock()
	defer testMutex.RUnlock()

	var allTests []NamedTest

	// Add root-level tests
	allTests = append(allTests, registeredTests...)

	// Add group tests
	var allGroups []*GroupContext = GetAllGroups()
	for _, group := range allGroups {
		group.mu.RLock()
		allTests = append(allTests, group.Tests...)
		group.mu.RUnlock()
	}

	return allTests
}

// GetTestsByGroup returns all tests for a specific group
func GetTestsByGroup(groupName string) []NamedTest {
	var group *GroupContext = GetGroupByName(groupName)
	if group == nil {
		return nil
	}

	group.mu.RLock()
	defer group.mu.RUnlock()

	// Return a copy of the tests
	var tests []NamedTest = make([]NamedTest, len(group.Tests))
	copy(tests, group.Tests)

	return tests
}

// GetTestByName finds a test by its full name (including group path)
func GetTestByName(testName string) *NamedTest {
	// Search in root-level tests first
	testMutex.RLock()
	defer testMutex.RUnlock()

	for _, test := range registeredTests {
		if test.Name == testName {
			return &test
		}
	}

	// Search in group tests
	var allGroups []*GroupContext = GetAllGroups()
	for _, group := range allGroups {
		group.mu.RLock()
		for _, test := range group.Tests {
			if test.Name == testName {
				return &test
			}
		}
		group.mu.RUnlock()
	}

	return nil
}

// CountTests returns the total number of registered tests
func CountTests() int {
	testMutex.RLock()
	defer testMutex.RUnlock()

	var count int = len(registeredTests)

	var allGroups []*GroupContext = GetAllGroups()
	for _, group := range allGroups {
		group.mu.RLock()
		count += len(group.Tests)
		group.mu.RUnlock()
	}

	return count
}

// PrintAllTests prints all registered tests for debugging
func PrintAllTests() {
	fmt.Println("🧪 Kexas Registered Tests:")
	fmt.Println("==========================")

	testMutex.RLock()
	defer testMutex.RUnlock()

	// Print root-level tests
	if len(registeredTests) > 0 {
		fmt.Println("📄 Root-level tests:")
		for _, test := range registeredTests {
			fmt.Printf("  🧪 %s\n", test.Name)
		}
	}

	// Print group tests
	var allGroups []*GroupContext = GetAllGroups()
	for _, group := range allGroups {
		group.mu.RLock()
		if len(group.Tests) > 0 {
			fmt.Printf("📁 Group '%s' tests:\n", group.Name)
			for _, test := range group.Tests {
				fmt.Printf("  🧪 %s\n", test.Name)
			}
		}
		group.mu.RUnlock()
	}

	fmt.Printf("📊 Total tests: %d\n", CountTests())
}
