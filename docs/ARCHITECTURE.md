# Architecture

How OpenMacUpdate works internally.

## Project Structure

```
OpenMacUpdate/
├── main.go                    # Entry point, --help, --version
├── tui/tui.go                 # Terminal UI (Bubble Tea)
├── scan/scan.go               # App scanner (Info.plist parsing)
├── update/
│   ├── update.go              # Sparkle appcast XML checker
│   ├── brew.go                # Homebrew cask integration
│   └── mas.go                 # Mac App Store integration
├── db/
│   ├── db.go                  # Database loader (remote + cache)
│   └── apps.json              # App database (60K+ entries)
├── cmd/
│   └── cask-sync/main.go      # Homebrew Cask sync tool
├── .goreleaser.yml            # GoReleaser config for releases
├── .github/workflows/
│   ├── release.yml            # Build + release on tag push
│   └── cask-sync.yml          # Weekly database sync
└── docs/                      # Documentation
```

## Data Flow

```
┌─────────────────────────────────────────────────────────┐
│                    Startup                               │
│                                                         │
│  1. Load Database                                       │
│     ├─ Check ~/.cache/openmacupdate/apps.json (< 24h)   │
│     ├─ If stale → fetch from GitHub raw URL             │
│     └─ If offline → use stale cache                     │
│                                                         │
│  2. Scan Applications                                   │
│     ├─ /Applications/*.app                              │
│     ├─ ~/Applications/*.app                             │
│     └─ Read Info.plist → BundleID, Version, SUFeedURL   │
│                                                         │
│  3. Match Against Database                              │
│     └─ BundleID → AppEntry (feed URL, cask name, etc.)  │
│                                                         │
│  4. Auto-Check (parallel)                               │
│     ├─ Sparkle: FetchFeed(feedURL) → appcast XML        │
│     ├─ Homebrew: brew outdated --cask → JSON            │
│     └─ App Store: mas outdated → text                   │
│                                                         │
│  5. Compare Versions                                    │
│     └─ Semantic version comparison (1.2.3 vs 1.2.4)     │
│                                                         │
│  6. Display Results                                     │
│     └─ TUI table with status, progress bar, logs        │
│                                                         │
│  7. Update (optional)                                   │
│     ├─ Sparkle: curl download → open DMG/ZIP            │
│     ├─ Homebrew: brew upgrade --cask <name>             │
│     └─ App Store: mas upgrade <id>                      │
└─────────────────────────────────────────────────────────┘
```

## Update Sources

### Sparkle (Primary)

Most macOS apps use the [Sparkle](https://sparkle-project.org/) framework for auto-updates.

1. App declares `SUFeedURL` in `Info.plist`
2. We fetch that URL → get appcast XML
3. Parse `<enclosure>` tags for version and download URL
4. Compare with installed version

```
<rss>
  <channel>
    <item>
      <title>Version 1.2.3</title>
      <enclosure url="https://..." version="1.2.3" shortVersionString="1.2.3"/>
    </item>
  </channel>
</rss>
```

### Homebrew Cask

For apps installed via Homebrew or tracked in the cask database.

1. Run `brew outdated --cask --json`
2. Parse JSON for outdated casks
3. Match against our database by cask name
4. Update via `brew upgrade --cask <name>`

### Mac App Store

For App Store apps.

1. Run `mas outdated`
2. Parse output for app ID and name
3. Match against our database
4. Update via `mas upgrade <id>`

## Version Comparison

Semantic version comparison handles:
- `1.2.3` vs `1.2.4`
- `1.2.3` vs `1.3.0`
- `1.2.3-beta` vs `1.2.3`
- `v1.2.3` vs `1.2.3` (strips `v` prefix)
- `1.2.3.4` vs `1.2.3.5` (4+ segments)

See `update/update.go` → `CompareVersions()`

## TUI (Terminal UI)

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles).

**State machine:**
```
stateLoading → stateReady
```

**Key bindings:**
- `↑↓` — navigate table
- `enter` — open selected app
- `space` — toggle selection
- `u` — scan all apps
- `U` — update selected
- `f` — cycle filter (updates → all → errors)
- `/` — search
- `q` — quit

**Layout:**
```
┌─ Header ──────────────────────────┐
│ OpenMacUpdate  [spinner] [bar]    │
│ 191 apps  8 updates  28 ok       │
├─ Table ───────────────────────────┤
│ * Firefox    130.0   131.0  UPDATE│
│ * Slack      4.39    4.40   UPDATE│
│   IINA       1.4.3   1.4.3  OK   │
├─ Log ─────────────────────────────┤
│ 22:41:48 Firefox 130.0 → 131.0   │
│ 22:41:49 Done — 8 updates found   │
├─ Help ────────────────────────────┤
│ q:quit  /:search  enter:open ...  │
└───────────────────────────────────┘
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `charmbracelet/bubbletea` | TUI framework |
| `charmbracelet/bubbles` | TUI components (table, spinner, progress, textinput) |
| `charmbracelet/lipgloss` | Terminal styling |
| `howett.net/plist` | macOS plist parsing |
| `encoding/xml` | Sparkle appcast parsing |
| `net/http` | HTTP requests |
| `os/exec` | Running brew, mas, open, curl |
