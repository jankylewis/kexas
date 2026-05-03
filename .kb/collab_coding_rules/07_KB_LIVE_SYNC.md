# 07 — KB live-sync from Claude memory

## The rule

Every new finding, solution, implementation, behavior quirk, or design decision saved into Claude Code's session memory **must be mirrored into `.kb/`** in the same turn — no batching, no end-of-day rollups.

If you wouldn't write it down, don't save it to memory either.

## Why

Memory is per-Claude, per-session, opaque to humans. `.kb/` is the project's durable knowledge — readable by humans, by future Claude sessions starting fresh, by anyone reading the repo. Findings that live only in memory are invisible to maintainers and lost when memory is wiped or compacted.

Throughout 2026-05-02 / 05-03 dogfood, several useful findings (strict-Find, WaitForLoadState polls title, x264 odd-dimensions) were nearly missed because they sat in conversation context for hours before getting written down. This rule closes that window.

## How to apply

When working in this repo and Claude detects something that would normally trigger a memory save:

1. **Save the memory file** as usual (preserves cross-session continuity).
2. **In the same turn**, write or update a `.kb/<TOPIC>.md` capturing the same insight in human-readable form.
3. Both writes happen before the assistant's reply ends.

The `.kb/` doc and the memory entry should reference each other where useful — e.g., the memory description line can name the `.kb/` doc it mirrors.

## What counts as "new finding"

- A kexas API behavior worth documenting (e.g., strict-single-match Find, WaitForLoadState quirks).
- A bug surfaced + the fix or workaround applied.
- A design decision made during implementation (why we chose A over B).
- A user preference or feedback that reshapes how kexas should work (already covered by feedback memories — those still go into `.kb/` too if technical).
- A new external resource or convention adopted (e.g., evermeet.cx as ffmpeg static-build source).

## What does NOT need a `.kb/` mirror

- User profile facts (`user_role.md`) — pure memory.
- Pointers to chat-session captures (`reference_chat_sessions.md`) — already an index.
- Project-state-of-the-day captures (`project_*.md`) — those decay quickly; recreate from `git log` / current code.
- In-conversation TODOs and ephemeral state (use TaskCreate, not memory or `.kb/`).

## Naming + placement

- Existing `.kb/<TOPIC>.md` doc on the same topic? **Update it** rather than create a new one. Avoids fragmentation.
- Brand-new topic? **New file**, ALL_UPPERCASED stem (Rule 06), ≤200 lines (Rule 06).
- Subtopic of a larger area? Place under `.kb/<area>/<TOPIC>.md` — e.g., `.kb/browsers/SCROLL_ACTIONS.md`, `.kb/codebase/infras/DESIGN_PATTERNS.md`.

## Audit checklist

- [ ] Is there a memory entry without a corresponding `.kb/` doc? → either write the doc or delete the memory.
- [ ] Is `.kb/` updated for every kexas API addition / bug fix / design decision in the latest commits?
- [ ] When user says "what did we change", can the answer come from `.kb/` alone (no need to ask Claude to recall from memory)?
