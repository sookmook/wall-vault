<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

<h1 align="center">🔐 wall-vault</h1>

<p align="center"><i>OpenClaw을 위한 AI 프록시 + API 키 금고 — 어떤 상황에서도 오픈클로가 끊기지 않게</i></p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3-blue.svg" alt="License: GPL v3"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg" alt="Go Version">
  <a href="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml"><img src="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <img src="https://img.shields.io/badge/languages-17-brightgreen.svg" alt="Languages">
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg" alt="Platform">
  <img src="https://img.shields.io/badge/clients-OpenClaw%20%7C%20NanoClaw%20%7C%20Claude%20Code%20%7C%20Cursor%20%7C%20VSCode-purple.svg" alt="Clients">
</p>

---

## Language · 언어

| [🇺🇸 English](#-origin-story-the-night-the-bots-died) | [🇰🇷 한국어](#-탄생-배경-봇들이-죽던-날) | [🇨🇳 中文](#-项目简介) | [🇯🇵 日本語](#-はじめに) | [🇪🇸 Español](#-introducción) | [🇫🇷 Français](#-présentation) | [🇩🇪 Deutsch](#-über-das-projekt) |

---

## 💀 Origin Story: The Night the Bots Died

It was the middle of the night.

One alert. Something felt wrong.

Logged in — all API keys invalidated. vault.json empty. The bots had gone silent. Of course they had. **Their memories had been completely erased.**

> *Alpha knew my work style inside-out. Beta prepared morning briefings every day. Gamma handled everything quietly from a Raspberry Pi in the corner. Two weeks of careful cultivation. Two weeks of patience, tuning, and personality-shaping.*
>
> *A single hacker broke into the lab's internal network and torched all of it.*
>
> *It felt like coming home to find a beloved pet had simply vanished.*

It took a week to restore most of the memories. Not all of them came back.

This could never happen again.

So I built something. **A vault for the keys. A wall for the bots. A guarantee that no single attack could ever end everything again.**

---

## ⚔️ What It Actually Is

One line: **"OpenClaw's backbone — routes, rotates, and recovers so your AI never stops."**

wall-vault was built for **OpenClaw**. It sits between OpenClaw and the LLM APIs, handling everything that would otherwise interrupt a session:

```
Google key hits rate limit?          → Rotates to next key. OpenClaw keeps going.
OpenRouter quota runs out?           → Falls back to Ollama. No interruption.
Key gets stolen?                     → Vault blocks it. Next key takes over.
Running OpenClaw on 3 machines?      → Change model once. All 3 update in 1–3s via SSE.
Want to switch from Gemini to Kimi?  → One click in the vault dashboard. Done.
```

In more detail:

- 🦞 **OpenClaw Integration**: The main event. Live events over Unix socket to TUI. Auto-updates openclaw.json. SSE model sync in 1–3s.
- 🔐 **Key Vault**: AES-GCM encrypted storage. Round-robin rotation. Quota, cooldown, and error handling — all automatic.
- 🔀 **AI Proxy**: OpenClaw sends requests here, wall-vault routes them to Gemini / OpenRouter / Ollama. One dies, the next one picks up.
- ⚡ **SSE Real-time Sync**: Change anything in the vault, every connected proxy reflects it instantly. No restarts.
- 🛡️ **Security Filter**: Full function calling block. Stops external skills from hijacking your AI.

Single Go binary. Works with Claude Code, Cursor, VS Code too — but OpenClaw is what it's built for.

---

## 🔌 Works With Everything

wall-vault was built for OpenClaw — but since it speaks **four API formats**, other clients connect too.
Point any AI client at `http://your-host:56244` and it just works.

| Client | Endpoint | Format | Setup |
|--------|----------|--------|-------|
| **OpenClaw** ⭐ | `/google/v1beta/models/...` | Gemini | Built-in — just point to proxy port |
| **NanoClaw** ⭐ | `/google/v1beta/models/...` | Gemini | Same config format as OpenClaw (`~/.openclaw/openclaw.json`) |
| **Gemini CLI** | `/google/v1beta/models/...` | Gemini | `GEMINI_API_BASE_URL=http://localhost:56244` |
| **Antigravity IDE** | `/google/v1beta/models/...` | Gemini | `GEMINI_API_BASE_URL=http://localhost:56244` |
| **Claude Code** | `/v1/messages` | Anthropic | `ANTHROPIC_BASE_URL=http://localhost:56244` |
| **Cursor** | `/v1/chat/completions` | OpenAI | Set base URL in Cursor settings |
| **VS Code / Continue** | `/v1/chat/completions` | OpenAI | Set `apiBase` in Continue config |
| **LM Studio apps** | `/v1/chat/completions` | OpenAI | Any OpenAI-compatible client |
| **Custom scripts** | `/v1/chat/completions` | OpenAI | `curl -X POST .../v1/chat/completions` |

---

### Gemini CLI Setup

> **Note**: Gemini CLI v0.25.0 does not yet support custom API endpoints natively.
> Native support is tracked in [google-gemini/gemini-cli#1679](https://github.com/google-gemini/gemini-cli/issues/1679).
> Once merged, the setup will be one line:
> ```bash
> export GOOGLE_GEMINI_BASE_URL=http://localhost:56244
> gemini  # routes through wall-vault → key rotation + fallback
> ```

Current workaround — use `mitmproxy` to intercept Gemini API traffic:

```bash
# 1. Install mitmproxy
pip install mitmproxy

# 2. Run mitmproxy that forwards googleapis.com to wall-vault (port 56244)
mitmdump --mode reverse:https://generativelanguage.googleapis.com --listen-port 8080

# 3. Point Gemini CLI at the interceptor
# ~/.gemini/settings.json:
{
  "proxy": "http://localhost:8080"
}
```

> Once Gemini CLI adds native base URL support, this will simplify to a single env var.

---

### Antigravity IDE Setup

Antigravity is Google's agentic AI IDE (similar position to Claude Code). It uses the Gemini API format. Configuration depends on the version:

```bash
# If Antigravity respects GEMINI_API_KEY + endpoint env vars:
export GEMINI_API_KEY=<your-vault-agent-token>
export GOOGLE_GEMINI_BASE_URL=http://localhost:56244
# Then launch Antigravity
```

> Check Antigravity's documentation for the exact env var name — it varies by version.
> If it supports `GOOGLE_GEMINI_BASE_URL`, wall-vault handles the rest:
> key rotation, fallback to OpenRouter/Ollama, and live model switching.

---

### Claude Code Setup

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244
claude  # now routes through wall-vault → Gemini / OpenRouter / Ollama
```

> When Anthropic keys run dry, wall-vault silently falls back to Gemini or Ollama.
> Claude Code keeps working. You don't notice.

---

### Cursor / VS Code Setup

In Cursor **Settings → AI → OpenAI API**:
```
Base URL:  http://localhost:56244
API Key:   (any string — wall-vault ignores it)
Model:     gemini-2.5-flash  (or any model from /v1/models)
```

Same for VS Code + Continue extension:
```json
{
  "models": [{
    "title": "wall-vault",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://localhost:56244"
  }]
}
```

---

> **Behind the scenes**: all four formats are converted to a unified internal format,
> then dispatched through the same fallback chain: Primary → OpenRouter → Ollama.
> One vault. One config. Every client.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [OpenClaw Integration](#openclaw-integration) ⭐
- [Languages](#languages)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Supported Services](#supported-services)
- [API Reference](#api-reference)
- [Modes](#modes)
- [Auto-Start](#auto-start)
- [Internal Network Setup](#internal-network-setup)
- [Build](#build)
- [Project Structure](#project-structure)
- [License](#license)

---

## Features

| Feature | Description |
|---------|-------------|
| **[OpenClaw Integration](#openclaw-integration)** ⭐ | The primary use case — Unix socket TUI events, openclaw.json auto-config, SSE model sync |
| **NanoClaw Support** ⭐ | Lightweight OpenClaw-compatible agent — same `~/.openclaw/openclaw.json` config, default workdir `~/nanoclaw` |
| **AI Proxy** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **Client Support** | OpenClaw / NanoClaw / Claude Code / Gemini CLI / Antigravity IDE / Cursor / VS Code / LM Studio / scripts |
| **Key Vault** | API key management, usage monitoring, round-robin rotation |
| **AES-GCM Encryption** | Keys encrypted with master password, never stored in plaintext |
| **SSE Real-time Sync** | Vault ↔ proxy config sync within 1–3 seconds |
| **Tool Security Filter** | Block function calling (`strip_all` / `whitelist` / `passthrough`) |
| **Fallback Chain** | Auto-switch on service failure, final fallback to local Ollama |
| **Model Registry** | 340+ OpenRouter models + dynamic local model discovery |
| **Local AI Support** | Ollama / LM Studio / vLLM auto-detection + manual URL |
| **Service Management** | Add/edit/delete services from UI, custom service support |
| **Proxy Service Filter** | "프록시 사용" checkbox per service — only checked services appear in agent model dropdowns and are dispatched by the proxy |
| **Service Auto-check** | Dashboard load / key change → cloud services auto-enable/disable by key count |
| **Agent Management** | Per-agent service / model / IP whitelist / workdir / avatar (path or data URI) |
| **Agent Status** | 4-state: 🟢Online / 🟡Delayed / 🔴Offline / ⚫Disconnected |
| **Bidirectional Model Sync** | TUI model change → vault; vault change → TUI. All sources stay in sync. |
| **Per-type Config Copy** | 🦞 openclaw / 🦀 nanoclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — one-click config snippet |
| **Doctor** | Health check, auto-recovery, systemd/launchd/NSSM registration |
| **[17 Languages](#languages)** | Korean · English · Chinese · Japanese · Spanish · Hindi · Arabic · Portuguese · French · German · Thai · Mongolian · Swahili · Hausa · Zulu · Nepali · Indonesian |
| **Themes** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 / Autumn 🍂 / Winter ❄️ |
| **Cross-platform** | Linux / macOS / Windows / WSL |

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

## OpenClaw Integration

**OpenClaw** is a distributed AI agent TUI framework that runs personas with long-term memory across multiple devices. wall-vault was built specifically to serve OpenClaw — the two systems are deeply integrated.

### Step 1: Register an OpenClaw Agent

In the dashboard **Add Agent** modal, set agent type to `openclaw` or `nanoclaw`:
- `openclaw` — work directory auto-fills as `~/.openclaw`
- `nanoclaw` — lightweight variant, work directory auto-fills as `~/nanoclaw`; reuses the same `~/.openclaw/openclaw.json` config format
- wall-vault becomes the API key supplier and proxy for that agent

```bash
# Run proxy for OpenClaw (distributed mode)
VAULT_CLIENT_ID=bot-a \
VAULT_URL=http://192.168.x.x:56243 \
VAULT_TOKEN=your-agent-token \
wall-vault proxy
```

### Step 2: Point openclaw.json at the proxy

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },     // free 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

The dashboard's **🦞 OpenClaw 설정 복사** button generates this snippet automatically.

### Model routing with `wall-vault/` prefix

| Model prefix | Routes to |
|---|---|
| `wall-vault/gemini-*` | Google Gemini (direct) |
| `wall-vault/gpt-*` / `wall-vault/o3` / `wall-vault/o4*` | OpenAI (direct) |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free, 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| All `opencode/`, `moonshot/`, `kimi-coding/`, `groq/`, `deepseek/`, `qwen/`, `meta-llama/` | OpenRouter |
| `model-name:cloud` | `:cloud` stripped → OpenRouter |

### Unix Socket Events (TUI live notifications)

wall-vault sends real-time events to OpenClaw's TUI over a Unix socket:

```yaml
# wall-vault.yaml
hooks:
  openclaw_socket: ~/.openclaw/wall-vault.sock
```

| Event | Trigger |
|-------|---------|
| `model_changed` | Model/service switch |
| `key_exhausted` | API key daily limit reached |
| `service_down` | Service failure or all keys on cooldown |
| `ollama_waiting` | Waiting for local Ollama response |
| `tui_footer` | Status message pushed to TUI footer |

### SSE Auto-Sync

OpenClaw agents subscribe to the wall-vault SSE stream. Model or service changes in the vault are applied within **1–3 seconds** — no TUI restart needed.

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
| 582 Gateway Overload | 5 minutes |
| Network error | 10 minutes |

> 429 / 402 / 582 errors increment `today_attempts` (total call count) but not `today_usage` (successful tokens only).

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
| `WV_AVATAR` | Proxy avatar file path (relative to `~/.openclaw/`, e.g. `workspace/avatars/avatar.png`) |

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

## Internal Network Setup

One vault. Multiple proxies. All machines on the same internal network.

```
Internal Network (e.g. 10.0.0.x)
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  [Mac Mini :56243]          [WSL / Linux]               │
│  Key Vault                  Proxy A (bot-a)             │
│  vault.json                 → VAULT_URL=192.168.x.x     │
│                                                         │
│                             [Raspberry Pi]              │
│                             Proxy B (bot-b)             │
│                             → VAULT_URL=192.168.x.x     │
│                                                         │
│                             [Windows / Mac]             │
│                             Proxy C (bot-c)             │
│                             → VAULT_URL=127.0.0.1       │
└─────────────────────────────────────────────────────────┘
```

### Step 1: Run the vault (on the host machine)

```bash
# On the machine that will host the vault (e.g. 192.168.x.x)
./wall-vault vault

# Or with explicit config
./wall-vault vault --config ~/.wall-vault/vault.yaml
```

Default: vault listens on `0.0.0.0:56243` — accessible from all network interfaces.

### Step 2: Create client tokens (from the dashboard)

Open `http://192.168.x.x:56243` → **Add Agent**:

| Field | Example |
|-------|---------|
| ID | `bot-a` |
| Name | `봇 A` |
| Agent Type | `openclaw` |
| Default Service | `google` |
| Token | leave blank → auto-generated, shown once |
| IP Whitelist | `<proxy-ip>` (WSL IP) — optional but recommended |

Copy the generated token. You will need it for each proxy.

> **Tip:** Set IP whitelist per client to prevent token abuse if a machine is compromised.

### Step 3: Deploy each proxy

```bash
# Proxy on WSL (bot-a)
VAULT_URL=http://192.168.x.x:56243 \
VAULT_CLIENT_ID=bot-a \
VAULT_TOKEN=your-bot-a-token \
./wall-vault proxy

# Proxy on Raspberry Pi (bot-b)
VAULT_URL=http://192.168.x.x:56243 \
VAULT_CLIENT_ID=bot-b \
VAULT_TOKEN=your-bot-b-token \
./wall-vault proxy

# Proxy on the same machine as vault (bot-c — use localhost)
VAULT_URL=http://127.0.0.1:56243 \
VAULT_CLIENT_ID=bot-c \
VAULT_TOKEN=your-bot-c-token \
./wall-vault proxy
```

Each proxy:
1. Connects to the vault on startup
2. Subscribes to the SSE stream (`/api/events`)
3. Fetches the active API key for its assigned service
4. Reports heartbeat every 60 seconds

### Step 4: Verify SSE sync

```bash
# Watch the SSE stream from vault
curl -s http://192.168.x.x:56243/api/events --max-time 10

# Check proxy status (shows SSE connection state)
curl http://localhost:56244/status

# Change model on vault → confirm proxy reflects it within 3s
curl -X PUT http://192.168.x.x:56243/admin/clients/bot-a \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"default_model":"gemini-2.5-pro"}'

curl http://localhost:56244/status | grep model
```

### Firewall / Port Guide

| Port | Service | Open to |
|------|---------|---------|
| 56243 | Key Vault (dashboard + API) | internal network only |
| 56244 | AI Proxy (per machine) | localhost only (or per-bot clients) |
| 11434 | Ollama (if shared) | internal network |

```bash
# Linux — allow vault from internal network only
sudo ufw allow from 10.0.0.0/24 to any port 56243
sudo ufw deny 56243

# macOS — block vault from external
# In System Settings → Firewall, allow wall-vault only on LAN interface
```

### Config file approach (recommended for production)

Instead of environment variables, use a config file per machine:

```yaml
# ~/.wall-vault/proxy-bot-a.yaml
mode: distributed
proxy:
  port: 56244
  client_id: bot-a
  vault_url: http://192.168.x.x:56243
  vault_token: your-bot-a-token
  tool_filter: strip_all
```

```bash
./wall-vault proxy --config ~/.wall-vault/proxy-bot-a.yaml
```

### What syncs, what doesn't

| Setting | Syncs via SSE? | Stored where |
|---------|---------------|--------------|
| Active API key | ✅ Yes | vault.json (encrypted) |
| Service / Model per agent | ✅ Yes | vault.json |
| Proxy port | ❌ No | local config |
| Tool filter mode | ❌ No | local config |
| Theme / Language | ❌ No | vault.json (per vault) |

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

> *알파는 내 작업 스타일을 꿰뚫고 있었다. 베타는 매일 아침 브리핑을 준비했다. 감마는 라즈베리파이 위에서 묵묵히 모든 걸 처리했다. 2주 넘게 공들여 키운 AI 비서들이었다.*
>
> *해커 한 명이 내부망에 들어와서 그걸 전부 날려버렸다.*
>
> *잘 키운 반려동물이 하룻밤 사이에 사라진 것 같은 기분이었다.*

기억을 복원하는 데 일주일이 걸렸다. 완전하지도 않았다.

이건 두 번 다시 겪으면 안 됐다.

그래서 만들었다. **키를 잠그는 금고. 봇들을 지키는 벽. 다시는 해커 한 명 때문에 모든 게 끝나지 않도록.**

---

## ⚔️ 그래서, 이게 뭐냐면

한 줄 요약: **"오픈클로가 어떤 상황에서도 LLM 서비스를 끊김 없이 쓸 수 있게 하는 AI 프록시 + API 키 금고."**

오픈클로와 LLM API 사이에 앉아서, 세션을 방해할 모든 요소를 대신 처리한다:

```
구글 키 한도 초과    → 다음 키로 자동 전환. 오픈클로는 계속된다.
OpenRouter 크레딧 소진 → Ollama로 폴백. 끊김 없음.
키가 탈취됨          → 금고가 막는다. 다음 키가 투입된다.
머신 3대에서 오픈클로 실행 중 → 모델 변경 한 번. 3대 전부 1-3초 내 반영.
```

더 풀어쓰면:

- 🦞 **OpenClaw 전용 연동**: 핵심 목적. Unix 소켓으로 TUI에 실시간 이벤트 전달. openclaw.json 자동 갱신. SSE 모델 동기화 1-3초.
- 🔐 **키 금고(Vault)**: AES-GCM 암호화. 라운드 로빈 자동 순환. 할당량·오류·쿨다운 알아서 관리.
- 🔀 **AI 프록시(Proxy)**: 오픈클로가 여기로 요청을 보내면, Gemini / OpenRouter / Ollama로 라우팅. 하나 죽으면 다음 걸로.
- ⚡ **SSE 실시간 동기화**: 금고에서 뭔가 바꾸면 연결된 모든 프록시에 즉각 반영. 재시작 불필요.
- 🛡️ **보안 필터**: function calling 완전 차단. 외부 스킬이 오픈클로를 멋대로 조종하는 걸 막는다.

Go 바이너리 단 하나. Claude Code·Cursor·VS Code도 연결 가능하지만, 오픈클로가 본래 목적.

---

### 기능 목록

| 기능 | 설명 |
|------|------|
| **OpenClaw 전용 연동** ⭐ | 핵심 목적 — Unix 소켓 TUI 이벤트, openclaw.json 자동 설정, SSE 모델 동기화 |
| **NanoClaw 지원** ⭐ | 경량 OpenClaw 호환 에이전트 — 동일한 openclaw.json 설정 사용, 기본 작업 디렉토리 `~/nanoclaw` |
| **AI 프록시** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **클라이언트 지원** | OpenClaw / NanoClaw / Claude Code / Gemini CLI / Cursor / VS Code / 스크립트 |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 자동 순환 |
| **AES-GCM 암호화** | 마스터 비밀번호로 API 키 암호화 저장 |
| **SSE 실시간 동기화** | 금고 ↔ 프록시 1–3초 내 자동 반영 |
| **도구 보안 필터** | function calling 차단 (`strip_all` / `whitelist` / `passthrough`) |
| **폴백 체인** | 서비스 실패 시 자동 전환, 최종 폴백은 Ollama |
| **모델 레지스트리** | OpenRouter 340개+ + 로컬 모델 동적 감지 |
| **로컬 AI 지원** | Ollama / LM Studio / vLLM 자동 감지 + 수동 URL |
| **서비스 관리** | UI에서 서비스 추가·수정·삭제 |
| **프록시 서비스 필터** | 서비스 카드의 "프록시 사용" 체크박스 — 체크된 서비스만 에이전트 모델 드롭다운에 표시되고 프록시 dispatch에 포함됨 |
| **에이전트 관리** | 에이전트별 서비스·모델·IP·작업 디렉토리 설정 |
| **에이전트 상태** | 4단계 🟢실행중 / 🟡지연 / 🔴오프라인 / ⚫미연결 |
| **주치의(Doctor)** | 헬스체크, 자동복구, systemd/launchd/NSSM 등록 |
| **[17개 언어](#languages)** | 한국어·영어·중국어·일본어·스페인어·힌디어·아랍어·포르투갈어·프랑스어·독일어·태국어·몽골어·스와힐리어·하우사어·줄루어·네팔어·인도네시아어 기본 탑재 |
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

### 내부망 분산 운영 (금고 1대 + 프록시 N대)

내부망에서 금고 한 대를 두고 여러 프록시를 연결하는 구성입니다.

```
내부망 (예: 10.0.0.x)
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  [맥미니 :56243]            [WSL / Linux]               │
│  키 금고 (vault)            프록시 알파                  │
│  vault.json 저장            VAULT_URL=192.168.x.x       │
│                                                         │
│                             [라즈베리파이]              │
│                             프록시 감마                  │
│                             VAULT_URL=192.168.x.x       │
│                                                         │
│                             [맥미니 로컬]               │
│                             프록시 베타                  │
│                             VAULT_URL=127.0.0.1         │
└─────────────────────────────────────────────────────────┘
```

#### 1단계: 금고 실행 (호스트 머신)

```bash
# 금고 호스트 (예: 192.168.x.x)
./wall-vault vault
```

기본값으로 `0.0.0.0:56243`에서 수신 — 내부망 전체에서 접근 가능.

#### 2단계: 클라이언트 토큰 생성 (대시보드)

`http://192.168.x.x:56243` → **에이전트 추가**:

| 항목 | 예시 |
|------|------|
| ID | `bot-a` |
| 이름 | `봇 A` |
| 에이전트 종류 | `openclaw` |
| 기본 서비스 | `google` |
| 토큰 | 빈칸 → 자동 생성 (한 번만 표시) |
| 허용 IP | `<proxy-ip>` (WSL IP) — 선택, 권장 |

생성된 토큰을 복사해 각 프록시에 사용합니다.

#### 3단계: 각 머신에서 프록시 실행

```bash
# WSL (알파)
VAULT_URL=http://192.168.x.x:56243 \
VAULT_CLIENT_ID=bot-a \
VAULT_TOKEN=your-bot-a-token \
./wall-vault proxy

# 라즈베리파이 (감마)
VAULT_URL=http://192.168.x.x:56243 \
VAULT_CLIENT_ID=bot-b \
VAULT_TOKEN=your-bot-b-token \
./wall-vault proxy

# 맥미니 로컬 (베타)
VAULT_URL=http://127.0.0.1:56243 \
VAULT_CLIENT_ID=bot-c \
VAULT_TOKEN=your-bot-c-token \
./wall-vault proxy
```

#### 4단계: SSE 동기화 확인

```bash
# SSE 스트림 직접 확인
curl -s http://192.168.x.x:56243/api/events --max-time 10

# 프록시 상태 확인 (SSE 연결 포함)
curl http://localhost:56244/status

# 금고에서 모델 변경 → 3초 내 프록시에 반영되는지 확인
curl -X PUT http://192.168.x.x:56243/admin/clients/bot-a \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"default_model":"gemini-2.5-pro"}'
```

#### 포트 방화벽 설정

| 포트 | 서비스 | 오픈 범위 |
|------|--------|-----------|
| 56243 | 키 금고 (대시보드+API) | 내부망만 |
| 56244 | AI 프록시 (머신별) | localhost 권장 |
| 11434 | Ollama (공유 시) | 내부망 |

```bash
# Linux: 내부망에서만 금고 접근 허용
sudo ufw allow from 10.0.0.0/24 to any port 56243
sudo ufw deny 56243
```

#### 동기화 범위

| 설정 | SSE 동기화 | 저장 위치 |
|------|-----------|----------|
| 활성 API 키 | ✅ 실시간 | vault.json (암호화) |
| 서비스·모델 (에이전트별) | ✅ 실시간 | vault.json |
| 프록시 포트 | ❌ 없음 | 로컬 설정 |
| 도구 필터 모드 | ❌ 없음 | 로컬 설정 |
| 테마·언어 | ❌ 없음 | vault.json (금고별) |

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

*Last updated · 최종 업데이트: 2026-03-20 — v0.1.8 (NanoClaw agent type: 🦀 badge, config copy button, default workdir ~/nanoclaw; daily reset stale counter fix; drain-first key selection; streaming token count fix)*
