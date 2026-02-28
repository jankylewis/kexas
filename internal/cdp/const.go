package cdp

// Page Command Constants
const (
	CmdPageNavigate           = "Page.navigate"
	CmdPageWaitForLoadState   = "Page.waitForLoadState"
	CmdPageClose              = "Page.close"
	CmdPageCaptureScreenshot  = "Page.captureScreenshot"
	CmdPageSetDocumentContent = "Page.setDocumentContent"
	CmdPageGetFrameTree       = "Page.getFrameTree"
)

// DOM Command Constants
const (
	CmdDOMPerformSearch    = "DOM.performSearch"
	CmdDOMGetSearchResults = "DOM.getSearchResults"
	CmdDOMQuerySelector    = "DOM.querySelector"
	CmdDOMGetComputedStyle = "DOM.getComputedStyle"
	CmdDOMGetBoxModel      = "DOM.getBoxModel"
	CmdDOMGetAttributes    = "DOM.getAttributes"
	CmdDOMGetOuterHTML     = "DOM.getOuterHTML"
	CmdDOMDescribeNode     = "DOM.describeNode"
)

// Runtime Command Constants
const (
	CmdRuntimeEvaluate       = "Runtime.evaluate"
	CmdRuntimeCallFunctionOn = "Runtime.callFunctionOn"
)

// Input Command Constants
const (
	CmdInputDispatchKeyEvent = "Input.dispatchKeyEvent"
)

// DOM additional Command Constants
const (
	CmdDOMResolveNode = "DOM.resolveNode"
)

// Network Command Constants
const (
	CmdNetworkGetCookies          = "Network.getCookies"
	CmdNetworkSetCookie           = "Network.setCookie"
	CmdNetworkDeleteCookies       = "Network.deleteCookies"
	CmdNetworkClearBrowserCookies = "Network.clearBrowserCookies"
)

// Target Command Constants
const (
	CmdTargetCreateTarget       = "Target.createTarget"
	CmdTargetGetTargets         = "Target.getTargets"
	CmdTargetAttachToTarget     = "Target.attachToTarget"
	CmdTargetCloseTarget        = "Target.closeTarget"
	CmdTargetSetDiscoverTargets = "Target.setDiscoverTargets"
)

// Page Screencast Command Constants
const (
	CmdPageStartScreencast    = "Page.startScreencast"
	CmdPageStopScreencast     = "Page.stopScreencast"
	CmdPageScreencastFrameAck = "Page.screencastFrameAck"
	CmdPageBringToFront       = "Page.bringToFront"
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
		CmdPageSetDocumentContent,
		CmdPageStartScreencast,
		CmdPageStopScreencast,
		CmdPageScreencastFrameAck,
		CmdPageBringToFront,
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

	// Network Commands - Cookie and network management
	NetworkCommands = []string{
		CmdNetworkGetCookies,
		CmdNetworkSetCookie,
		CmdNetworkDeleteCookies,
		CmdNetworkClearBrowserCookies,
	}

	// Target Commands - Tab management
	TargetCommands = []string{
		CmdTargetCreateTarget,
		CmdTargetGetTargets,
		CmdTargetAttachToTarget,
		CmdTargetCloseTarget,
		CmdTargetSetDiscoverTargets,
	}

	// All Commands - Complete registry
	AllCommands = append(append(append(append(PageCommands, DOMCommands...), RuntimeCommands...), NetworkCommands...), TargetCommands...)
)
