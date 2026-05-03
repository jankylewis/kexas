# wikitests

Sample test-automation project against English Wikipedia (`https://en.wikipedia.org`), built on `kexas`. Three critical tests exercising search + article navigation.

## Why Wikipedia as a sample target

- No anti-bot, no CAPTCHA, no Cloudflare. Wikimedia explicitly welcomes programmatic access.
- Selectors have been stable for years (Vector skin's `#searchInput`, `#firstHeading`).
- Both anonymous and authenticated paths exist (this suite covers anonymous only).

## Run

```bash
cd samplings/wikitests
go test -count=1 -timeout=60s ./tests/...
```

Outputs land in `test-results/`: `report.html`, `screenshots/`, `videos/<TestName>.gif`.

## Tests

| # | Test | Verifies |
|---|---|---|
| 1 | `TestAMainPageHasSearchInput` | The Vector-skin search input renders on the main page (anchor for every other test) |
| 2 | `TestBSearchForGoLanguageOpensCorrectArticle` | Searching for "Go (programming language)" lands on the article whose H1 matches the query |
| 3 | `TestCRandomArticleHasNonEmptyTitle` | The "Random article" sidebar link navigates to an article with a non-empty H1 — exercises the project's content randomness |
