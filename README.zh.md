# OpenMacUpdate

**Mac 应用批量更新工具 — 免费开源，无需注册**

OpenMacUpdate 扫描已安装应用，通过 **Sparkle**、**Homebrew Casks** 和 **Mac App Store** 检查更新 — 全部在一个终端界面完成。

> 数据库包含 **60,000+ 应用**。完全开源。无注册、无追踪、无限制。

![OpenMacUpdate TUI](docs/screenshot.png)

---

## 功能特点

- **🔍 智能扫描** — 扫描 `/Applications` 和 `~/Applications`，解析 `Info.plist`
- **📡 Sparkle 检查** — 自动获取 appcast XML 源
- **🍺 Homebrew Casks** — 通过 `brew outdated --cask` 检查
- **🏪 Mac App Store** — 通过 `mas outdated` 检查
- **⚡ 并行检查** — 所有源同时检查，不逐一等待
- **🎯 ARM/Intel** — 自动选择正确的架构源
- **📦 远程数据库** — 始终从 GitHub 获取最新数据库
- **💾 智能缓存** — 本地缓存 24 小时，支持离线使用
- **🎨 精美界面** — 彩色状态、进度条、实时日志

---

## 数据库

数据库包含 **60,000+ macOS 应用**：

| 字段 | 描述 |
|------|------|
| Sparkle feed URLs | appcast XML 直链 |
| ARM/Intel 源 | 特定架构的更新源 |
| Homebrew cask 名称 | 用于 `brew upgrade --cask` 集成 |
| 忽略规则 | 系统应用、安装程序等 |
| 代码签名 | 开发者身份验证 |
| 更新日志链接 | 更新说明直链 |

数据库每次运行时从仓库拉取，本地缓存 24 小时。欢迎提交 PR 添加应用。

---

## 安装

### 通过 Homebrew 安装（推荐）

```bash
brew tap yakushevhk/tap
brew install --cask macbee
```

### 从源码安装

```bash
git clone https://github.com/yakushevhk/OpenMacUpdate.git
cd OpenMacUpdate
go build -o macbee .
./macbee
```

### 系统要求

- **macOS**
- **Homebrew**（可选，用于 cask 集成）
- **mas-cli**（可选，用于 App Store 集成）

```bash
brew install mas
```

### 更新

```bash
brew upgrade --cask macbee
```

---

## 使用方法

| 按键 | 功能 |
|------|------|
| `↑↓` | 导航 |
| `enter` | 打开应用 |
| `u` | 检查所有应用 |
| `space` | 选择/取消选择 |
| `a` | 全选/全不选 |
| `U` | 更新选中的应用 |
| `f` | 过滤器（更新 → 全部 → 错误） |
| `/` | 搜索 |
| `r` | 打开更新日志 |
| `q` | 退出 |

---

## 工作原理

```
┌─────────────────────────────────────────────────┐
│  1. 扫描 /Applications + ~/Applications        │
│     └─ 读取 Info.plist → BundleID、版本         │
│                                                 │
│  2. 加载数据库 (GitHub → 缓存 → 离线)           │
│     └─ 匹配 BundleID → Feed URL                │
│                                                 │
│  3. 并行检查更新                                │
│     ├─ Sparkle: 获取 appcast XML               │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. 版本比较与显示                              │
│     └─ 语义化版本比较                           │
│                                                 │
│  5. 更新（可选）                                │
│     ├─ Sparkle: curl + 打开 DMG/ZIP            │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| TUI 框架 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI 组件 | [Bubbles](https://github.com/charmbracelet/bubbles) |
| 样式 | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| plist 解析 | [howett.net/plist](https://howett.net/plist) |
| 数据库 | JSON（60K+ 应用，社区维护） |

---

## 参与贡献

项目采用 MIT 许可证，**完全开源**。欢迎任何形式的贡献。

- **添加应用** — 编辑 `db/apps.json`，提交 PR
- **报告问题** — 创建 issue
- **添加功能** — fork、分支、PR
- **改进 TUI** — 代码在 `tui/tui.go`

---

## 许可证

[MIT License](LICENSE) — 随意使用。

---

**使用 Go + Bubble Tea 构建** · 数据库来自开源社区
