# 01 — File size limits

## The rule

| File type | Max lines |
|---|---|
| `.go` source files (non-test) | **230** |
| `_test.go` test files | **350** |
| Embedded asset files (`report_assets.go`, `report_script.go`) | exempt |
| Generated files | exempt |

When a file exceeds its limit, **split into multiple files within the same package** with meaningful names.

## Why these numbers

- **230 for source** — forces focused, single-purpose files. Easier code review (small diffs, focused scope). Aligns with the project's parallel rule for `.md` files (200) and is consistent with Clean Code-style guidance.
- **350 for tests** — table-driven test cases legitimately grow longer than logic code. 350 gives breathing room without becoming unmanageable.
- **Embedded asset exemption** — files holding raw CSS/JS strings (the HTML report's embedded assets) shouldn't be split; they're effectively configuration data.

## How to split

Use **same-package splits**: multiple files in the same directory share the same `package` name and can see each other's identifiers (types, vars, funcs, helpers). This is Go's natural unit of cohesion — you don't lose anything by splitting a file into multiple files in the same directory.

### Source-file split example

`browser.go` (over-size) →
- `browser.go` — main `Browser` struct, `Launch`, `Close`, `NewPage`
- `browser_pages.go` — `Pages`, `PageCount`, `PageByIndex`, `PageByURL`, `WaitForNewPage`, `CloseAllPagesExcept`
- `browser_stealth.go` — the stealth fingerprint masking JS injection

### Test-file split example

`kapi_test.go` (1029 lines) was actually split into:
- `helpers_test.go` — `newTestServer` helper
- `client_construction_test.go` — `TestNewClient_*`
- `http_methods_test.go` — Get/Post/Put/Delete/Head/Options
- `request_builder_test.go` — `TestRequestBuilder_*`
- `response_test.go` — `TestResponse_*`
- `runtime_test.go` — runtime mutations + error handling + parallel + redirect

Naming: pick the most meaningful grouping — by feature, by section header, or by test-name prefix.

## Same-package visibility (critical for test splits)

When splitting tests within a folder (e.g. `tests/kapi/`), every `_test.go` file shares the same package (e.g. `kapi_test`). All package-level identifiers — helper functions, mock types, sentinel vars — defined in any one file are visible to all other files in the same folder. So when splitting:

- Helper functions stay in one file (e.g. `helpers_test.go`); other files use them directly.
- Mock types like `mockT`, `mockKTestT`, `MockSuite` defined in one file are usable from all other files in the same folder.
- No need to extract helpers into a separate package.

## When NOT to split

- If splitting would create a file under 50 lines (too granular).
- If two pieces of logic are so coupled that reading them apart hurts comprehension more than the size limit hurts. In that case, consider whether they should be merged into the same function or extracted as a higher-level helper, instead of split into separate files.

## Audit checklist

- [ ] No `.go` file (except exemptions) exceeds 230 lines
- [ ] No `_test.go` file exceeds 350 lines
- [ ] Split files use meaningful names (not `_part1.go`, `_part2.go`)
- [ ] Same-package splits don't duplicate helpers
