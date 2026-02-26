package cdp

// Page Command Constants
const (
	CmdPageNavigate         = "Page.navigate"
	CmdPageWaitForLoadState = "Page.waitForLoadState"
	CmdPageClose            = "Page.close"
	CmdPageCaptureScreenshot = "Page.captureScreenshot"
)

// DOM Command Constants
const (
	CmdDOMPerformSearch    = "DOM.performSearch"
	CmdDOMGetSearchResults = "DOM.getSearchResults"
	CmdDOMQuerySelector     = "DOM.querySelector"
	CmdDOMGetComputedStyle  = "DOM.getComputedStyle"
	CmdDOMGetBoxModel       = "DOM.getBoxModel"
	CmdDOMGetAttributes     = "DOM.getAttributes"
	CmdDOMGetOuterHTML      = "DOM.getOuterHTML"
	CmdDOMDescribeNode      = "DOM.describeNode"
)

// Runtime Command Constants
const (
	CmdRuntimeEvaluate      = "Runtime.evaluate"
	CmdRuntimeCallFunctionOn = "Runtime.callFunctionOn"
)

// Command Categories
const (
	CategoryPage    = "Page"
	CategoryDOM     = "DOM"
	CategoryRuntime = "Runtime"
)

// Command Groups for better organization
var (
	// Page Commands - Navigation and lifecycle
	PageCommands = []string{
		CmdPageNavigate,
		CmdPageWaitForLoadState,
		CmdPageClose,
		CmdPageCaptureScreenshot,
	}
	
	// DOM Commands - Document Object Model manipulation
	DOMCommands = []string{
		CmdDOMPerformSearch,
		CmdDOMGetSearchResults,
		CmdDOMQuerySelector,
		CmdDOMGetComputedStyle,
		CmdDOMGetBoxModel,
		CmdDOMGetAttributes,
		CmdDOMGetOuterHTML,
		CmdDOMDescribeNode,
	}
	
	// Runtime Commands - JavaScript execution
	RuntimeCommands = []string{
		CmdRuntimeEvaluate,
		CmdRuntimeCallFunctionOn,
	}
	
	// All Commands - Complete registry
	AllCommands = append(append(PageCommands, DOMCommands...), RuntimeCommands...)
)
