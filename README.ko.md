# OpenMacUpdate

**Mac 앱 전체 업데이트 — 무료, 오픈소스, 가입 불필요**

OpenMacUpdate는 설치된 앱을 스캔하고 **Sparkle**, **Homebrew Casks**, **Mac App Store**에서 업데이트를 확인합니다 — 하나의 터미널에서 모두 처리.

> 데이터베이스에 **60,000개 이상의 앱** 포함. 완전 오픈소스. 가입 불필요, 추적 없음, 제한 없음.

![OpenMacUpdate TUI](docs/screenshot.png)

---

## 기능

- **🔍 스마트 스캔** — `/Applications`와 `~/Applications` 스캔, `Info.plist` 파싱
- **📡 Sparkle 확인** — appcast XML 피드 자동 가져오기
- **🍺 Homebrew Casks** — `brew outdated --cask`로 확인
- **🏪 Mac App Store** — `mas outdated`로 확인
- **⚡ 병렬 확인** — 모든 피드를 동시에 확인
- **🎯 ARM/Intel** — 아키텍처에 맞는 피드 자동 선택
- **📦 원격 데이터베이스** — GitHub에서 항상 최신 데이터베이스 가져오기
- **💾 스마트 캐시** — 24시간 로컬 캐시, 오프라인 지원
- **🎨 리치 TUI** — 컬러 상태 표시, 진행률 바, 실시간 로그

---

## 데이터베이스

**60,000개 이상의 macOS 앱** 포함:

| 필드 | 설명 |
|------|------|
| Sparkle feed URLs | appcast XML 직접 링크 |
| ARM/Intel 피드 | 아키텍처별 피드 |
| Homebrew cask 이름 | `brew upgrade --cask` 연동용 |
| 무시 규칙 | 시스템 앱, 설치 프로그램 등 |
| 코드 서명 | 개발자 신원 확인 |
| 릴리스 노트 URL | 변경 로그 직접 링크 |

데이터베이스는 매 실행마다 GitHub에서 가져와 24시간 로컬에 캐시됩니다. 앱 추가 PR 환영.

---

## 설치

```bash
git clone https://github.com/yakushevhk/OpenMacUpdate.git
cd OpenMacUpdate
go build -o macbee .
./macbee
```

### 요구 사항

- **macOS**
- **Go 1.21+** (소스에서 빌드하는 경우)
- **Homebrew** (선택, cask 연동용)
- **mas-cli** (선택, App Store 연동용)

```bash
brew install mas
```

---

## 사용법

| 키 | 동작 |
|----|------|
| `↑↓` | 탐색 |
| `enter` | 앱 열기 |
| `u` | 모든 앱 확인 |
| `space` | 선택/선택 해제 |
| `a` | 전체 선택/해제 |
| `U` | 선택한 앱 업데이트 |
| `f` | 필터 (업데이트 → 전체 → 오류) |
| `/` | 검색 |
| `r` | 릴리스 노트 열기 |
| `q` | 종료 |

---

## 작동 방식

```
┌─────────────────────────────────────────────────┐
│  1. /Applications + ~/Applications 스캔         │
│     └─ Info.plist 읽기 → BundleID, 버전         │
│                                                 │
│  2. 데이터베이스 로드 (GitHub → 캐시 → 오프라인) │
│     └─ BundleID → Feed URL 매핑                │
│                                                 │
│  3. 업데이트 확인 (병렬)                        │
│     ├─ Sparkle: appcast XML 가져오기            │
│     ├─ Homebrew: brew outdated --cask           │
│     └─ App Store: mas outdated                  │
│                                                 │
│  4. 버전 비교 및 표시                           │
│     └─ 시맨틱 버전 비교                         │
│                                                 │
│  5. 업데이트 (선택사항)                         │
│     ├─ Sparkle: curl + DMG/ZIP 열기            │
│     ├─ Homebrew: brew upgrade --cask            │
│     └─ App Store: mas upgrade                   │
└─────────────────────────────────────────────────┘
```

---

## 기술 스택

| 구성 요소 | 기술 |
|-----------|------|
| 언어 | Go 1.26 |
| TUI 프레임워크 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| UI 컴포넌트 | [Bubbles](https://github.com/charmbracelet/bubbles) |
| 스타일링 | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| plist 파싱 | [howett.net/plist](https://howett.net/plist) |
| 데이터베이스 | JSON (60K+ 앱, 커뮤니티 유지관리) |

---

## 기여하기

프로젝트는 MIT 라이선스 하에 **완전 오픈소스**입니다. 원하는 대로 사용하세요.

- **앱 추가** — `db/apps.json` 편집, PR 제출
- **버그 신고** — issue 생성
- **기능 추가** — fork, 브랜치, PR
- **TUI 개선** — 코드는 `tui/tui.go`

---

## 라이선스

[MIT License](LICENSE) — 원하는 대로 사용하세요.

---

**Go + Bubble Tea로 제작** · 오픈소스 커뮤니티의 데이터베이스
