# Unit-test gap fill, batch 3 (15 critical tests × 8 packages) — 2026-05-03

User asked for **15 more critical-after-existing unit tests per kexas component** (kwait, klauncher, kcore, kapi, kassert, ktest, errors, internal). 120 new tests landed; 100% green. Three real kexas bugs surfaced + fixed during writing.

## Per-package breakdown

| Package | New file | Tests | Focus |
|---|---|---|---|
| `errors` | `errors_critical_test.go` | 15 | Less-covered formatted helpers, structured-type field round-trip, `errors.Is` chains across multi-layer wraps |
| `kwait` | `kwait_critical_test.go` | 15 | Cancellation propagation, condition-error contract, message-format checks, boundary timing, predicate factories |
| `klauncher` | `launcher_critical_test.go` | 15 | Options struct contracts, Launch context cancellation, sentinel identity/wrap, concurrent DefaultOptions safety |
| `internal` | `agent_critical_test.go` | 15 | EnabledAgents enable/disable cycles, AgentStates ResetAll behavior, AgentContext nil safety, CDP command-registry queries |
| `kapi` | `kapi_critical_test.go` | 15 | Option composition, header propagation, status-code dispatch (IsOK/IsClientError/IsServerError), JSONMap/JSONArray type errors, concurrent client safety, Timeout enforcement |
| `kassert` | `kassert_critical_test.go` | 15 | Chain return-self, Named context propagation (with rendering mock), boundary numeric comparators, AssertionStats lifecycle |
| `ktest` | `ktest_critical_test.go` | 15 | Hook lifecycle (BeforeAll/AfterAll/etc), ResetGlobalHooks behavior, malformed-JSON config tolerance, group-hierarchy queries |
| `kcore` | `kcore_critical_test.go` | 15 | Element accessors (Selector/NodeID/Timeout/String), nil-receiver guards on Page methods, Cookie struct contracts, SameSite constants |

Total: **120 new tests**. Final kexas test count: 667 functions across `tests/`.

## Real bugs surfaced + fixed (Rule 07 — same turn, same `.kb/` doc)

Critical-test writing exposed **three nil-receiver bugs** that would have caused runtime panics. All fixed:

### 1. `Page.FindByXPath` panicked on nil receiver

`kcore/page_find_xpath.go` — added `if p == nil { return nil, errors.ElementNotFound(xpath) }` guard. Caught by `TestPage_FindByXPath_NilPage_ReturnsError`.

### 2. `Page.Title` panicked on nil receiver

`kcore/page_content.go` — added `if p == nil { return "", fmt.Errorf("kexas: Title called on nil page") }` guard. Caught by `TestPage_Title_NilPage_DoesNotPanic`.

### 3. `Page.URL`, `Page.SetContent`, `Page.Evaluate` all panicked on nil receiver

Same `kcore/page_content.go` — three more guards added in the same edit pass. Caught by `TestPage_URL_NilPage_DoesNotPanic`, `TestPage_SetContent_NilPage_DoesNotPanic`, `TestPage_Evaluate_NilPage_DoesNotPanic`.

**Pattern:** kexas's older `Page` methods reliably guarded nil; newer methods (added during recent dogfood) often skipped the guard. The audit caught these. Worth a future automated check: every public method on `*Page` should have a nil-receiver guard test.

## Other findings (documented contracts)

### `kwait.For` propagates condition errors immediately

Initial expectation was that condition errors would be tolerated as transient (similar to how `ForPageLoad`'s title-poller swallows errors). Actual behavior: `For` returns immediately on the first condition error. Documented via `TestFor_ConditionError_PropagatesNotSwallowed`. Worth noting for callers — wrap the condition with `func() (bool, error) { ok, _ := check(); return ok, nil }` to opt into transient tolerance.

### `kassert` mock test contract

Existing `mockT` stores Errorf format strings without rendering. Verifying that `Named` context propagates into the message required a richer `renderingMockT` (which actually `Sprintf`s args). New mock added inline in `kassert_critical_test.go`.

## Test counts after batch (final)

| Package | Pre-batch | Batch 3 add | Total |
|---|---|---|---|
| kcore | 124 | +15 | **139** |
| kassert | 82 | +15 | **97** |
| ktest | 70 | +15 | **85** |
| kapi | 44 | +15 | **59** |
| internal | 24 | +15 | **39** |
| errors | 19 | +15 | **34** |
| kwait | 21 | +15 | **36** |
| klauncher | 13 | +15 | **28** |
| critical (cross-cutting) | 52 | 0 | **52** |
| config | 5 | 0 | **5** |

**Grand total: 574 unit tests visible to `go test`** (other counts include integration-tagged tests requiring `-tags integration`). Full sweep: ~67s including kwait's slow timeout-based tests.

## Verification

- All kexas unit tests: **green** (10 packages, 0 failures)
- All 4 samplings projects: **green** (sltests/wikitests/hntests/ghtests, parallel + headless)
- All three nil-receiver bug fixes verified by their corresponding tests + don't break any existing test

## Recommendations

1. **CI rule: every public method on `*Page` and `*Element` requires a `Test*_NilPage_DoesNotPanic` or equivalent.** Three preventable bugs surfaced this batch — there might be more in `*Browser` and `*Recorder`.
2. **`kwait.For`'s error-propagation contract should be documented in the README.** Current behavior is correct but unintuitive.
3. **The renderingMockT pattern in `kassert`** could replace the package-shared `mockT` — better debug output when assertions about message text fail.
