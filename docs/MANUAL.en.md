# wall-vault User Manual
*(Last updated: 2026-04-08 — v0.1.25)*

---

## Table of Contents

1. [What is wall-vault?](#what-is-wall-vault)
2. [Installation](#installation)
3. [Getting Started (setup wizard)](#getting-started)
4. [Registering API Keys](#registering-api-keys)
5. [Using the Proxy](#using-the-proxy)
6. [Key Vault Dashboard](#key-vault-dashboard)
7. [Distributed Mode (Multi-Bot)](#distributed-mode-multi-bot)
8. [Auto-Start Configuration](#auto-start-configuration)
9. [Doctor](#doctor)
10. [RTK Token Saving](#rtk-token-saving)
11. [Environment Variables Reference](#environment-variables-reference)
12. [Troubleshooting](#troubleshooting)

---

## What is wall-vault?

**wall-vault = AI proxy + API key vault for OpenClaw**

To use AI services, you need **API keys**. An API key is like a **digital pass** that proves "this person is authorized to use this service." However, these passes have daily usage limits, and if mismanaged, they risk being exposed.

wall-vault stores these passes in a secure vault and acts as a **proxy** between OpenClaw and AI services. In simple terms, OpenClaw only needs to connect to wall-vault, and wall-vault handles everything else automatically.

Problems wall-vault solves:

- **Automatic API key rotation**: When one key hits its usage limit or gets temporarily blocked (cooldown), it silently switches to the next key. OpenClaw continues working without interruption.
- **Automatic service fallback**: If Google doesn't respond, it switches to OpenRouter; if that fails too, it switches to locally installed Ollama, LM Studio, or vLLM (local AI) on your computer. Sessions never drop. When the original service recovers, it automatically switches back on the next request (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Real-time sync (SSE)**: When you change a model in the vault dashboard, it reflects in the OpenClaw screen within 1-3 seconds. SSE (Server-Sent Events) is a technology where the server pushes changes to clients in real time.
- **Real-time notifications**: Events like key exhaustion or service outages are immediately displayed at the bottom of the OpenClaw TUI (terminal screen).

> 💡 **Claude Code, Cursor, and VS Code** can also be connected, but wall-vault's primary purpose is to be used with OpenClaw.

```
OpenClaw (TUI terminal screen)
        │
        ▼
  wall-vault proxy (:56244)   ← key management, routing, fallback, events
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ models)
        ├─ Ollama / LM Studio / vLLM (your computer, last resort)
        └─ OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Open a terminal and paste the following commands:

```bash
# Linux (regular PC, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Downloads a file from the internet.
- `chmod +x` — Makes the downloaded file "executable." If you skip this step, you'll get a "permission denied" error.

### Windows

Open PowerShell (as administrator) and run the following commands:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Add to PATH (takes effect after PowerShell restart)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **What is PATH?** It's the list of folders where your computer looks for commands. You need to add wall-vault to PATH so you can run `wall-vault` from any folder.

### Building from Source (for developers)

Only applicable if you have a Go development environment installed.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build timestamp version**: When built with `make build`, the version is automatically generated in a format like `v0.1.25.20260408.022325` that includes the date and time. If you build directly with `go build ./...`, the version will only show as `"dev"`.

---

## Getting Started

### Running the setup wizard

After installation, you must run the **setup wizard** with the following command. The wizard guides you through each required setting step by step.

```bash
wall-vault setup
```

The wizard walks you through these steps:

```
1. Language selection (10 languages including Korean)
2. Theme selection (light / dark / gold / cherry / ocean)
3. Operation mode — standalone (single use) or distributed (multi-machine)
4. Bot name — the name displayed on the dashboard
5. Port settings — defaults: proxy 56244, vault 56243 (press Enter to keep defaults)
6. AI service selection — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool security filter settings
8. Admin token — a password to lock dashboard admin functions. Can be auto-generated
9. API key encryption password — for extra-secure key storage (optional)
10. Config file save location
```

> ⚠️ **Make sure to remember your admin token.** You'll need it later to add keys or change settings in the dashboard. If you lose it, you'll have to edit the config file directly.

Once the wizard completes, a `wall-vault.yaml` config file is automatically generated.

### Running

```bash
wall-vault start
```

Two servers start simultaneously:

- **Proxy** (`http://localhost:56244`) — the agent connecting OpenClaw and AI services
- **Key Vault** (`http://localhost:56243`) — API key management and web dashboard

Open `http://localhost:56243` in your browser to access the dashboard.

---

## Registering API Keys

There are four ways to register API keys. **Method 1 (environment variables) is recommended for beginners.**

### Method 1: Environment Variables (recommended — simplest)

Environment variables are **pre-set values** that programs read when they start. Enter them in your terminal like this:

```bash
# Register a Google Gemini key
export WV_KEY_GOOGLE=AIzaSy...

# Register an OpenRouter key
export WV_KEY_OPENROUTER=sk-or-v1-...

# Start after registration
wall-vault start
```

If you have multiple keys, connect them with commas. wall-vault will automatically rotate through them (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tip**: The `export` command only applies to the current terminal session. To persist across reboots, add the lines to your `~/.bashrc` or `~/.zshrc` file.

### Method 2: Dashboard UI (point and click)

1. Open `http://localhost:56243` in your browser
2. Click `[+ Add]` in the **🔑 API Keys** card at the top
3. Enter the service type, key value, label (descriptive name), and daily limit, then save

### Method 3: REST API (for automation/scripting)

REST API is a method for programs to exchange data over HTTP. Useful for automated registration via scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Main Key",
    "daily_limit": 1000
  }'
```

### Method 4: Proxy Flags (for quick testing)

For temporary testing without formal registration. Keys are lost when the program exits.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Using the Proxy

### Using with OpenClaw (primary purpose)

Here's how to configure OpenClaw to connect to AI services through wall-vault.

Open `~/.openclaw/openclaw.json` and add the following:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault agent token
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // free 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Easier method**: Click the **🦞 Copy OpenClaw Config** button on the dashboard agent card. A snippet with the token and address already filled in will be copied to your clipboard. Just paste it.

**Where does the `wall-vault/` prefix in model names route to?**

wall-vault automatically determines which AI service to route requests to based on the model name:

| Model Format | Target Service |
|-------------|---------------|
| `wall-vault/gemini-*` | Direct to Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Direct to OpenAI |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free 1M token context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/model`, `openai/model`, `anthropic/model`, etc. | Direct to the corresponding service |
| `custom/google/model`, `custom/openai/model`, etc. | Strips `custom/` prefix and re-routes |
| `model:cloud` | Strips `:cloud` suffix and routes to OpenRouter |

> 💡 **What is context?** It's the amount of conversation an AI can remember at once. 1M (one million tokens) means it can process very long conversations or documents in a single pass.

### Direct Gemini API Format Connection (for existing tool compatibility)

If you have tools that were using the Google Gemini API directly, just change the address to wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Or if the tool specifies URLs directly:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Using with OpenAI SDK (Python)

You can also connect wall-vault from Python code that uses AI. Just change the `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault manages API keys for you
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # use provider/model format
    messages=[{"role": "user", "content": "Hello"}]
)
```

### Changing Models While Running

To change the AI model while wall-vault is already running:

```bash
# Change model by requesting the proxy directly
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# In distributed mode (multi-bot), change from the vault server → instantly synced via SSE
curl -X PUT http://localhost:56243/admin/clients/my-bot-id \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Checking Available Models

```bash
# View full list
curl http://localhost:56244/api/models | python3 -m json.tool

# View only Google models
curl "http://localhost:56244/api/models?service=google"

# Search by name (e.g., models containing "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Key models by service:**

| Service | Key Models |
|---------|-----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context free, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Auto-detects models from your local server |
| LM Studio | Local server (port 1234) |
| vLLM | Local server (port 8000) |

---

## Key Vault Dashboard

Open `http://localhost:56243` in your browser to access the dashboard.

**Layout:**
- **Top bar (fixed)**: Logo, language/theme selector, SSE connection status
- **Card grid**: Agent, service, and API key cards arranged in a tile layout

### API Key Card

A card for managing all registered API keys at a glance.

- Shows key lists organized by service.
- `today_usage`: Number of tokens (characters read/written by AI) successfully processed today
- `today_attempts`: Total number of calls today (including both successes and failures)
- `[+ Add]` button to register new keys, `✕` to delete keys.

> 💡 **What is a token?** It's a unit AI uses to process text. It roughly corresponds to one English word, or 1-2 Korean characters. API pricing is usually calculated based on token count.

### Agent Card

A card showing the status of bots (agents) connected to the wall-vault proxy.

**Connection status is displayed in 4 levels:**

| Indicator | Status | Meaning |
|-----------|--------|---------|
| 🟢 | Running | Proxy is operating normally |
| 🟡 | Delayed | Responding but slow |
| 🔴 | Offline | Proxy is not responding |
| ⚫ | Not connected / Disabled | Proxy has never connected to vault or is disabled |

**Agent card bottom button guide:**

When you register an agent and specify the **agent type**, convenience buttons matching that type appear automatically.

---

#### 🔘 Copy Config Button — automatically generates connection settings

Clicking the button copies a config snippet with the agent's token, proxy address, and model information already filled in to your clipboard. Just paste the copied content into the location shown in the table below to complete the connection setup.

| Button | Agent Type | Paste Location |
|--------|-----------|---------------|
| 🦞 Copy OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copy NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copy Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copy Cursor Config | `cursor` | Cursor → Settings → AI |
| 💻 Copy VSCode Config | `vscode` | `~/.continue/config.json` |

**Example — for Claude Code type, this content is copied:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "this-agent's-token"
}
```

**Example — for VSCode (Continue) type:**

```yaml
# ~/.continue/config.yaml  ← paste into config.yaml, not config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: this-agent's-token
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **The latest version of Continue uses `config.yaml`.** If `config.yaml` exists, `config.json` is completely ignored. Make sure to paste into `config.yaml`.

**Example — for Cursor type:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : this-agent's-token

// Or environment variables:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=this-agent's-token
```

> ⚠️ **If clipboard copy doesn't work**: Browser security policies may block copying. If a popup text box appears, use Ctrl+A to select all, then Ctrl+C to copy.

---

#### ⚡ Auto-Apply Button — one click to complete setup

For agent types `cline`, `claude-code`, `openclaw`, or `nanoclaw`, an **⚡ Apply Config** button appears on the agent card. Clicking this button automatically updates the agent's local config file.

| Button | Agent Type | Target File |
|--------|-----------|------------|
| ⚡ Apply Cline Config | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Apply Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| ⚡ Apply OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Apply NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ This button sends requests to **localhost:56244** (local proxy). The proxy must be running on that machine for it to work.

---

#### 🔀 Drag-and-Drop Card Sorting (v0.1.17, improved v0.1.25)

You can **drag** dashboard agent cards to rearrange them in your preferred order.

1. Grab the **traffic light (●)** area at the top-left of the card with your mouse and drag
2. Drop it onto another card to swap positions

> 💡 The card body (input fields, buttons, etc.) is not draggable. You can only grab from the traffic light area.

#### 🟠 Agent Process Detection (v0.1.25)

When the proxy is running normally but the local agent process (NanoClaw, OpenClaw) has died, the card traffic light turns **orange (blinking)** and displays an "Agent process stopped" message.

- 🟢 Green: Proxy + agent both normal
- 🟠 Orange (blinking): Proxy normal, agent dead
- 🔴 Red: Proxy offline
3. The changed order is **immediately saved to the server** and persists after refresh

> 💡 Touch devices (mobile/tablet) are not yet supported. Please use a desktop browser.

---

#### 🔄 Bidirectional Model Sync (v0.1.16)

When you change an agent's model in the vault dashboard, the agent's local config is automatically updated.

**For Cline:**
- Change model in vault → SSE event → proxy updates model fields in `globalState.json`
- Updated fields: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` and API key are not touched
- **VS Code reload (`Ctrl+Alt+R` or `Ctrl+Shift+P` → `Developer: Reload Window`) is required**
  - Because Cline doesn't re-read config files while running

**For Claude Code:**
- Change model in vault → SSE event → proxy updates `model` field in `settings.json`
- Automatically searches both WSL and Windows paths (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Reverse direction (agent → vault):**
- When an agent (Cline, Claude Code, etc.) sends a request to the proxy, the proxy includes that client's service/model info in the heartbeat
- The agent card on the vault dashboard shows the currently used service/model in real time

> 💡 **Key point**: The proxy identifies agents by the Authorization token in requests and automatically routes to the service/model configured in the vault. Even if Cline or Claude Code sends a different model name, the proxy overrides it with the vault's configuration.

---

### Using Cline in VS Code — Detailed Guide

#### Step 1: Install Cline

Install **Cline** (ID: `saoudrizwan.claude-dev`) from the VS Code extension marketplace.

#### Step 2: Register the Agent in the Vault

1. Open the vault dashboard (`http://VAULT_IP:56243`)
2. Click **+ Add** in the **Agents** section
3. Enter the following:

| Field | Value | Description |
|-------|-------|-------------|
| ID | `my_cline` | Unique identifier (alphanumeric, no spaces) |
| Name | `My Cline` | Name displayed on dashboard |
| Agent Type | `cline` | ← must select `cline` |
| Service | Select desired service (e.g., `google`) | |
| Model | Enter desired model (e.g., `gemini-2.5-flash`) | |

4. Click **Save** — a token is automatically generated

#### Step 3: Connect to Cline

**Method A — Auto-Apply (recommended)**

1. Verify that the wall-vault **proxy** is running on that machine (`localhost:56244`)
2. Click the **⚡ Apply Cline Config** button on the agent card in the dashboard
3. Success when you see the "Config applied!" notification
4. Reload VS Code (`Ctrl+Alt+R`)

**Method B — Manual Setup**

Open settings (⚙️) in the Cline sidebar:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://PROXY_ADDRESS:56244/v1`
  - Same machine: `http://localhost:56244/v1`
  - Different machine (e.g., Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Token issued from the vault (copy from agent card)
- **Model ID**: Model set in the vault (e.g., `gemini-2.5-flash`)

#### Step 4: Verify

Send any message in the Cline chat. If everything is working:
- The agent card on the vault dashboard shows **green dot (● Running)**
- The card displays the current service/model (e.g., `google / gemini-2.5-flash`)

#### Changing Models

When you want to change Cline's model, change it from the **vault dashboard**:

1. Change the service/model dropdown on the agent card
2. Click **Apply**
3. Reload VS Code (`Ctrl+Alt+R`) — the model name in the Cline footer updates
4. The new model is used from the next request

> 💡 In practice, the proxy identifies Cline's requests by token and routes to the model configured in the vault. Even without VS Code reload, **the actual model being used changes immediately** — the reload is just to update the model display in the Cline UI.

#### Disconnect Detection

When you close VS Code, the agent card on the vault dashboard turns yellow (delayed) after about **90 seconds**, then red (offline) after **3 minutes**. (From v0.1.18, 15-second interval status checks make offline detection faster.)

#### Troubleshooting

| Symptom | Cause | Solution |
|---------|-------|----------|
| "Connection failed" error in Cline | Proxy not running or wrong address | Check proxy with `curl http://localhost:56244/health` |
| Green dot not showing in vault | API key (token) not configured | Click **⚡ Apply Cline Config** button again |
| Cline footer model doesn't change | Cline caches config | Reload VS Code (`Ctrl+Alt+R`) |
| Wrong model name displayed | Old bug (fixed in v0.1.16) | Update proxy to v0.1.16+ |

---

#### 🟣 Copy Deploy Command Button — for installing on new machines

Used when first installing the wall-vault proxy on a new computer and connecting to the vault. Clicking the button copies the entire install script. Paste it into the terminal on the new computer and run it — the following is handled all at once:

1. Install the wall-vault binary (skipped if already installed)
2. Auto-register systemd user service
3. Start service and auto-connect to vault

> 💡 The script already contains this agent's token and vault server address, so you can run it immediately after pasting without any modifications.

---

### Service Card

A card for enabling/disabling and configuring AI services.

- Toggle switches to enable/disable each service
- Enter the address of a local AI server (Ollama, LM Studio, vLLM, etc. running on your computer) to automatically discover available models.
- **Local service connection status**: The ● dot next to the service name is **green** when connected, **gray** when not connected
- **Local service auto traffic light** (v0.1.23+): Local services (Ollama, LM Studio, vLLM) are automatically enabled/disabled based on connection availability. When a service connects, the ● dot turns green and the checkbox is checked within 15 seconds; when disconnected, it's automatically disabled. This works the same way as cloud services (Google, OpenRouter, etc.) auto-toggling based on API key availability.

> 💡 **If your local service is running on another computer**: Enter that computer's IP in the service URL field. Example: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). If the service is bound only to `127.0.0.1` instead of `0.0.0.0`, external IP access won't work — check the binding address in the service settings.

### Admin Token Entry

When you try to use important features like adding or deleting keys on the dashboard, an admin token input popup appears. Enter the token you set during the setup wizard. Once entered, it persists until you close the browser.

> ⚠️ **If authentication failures exceed 10 within 15 minutes, the IP is temporarily blocked.** If you've forgotten the token, check the `admin_token` field in the `wall-vault.yaml` file.

---

## Distributed Mode (Multi-Bot)

A configuration for **sharing a single key vault** when running OpenClaw on multiple computers simultaneously. It's convenient because you only need to manage keys in one place.

### Configuration Example

```
[Key Vault Server]
  wall-vault vault    (key vault :56243, dashboard)

[WSL Alpha]          [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE sync            ↕ SSE sync              ↕ SSE sync
```

All bots point to the central vault server, so when you change a model or add a key in the vault, it's instantly reflected across all bots.

### Step 1: Start the Key Vault Server

Run this on the computer that will serve as the vault server:

```bash
wall-vault vault
```

### Step 2: Register Each Bot (Client)

Pre-register information for each bot that will connect to the vault server:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Step 3: Start the Proxy on Each Bot Computer

On each computer with a bot, run the proxy with the vault server address and token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Replace **`192.168.x.x`** with the actual internal IP address of the vault server computer. You can check it via your router settings or the `ip addr` command.

---

## Auto-Start Configuration

If it's tedious to manually start wall-vault every time you restart your computer, register it as a system service. Once registered, it starts automatically on boot.

### Linux — systemd (most Linux distributions)

systemd is the system that automatically starts and manages programs on Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

View logs:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

The system responsible for auto-starting programs on macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Download NSSM from [nssm.cc](https://nssm.cc/download) and add it to PATH.
2. In an administrator PowerShell:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor

The `doctor` command is a tool that **self-diagnoses and repairs** wall-vault's configuration.

```bash
wall-vault doctor check   # Diagnose current state (read-only, changes nothing)
wall-vault doctor fix     # Automatically fix problems
wall-vault doctor all     # Diagnose + auto-fix in one step
```

> 💡 If something seems off, try running `wall-vault doctor all` first. It catches many problems automatically.

---

## RTK Token Saving

*(v0.1.24+)*

**RTK (Token Saving Tool)** automatically compresses the output of shell commands run by AI coding agents (such as Claude Code), reducing token usage. For example, 15 lines of `git status` output gets compressed to a 2-line summary.

### Basic Usage

```bash
# Wrap commands with wall-vault rtk to auto-filter output
wall-vault rtk git status          # shows only changed file list
wall-vault rtk git diff HEAD~1     # changed lines + minimal context only
wall-vault rtk git log -10         # hash + one-line message each
wall-vault rtk go test ./...       # shows only failed tests
wall-vault rtk ls -la              # unsupported commands are auto-truncated
```

### Supported Commands and Savings

| Command | Filter Method | Savings |
|---------|--------------|---------|
| `git status` | Changed file summary only | ~87% |
| `git diff` | Changed lines + 3-line context | ~60-94% |
| `git log` | Hash + first line message | ~90% |
| `git push/pull/fetch` | Remove progress, summary only | ~80% |
| `go test` | Show failures only, count passes | ~88-99% |
| `go build/vet` | Show errors only | ~90% |
| All other commands | First 50 + last 50 lines, max 32KB | Variable |

### 3-Stage Filter Pipeline

1. **Command-specific structural filter** — Understands output formats of git, go, etc. and extracts only meaningful parts
2. **Regex post-processing** — Removes ANSI color codes, collapses blank lines, aggregates duplicate lines
3. **Passthrough + truncation** — Unsupported commands keep only first/last 50 lines

### Claude Code Integration

You can set up all shell commands to automatically go through RTK using Claude Code's `PreToolUse` hook.

```bash
# Install hook (auto-added to Claude Code settings.json)
wall-vault rtk hook install
```

Or manually add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "command": "wall-vault rtk rewrite"
    }]
  }
}
```

> 💡 **Exit code preservation**: RTK returns the original command's exit code as-is. If a command fails (exit code ≠ 0), the AI accurately detects the failure.

> 💡 **Forced English output**: RTK runs commands with `LC_ALL=C` to always produce English output regardless of system language settings. This ensures filters work accurately.

---

## Environment Variables Reference

Environment variables are a way to pass configuration values to programs. Enter them in your terminal as `export VARIABLE=value`, or add them to your auto-start service file for permanent effect.

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `WV_LANG` | Dashboard language | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard theme | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API key (comma-separated for multiple) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API key | `sk-or-v1-...` |
| `WV_VAULT_URL` | Vault server address in distributed mode | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client (bot) authentication token | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin token | `admin-token-here` |
| `WV_MASTER_PASS` | API key encryption password | `my-password` |
| `WV_AVATAR` | Avatar image file path (relative to `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama local server address | `http://192.168.x.x:11434` |

---

## Troubleshooting

### When the Proxy Won't Start

The port is often already in use by another program.

```bash
ss -tlnp | grep 56244   # Check what's using port 56244
wall-vault proxy --port 8080   # Start with a different port number
```

### API Key Errors (429, 402, 401, 403, 582)

| Error Code | Meaning | Solution |
|-----------|---------|----------|
| **429** | Too many requests (usage exceeded) | Wait a moment or add more keys |
| **402** | Payment required or credits depleted | Top up credits on that service |
| **401 / 403** | Invalid key or no permission | Re-check key value and re-register |
| **582** | Gateway overload (5-minute cooldown) | Auto-resolves after 5 minutes |

```bash
# Check registered key list and status
curl -H "Authorization: Bearer YOUR_ADMIN_TOKEN" http://localhost:56243/admin/keys

# Reset key usage counters
curl -X POST -H "Authorization: Bearer YOUR_ADMIN_TOKEN" http://localhost:56243/admin/keys/reset
```

### When Agent Shows "Not Connected"

"Not connected" means the proxy process is not sending heartbeat signals to the vault. **It does not mean settings are not saved.** The proxy must be running with the vault server address and token to show as connected.

```bash
# Start proxy with vault server address, token, and client ID
WV_VAULT_URL=http://VAULT_SERVER:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Once connected successfully, the dashboard shows 🟢 Running within about 20 seconds.

### When Ollama Won't Connect

Ollama is a program that runs AI directly on your computer. First, check if Ollama is running.

```bash
curl http://localhost:11434/api/tags   # If model list appears, it's working
export OLLAMA_URL=http://192.168.x.x:11434   # If running on another computer
```

> ⚠️ If Ollama isn't responding, start it first with the `ollama serve` command.

> ⚠️ **Large models are slow**: Large models like `qwen3.5:35b` or `deepseek-r1` can take several minutes to generate a response. Even if it seems like there's no response, it may be processing normally — please wait.

---

## Recent Changes (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Agent process detection**: Proxy detects whether local agents (NanoClaw/OpenClaw) are alive and displays an orange traffic light on the dashboard.
- **Drag handle improvement**: Card sorting now only grabs from the traffic light (●) area. Prevents accidental dragging from input fields or buttons.

### v0.1.24 (2026-04-06)
- **RTK token saving subcommand**: `wall-vault rtk <command>` auto-filters shell command output, reducing AI agent token usage by 60-90%. Built-in filters for major commands like git and go, with auto-truncation for unsupported commands. Integrates transparently with Claude Code `PreToolUse` hooks.

### v0.1.23 (2026-04-06)
- **Ollama model change fix**: Fixed issue where changing Ollama models in the vault dashboard didn't actually reflect in the proxy. Previously only used environment variable (`OLLAMA_MODEL`), now vault settings take priority.
- **Local service auto traffic light**: Ollama, LM Studio, and vLLM auto-enable when connectable and auto-disable when disconnected. Same mechanism as cloud services' key-based auto-toggle.

### v0.1.22 (2026-04-05)
- **Empty content field fix**: Fixed issue where thinking models (gemini-3.1-pro, o1, claude thinking, etc.) that used all max_tokens on reasoning without producing actual responses caused the proxy to omit `content`/`text` fields via `omitempty`, crashing OpenAI/Anthropic SDK clients with `Cannot read properties of undefined (reading 'trim')` errors. Changed to always include fields per official API spec.

### v0.1.21 (2026-04-05)
- **Gemma 4 model support**: Gemma models like `gemma-4-31b-it` and `gemma-4-26b-a4b-it` can now be used via the Google Gemini API.
- **LM Studio / vLLM official support**: Previously these services were missing from proxy routing and always fell back to Ollama. Now properly routed via OpenAI-compatible API.
- **Dashboard service display fix**: Dashboard always shows the user's configured service even when fallback occurs.
- **Local service status display**: Shows local service (Ollama, LM Studio, vLLM, etc.) connection status via ● dot color on dashboard load.
- **Tool filter environment variable**: Tool passing mode can be set with `WV_TOOL_FILTER=passthrough` environment variable.

### v0.1.20 (2026-03-28)
- **Comprehensive security hardening**: XSS prevention (41 points), constant-time token comparison, CORS restriction, request size limits, path traversal prevention, SSE authentication, rate limiter hardening, and 12 total security improvements.

### v0.1.19 (2026-03-27)
- **Claude Code online detection**: Claude Code running without going through the proxy is now shown as online on the dashboard.

### v0.1.18 (2026-03-26)
- **Fallback service stuck fix**: After temporary error fallback to Ollama, automatically returns to original service when it recovers.
- **Offline detection improvement**: 15-second interval status checks make proxy shutdown detection faster.

### v0.1.17 (2026-03-25)
- **Drag-and-drop card sorting**: Agent cards can be dragged to reorder.
- **Inline config apply buttons**: [⚡ Apply Config] buttons appear on offline agent cards.
- **cokacdir agent type added**.

### v0.1.16 (2026-03-25)
- **Bidirectional model sync**: Changing Cline or Claude Code models from the vault dashboard is automatically reflected.

---

*For more detailed API information, see [API.md](API.md).*
