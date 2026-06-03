# OpenMacUpdate

**Alle Mac-Apps aktualisieren — kostenlos, Open Source, ohne Anmeldung**

OpenMacUpdate scannt installierte Apps und prüft auf Updates über **Sparkle**, **Homebrew Casks** und **Mac App Store** — alles in einer Terminal-Oberfläche.

> **60.000+ Apps** in der Datenbank. Vollständig Open Source. Keine Anmeldung, kein Tracking, keine Einschränkungen.

![OpenMacUpdate TUI](docs/screenshot.png)

---

## Funktionen

- **🔍 Intelligentes Scannen** — scannt `/Applications` und `~/Applications`, liest `Info.plist`
- **📡 Sparkle-Prüfung** — automatischer Abruf von appcast XML-Feeds
- **🍺 Homebrew Casks** — Prüfung via `brew outdated --cask`
- **🏪 Mac App Store** — Prüfung via `mas outdated`
- **⚡ Parallele Prüfung** — alle Feeds gleichzeitig, nicht nacheinander
- **🎯 ARM/Intel** — automatische Auswahl des richtigen Feeds für Ihre Architektur
- **📦 Remote-Datenbank** — immer aktuelle Datenbank von GitHub
- **💾 Smart Caching** — lokaler Cache mit 24h TTL, funktioniert offline
- **🎨 Rich TUI** — farbige Statusanzeigen, Fortschrittsbalken, Live-Log

---

## Datenbank

**60.000+ macOS-Apps** enthalten:

| Feld | Beschreibung |
|------|--------------|
| Sparkle feed URLs | Direkte appcast XML-Links |
| ARM/Intel Feeds | Architekturspezifische Feeds |
| Homebrew cask Namen | für `brew upgrade --cask` Integration |
| Ignorier-Regeln | System-Apps, Installer usw. |
| Code-Signature | Entwickler-Verifizierung |
| Release-Notes-URLs | Direkte Links zu Änderungsprotokollen |

Die Datenbank wird bei jedem Start aus dem Repository geladen und 24 Stunden lokal gecacht. PRs zum Hinzufügen von Apps sind willkommen.

---

## Installation

```bash
git clone https://github.com/yakushevhk/OpenMacUpdate.git
cd OpenMacUpdate
go build -o macbee .
./macbee
```

### Voraussetzungen

- **macOS**
- **Go 1.21+** (für Build aus Quellcode)
- **Homebrew** (optional, für Cask-Integration)
- **mas-cli** (optional, für App Store Integration)

```bash
brew install mas
```

---

## Verwendung

| Taste | Aktion |
|-------|--------|
| `↑↓` | Navigation |
| `enter` | App öffnen |
| `u` | Alle Apps prüfen |
| `space` | App auswählen/abwählen |
| `a` | Alle auswählen/abwählen |
| `U` | Ausgewählte Apps aktualisieren |
| `f` | Filter (Updates → Alle → Fehler) |
| `/` | Suche |
| `r` | Release-Notes öffnen |
| `q` | Beenden |

---

## So funktioniert es

```
┌─────────────────────────────────────────────────┐
│  1. Scan /Applications + ~/Applications         │
│     └─ Info.plist lesen → BundleID, Version     │
│                                                 │
│  2. Datenbank laden (GitHub → Cache → Offline)  │
│     └─ BundleID → Feed URL zuordnen             │
│                                                 │
│  3. Updates prüfen (parallel)                   │
│     ├─ Sparkle: appcast XML abrufen            │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. Versionen vergleichen & anzeigen            │
│     └─ Semantischer Versionsvergleich           │
│                                                 │
│  5. Aktualisierung (optional)                   │
│     ├─ Sparkle: curl + DMG/ZIP öffnen          │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## Technologie-Stack

| Komponente | Technologie |
|------------|-------------|
| Sprache | Go 1.26 |
| TUI-Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI-Komponenten | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| plist-Parsing | [howett.net/plist](https://howett.net/plist) |
| Datenbank | JSON (60K+ Apps, community-betrieben) |

---

## Mitmachen

Das Projekt ist **vollständig Open Source** unter der MIT-Lizenz. Sie können alles damit machen.

- **Apps hinzufügen** — `db/apps.json` bearbeiten, PR senden
- **Bugs melden** — Issue erstellen
- **Funktionen hinzufügen** — Fork, Branch, PR
- **TUI verbessern** — Code in `tui/tui.go`

---

## Lizenz

[MIT License](LICENSE) — Machen Sie, was Sie wollen.

---

**Gebaut mit Go + Bubble Tea** · Datenbank von der Open-Source-Community
