# 04 — Naming for acronyms

## The rule

**True acronyms are ALL CAPS. Abbreviations are camelCase.**

This is the Go community convention (Effective Go) and the existing house style throughout kexas.

## What's an "acronym"

A **true acronym**: each letter stands for a separate word.

| Acronym | Stands for |
|---|---|
| `DOM` | Document Object Model |
| `URL` | Uniform Resource Locator |
| `HTTP` | HyperText Transfer Protocol |
| `ID` | IDentifier |
| `JSON` | JavaScript Object Notation |
| `CSS` | Cascading Style Sheets |
| `API` | Application Programming Interface |
| `WS` | WebSocket |
| `JS` | JavaScript |
| `XML` | eXtensible Markup Language |
| `RPC` | Remote Procedure Call |
| `TCP` / `UDP` | Transmission/User Datagram Protocol |
| `CDP` | Chrome DevTools Protocol |

These get **ALL CAPS** wherever they appear in identifiers (preserving exported/unexported case at the start of unexported names — see "leading acronym" section).

## What's an "abbreviation"

An **abbreviation / shortening**: one word with letters dropped.

| Abbreviation | Long form |
|---|---|
| `Cmd` | Command |
| `Ctx` | Context |
| `Doc` | Document |
| `Tmp` | Temporary |
| `Cfg` | Config |
| `Ptr` | Pointer |
| `Idx` | Index |
| `Buf` | Buffer |

These get **camelCase** — first letter for the position (lower if unexported, upper if exported), rest lowercase.

## Examples

```go
// CORRECT
CmdDOMPerformSearch          // Cmd (abbr) + DOM (acronym) + PerformSearch
CmdRuntimeEvaluate           // Cmd (abbr) + Runtime (regular word) + Evaluate
extractPortFromWSURL         // extract + Port + From + WS + URL — both acronyms ALL CAPS
WithHTTPClient               // With + HTTP (acronym) + Client (regular)
parseJSONConfig              // parse + JSON (acronym) + Config
userID, nodeID, sessionID    // user/node/session (regular) + ID (acronym)
ctxKey                       // ctx (abbr) + Key (regular)
tmpDir                       // tmp (abbr) + Dir (regular)
```

```go
// WRONG
CMDDOMPerformSearch          // CMD treats Cmd as if acronym (Cmd is abbr, not acronym)
CmdDomPerformSearch          // Dom treats DOM as word (DOM is acronym → ALL CAPS)
ParseJsonConfig              // Json treats JSON as word (JSON is acronym)
UserId                       // Id treats ID as abbreviation (ID is acronym)
TmpDIR                       // DIR treats dir as acronym (dir is regular word)
```

## Leading acronym in unexported names

Lowercase the WHOLE acronym:

```go
htmlParser    // not HTMLParser (would be exported) or hTMLParser (broken)
jsonDecoder   // not JSONDecoder (exported) or jSONDecoder (broken)
urlBuilder    // not URLBuilder (exported)
```

## Why option 1 (and not the alternatives)

| Option | Verdict |
|---|---|
| **`CmdDOMPerformSearch`** ✅ | This rule. Matches Go community convention. Already used throughout kexas (`CmdDOMQuerySelector`, `kapi.WithHTTPClient`, `extractPortFromWSURL`, `cdp.NodeID`, `AgentDOM`). |
| `CMDDOMPerformSearch` ❌ | Stacks acronyms unreadably ("CMDDOM" runs together). Treats `Cmd` as acronym (it's not). |
| `CmdDomPerformSearch` ❌ | Treats acronym as word. Java/C# style; non-idiomatic Go. |

## Identifying boundary cases

Some terms are ambiguous. Decide once and document:

- `Os` (operating system)? — In kexas, treated as a regular word: `osName`, `OSPath`. **In doubt, lean ALL CAPS** if you'd say it letter-by-letter ("O-S"); camelCase if you'd say it as a word ("os" / "oss").
- `Ip` (internet protocol)? — ALL CAPS: `IPAddress`, `serverIP`.

## Audit checklist

- [ ] All acronyms (DOM, URL, HTTP, ID, JSON, CSS, API, WS, JS, XML, RPC, TCP, UDP, CDP) appear ALL CAPS in identifiers
- [ ] All abbreviations (Cmd, Ctx, Doc, Tmp, Cfg, Ptr, Idx, Buf) appear camelCase
- [ ] Unexported names with leading acronym lowercase the entire acronym (`htmlParser`, not `hTMLParser`)
- [ ] Multi-acronym names don't run together unreadably (use a regular word in between if needed)
