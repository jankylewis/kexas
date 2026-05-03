# Package rename: launcher → klauncher — 2026-05-03

The `launcher/` package was renamed to `klauncher/` for naming consistency with the rest of kexas's public packages (`kcore`, `kapi`, `kassert`, `ktest`, `kwait`).

## What changed

- `launcher/` → `klauncher/` (directory + package decl)
- `tests/launcher/` → `tests/klauncher/` (test directory)
- All import paths: `github.com/jankylewis/kexas/launcher` → `github.com/jankylewis/kexas/klauncher`
- All package qualifier sites: `launcher.Launch`, `launcher.Options`, etc. → `klauncher.Launch`, `klauncher.Options`, etc.

Files touched (10 source + 3 test sites + the package itself):

- `options.go` (root re-export)
- `kcore/browser.go`
- `ktest/ktest_main.go`
- `ktest/ktest_autorun.go`
- `ktest/ktest_parallel_launch.go`
- `tests/kcore/browser_test.go`
- `tests/kcore/page_find_strict_test.go`
- `tests/klauncher/launcher_test.go` (also renamed `package launcher_test` → `package klauncher_test`)
- All files inside `klauncher/` (10 files: 1 package decl change each)

## Why now

Discovered during the kapi-samplings audit (Q1 of 2026-05-03 session) that every kexas public package uses the `k` prefix except `launcher`. Single inconsistency stood out. Rename is breaking for any external user, but kexas hasn't been published anywhere yet — internal consistency wins.

## Verification

All kexas unit tests + 4 samplings projects (sltests/wikitests/hntests/ghtests) green after the rename.

## Migration mechanics

For any future external user:

```bash
# Find call sites
grep -r "kexas/launcher" .
grep -r "launcher\.\(Launch\|Options\|Browser\)" .

# Mass-replace (macOS sed)
find . -name "*.go" | xargs sed -i '' \
  -e 's|github.com/jankylewis/kexas/launcher|github.com/jankylewis/kexas/klauncher|g' \
  -e 's|launcher\.|klauncher.|g' \
  -e 's|kklauncher\.|klauncher.|g'  # cleanup if any double-k slipped in
```

## No deprecation alias

There is no `launcher` shim. Users must update import paths in their own code. Pre-publication, this is fine; once kexas hits a public release, future renames should ship with a one-version-deprecation alias instead.
