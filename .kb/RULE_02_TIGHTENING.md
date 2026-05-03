# Rule 02 tightening 50 → 40 logical lines + mass refactor — 2026-05-03

## What changed

Rule 02 (`.kb/collab_coding_rules/02_FUNCTION_SIZE.md`) now caps function bodies at **40 logical lines**, down from 50. INDEX.md updated to match.

## Refactored: 18 violators → 0

Survey found 18 non-test functions over 40 logical lines. All refactored, kexas unit tests + all 4 samplings projects still 100% green.

| File | Function | Old (lines) | Approach |
|---|---|---|---|
| `klauncher/launcher_chromium.go` | `findChromium` | 42 | Split into `tryCachedChromium`, `tryDownloadChromium`, `findSystemChromium`, `systemChromiumCandidatePaths` |
| `klauncher/launcher_chromium.go` | `buildArgs` | 50 | Split into `generateUniqueUserDataDir`, `resolveWindowSize`, `baseChromeArgs` |
| `klauncher/launcher_url.go` | `extractDebuggerURLFromPipe` | 46 | Split into `scanPipeForDebuggerURL`, `classifyPipeLine`; new type `pipeReadResult` |
| `klauncher/launcher_run.go` | `Launch` | 46 | Extracted `launchChromiumProcess` for the post-resolution wiring |
| `internal/agent/agent_states.go` | `ResetAll` | 50 | Single `resetState(s *AgentState)` helper called per agent |
| `internal/agent/agent_context.go` | `GetContext` | 46 | Generic `nilOrTyped[T]` helper folds nil-check + typed return |
| `internal/agent/smart_enablement.go` | `EnsureAgent` | 41 | Split into `touchAgentLastUsed`, `sendEnableCommand`, `markAgentEnabled` |
| `internal/cdp/cdp_send.go` | `SendToSession` | 47 | Split into `registerPending`, `unregisterPending`, `marshalAndWrite`, `awaitResponse` |
| `kwait/kwait.go` | `ForPageLoad` | 48 | Extracted predicate factories `titleChangedFromNewTab`, `titleStableAndSet`, `waitForNetworkIdleViaTitle`; `newPageLoadOptions` for opts construction |
| `kcore/page_navigate.go` | `Navigate` | 48 | Extracted `waitForNavigationComplete` and `pollReadyState` |
| `kcore/browser_pages.go` | `firstPageAttempt` | 45 | Split into `fetchTargetInfos`, `extractPageTargetID`, `attachAndRegisterPage` |
| `kcore/cookie.go` | `SetCookie` | 43 | Extracted `buildSetCookieParams` |
| `kcore/element_interaction.go` | `WaitAndClickFor` | 42 | Extracted `checkClickArgs` + `dispatchClickViaObjectID`; **bonus refactor** of sibling `Click` and `WaitAndClick` to use the same helpers (eliminates triplicate code) |
| `kcore/recorder_gif.go` | `SaveAnimatedGIF` | 41 | Split into `framesToGIF` + `writeGIFToPath` |
| `kcore/page_find.go` | `Find` | 41 | Extracted `dispatchFind` (selector-type routing) and `logFindTimeout` (page-title diagnostic) |
| `kcore/page_find_all.go` | `wrapArrayPropertiesAsElements` | 46 | Split into `collectIndexedObjectIDs`, `insertionSortByIndex`, `objectIDsToElements`; new type `indexedObjectID` |
| `ktest/ktest_config.go` | `applyJSONConfig` | 47 | Split into `applyExecutionFlagsFromJSON`, `applyArtifactPathsFromJSON`, `applyBrowserOptionsFromJSON` |
| `ktest/ktest_impl.go` | `runSingleTest` | 44 | Extracted `runWithRetries` (the retry-loop bookkeeping) |

## Why 40

The user's call. 50 was the previous cap; tightening to 40 forces helpers earlier and rewards extraction. Most functions naturally fit ≤30; the 40 ceiling targets the orchestration / dispatch layers where 30 is genuinely too tight.

## Bonus wins from the refactor

- **`Click` / `WaitAndClick` / `WaitAndClickFor` deduplicated.** All three had the same arg-validation block + the same CDP click dispatch. Now they share `checkClickArgs` + `dispatchClickViaObjectID`. The original three were 51 / 57 / 62 lines of largely identical code; now 8 / 12 / 13 lines respectively, with the shared helpers at ~20 lines each.
- **`agent_states.go` `ResetAll`** went from 50 lines of repeated `if as.X != nil { as.X.Enabled = false; ...; }` blocks to 9 lines + a 7-line nil-safe `resetState(s *AgentState)` helper.
- **Generic `nilOrTyped[T]`** in `agent_context.go` collapsed 8 near-identical case branches into a single helper call per case.

## What didn't get refactored

`attachToPage` in `kcore/browser_pages.go` is 103 raw lines but ≤40 logical lines — a giant embedded JS string literal (the stealth-injection script) is exempt per Rule 02's embedded-string clause. Left as-is.

## Verification

All kexas unit tests pass (10 packages, 0 failures). All 4 samplings projects pass with the post-rename + post-refactor codebase:

| Project | Tests | Runtime |
|---|---|---|
| sltests | 53 (3 packages) | ~95s combined |
| wikitests | 13 + 3 (kapi) = 16 | ~17s |
| hntests | 13 | ~13s |
| ghtests | 13 | ~26s |
