# Page.FindAll — 2026-05-03

Closes the biggest API gap surfaced by the samplings dogfood: kexas's `Page.Find` is strict-single-match (errors on >1), and the only escape hatches were `FindByXPath` with `(...)[1]` for "first match" or dropping into raw `Page.Evaluate("document.querySelectorAll(...)")` JS for anything multi-element. `FindAll` returns `[]*Element` for any matching count.

## API

```go
func (p *Page) FindAll(selector string) ([]*Element, error)
```

Selector type auto-detected the same way as `Find`: leading `//` or `(` → XPath, anything else → CSS. ID-shortcut form (`#x`) goes through CSS.

Behavior:

- Returns matched elements in **document order**.
- Empty selector match returns `([]*Element{}, nil)` — not an error. Caller uses `len() == 0` to check.
- Does **not** auto-wait. Unlike `Find`'s 7s retry loop, `FindAll` is one-shot. For "wait for at least N to be present", combine with `WaitForElement` on a parent or use `Page.Evaluate` directly.
- Each returned `*Element` is bound to a Runtime objectId only (no nodeID). Methods that depend on nodeID (rare — most things use objectID dispatch) need the objectID-fallback path.

## Implementation

`kcore/page_find_all.go` — three-step flow:

1. **Build a JS array of matching nodes** via `Runtime.evaluate`:
   - CSS: `Array.from(document.querySelectorAll(<sel>))`
   - XPath: `document.evaluate(<xpath>, document, null, 7, null).snapshotItem(i)` for `i` in `0..snapshotLength` (constant 7 = `XPathResult.ORDERED_NODE_SNAPSHOT_TYPE`)
2. **Enumerate the array's properties** via `Runtime.getProperties` with `ownProperties: true`. Returns descriptors for indices `0`, `1`, `2`, … plus `length` and prototype noise.
3. **Filter to numeric-named properties** and pull each `value.objectId`. Sort by index (CDP ordering not guaranteed). Wrap each into `*Element` via `NewElementWithObject`.

CDP plumbing: added `CmdRuntimeGetProperties = "Runtime.getProperties"` constant + registered in `RuntimeCommands` slice (`internal/cdp/const.go`).

## Samplings tests added

| Project | Test | Selector | Asserts |
|---|---|---|---|
| hntests | `TestNFindAllTitleLinesReturns30Elements` | CSS `.titleline` | `len == 30`; `GetText` non-empty on each |
| ghtests | `TestNFindAllOcticonReturnsManySVGs` | CSS `.octicon` | `len > 20`; `GetAttribute("class")` works on every SVG (~149) |
| ghtests | `TestOFindAllByXPathFindsAnchorsInsideNav` | XPath `//nav//a` | `len >= 5` (multiple navs on GH) |
| ghtests | `TestPFindAllOnNoMatchReturnsEmptySliceNotError` | CSS `.kexas-totally-fake-class-xyz` | `err == nil`, `len == 0` |
| wikitests | `TestNFindAllParagraphsInArticleViaXPath` | XPath that previously matched 8 (now via FindAll) | `len > 1`; first paragraph contains "Go" (proves document order) |
| wikitests | `TestOFindAllByCSSReturnsExternalLinksList` | CSS `.mw-parser-output ul li` | `len >= 5` |

All pass. The HN test takes ~3.7s because each `GetText` is a CDP roundtrip (30 elements × ~120ms = ~3.6s). Worth a future optimization — possibly batch GetText via `Runtime.callFunctionOn` over the original array — but a separate concern from FindAll itself.

## Bug surfaced during verification

GitHub's repo nav doesn't use `role="tablist"` (the original selector). Switched to the more robust `//nav//a`. Not a kexas issue — selector-design problem only.

## Related findings (still open)

- `Find` waits the full 7s timeout on no-match. `FindAll` doesn't have this issue (one-shot), so users can use `len(FindAll(sel)) == 0` as a fast no-match check.
- `WaitForElementAll(selector, timeout)` would be the next natural addition — wait until at least one element matches, then return all. Skipping for now; combining `WaitForElement` (single) + `FindAll` covers the common case.

## Recommendations for next iteration

1. **Element batch-text via single CDP call** — `Page.GetTextAll(selector)` could return `[]string` in one roundtrip, much faster than iterating `FindAll` + `GetText`.
2. **Update README** to point users at `FindAll` for any multi-element flow. Today's docs (and the strict-Find error message) don't mention it.
3. **Consider deprecating the strict-Find error message text** — replacing "matched N elements, expected exactly 1" with "matched N elements; use FindAll if you want all of them".
