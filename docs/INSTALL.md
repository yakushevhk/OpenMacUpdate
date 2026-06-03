# Installation

## Via Homebrew (Recommended)

```bash
brew tap yakushevhk/tap
brew install --cask macbee
```

Update:
```bash
brew upgrade --cask macbee
```

## From Source

Requirements:
- Go 1.21+
- macOS

```bash
git clone https://github.com/yakushevhk/macupdate.git
cd macupdate
go build -o macbee .
./macbee
```

## Optional Dependencies

```bash
# Mac App Store integration
brew install mas
```

Homebrew is used for cask integration — if you have Homebrew installed, it works automatically.

## Binary Releases

Download pre-built binaries from [Releases](https://github.com/yakushevhk/macupdate/releases).

Available:
- `MacUpdate_X.Y.Z_darwin_arm64.zip` — Apple Silicon
- `MacUpdate_X.Y_Z_darwin_amd64.zip` — Intel

After downloading:
```bash
unzip MacUpdate_*.zip
chmod +x macbee
./macbee
```

## Verify Installation

```bash
macbee --version
macbee --help
```

## Configuration

No configuration file needed. The app works out of the box.

**Environment variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `MACBEE_DB` | Path to custom database file | Remote from GitHub |

**Cache location:**
- `~/.cache/macupdate/apps.json` — cached database (24h TTL)
