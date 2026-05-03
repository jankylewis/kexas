# 06 — Markdown file constraints

## The rule

**Every `.md` file: <200 lines. Filename stem: ALL_UPPERCASED.**

## Scope

All project `.md`. **Exempt:** `README.md`, `LICENSE.md`, `.kb/collab_coding_rules/INDEX.md`, `.claudoholic/ctx_summarizing/**/SUMMARY.md`, `.kb/codebase/**/*.md` (architectural reference docs), `.kb/browsers/*.md` (article-grade browser/CDP docs).

## Why

- 200 lines is roughly a single concept the reader can hold in their head without scrolling-to-rebuild context. Larger docs become reference manuals nobody reads end-to-end.
- ALL_UPPERCASED stems make titles glanceable in any file listing and match the project's house style.

## How to apply

- New files: comply from day one. Over 200 lines → split into multiple meaningful `.md` files (semantic split, not blind line-range slicing).
- Lowercased filenames → rename via `git mv lower.md UPPER.md`.
