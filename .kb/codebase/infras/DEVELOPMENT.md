# Development Workflow

> Guidelines for contributing to Kexas

**Last Updated:** February 24, 2026

---

## Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/kexas-project/kexas.git
cd kexas
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Verify Setup

```bash
go test ./tests/... -v
go build ./cmd/kexas
```

---

## Adding a New Feature

### Step 1: Plan

Create task file in `.tasks/` with checklist.

**Example**: `.tasks/02242026/PHASE_6_ELEMENT_INTERACTION.md`

```markdown
## Task 1.1: Implement DOM.getDocument
- [ ] Add `getDocument()` method to Page
- [ ] Use CDP command `DOM.getDocument`
- [ ] Return root node ID
- [ ] Handle errors properly

**Status:** Not started
```

---

### Step 2: Design

Update `.kb/` docs with design decisions.

**Example**: `.kb/codebase/ELEMENT_DESIGN.md`

```markdown
# Element Design

## API

```go
type Element struct {
    nodeId int
    page   *Page
}

func (e *Element) Click() error
func (e *Element) Fill(text string) error
```

## CDP Commands

- `DOM.getDocument` - Get root node
- `DOM.querySelector` - Find element
- `Input.dispatchMouseEvent` - Click element
```

---

### Step 3: Implement

Write code following `CODING_RULES.md`.

**Rules**:
- Function size < 80 lines
- Nesting depth < 2 levels
- File size < 500 lines
- Explicit types everywhere
- Errors properly wrapped

**Example**:
```go
// element.go
package kexas

type Element struct {
    nodeID int
    page   *Page
}

func (e *Element) Click() error {
    // Implementation
    return nil
}
```

---

### Step 4: Test

Write unit tests in `tests/`.

**Example**: `tests/element_test.go`

```go
func TestElement_Click_Success(t *testing.T) {
    t.Skip("requires real browser - integration test")
    
    browser := SetupBrowser(t)
    page, _ := browser.NewPage()
    
    err := page.Navigate("https://example.com")
    kassert.ThatError(t, err).IsNil()
    
    element, err := page.QuerySelector("button")
    kassert.ThatError(t, err).IsNil()
    
    err = element.Click()
    kassert.ThatError(t, err).IsNil()
}
```

---

### Step 5: Document

Update `README.md` and examples.

**README.md**:
```markdown
## Element Interaction

```go
element, err := page.QuerySelector("button")
if err != nil {
    log.Fatal(err)
}

err = element.Click()
```

**Example**: `examples/element_interaction.go`

```go
package main

import (
    "log"
    "github.com/kexas-project/kexas"
)

func main() {
    browser, _ := kexas.Launch(&kexas.LaunchOptions{Headless: true})
    defer browser.Close()
    
    page, _ := browser.NewPage()
    page.Navigate("https://example.com")
    
    element, _ := page.QuerySelector("button")
    element.Click()
}
```

---

### Step 6: Verify

Run checks before committing.

```bash
# Run tests
go test ./tests/... -v

# Check for errors
go vet ./...

# Format code
go fmt ./...

# Check diagnostics
go build ./...
```

---

## Code Review Checklist

Before submitting PR:

- [ ] Follows `CODING_RULES.md`
- [ ] Function size < 80 lines
- [ ] Nesting depth < 2 levels
- [ ] File size < 500 lines
- [ ] Explicit types everywhere
- [ ] Errors properly wrapped
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Examples work
- [ ] No `go vet` warnings
- [ ] Code formatted with `go fmt`

---

## Common Tasks

### Running Tests

```bash
# All tests
go test ./tests/... -v

# Specific test
go test ./tests -run TestBrowserLaunch_Success -v

# With coverage
go test ./tests/... -cover

# Integration tests (remove skip statements first)
go test ./tests/... -v -tags=integration
```

---

### Debugging

#### Enable Debug Logging

```go
import "github.com/kexas-project/kexas/internal/logger"

