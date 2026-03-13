<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

<h1 align="center">🔐 wall-vault</h1>

<p align="center"><i>AI Proxy + Key Vault — keep your bots alive, no matter what</i></p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3-blue.svg" alt="License: GPL v3"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg" alt="Go Version">
  <a href="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml"><img src="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <img src="https://img.shields.io/badge/languages-17-brightgreen.svg" alt="Languages">
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg" alt="Platform">
</p>

---

## Language · 언어

| [🇺🇸 English](#-origin-story-the-night-the-bots-died) | [🇰🇷 한국어](#-탄생-배경-봇들이-죽던-날) | [🇨🇳 中文](#-项目简介) | [🇯🇵 日本語](#-はじめに) | [🇪🇸 Español](#-introducción) | [🇫🇷 Français](#-présentation) | [🇩🇪 Deutsch](#-über-das-projekt) |

---

## 💀 Origin Story: The Night the Bots Died

It was the middle of the night.

One alert. Something felt wrong.

Logged in — all API keys invalidated. vault.json empty. The bots had gone silent. Of course they had. **Their memories had been completely erased.**

> *Motoko knew my work style inside-out. Mini prepared morning briefings every day. Raz handled everything quietly from a Raspberry Pi in the corner. Two weeks of careful cultivation. Two weeks of patience, tuning, and personality-shaping.*
>
> *A single hacker broke into the lab's internal network and torched all of it.*
>
> *It felt like coming home to find a beloved pet had simply vanished.*

It took a week to restore most of the memories. Not all of them came back.

This could never happen again.

So I built something. **A vault for the keys. A wall for the bots. A guarantee that no single attack could ever end everything again.**

---

## ⚔️ What It Actually Is

One line: **"A bodyguard that keeps your AI bots alive no matter what."**

```
Hacker steals a key?       → Vault blocks it. Rotates to the next.
Key hits its daily limit?  → Automatically switches. No downtime.
Service goes dark?         → Falls back: Gemini → OpenAI → Ollama
Running 100 bots?          → Change one setting. All bots updated in 1–3s.
```

In more detail:

- 🔐 **Key Vault**: AES-GCM encrypted storage. Round-robin rotation. Quota, cooldown, and error handling — all automatic.
- 🔀 **AI Proxy**: Accepts requests from OpenClaw, Claude Code, VS Code, your scripts — routes them to Gemini / OpenAI / Ollama. One dies, the next one picks up.
- ⚡ **SSE Real-time Sync**: Change anything in the vault, every connected bot reflects it instantly. No restarts.
- 🛡️ **Security Filter**: Full function calling block. Stops external skills from hijacking your AI.
- 🦞 **OpenClaw Integration**: Live events over Unix socket to TUI. Auto-updates openclaw.json.

Single Go binary. One bot or a dozen — fully covered.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Languages](#languages)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Supported Services](#supported-services)
- [API Reference](#api-reference)
- [Modes](#modes)
- [Auto-Start](#auto-start)
- [OpenClaw Integration](#openclaw-integration)
- [Build](#build)
- [Project Structure](#project-structure)
- [License](#license)

---

## Features

| Feature | Description |
|---------|-------------|
| **AI Proxy** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **Key Vault** | API key management, usage monitoring, round-robin rotation |
| **AES-GCM Encryption** | Keys encrypted with master password, never stored in plaintext |
| **SSE Real-time Sync** | Vault ↔ proxy config sync within 1–3 seconds |
| **Tool Security Filter** | Block function calling (`strip_all` / `whitelist` / `passthrough`) |
| **Fallback Chain** | Auto-switch on service failure, final fallback to local Ollama |
| **Model Registry** | 340+ OpenRouter models + dynamic local model discovery |
| **Local AI Support** | Ollama / LM Studio / vLLM auto-detection + manual URL |
| **Service Management** | Add/edit/delete services from UI, custom service support |
| **Service Auto-check** | Dashboard load / key change → cloud services auto-enable/disable by key count; local services probed for connectivity |
| **Agent Management** | Per-agent service / model / IP whitelist / workdir |
| **Agent Status** | 4-state: 🟢Online / 🟡Delayed / 🔴Offline / ⚫Disconnected |
| **Bidirectional Model Sync** | TUI model change → vault; vault change → TUI. All sources stay in sync. |
| **Per-type Config Copy** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — one-click config snippet |
| **Doctor** | Health check, auto-recovery, systemd/launchd/NSSM registration |
| **[17 Languages](#languages)** | Korean · English · Chinese · Japanese · Spanish · Hindi · Arabic · Portuguese · French · German · Thai · Mongolian · Swahili · Hausa · Zulu · Nepali · Indonesian — drop a JSON file to add any language, zero code changes |
| **Themes** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |
| **Cross-platform** | Linux / macOS / Windows / WSL |
| **[OpenClaw Integration](#openclaw-integration)** | Unix socket TUI events, agent auto-config |

---

## Quick Start

### Linux / macOS

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Interactive setup wizard
./wall-vault setup

# Launch (proxy + vault)
./wall-vault start
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe

.\wall-vault.exe setup
.\wall-vault.exe start
```

Open `http://localhost:56243` to access the dashboard.

---

## Languages

wall-vault speaks **17 languages** out of the box — and you can add any language you want with a single JSON file. No Go code to write. No recompilation. Just drop the file and rebuild.

| Code | Language | Code | Language | Code | Language |
|------|----------|------|----------|------|----------|
| `ko` | 한국어 | `ar` | العربية | `th` | ภาษาไทย |
| `en` | English | `pt` | Português | `mn` | Монгол |
| `zh` | 中文 | `fr` | Français | `sw` | Kiswahili |
| `ja` | 日本語 | `de` | Deutsch | `ha` | Hausa |
| `es` | Español | `id` | Indonesia | `zu` | IsiZulu |
| `hi` | हिन्दी | `ne` | नेपाली | | |

Language coverage includes:
- **Setup wizard** — interactive CLI prompts, help text, error messages
- **Dashboard UI** — all labels, buttons, status messages
- **System messages** — key events, service status, health reports
- **Live switching** — change language in the dashboard with no page reload

```bash
WV_LANG=en ./wall-vault setup   # English setup
WV_LANG=ja ./wall-vault setup   # Japanese setup
WV_LANG=zh ./wall-vault setup   # Chinese setup
WV_LANG=ko ./wall-vault setup   # Korean setup
```

### Adding a New Language

Want Turkish? Vietnamese? Swahili wasn't enough? **It takes about 5 minutes.**

1. Copy an existing locale file as a starting point:

```bash
cp internal/i18n/locales/en.json internal/i18n/locales/tr.json
```

2. Translate the values (keys stay in English):

```json
{
  "lang_label": "Türkçe",
  "lang_emoji": "🇹🇷",
  "unknown_command": "Bilinmeyen komut",
  "setup_welcome": "wall-vault kurulum sihirbazına hoş geldiniz",
  "setup_done": "Kurulum tamamlandı! wall-vault start komutunu çalıştırın",
  "title": "AI Proxy Anahtar Kasası",
  ...
}
```

3. Rebuild:

```bash
make build
```

That's it. The new language appears automatically in the setup wizard, dashboard language picker, and all UI text. The i18n system uses Go's `embed.FS` — any `.json` file in `internal/i18n/locales/` is picked up at compile time, no registration required.

> **Contributions welcome.** If you add a language, consider opening a PR — every language helps someone run their bots in their native tongue.

---

## Architecture

```
              ┌──────────────────────────┐
              │   Key Vault (:56243)     │
              │   AES-GCM encrypted      │
              │   SSE broadcaster        │
              └───────────┬──────────────┘
                          │ SSE real-time sync (1–3s)
       ┌──────────────────┼──────────────────┐
       ▼                  ▼                  ▼
  Bot A (:56244)     Bot B (:56244)     Bot C (:56244)
   (proxy)            (proxy)            (proxy)
       │                  │                  │
       └──────────────────┴──────────────────┘
                          │ fallback chain
       ┌───────────┬───────┴───────┬──────────────┐
       ▼           ▼               ▼              ▼
    Google      OpenAI        OpenRouter      Ollama (final)
```

### Fallback Chain

```
Step 1: Assigned service (per client config)
Step 2: Remaining services in order
Step 3: Ollama (final fallback — survives internet outages)
```

### Cooldown

| HTTP Error | Wait |
|------------|------|
| 429 Too Many Requests | 30 minutes |
| 400 / 401 / 402 / 403 | 24 hours |
| Network error | 10 minutes |

---

## Configuration

```bash
# Recommended: interactive wizard
./wall-vault setup

# Or copy and edit manually
cp configs/example-standalone.yaml wall-vault.yaml
```

### Key Settings (`wall-vault.yaml`)

```yaml
mode: standalone   # standalone | distributed
lang: en           # en | ko | zh | ja | es | hi | ar | pt | fr | de | th | mn | sw | ha | zu | ne | id
theme: cherry      # light | dark | gold | cherry | ocean | autumn | winter

proxy:
  port: 56244
  client_id: my-bot
  vault_url: ""              # distributed mode: http://vault-server:56243
  vault_token: ""            # distributed mode: client token
  tool_filter: strip_all     # strip_all | whitelist | passthrough
  services: [google, openrouter, ollama]
  timeout: 60s

vault:
  port: 56243
  admin_token: ""            # empty = no auth (local dev only)
  master_password: ""        # API key encryption password
  data_dir: ~/.wall-vault/data
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `WV_LANG` | Language code |
| `WV_THEME` | Theme name |
| `WV_PROXY_PORT` | Proxy port override |
| `WV_VAULT_PORT` | Vault port override |
| `WV_VAULT_URL` | Key vault URL (distributed mode) |
| `WV_VAULT_TOKEN` | Proxy auth token |
| `WV_ADMIN_TOKEN` | Admin token |
| `WV_MASTER_PASS` | Encryption master password |
| `WV_KEY_GOOGLE` | Google API key (comma-separated for multiple) |
| `WV_KEY_OPENROUTER` | OpenRouter API key |

---

## Supported Services

### Cloud APIs

| Service ID | Name | Models |
|------------|------|--------|
| `google` | Google Gemini | 6 fixed |
| `openai` | OpenAI | 8 fixed |
| `anthropic` | Anthropic | 6 fixed |
| `openrouter` | OpenRouter | 340+ dynamic |
| `github-copilot` | GitHub Copilot | 6 fixed |

### Local AI

| Service ID | Name | Default Port |
|------------|------|--------------|
| `ollama` | Ollama | 11434 |
| `lmstudio` | LM Studio | 1234 |
| `vllm` | vLLM | 8000 |
| (custom) | Add your own | any |

Local services: set URL and auto-detect models from the **Services** card in the dashboard.

---

## API Reference

Full docs: [docs/API.md](docs/API.md)

### Proxy (`:56244`)

| Path | Description |
|------|-------------|
| `POST /google/v1beta/models/{m}:generateContent` | Gemini API proxy |
| `POST /google/v1beta/models/{m}:streamGenerateContent` | Gemini streaming |
| `POST /v1/chat/completions` | OpenAI-compatible API |
| `GET /health` | Health check |
| `GET /status` | Status |
| `GET /api/models` | Model list |
| `PUT /api/config/model` | Change model (write-through to vault) |
| `POST /reload` | Reload config from vault |

### Key Vault (`:56243`)

| Path | Auth | Description |
|------|------|-------------|
| `GET /` | — | Dashboard UI |
| `GET /api/status` | — | Status |
| `GET /api/events` | — | SSE stream |
| `GET /api/keys` | Client token | Decrypted key list (IP whitelist applied) |
| `POST /api/heartbeat` | Client token | Report proxy status |
| `PUT /api/config` | Client token | **Bidirectional sync** — update own service/model, triggers SSE broadcast |
| `GET /admin/keys` | Admin | Key list |
| `POST /admin/keys` | Admin | Add key |
| `DELETE /admin/keys/{id}` | Admin | Delete key |
| `GET /admin/clients` | Admin | Client list |
| `POST /admin/clients` | Admin | Add client |
| `PUT /admin/clients/{id}` | Admin | Update client + SSE broadcast |
| `DELETE /admin/clients/{id}` | Admin | Delete client |
| `GET /admin/services` | Admin | Service list |
| `POST /admin/services` | Admin | Add custom service |
| `PUT /admin/services/{id}` | Admin | Update service |
| `DELETE /admin/services/{id}` | Admin | Delete service |
| `PUT /admin/theme` | Admin | Change theme |
| `PUT /admin/lang` | Admin | Change language |

---

## Modes

### Standalone (single bot)

```bash
# Pass key via env
WV_KEY_GOOGLE=AIza... ./wall-vault start

# Pass key via flag
./wall-vault proxy --key-google=AIza...
```

### Distributed (multi-bot)

```bash
# [Vault server] run key vault
./wall-vault vault

# [Each bot] connect to vault
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=my-bot-token \
./wall-vault proxy
```

Change a setting in the vault → all bots update within 1–3 seconds via SSE. **No restart required.**

#### Bidirectional Model Sync

Model changes are fully bidirectional:

```
TUI (OpenClaw) changes model
  → PUT /api/config/model on proxy
  → proxy writes through to vault (PUT /api/config)
  → vault broadcasts SSE config_change
  → dashboard agent card updates immediately
  → all other proxies for that client reflect the change

Dashboard changes model
  → PUT /admin/clients/{id} on vault
  → vault broadcasts SSE config_change
  → proxy receives SSE → updates local state
  → TUI footer updated via Unix socket event
```

---

## Auto-Start

### Linux — systemd

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

### macOS — launchd

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Download [NSSM](https://nssm.cc/download) and add to PATH
2. Generate service script: `.\wall-vault.exe doctor deploy windows`
3. Run as Administrator: `%USERPROFILE%\install-wall-vault-service.bat`

---

## OpenClaw Integration

**OpenClaw** is a distributed AI agent framework that runs personas with long-term memory across multiple devices. wall-vault was born to serve OpenClaw — the two systems are deeply integrated.

### Register an OpenClaw Agent

In the dashboard **Add Agent** modal, set the agent type to `openclaw`:
- Work directory auto-fills as `~/.openclaw`
- wall-vault becomes the API key supplier and proxy for that agent

```bash
VAULT_CLIENT_ID=bot-a \
VAULT_URL=http://192.168.x.x:56243 \
wall-vault proxy
```

### Unix Socket Events (TUI Live Notifications)

wall-vault sends real-time JSON events over a Unix domain socket to OpenClaw's TUI.

```yaml
# wall-vault.yaml
hooks:
  openclaw_socket: ~/.openclaw/wall-vault.sock
```

| Event | Trigger |
|-------|---------|
| `model_changed` | Model switch |
| `key_exhausted` | API key daily limit reached |
| `service_down` | Service failure / cooldown |
| `ollama_waiting` | Waiting for local Ollama response |
| `ollama_done` | Ollama response complete |
| `tui_footer` | Status message to TUI footer |

### openclaw.json Provider Config

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/gemini-2.0-flash" }
        ]
      }
    }
  }
}
```

- Prefix model IDs with `wall-vault/` for automatic routing
- `wall-vault/gemini-*` → Google Gemini (direct)
- `wall-vault/gpt-*` / `wall-vault/o3` → OpenAI (direct)
- `wall-vault/claude-*` → Anthropic via OpenRouter
- All OpenClaw provider prefixes supported: `opencode/`, `moonshot/`, `kimi-coding/`, `groq/`, `mistral/`, `deepseek/`, `qwen/`, `meta-llama/`, etc.

### SSE Auto-Sync

OpenClaw agents subscribe to the wall-vault SSE stream and apply model/service changes within **1–3 seconds** — no restart needed.

---

## Build

```bash
# Current OS
make build

# All platforms (cross-compile)
make build-all
# → bin/wall-vault-linux-amd64
# → bin/wall-vault-linux-arm64
# → bin/wall-vault-darwin-amd64
# → bin/wall-vault-darwin-arm64
# → bin/wall-vault-windows-amd64.exe

# Tests (39 cases)
make test

# Local install
make install  # → ~/.local/bin/wall-vault
```

---

## Project Structure

```
wall-vault/
├── main.go                      # entry point + subcommand router
├── cmd/
│   ├── proxy/proxy.go           # proxy subcommand
│   ├── vault/vault.go           # vault subcommand
│   ├── setup/setup.go           # interactive setup wizard
│   └── doctor/doctor.go         # doctor subcommand
├── internal/
│   ├── config/
│   │   ├── config.go            # config load/save
│   │   └── services.go          # service plugin loader
│   ├── proxy/
│   │   ├── server.go            # proxy HTTP server + fallback chain
│   │   ├── keymgr.go            # round-robin key manager
│   │   ├── convert.go           # Gemini↔OpenAI↔Ollama conversion
│   │   └── toolfilter.go        # tool security filter
│   ├── vault/
│   │   ├── server.go            # vault HTTP server + rate limiter
│   │   ├── store.go             # AES-GCM encrypted store
│   │   ├── models.go            # data models
│   │   ├── broker.go            # SSE broadcaster
│   │   └── ui.go                # dashboard HTML (themes + i18n)
│   ├── doctor/doctor.go         # auto-recovery
│   ├── models/registry.go       # model registry (340+)
│   ├── i18n/
│   │   ├── i18n.go              # embed.FS loader — auto-discovers locales
│   │   └── locales/             # ← drop xx.json here to add a language
│   │       ├── ko.json
│   │       ├── en.json
│   │       └── ...              # 17 languages total
│   └── hooks/hooks.go           # event hook system
├── configs/
│   ├── services/                # service plugin YAML
│   ├── example-standalone.yaml
│   └── example-distributed.yaml
└── docs/
    ├── logo.png
    ├── API.md
    └── MANUAL.md
```

---

## 🤓 Tech Stack

- **Language**: Go 1.22+ (single binary, zero runtime dependencies)
- **Encryption**: AES-256-GCM (crypto/rand nonce)
- **Realtime**: Server-Sent Events (SSE)
- **UI**: Server-rendered HTML (no frontend framework, no npm)
- **I18N**: embed.FS JSON files — add a language without touching Go code
- **Tests**: 39 unit tests (crypto / proxy / vault / middleware / hooks)
- **CI/CD**: GitHub Actions (5-platform cross-compile + auto Release)

---

## License

This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

Personal use and educational use are fully permitted.

If you wish to distribute modified versions or use this commercially, please contact the author beforehand.

> The author is a lazy daydreamer who loves to play, so whether new feature requests will make it into a release is anybody's guess — but keep nagging and maybe someday they'll get done. Once motivated, though, the work gets done well. lol

---

## 💀 탄생 배경: 봇들이 죽던 날

새벽이었다.

알림 하나가 떴다. 평소와 달랐다.

로그인해 보니 — API 키 전부 무효. vault.json 공백. 봇들은 아무 말도 없었다. 할 수가 없었다. **기억이 통째로 지워져 있었으니까.**

> *Bravo는 내 작업 스타일을 꿰뚫고 있었다. 미니는 매일 아침 브리핑을 준비했다. Charlie는 라즈베리파이 위에서 묵묵히 모든 걸 처리했다. 2주 넘게 공들여 키운 AI 비서들이었다.*
>
> *해커 한 명이 내부망에 들어와서 그걸 전부 날려버렸다.*
>
> *잘 키운 반려동물이 하룻밤 사이에 사라진 것 같은 기분이었다.*

기억을 복원하는 데 일주일이 걸렸다. 완전하지도 않았다.

이건 두 번 다시 겪으면 안 됐다.

그래서 만들었다. **키를 잠그는 금고. 봇들을 지키는 벽. 다시는 해커 한 명 때문에 모든 게 끝나지 않도록.**

---

## ⚔️ 그래서, 이게 뭐냐면

한 줄 요약: **"AI 봇들이 절대 죽지 않게 만드는 보디가드."**

```
해커가 키를 털어도  → 금고가 막는다
키 한도가 차도      → 다음 키로 알아서 넘긴다
서비스가 다운돼도   → Gemini → OpenAI → Ollama 순서로 폴백
봇이 100대여도      → 설정 하나 바꾸면 1-3초 내 전원에 반영
```

더 풀어쓰면:

- 🔐 **키 금고(Vault)**: AES-GCM 암호화. 라운드 로빈 자동 순환. 할당량·오류·쿨다운 알아서 관리.
- 🔀 **AI 프록시(Proxy)**: OpenClaw·Claude Code·VS Code·내 스크립트 — 어디서 오든 Gemini / OpenAI / Ollama로 중계. 하나 죽으면 다음 걸로.
- ⚡ **SSE 실시간 동기화**: 금고에서 뭔가 바꾸면 연결된 모든 봇에 즉각 반영. 재시작 불필요.
- 🛡️ **보안 필터**: function calling 완전 차단. 외부 스킬이 내 AI를 멋대로 조종하는 걸 막는다.
- 🦞 **OpenClaw 전용 연동**: Unix 소켓으로 TUI에 실시간 이벤트 전달. openclaw.json 자동 갱신.

Go 바이너리 단 하나. 봇 한 대부터 분산 다중 봇까지 전부 커버.

---

### 기능 목록

| 기능 | 설명 |
|------|------|
| **AI 프록시** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 자동 순환 |
| **AES-GCM 암호화** | 마스터 비밀번호로 API 키 암호화 저장 |
| **SSE 실시간 동기화** | 금고 ↔ 프록시 1–3초 내 자동 반영 |
| **도구 보안 필터** | function calling 차단 (`strip_all` / `whitelist` / `passthrough`) |
| **폴백 체인** | 서비스 실패 시 자동 전환, 최종 폴백은 Ollama |
| **모델 레지스트리** | OpenRouter 340개+ + 로컬 모델 동적 감지 |
| **로컬 AI 지원** | Ollama / LM Studio / vLLM 자동 감지 + 수동 URL |
| **서비스 관리** | UI에서 서비스 추가·수정·삭제 |
| **에이전트 관리** | 에이전트별 서비스·모델·IP·작업 디렉토리 설정 |
| **에이전트 상태** | 4단계 🟢실행중 / 🟡지연 / 🔴오프라인 / ⚫미연결 |
| **주치의(Doctor)** | 헬스체크, 자동복구, systemd/launchd/NSSM 등록 |
| **[17개 언어](#languages)** | 한국어·영어·중국어·일본어·스페인어·힌디어·아랍어·포르투갈어·프랑스어·독일어·태국어·몽골어·스와힐리어·하우사어·줄루어·네팔어·인도네시아어 기본 탑재. `locales/xx.json` 파일 하나로 어떤 언어든 추가 가능 — 코드 수정 불필요 |
| **테마** | 라이트 ☀️ / 다크 🌑 / 골드 ✨ / 벚꽃 🌸 / 오션 🌊 / 가을 🍂 / 겨울 ❄️ |
| **크로스 플랫폼** | Linux / macOS / Windows / WSL |

### 빠른 시작

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

./wall-vault setup   # 대화형 설치 마법사
./wall-vault start   # 프록시 + 키 금고 실행
```

브라우저에서 `http://localhost:56243` 을 열면 키 금고 대시보드가 나타납니다.

---

### 다국어 지원

wall-vault는 **17개 언어**를 기본 탑재하며, JSON 파일 하나만 추가하면 어떤 언어든 지원 가능합니다. Go 코드를 건드릴 필요가 없습니다.

| 코드 | 언어 | 코드 | 언어 | 코드 | 언어 |
|------|------|------|------|------|------|
| `ko` | 한국어 | `ar` | العربية | `th` | ภาษาไทย |
| `en` | English | `pt` | Português | `mn` | Монгол |
| `zh` | 中文 | `fr` | Français | `sw` | Kiswahili |
| `ja` | 日本語 | `de` | Deutsch | `ha` | Hausa |
| `es` | Español | `id` | Indonesia | `zu` | IsiZulu |
| `hi` | हिन्दी | `ne` | नेपाली | | |

- 설치 마법사, 대시보드 UI, 시스템 메시지 전체에 적용
- 대시보드에서 언어 즉시 변경 (페이지 새로고침 불필요)

**새 언어 추가 방법 (5분이면 충분)**:

```bash
# 1. 기존 파일을 복사해 시작
cp internal/i18n/locales/ko.json internal/i18n/locales/tr.json

# 2. 번역 (키는 그대로, 값만 번역)
# "lang_label": "Türkçe", "lang_emoji": "🇹🇷", ...

# 3. 빌드
make build
```

빌드만 다시 하면 새 언어가 마법사·대시보드·UI 전체에 자동 반영됩니다.

---

## 🇨🇳 项目简介

> *"上个月，一名黑客入侵了我们实验室的内网，造成了严重破坏。"*
>
> *"精心培育了两周多的 AI 助手机器人的所有记忆，瞬间全部消失。"*
>
> *"就是为了这个，才有了这个项目。"*

**wall-vault** 是一个 AI 代理 + 密钥保险库一体化系统。

### 主要功能

| 功能 | 说明 |
|------|------|
| **AI 代理** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **密钥保险库** | API 密钥 AES-GCM 加密存储，自动轮换 |
| **实时同步** | 保险库设置变更后 1–3 秒内同步到所有代理 |
| **安全过滤** | 阻断外部 function calling（strip_all 模式）|
| **自动故障转移** | 服务失败时自动切换，最终回退到本地 Ollama |
| **多语言** | 17 种语言，添加 JSON 文件即可扩展 |
| **主题** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |

### 快速开始

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

打开 `http://localhost:56243` 即可访问控制面板。

---

## 🇯🇵 はじめに

> *「2週間以上かけて育ててきた AI アシスタントボットの記憶が、ハッカーに一瞬で全て消されてしまった。」*
>
> *「だから、このプロジェクトを始めた。」*

**wall-vault** は AI プロキシ + キー金庫の統合システムです。

### 主な機能

| 機能 | 説明 |
|------|------|
| **AI プロキシ** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **キー金庫** | API キーを AES-GCM で暗号化・自動ローテーション |
| **リアルタイム同期** | 設定変更が 1〜3 秒以内に全プロキシへ反映 |
| **セキュリティフィルタ** | 外部 function calling を完全ブロック |
| **多言語** | 17言語対応、JSON ファイルを追加するだけで拡張可能 |
| **テーマ** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |

### クイックスタート

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

ブラウザで `http://localhost:56243` を開くとダッシュボードが表示されます。

---

## 🇪🇸 Introducción

**wall-vault** es un sistema integrado de proxy de IA y bóveda de claves API.

### Características principales

| Función | Descripción |
|---------|-------------|
| **Proxy de IA** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Bóveda de claves** | Cifrado AES-GCM, rotación automática |
| **Sync en tiempo real** | Cambios reflejados en 1–3 segundos (SSE) |
| **Filtro de seguridad** | Bloqueo total de function calling externo |
| **17 idiomas** | Añade un archivo JSON para soportar un nuevo idioma |
| **Temas** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |

### Inicio rápido

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Abra `http://localhost:56243` en el navegador.

---

## 🇫🇷 Présentation

**wall-vault** est un système intégré proxy IA + coffre-fort de clés API.

### Fonctionnalités

| Fonction | Description |
|----------|-------------|
| **Proxy IA** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Coffre-fort** | Chiffrement AES-GCM, rotation automatique des clés |
| **Sync temps réel** | Changements reflétés en 1–3 secondes (SSE) |
| **Filtre sécurité** | Blocage total du function calling externe |
| **17 langues** | Ajoutez un fichier JSON pour une nouvelle langue |
| **Thèmes** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |

### Démarrage rapide

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Ouvrez `http://localhost:56243` dans votre navigateur.

---

## 🇩🇪 Über das Projekt

**wall-vault** ist ein integriertes KI-Proxy- und API-Schlüssel-Tresor-System.

### Hauptfunktionen

| Funktion | Beschreibung |
|----------|--------------|
| **KI-Proxy** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Schlüsseltresor** | AES-GCM-Verschlüsselung, automatische Rotation |
| **Echtzeit-Sync** | Änderungen in 1–3 Sekunden übertragen (SSE) |
| **Sicherheitsfilter** | Vollständige Blockierung externen Function Callings |
| **17 Sprachen** | Neue Sprache per JSON-Datei hinzufügen |
| **Designs** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |

### Schnellstart

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Öffnen Sie `http://localhost:56243` im Browser.

---

<p align="center">
  <b>sookmook · Sookmook Future Informatics Foundation</b><br>
  <i>"An AI bot's memory is precious. Protect it."</i><br>
  <i>"AI 봇의 기억은 소중하다. 지키자."</i>
</p>

---

*Last updated · 최종 업데이트: 2026-03-13*
