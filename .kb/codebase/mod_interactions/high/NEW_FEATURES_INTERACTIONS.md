# New Feature Module Interactions — High-Level

---

## How New Features Connect to Existing Modules

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│    kapi/     │────▶│   kexas/     │────▶│  internal/   │
│  flow.go     │     │  cookie.go   │     │  cdp/const.go│
│  patterns.go │     │  storage.go  │     │  cdp/client  │
│              │     │  recorder.go │     │  agent/      │
│              │     │  browser.go  │     │              │
│              │     │  page.go     │     │              │
└─────────────┘     └──────────────┘     └──────────────┘
                           │                     │
                           ▼                     ▼
                    ┌──────────────┐     ┌──────────────┐
                    │   errors/    │     │   launcher/  │
                    │  errors.go   │     │  browser.go  │
                    └──────────────┘     └──────────────┘
```

## Feature → Module Dependencies

| Feature | Files Modified/Created | Depends On |
|---------|----------------------|------------|
| **Cookie** | `cookie.go` (new) | `page.sendCommand`, `cdp.CmdNetwork*`, `agent.AgentNetwork`, `errors.ErrCookie*` |
| **Storage** | `storage.go` (new) | `page.Evaluate` (Runtime.evaluate), `errors.ErrStorage*` |
| **Multi-Tab** | `browser.go` (modified), `page.go` (modified) | `cdp.CmdTarget*`, `cdp.CmdPageBringToFront`, `agent.AgentPage`, `errors.ErrPage*` |
| **Recorder** | `recorder.go` (new) | `page.sendCommand`, `page.Screenshot`, `cdp.CmdPageScreencast*`, `logger`, `os/exec` (ffmpeg) |
| **KAPI** | `kapi/flow.go`, `kapi/patterns.go` (new package) | `kexas.Browser`, `kexas.Page`, `kexas.Element`, `kexas.Cookie`, `kexas.Storage`, `logger` |

## Cross-Feature Interactions

| Interaction | Description |
|-------------|-------------|
| **Cookie ↔ Multi-Tab** | Cookies are browser-level; all tabs share cookies for the same domain |
| **Storage ↔ Multi-Tab** | localStorage shared across tabs; sessionStorage isolated per tab |
| **Recorder ↔ Page** | Recorder captures from one page; switching tabs changes what's recorded |
| **KAPI ↔ All** | KAPI wraps cookie, storage, page, element, and browser operations |
| **Agent Manager ↔ Cookie** | Network agent auto-enabled when cookie commands are sent |
| **Agent Manager ↔ Recorder** | Page agent auto-enabled when screencast commands are sent |

## Data Flow: Cookie Set Example

```
User code: page.SetCookie(cookie)
  → cookie.go: validate name, build params
    → page.go: sendCommand(CmdNetworkSetCookie, params)
      → page.go: ensureAgentsForCommand("Network.setCookie")
        → agent_manager.go: EnsureAgent(AgentNetwork)
          → cdp_client: Send("Network.enable", {})  [if not already enabled]
      → cdp_client: Send("Network.setCookie", params)
        → WebSocket → Chrome → WebSocket response
      ← parse response, check success
    ← return nil or error
  ← return nil or wrapped error
```

## Data Flow: KAPI Chain Example

```
User code: kapi.Open(url).Find("#btn").Click()
  → flow.go: Open(url)
    → kexas.Launch(nil) → browser
    → browser.FirstPage() → page
    → page.Navigate(url)
  → flow.go: Find("#btn")
    → page.Find("#btn") → element (with 7s auto-retry)
    → flow.element = element
  → flow.go: Click()
    → element.Click() → CDP Input.dispatchMouseEvent
  ← *Flow (errors accumulated internally)
```