log := logger.New("cdp").WithLevel(logger.LevelDebug)
```

#### View CDP Messages

Debug logs show all CDP commands and responses:

```
2026-02-24T22:00:00+07:00 [DEBUG] [cdp] → Page.navigate {"url":"https://example.com"}
2026-02-24T22:00:01+07:00 [DEBUG] [cdp] ← {"id":1,"result":{"frameId":"..."}}
```

#### Common Issues

**Connection refused**:
- Browser not started or wrong port
- Check if browser process is running
- Verify port 9222 is available

**Timeout**:
- Page load taking too long
- Increase timeout: `page.WaitForLoadState(strategy, 60*time.Second)`
- Try different wait strategy

**Session not found**:
- Page closed or browser crashed
- Check browser logs
- Verify page is still open

---

### Adding CDP Commands

#### Step 1: Find CDP Method

Check [CDP documentation](https://chromedevtools.github.io/devtools-protocol/)

Example: `DOM.querySelector`

#### Step 2: Add Method to Page

```go
func (p *Page) QuerySelector(selector string) (*Element, error) {
    var params map[string]interface{} = map[string]interface{}{
        "nodeId":   rootNodeID,
        "selector": selector,
    }
    
    var result map[string]interface{}
    var err error
    result, err = p.sendCommand("DOM.querySelector", params)
    if err != nil {
        return nil, fmt.Errorf("kexas: querySelector failed: %w", err)
    }
    
    var nodeID int
    var ok bool
    nodeID, ok = result["nodeId"].(int)
    if !ok {
        return nil, fmt.Errorf("kexas: invalid nodeId in response")
    }
    
    return &Element{nodeID: nodeID, page: p}, nil
}
```

#### Step 3: Test

```go
func TestPage_QuerySelector_Success(t *testing.T) {
    t.Skip("requires real browser - integration test")
    
    browser := SetupBrowser(t)
    page, _ := browser.NewPage()
    page.Navigate("https://example.com")
    
    element, err := page.QuerySelector("h1")
    kassert.ThatError(t, err).IsNil()
    kassert.That(t, element).IsNotNil()
}
```

---

### Refactoring Large Files

If file exceeds 500 lines, split it:

**Before**:
```
page.go (600 lines)
  - Navigate
  - Screenshot
  - Click
  - Fill
  - Select
```

**After**:
```
page.go (200 lines)
  - Core Page type
  - Navigate
  - Close

page_screenshot.go (100 lines)
  - Screenshot
  - PDF

page_interaction.go (300 lines)
  - Click
  - Fill
  - Select
```

---

## Git Workflow

### Branch Naming

```
feature/element-interaction
fix/navigation-timeout
docs/update-readme
refactor/split-page-file
```

### Commit Messages

```
feat: add element.Click() method

- Implement DOM.querySelector
- Add Input.dispatchMouseEvent
- Add tests for click interaction

Closes #123
```

**Format**:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Add tests
- `chore:` - Maintenance

---

## Release Process

### 1. Update Version

```go
// kexas.go
const Version string = "0.2.0"
```

### 2. Update Changelog

```markdown
# Changelog

## [0.2.0] - 2026-02-25

### Added
- Element interaction (Click, Fill, Select)
- Wait for selector
- Auto-wait before actions

### Fixed
- Navigation timeout handling
- Screenshot memory limit

### Changed
- Improved error messages
```

### 3. Tag Release

```bash
git tag v0.2.0
git push origin v0.2.0
```

---

## Documentation

### Where to Document

- **README.md**: Quick start, basic usage
- **.kb/**: Detailed guides, architecture
- **Code comments**: Function/type documentation
- **Examples**: Runnable code samples

### Documentation Style

```go
// QuerySelector finds the first element matching the CSS selector.
// Returns ErrElementNotFound if no element matches.
//
// Example:
//
//	element, err := page.QuerySelector("button.submit")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Page) QuerySelector(selector string) (*Element, error) {
    // Implementation
}
```

---

## Summary

Development workflow:
1. **Plan** - Create task file
2. **Design** - Document decisions
3. **Implement** - Follow coding rules
4. **Test** - Write unit tests
5. **Document** - Update docs and examples
6. **Verify** - Run checks

Key commands:
- `go test ./tests/... -v` - Run tests
- `go vet ./...` - Check for errors
- `go fmt ./...` - Format code
- `go build ./...` - Verify compilation

Always:
- Follow `CODING_RULES.md`
- Write tests
- Update documentation
- Run verification before committing

