# OpenMacUpdate

## The Open-Source macOS App Update Checker

OpenMacUpdate scans your installed applications and checks for updates across **Sparkle feeds**, **Homebrew Casks**, and the **Mac App Store** — all from a single fast terminal UI.

> 60,000+ apps in the database. Fully open source. No tracking, no accounts, no BS.

![OpenMacUpdate TUI](docs/screenshot.png)

---

## Features

- **🔍 Smart Scanning** — reads `/Applications` and `~/Applications`, parses `Info.plist` for bundle IDs, versions, and Sparkle feed URLs
- **📡 Sparkle Feed Checking** — fetches appcast XML feeds to compare installed vs latest versions
- **🍺 Homebrew Cask Integration** — detects outdated casks via `brew outdated --cask`
- **🏪 Mac App Store Integration** — checks App Store updates via `mas outdated`
- **⚡ Parallel Checking** — all feeds checked simultaneously, not one-by-one
- **🎯 ARM/Intel Feed Selection** — automatically picks the right feed for your architecture
- **📦 Remote Database** — always-fresh app database synced from this repo via GitHub raw URL
- **💾 Smart Caching** — local cache with 24h TTL, works offline with stale cache
- **🎨 Rich TUI** — colored status indicators, progress bar, live log panel

---

## Database

The database contains **60,000+ macOS applications** with:

| Field | Description |
|-------|-------------|
| Sparkle feed URLs | Direct appcast XML links for version checking |
| ARM/Intel feeds | Architecture-specific feeds |
| Homebrew cask names | For `brew upgrade --cask` integration |
| Ignore rules | Bundle ID prefixes/suffixes to skip (system apps, installers, etc.) |
| Code signatures | Developer identity verification |
| Release notes URLs | Direct links to changelogs |

The database is pulled from this repository on every run and cached locally for 24 hours. Submit a PR to add or update app entries.

---

## Getting Started

### Install

```bash
git clone https://github.com/yakushevhk/OpenMacUpdate.git
cd OpenMacUpdate
go build -o macbee .
./macbee
```

### Requirements

- **macOS** (primary target)
- **Go 1.21+** (for building from source)
- **Homebrew** (optional, for cask integration)
- **mas-cli** (optional, for App Store integration)

```bash
# Optional dependencies
brew install mas
```

---

## Usage

| Key | Action |
|-----|--------|
| `↑↓` | Navigate |
| `enter` | Open selected app |
| `u` | Scan all apps |
| `space` | Select/deselect app |
| `a` | Select/deselect all |
| `U` | Update selected apps |
| `f` | Cycle filter (updates → all → errors) |
| `/` | Search |
| `r` | Open release notes |
| `q` | Quit |

### Filters

- **updates** (default) — only apps with available updates
- **all** — every scanned application
- **errors** — apps that failed to check

---

## How It Works

```
┌─────────────────────────────────────────────────┐
│  1. Scan /Applications + ~/Applications         │
│     └─ Read Info.plist → BundleID, Version      │
│                                                 │
│  2. Load Database (remote → cache → offline)    │
│     └─ Match BundleID → Feed URL, Cask Name     │
│                                                 │
│  3. Check Updates (parallel)                    │
│     ├─ Sparkle: fetch appcast XML               │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. Compare Versions & Display                  │
│     └─ Semantic version comparison              │
│                                                 │
│  5. Update (optional)                           │
│     ├─ Sparkle: curl + open DMG/ZIP             │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI Components | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Plist Parsing | [howett.net/plist](https://howett.net/plist) |
| Database | JSON (60K+ apps, community-maintained) |

---

## Contributing

This project is **fully open source** under the MIT License. You can do whatever you want with it.

Contributions welcome:

- **Add apps to the database** — edit `db/apps.json`, submit a PR
- **Report bugs** — open an issue
- **Add features** — fork, branch, PR
- **Improve the TUI** — the code is in `tui/tui.go`

### Project Structure

```
OpenMacUpdate/
├── main.go              # Entry point
├── tui/tui.go           # Terminal UI (Bubble Tea)
├── scan/scan.go         # App scanner (Info.plist parsing)
├── update/
│   ├── update.go        # Sparkle feed checker
│   ├── brew.go          # Homebrew cask integration
│   └── mas.go           # Mac App Store integration
├── db/
│   ├── db.go            # Database loader (remote + cache)
│   └── apps.json        # App database (60K+ entries)
└── docs/
    └── screenshot.png   # TUI screenshot
```

---

## License

[MIT License](LICENSE) — do whatever you want.

---

**Built with Go + Bubble Tea** · Database from the open-source community
