package cdp

// Command represents a Chrome DevTools Protocol command with metadata.
type Command struct {
	// Method is the CDP method name (e.g., "DOM.querySelector")
	Method string

	// Description explains what this command does
	Description string

	// Category groups related commands (e.g., "DOM", "Page", "Runtime")
	Category string

	// Deprecated indicates if this command is deprecated
	Deprecated bool

	// Since indicates the Chrome version when this command was introduced
	Since string
}

// Commands Registry - All CDP commands used by Kexas
var Commands = map[string]Command{
	// Page Commands - Page navigation and lifecycle
	"Page.navigate": {
		Method:      "Page.navigate",
		Description: "Navigates the current page to the specified URL",
		Category:    "Page",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"Page.waitForLoadState": {
		Method:      "Page.waitForLoadState",
		Description: "Waits for the page to reach a specific load state",
		Category:    "Page",
		Deprecated:  false,
		Since:       "Chrome 94",
	},
	"Page.close": {
		Method:      "Page.close",
		Description: "Closes the current page",
		Category:    "Page",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"Page.captureScreenshot": {
		Method:      "Page.captureScreenshot",
		Description: "Captures a screenshot of the current page",
		Category:    "Page",
		Deprecated:  false,
		Since:       "Chrome 63",
	},

	// DOM Commands - Document Object Model manipulation
	"DOM.performSearch": {
		Method:      "DOM.performSearch",
		Description: "Searches for nodes in the DOM using XPath or CSS selectors",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.getSearchResults": {
		Method:      "DOM.getSearchResults",
		Description: "Returns search results from a previous DOM.performSearch call",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.querySelector": {
		Method:      "DOM.querySelector",
		Description: "Finds the first element matching the specified CSS selector",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.getComputedStyle": {
		Method:      "DOM.getComputedStyle",
		Description: "Returns the computed style for the specified DOM node",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.getBoxModel": {
		Method:      "DOM.getBoxModel",
		Description: "Returns the box model for the specified DOM node",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.getAttributes": {
		Method:      "DOM.getAttributes",
		Description: "Returns all attributes of the specified DOM node",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.getOuterHTML": {
		Method:      "DOM.getOuterHTML",
		Description: "Returns the outer HTML of the specified DOM node",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"DOM.describeNode": {
		Method:      "DOM.describeNode",
		Description: "Describes the specified DOM node",
		Category:    "DOM",
		Deprecated:  false,
		Since:       "Chrome 63",
	},

	// Runtime Commands - JavaScript execution
	"Runtime.evaluate": {
		Method:      "Runtime.evaluate",
		Description: "Evaluates JavaScript expression in the page context",
		Category:    "Runtime",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
	"Runtime.callFunctionOn": {
		Method:      "Runtime.callFunctionOn",
		Description: "Calls a JavaScript function on the specified DOM node",
		Category:    "Runtime",
		Deprecated:  false,
		Since:       "Chrome 63",
	},
}

// GetCommand returns a command from the registry.
func GetCommand(method string) (Command, bool) {
	var cmd Command
	var exists bool
	cmd, exists = Commands[method]
	return cmd, exists
}

// ListCommands returns all commands in the specified category.
func ListCommands(category string) []Command {
	var result []Command
	for _, cmd := range Commands {
		if cmd.Category == category {
			result = append(result, cmd)
		}
	}
	return result
}

// AllCategories returns all command categories.
func AllCategories() []string {
	var categories map[string]bool = make(map[string]bool)
	for _, cmd := range Commands {
		categories[cmd.Category] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// IsDeprecated checks if a command is deprecated.
func IsDeprecated(method string) bool {
	var cmd Command
	var exists bool
	cmd, exists = Commands[method]
	return exists && cmd.Deprecated
}

// GetCommandDescription returns a human-readable description of the command.
func GetCommandDescription(method string) string {
	var cmd Command
	var exists bool
	cmd, exists = Commands[method]
	if !exists {
		return "Unknown command"
	}
	return cmd.Description
}
