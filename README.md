# OpenMacUpdate

**Update All Apps on Mac — Free, Open Source, No Signup Required**

[Русский](docs/README.ru.md) · [中文](docs/README.zh.md) · [日本語](docs/README.ja.md) · [Deutsch](docs/README.de.md) · [Español](docs/README.es.md) · [한국어](docs/README.ko.md) · [Français](docs/README.fr.md)

## The Open-Source macOS App Update Checker

OpenMacUpdate scans your installed applications and checks for updates across **Sparkle feeds**, **Homebrew Casks**, and the **Mac App Store** — all from a single fast terminal UI.

> 60,000+ apps in the database. Fully open source. No tracking, no accounts, no BS.

![OpenMacUpdate TUI](docs/screenshot.png)

---

## Install

```bash
brew tap yakushevhk/tap
brew install --cask macbee
```

[More install options →](docs/INSTALL.md)

---

## Features

- **🔍 Smart Scanning** — `/Applications` + `~/Applications`, parses `Info.plist`
- **📡 Sparkle Feeds** — 4,000+ apps with appcast XML checking
- **🍺 Homebrew Casks** — 1,900+ casks via `brew outdated --cask`
- **🏪 Mac App Store** — via `mas outdated`
- **⚡ Parallel Checking** — all sources simultaneously
- **🎯 ARM/Intel** — auto-selects correct feed for your architecture
- **📦 Remote Database** — always fresh from GitHub, cached 24h locally
- **🎨 Rich TUI** — colors, progress bar, live log

---

## Usage

| Key | Action |
|-----|--------|
| `↑↓` | Navigate |
| `enter` | Open app |
| `u` | Scan all |
| `space` | Select/deselect |
| `a` | Select all |
| `U` | Update selected |
| `f` | Filter (updates → all → errors) |
| `/` | Search |
| `r` | Release notes |
| `q` | Quit |

---

## Documentation

| File | Description |
|------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | How it works — data flow, update sources, version comparison, TUI |
| [DATABASE.md](docs/DATABASE.md) | Database structure, fields, sources, how to add apps |
| [CASK-SYNC.md](docs/CASK-SYNC.md) | Homebrew Cask sync tool — automated database updates |
| [INSTALL.md](docs/INSTALL.md) | Installation options — Homebrew, source, binaries |

---

## Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Plist | [howett.net/plist](https://howett.net/plist) |
| Database | JSON (60K+ apps, auto-synced from Homebrew Cask) |

---

## Contributing

Fully open source under MIT License. Do whatever you want.

- **Add apps** — edit `db/apps.json`, submit a PR
- **Report bugs** — open an issue
- **Add features** — fork, branch, PR

---

## License

[MIT License](LICENSE)

---

**Built with Go + Bubble Tea** · Database auto-synced from Homebrew Cask community
