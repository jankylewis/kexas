# Kexas Configuration File

## Overview

Kexas supports a `kexas.config.json` file at the project root for configuring test execution behavior. This is similar to Playwright's `playwright.config.ts`.

## Location

Place the config file at the root of your project:

```
your-project/
├── kexas.config.json  ← here
├── go.mod
├── tests/
└── ...
```

## Configuration Options

### `headless` (boolean)
- **Default:** `true`
- **Description:** Run browser in headless mode (no visible window)
- **Example:** `"headless": false` to see the browser

### `timeout` (number)
- **Default:** `30000` (30 seconds)
- **Description:** Maximum time in milliseconds for each test
- **Example:** `"timeout": 60000` for 60 seconds

### `retries` (number)
- **Default:** `0`
- **Description:** Number of times to retry failed tests
- **Example:** `"retries": 2` to retry twice

### `parallel` (boolean)
- **Default:** `false`
- **Description:** Run tests in parallel
- **Example:** `"parallel": true` to enable parallel execution

### `screenshotOnFail` (boolean)
- **Default:** `true`
- **Description:** Automatically capture screenshot when test fails
- **Example:** `"screenshotOnFail": false` to disable

### `screenshotDir` (string)
- **Default:** `"./test-results/screenshots"`
- **Description:** Directory to save screenshots
- **Example:** `"screenshotDir": "./screenshots"`

### `videoDir` (string)
- **Default:** `"./test-results/videos"`
- **Description:** Directory to save videos (future feature)
- **Example:** `"videoDir": "./videos"`

### `slowMo` (number)
- **Default:** `0`
- **Description:** Slow down operations by specified milliseconds (for debugging)
- **Example:** `"slowMo": 100` to add 100ms delay

### `baseURL` (string)
- **Default:** `""`
- **Description:** Base URL for relative navigation
- **Example:** `"baseURL": "https://example.com"`

### `browserExecutable` (string)
- **Default:** `""` (auto-detect)
- **Description:** Path to custom browser executable
- **Example:** `"browserExecutable": "/path/to/chrome"`

## Example Configuration

```json
{
  "headless": false,
  "timeout": 60000,
  "retries": 2,
  "parallel": false,
  "screenshotOnFail": true,
  "screenshotDir": "./test-results/screenshots",
  "videoDir": "./test-results/videos",
  "slowMo": 0,
  "baseURL": "https://staging.example.com",
  "browserExecutable": ""
}
```

## Usage

### Automatic Loading

When you use `ktest.Run()`, the config file is automatically loaded:

```go
func TestAmazon(t *testing.T) {
    ktest.Run(t, &AmazonSuite{})  // Loads kexas.config.json automatically
}
```

### Manual Override

You can still override config programmatically:

```go
func TestAmazon(t *testing.T) {
    var config *ktest.Config = ktest.DefaultConfig()
    config.Headless = false  // Override config file
    
    ktest.RunWithConfig(t, &AmazonSuite{}, config)
}
```

## Fallback Behavior

If `kexas.config.json` is not found, ktest uses default values:
- `headless: true`
- `timeout: 30s`
- `retries: 0`
- `screenshotOnFail: true`
- etc.

## Environment-Specific Configs

You can create different config files for different environments:

```bash
# Development
cp kexas.config.dev.json kexas.config.json

# CI/CD
cp kexas.config.ci.json kexas.config.json
```

## Notes

- Config file is optional - defaults work out of the box
- JSON format only (no comments allowed in JSON)
- All fields are optional - omitted fields use defaults
- Programmatic config via `RunWithConfig()` takes precedence over file
