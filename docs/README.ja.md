# MacUpdate

**Macアプリを一括更新 — 無料・オープンソース・登録不要**

MacUpdateはインストール済みアプリをスキャンし、**Sparkle**、**Homebrew Casks**、**Mac App Store**の更新をチェック — すべて一つのターミナルで完結。

> データベースに**60,000以上のアプリ**を収録。完全オープンソース。登録不要、トラッキングなし。

![MacUpdate TUI](docs/screenshot.png)

---

## 機能

- **🔍 スマートスキャン** — `/Applications`と`~/Applications`をスキャン、`Info.plist`を解析
- **📡 Sparkleチェック** — appcast XMLフィードを自動取得
- **🍺 Homebrew Casks** — `brew outdated --cask`でチェック
- **🏪 Mac App Store** — `mas outdated`でチェック
- **⚡ 並列チェック** — すべてのフィードを同時チェック
- **🎯 ARM/Intel** — アーキテクチャに応じたフィードを自動選択
- **📦 リモートデータベース** — GitHubから常に最新のデータベースを取得
- **💾 スマートキャッシュ** — 24時間のローカルキャッシュ、オフライン対応
- **🎨 リッチTUI** — カラーステータス、プログレスバー、ライブログ

---

## データベース

**60,000以上のmacOSアプリ**を収録：

| フィールド | 説明 |
|------------|------|
| Sparkle feed URLs | appcast XMLのダイレクトリンク |
| ARM/Intelフィード | アーキテクチャ固有のフィード |
| Homebrew cask名 | `brew upgrade --cask`連携用 |
| 無視ルール | システムアプリ、インストーラーなど |
| コード署名 | 開発者の身元確認 |
| リリースノートURL | 変更履歴への直接リンク |

データベースは毎回GitHubから取得し、24時間ローカルにキャッシュ。アプリの追加はPR歓迎。

---

## インストール

### Homebrewでインストール（推奨）

```bash
brew tap yakushevhk/tap
brew install --cask macbee
```

### ソースからインストール

```bash
git clone https://github.com/yakushevhk/macupdate.git
cd macupdate
go build -o macbee .
./macbee
```

### 必要要件

- **macOS**
- **Homebrew**（オプション、cask連携用）
- **mas-cli**（オプション、App Store連携用）

```bash
brew install mas
```

### アップデート

```bash
brew upgrade --cask macbee
```

---

## 使い方

| キー | アクション |
|------|------------|
| `↑↓` | ナビゲーション |
| `enter` | アプリを開く |
| `u` | すべてチェック |
| `space` | 選択/選択解除 |
| `a` | 全選択/全解除 |
| `U` | 選択したアプリを更新 |
| `f` | フィルター（更新 → すべて → エラー） |
| `/` | 検索 |
| `r` | リリースノートを開く |
| `q` | 終了 |

---

## 仕組み

```
┌─────────────────────────────────────────────────┐
│  1. /Applications + ~/Applicationsをスキャン    │
│     └─ Info.plist読み取り → BundleID、バージョン │
│                                                 │
│  2. データベース読み込み (GitHub → キャッシュ)   │
│     └─ BundleID → Feed URLのマッピング          │
│                                                 │
│  3. 並列更新チェック                            │
│     ├─ Sparkle: appcast XML取得                │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. バージョン比較と表示                         │
│     └─ セマンティックバージョン比較              │
│                                                 │
│  5. 更新（オプション）                          │
│     ├─ Sparkle: curl + DMG/ZIPを開く           │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## 技術スタック

| コンポーネント | 技術 |
|----------------|------|
| 言語 | Go 1.26 |
| TUIフレームワーク | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UIコンポーネント | [Bubbles](https://github.com/charmbracelet/bubbles) |
| スタイリング | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| plistパース | [howett.net/plist](https://howett.net/plist) |
| データベース | JSON（60K+アプリ、コミュニティ管理） |

---

## コントリビューション

MITライセンスの**完全オープンソース**プロジェクト。自由に参加可能。

- **アプリ追加** — `db/apps.json`を編集、PR送信
- **バグ報告** — issue作成
- **機能追加** — fork、ブランチ、PR
- **TUI改善** — コードは`tui/tui.go`

---

## ライセンス

[MIT License](LICENSE) — 自由にお使いください。

---

**Go + Bubble Teaで構築** · オープンソースコミュニティのデータベース
