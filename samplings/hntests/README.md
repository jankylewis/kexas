# hntests

Sample test-automation project against Hacker News (`https://news.ycombinator.com`), built on `kexas`. Three critical tests against the front page + comment thread.

## Why HN as a sample target

- Pure HTML, no JS framework, no telemetry, no anti-bot. Selectors unchanged since 2007.
- Most contrasting to saucedemo: no React, no localStorage tricks, no flaky timing.
- Simplest possible "real production website" — what your tests look like when the SUT isn't fighting you.

## Run

```bash
cd samplings/hntests
go test -count=1 -timeout=60s ./tests/...
```

Outputs land in `test-results/`: `report.html`, `screenshots/`, `videos/<TestName>.gif`.

## Tests

| # | Test | Verifies |
|---|---|---|
| 1 | `TestAFrontPageLoads` | Front page returns with the canonical "Hacker News" tab title — anchor for every other test |
| 2 | `TestBFrontPageHas30Stories` | Front page renders exactly 30 stories (`.titleline`), unchanged for 15+ years |
| 3 | `TestCNewestPageHasDifferentStories` | `/newest` URL returns a page whose first-story title differs from the front page's first — proves the route + ordering work |
