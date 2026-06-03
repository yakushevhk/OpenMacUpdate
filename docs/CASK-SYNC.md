# Cask Sync

How the automated Homebrew Cask synchronization works.

## Overview

`cmd/cask-sync/main.go` cross-references our database with Homebrew Cask data to discover:
- **Cask names** — enables `brew upgrade --cask` integration
- **Sparkle feed URLs** — from cask `livecheck` blocks
- **Lookup URLs** — non-Sparkle version check endpoints

## How It Works

```
┌─────────────────────────────────────────────────────┐
│  1. Fetch Homebrew Cask API                         │
│     URL: formulae.brew.sh/api/cask.json             │
│     Result: ~7,700 casks                            │
│                                                     │
│  2. Extract Bundle IDs                              │
│     Source: artifacts → uninstall → quit             │
│     Result: bundleID → cask mapping                 │
│                                                     │
│  3. Match Against Our Database                      │
│     Match by: bundle identifier                     │
│     Action: add cask name if missing                │
│                                                     │
│  4. Fetch Livecheck URLs (parallel, 16 workers)     │
│     Source: raw.githubusercontent.com/.../cask.rb   │
│     Parse: livecheck block for URL + strategy       │
│                                                     │
│  5. Update Database                                 │
│     :sparkle strategy → add as FeedURL (fe)         │
│     other strategy    → add as LookupURL (lu)       │
│                                                     │
│  6. Save Changes                                    │
│     Write updated db/apps.json                      │
└─────────────────────────────────────────────────────┘
```

## Livecheck Strategies

Homebrew casks use various strategies to check for updates:

| Strategy | Description | Our Field |
|----------|-------------|-----------|
| `:sparkle` | Sparkle appcast XML | `fe` (FeedURL) |
| `:header_match` | HTTP redirect/headers | `lu` (LookupURL) |
| `:json` | JSON API endpoint | `lu` |
| `:electron_builder` | Electron YAML endpoint | `lu` |
| `:github_latest` | GitHub Releases latest | `lu` |
| `:github_releases` | GitHub Releases list | `lu` |
| `:page_match` | HTML page scraping | `lu` |

## Running Manually

```bash
# Build
go build -o cask-sync ./cmd/cask-sync

# Run (modifies db/apps.json in place)
./cask-sync
```

Output:
```
=== MacUpdate Cask Sync ===

Loaded database: 60623 entries
Fetching Homebrew Cask API...
Fetched 7694 casks
Extracting bundle IDs from casks...
Extracted bundle IDs from 10093 casks
Matching against database...
Fetching livecheck URLs from cask source files...

=== Sync Complete ===
Cask names added:  1934
Sparkle feeds:     152
Lookup URLs:       613
Total matched:     2699
```

## Automated Sync

GitHub Action: `.github/workflows/cask-sync.yml`

**Schedule:** Every Monday at 06:00 UTC

**What happens:**
1. Checks out the repo
2. Runs `go run ./cmd/cask-sync`
3. If `db/apps.json` changed → creates a PR
4. PR title: `Auto-sync: Update database from Homebrew Cask`

**Manual trigger:**
Go to Actions → "Sync Homebrew Cask Data" → "Run workflow"

## Adding New Cask Sources

To add support for a new cask livecheck strategy:

1. Edit `cmd/cask-sync/main.go`
2. Add regex pattern for the new strategy in `fetchLivecheck()`
3. Map it to the appropriate database field

```go
// Example: adding a new strategy
if strings.Contains(content, "strategy :my_strategy") {
    lc.Strategy = "my_strategy"
}
```

## Limitations

- **Rate limits:** GitHub raw content has rate limits. We use 16 parallel workers with 10s timeout each.
- **Stale feeds:** Some cask livecheck URLs may go stale. A health check can identify these.
- **Bundle ID matching:** Some casks don't have `uninstall -> quit` entries. We fall back to path-based extraction from `zap -> trash`.
- **Multiple apps per cask:** Some casks install multiple apps. We match the first bundle ID found.
