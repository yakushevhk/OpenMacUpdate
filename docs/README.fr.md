# OpenMacUpdate

**Mettez à jour toutes vos apps Mac — gratuit, open source, sans inscription**

OpenMacUpdate scanne vos applications installées et vérifie les mises à jour via **Sparkle**, **Homebrew Casks** et **Mac App Store** — le tout dans une seule interface terminal.

> **60 000+ applications** dans la base de données. Entièrement open source. Sans inscription, sans tracking, sans limitation.

![OpenMacUpdate TUI](docs/screenshot.png)

---

## Fonctionnalités

- **🔍 Scan intelligent** — scanne `/Applications` et `~/Applications`, lit `Info.plist`
- **📡 Vérification Sparkle** — récupération automatique des flux appcast XML
- **🍺 Homebrew Casks** — vérification via `brew outdated --cask`
- **🏪 Mac App Store** — vérification via `mas outdated`
- **⚡ Vérification parallèle** — tous les flux vérifiés simultanément
- **🎯 ARM/Intel** — sélection automatique du bon flux pour votre architecture
- **📦 Base de données distante** — toujours à jour depuis GitHub
- **💾 Cache intelligent** — cache local de 24h, fonctionne hors ligne
- **🎨 TUI riche** — statuts colorés, barre de progression, journal en direct

---

## Base de données

**60 000+ applications macOS** incluses :

| Champ | Description |
|-------|-------------|
| Sparkle feed URLs | Liens directs vers appcast XML |
| Flux ARM/Intel | Flux spécifiques à l'architecture |
| Noms Homebrew cask | Pour l'intégration `brew upgrade --cask` |
| Règles d'ignorance | Apps système, installateurs, etc. |
| Signatures de code | Vérification de l'identité du développeur |
| URLs des notes de version | Liens directs vers les changelogs |

La base de données est téléchargée depuis le dépôt à chaque lancement et mise en cache localement pendant 24h. Les PRs pour ajouter des applications sont les bienvenus.

---

## Installation

### Via Homebrew (recommandé)

```bash
brew tap yakushevhk/tap
brew install --cask macbee
```

### Depuis les sources

```bash
git clone https://github.com/yakushevhk/OpenMacUpdate.git
cd OpenMacUpdate
go build -o macbee .
./macbee
```

### Prérequis

- **macOS**
- **Homebrew** (optionnel, pour l'intégration cask)
- **mas-cli** (optionnel, pour l'intégration App Store)

```bash
brew install mas
```

### Mise à jour

```bash
brew upgrade --cask macbee
```

---

## Utilisation

| Touche | Action |
|--------|--------|
| `↑↓` | Navigation |
| `enter` | Ouvrir l'application |
| `u` | Vérifier toutes les apps |
| `space` | Sélectionner/désélectionner |
| `a` | Tout sélectionner/désélectionner |
| `U` | Mettre à jour la sélection |
| `f` | Filtre (mises à jour → tout → erreurs) |
| `/` | Rechercher |
| `r` | Ouvrir les notes de version |
| `q` | Quitter |

---

## Comment ça fonctionne

```
┌─────────────────────────────────────────────────┐
│  1. Scanner /Applications + ~/Applications      │
│     └─ Lire Info.plist → BundleID, version      │
│                                                 │
│  2. Charger la base (GitHub → cache → offline)  │
│     └─ Mapper BundleID → Feed URL               │
│                                                 │
│  3. Vérifier les mises à jour (en parallèle)    │
│     ├─ Sparkle: récupérer appcast XML           │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. Comparer les versions et afficher           │
│     └─ Comparaison sémantique des versions      │
│                                                 │
│  5. Mettre à jour (optionnel)                   │
│     ├─ Sparkle: curl + ouvrir DMG/ZIP           │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.26 |
| Framework TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Composants UI | [Bubbles](https://github.com/charmbracelet/bubbles) |
| Styles | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Parsing plist | [howett.net/plist](https://howett.net/plist) |
| Base de données | JSON (60K+ apps, maintenue par la communauté) |

---

## Contribuer

Le projet est **entièrement open source** sous licence MIT. Vous pouvez en faire ce que vous voulez.

- **Ajouter des apps** — éditez `db/apps.json`, envoyez un PR
- **Signaler des bugs** — créez une issue
- **Ajouter des fonctionnalités** — fork, branche, PR
- **Améliorer le TUI** — le code est dans `tui/tui.go`

---

## Licence

[MIT License](LICENSE) — Faites ce que vous voulez.

---

**Construit avec Go + Bubble Tea** · Base de données de la communauté open source
