# ghtests

Sample test-automation project against GitHub public repository pages, built on `kexas`. Three critical tests against the `golang/go` repository.

## Why GitHub as a sample target

- Real-world complex modern web app — closer to what most production apps look like today (vs HN's static HTML or Wikipedia's MediaWiki skin).
- Anonymous public-page browsing is rate-limited but not actively blocked. No Cloudflare/CAPTCHA challenges.
- Selectors evolve (active product) — useful real-world maintenance challenge for the test suite.

## Run

```bash
cd samplings/ghtests
go test -count=1 -timeout=60s ./tests/...
```

Outputs land in `test-results/`: `report.html`, `screenshots/`, `videos/<TestName>.gif`.

## Tests

| # | Test | Verifies |
|---|---|---|
| 1 | `TestAGolangGoRepoLoadsWithCorrectTitle` | The `golang/go` repo page returns with a `<title>` containing both "golang" and "Go" |
| 2 | `TestBRepoHasStarCounter` | The repo's star counter (`#repo-stars-counter-star`) renders with a non-empty value |
| 3 | `TestCIssuesPageOpensWithIssuesList` | `/golang/go/issues` opens and shows at least one issue row — exercises navigation to a sub-page |

## Maintenance note

GitHub redesigns frequently. If a test breaks, the most likely cause is a selector change: check the current GitHub HTML for the new `id` / `data-test-id` and update the constants in `pages/repo_page.go`.
