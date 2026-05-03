# File size audit — 2026-05-03

User asked about `.go` file-size limits and whether any violators exist. Quick audit answers both.

## The rule

Per `.kb/collab_coding_rules/01_FILE_SIZE.md`:

| File type | Max lines |
|---|---|
| `.go` source files (non-test) | 230 |
| `_test.go` test files | 350 |
| Embedded asset files (`report_assets.go`, `report_script.go`) | exempt |

## Audit result

**Zero violators.** Every `.go` file in the kexas codebase (excluding samplings + tests + embedded assets) fits its respective limit.

The audit script (Python AST-style line counting):

```python
for root, dirs, files in os.walk('.'):
    skip = ('.git', 'test-results', '.kexas', 'samplings/')
    if any(s in root for s in skip): continue
    for f in files:
        if not f.endswith('.go'): continue
        is_test = f.endswith('_test.go')
        is_asset = 'report_assets' in f or 'report_script' in f
        if is_asset: continue
        limit = 350 if is_test else 230
        with open(...) as fh: n = len(fh.readlines())
        if n > limit: report(...)
# Result: 0 violators
```

This is a pleasant side effect of consistently splitting files when adding methods (e.g., the recent `recorder_ffmpeg_install.go`, `page_find_all.go`, `page_wait.go` additions all added new files rather than growing existing ones).

## What about `.md` files? (Rule 06)

Rule 06 caps `.md` files at 200 lines. The new `.kb/` docs added during the recent dogfood marathons (FFMPEG_AUTO_INSTALL.md, RULE_02_TIGHTENING.md, etc.) all comply. **However**, several pre-Rule-06 docs in `.kb/browsers/` and `.kb/codebase/infras/` exceed 200 lines:

| File | Lines |
|---|---|
| `.kb/codebase/infras/TESTING.md` | 671 |
| `.kb/codebase/infras/DATA_FLOW.md` | 578 |
| `.kb/codebase/infras/DESIGN_PATTERNS.md` | 552 |
| `.kb/codebase/infras/KEY_CONCEPTS.md` | 509 |
| `.kb/codebase/mod_interactions/detailed/INTERACTIONS.md` | 490 |
| `.kb/codebase/infras/DEVELOPMENT.md` | 481 |
| `.kb/codebase/infras/CORE_COMPONENTS.md` | 468 |
| `.kb/browsers/AGENT_ARCHITECTURE_DEEP_DIVE.md` | 403 |
| ... 12 more in `.kb/browsers/` and `.kb/codebase/` |

These predate Rule 06 (added 2026-05-02). Not blocking — they're pre-existing reference docs. Splitting them is a future cleanup task. Out of scope for the current `.go`-focused audit.

## Recommendation

- **No `.go` action needed today.**
- Add an automated check in CI: `find . -name "*.go" | xargs wc -l | awk '$1 > 230 && $1 != "total" && /\\/[^/]*\\.go/ && !/_test\\.go/ {print}'` (with similar for `_test.go` at 350).
- Schedule a future pass for the pre-Rule-06 `.md` docs in `.kb/browsers/` and `.kb/codebase/`.
