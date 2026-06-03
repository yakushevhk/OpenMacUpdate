# Database

How the MacUpdate database works.

## Overview

The database (`db/apps.json`) maps macOS **bundle identifiers** to app metadata. It contains **60,000+ entries** and is the core of update checking.

```
BundleID → { Sparkle feed URL, Cask name, Version check URL, ... }
```

The database is:
- Pulled from GitHub on every run
- Cached locally for 24 hours in `~/.cache/macupdate/apps.json`
- Updated weekly via automated Homebrew Cask sync

## Structure

```json
{
  "bundleIDMapping": {
    "com.example.app": {
      "fe": "https://example.com/appcast.xml",
      "ca": "example-app",
      "cs": "Developer Name (TEAMID)",
      "lu": "https://api.example.com/releases/latest",
      "ig": 0
    }
  },
  "ignoredBundleIDPrefixes": ["com.apple.", "automator."],
  "ignoredBundleIDSuffixes": [],
  "ignoredBundleIDSubstrings": ["uninstall", "installer"]
}
```

## Fields

| Key | Field | Description |
|-----|-------|-------------|
| `fe` | FeedURL | Sparkle appcast XML URL — primary update source |
| `feas` | FeedURLARM | ARM-specific Sparkle feed (Apple Silicon) |
| `ca` | CaskName | Homebrew cask name for `brew upgrade` |
| `lu` | LookupURL | Non-Sparkle version check URL (GitHub Releases, JSON API, etc.) |
| `cs` | CodeSignature | Developer identity for verification |
| `or` | OverrideBID | Redirect — the app was renamed, check this bundle ID instead |
| `ig` | IsIgnored | `1` = skip this app (system, installer, malware) |
| `ie` | IgnoreWhy | Reason for ignoring (`aux` = auxiliary, `pri` = private, `mal` = malware) |
| `rn` | ReleaseNotes | URL to release notes / changelog |
| `pu` | PurchaseURL | URL to purchase/download the app |
| `fn` | ForceName | Override display name |
| `pl` | IsPaid | `1` = paid app |
| `fp` | ForcePremium | `1` = force premium status |
| `mp` | MultiPackage | Number of packages in multi-package install |

## Sources

### 1. Homebrew Cask API

The primary automated source. We fetch `formulae.brew.sh/api/cask.json` which contains ~7,700 casks.

For each cask we extract:
- **Bundle IDs** from `artifacts -> uninstall -> quit`
- **Livecheck URLs** from the Ruby source file's `livecheck` block
- **Strategy** (`:sparkle`, `:header_match`, `:json`, `:github_latest`, etc.)

Run manually:
```bash
go run ./cmd/cask-sync
```

### 2. Sparkle Feed Discovery

When scanning installed apps, we read `SUFeedURL` from `Info.plist`. If a feed URL is found that's not in the database, it can be contributed.

### 3. Community Contributions

Submit PRs to add or update app entries in `db/apps.json`.

## Automated Sync

A GitHub Action runs weekly (Monday 6:00 UTC):

```yaml
# .github/workflows/cask-sync.yml
- Fetches Homebrew Cask API
- Extracts bundle IDs and livecheck URLs
- Matches against our database
- Creates PR if changes found
```

Current coverage after sync:

| Source | Count |
|--------|-------|
| Sparkle feeds | ~4,000 |
| Homebrew cask names | ~1,945 |
| Lookup URLs | ~4,000 |
| Ignored | ~9,500 |
| Redirects | ~19,000 |

## Adding a New App

1. Find the app's bundle ID: `mdls -name kMDItemCFBundleIdentifier /Applications/App.app`
2. Find the Sparkle feed URL: `defaults read /Applications/App.app/Contents/Info.plist SUFeedURL`
3. Add to `db/apps.json`:
   ```json
   "com.example.app": {
     "fe": "https://example.com/appcast.xml"
   }
   ```
4. Submit a PR

## Health Checking

Feed URLs can go stale. A health check script can verify feeds are still accessible:

```bash
# Check if a feed URL returns valid XML
curl -s "https://example.com/appcast.xml" | xmllint --noout -
```
