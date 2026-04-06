# wall-vault User Manual
*(Last updated: 2026-04-06 — v0.1.23)*

---

## Table of Contents

1. [What is wall-vault?](#what-is-wall-vault)
2. [Installation](#installation)
3. [Getting Started (setup wizard)](#getting-started)
4. [Registering API Keys](#registering-api-keys)
5. [Using the Proxy](#using-the-proxy)
6. [The Key Vault Dashboard](#the-key-vault-dashboard)
7. [Distributed Mode (multi-bot)](#distributed-mode-multi-bot)
8. [Auto-start Setup](#auto-start-setup)
9. [Doctor — Self-diagnosis Tool](#doctor--self-diagnosis-tool)
10. [Environment Variables Reference](#environment-variables-reference)
11. [Troubleshooting](#troubleshooting)

---

## What is wall-vault?

**wall-vault = an AI proxy + API key vault for OpenClaw**

To use AI services, you need **API keys** — think of them as a **digital pass** that proves you're authorized to use a particular service. These passes have daily usage limits, and if they're mishandled, they can be exposed or compromised.

wall-vault keeps your passes safe in an encrypted vault, and acts as a **proxy (middleman)** between OpenClaw and the AI services. In short, OpenClaw only needs to talk to wall-vault — wall-vault handles all the complicated stuff behind the scenes.

Here's what wall-vault takes care of for you:

- **Automatic key rotation**: When one key hits its limit or gets temporarily blocked (cooldown), wall-vault quietly switches to the next key. OpenClaw keeps working without interruption.
- **Automatic service fallback**: If Google doesn't respond, it falls back to OpenRouter. If that fails too, it automatically switches to Ollama, LM Studio, or vLLM (local AI) running on your machine. Your session never drops. When the original service recovers, it automatically switches back from the next request onward (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Real-time sync (SSE)**: Change the model in the vault dashboard, and OpenClaw reflects it within 1–3 seconds. SSE (Server-Sent Events) is a technology where the server pushes updates to clients in real time.
- **Real-time notifications**: Events like key exhaustion or service outages appear immediately in OpenClaw's TUI (terminal UI) at the bottom of the screen.

> 💡 **Claude Code, Cursor, and VS Code** can also be connected, but wall-vault's primary purpose is to work alongside OpenClaw.

```
OpenClaw (TUI terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← key management, routing, fallback, events
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ models)
        ├─ Ollama / LM Studio / vLLM (your local machine, last resort)
        └─ OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Open a terminal and paste the following commands:

```bash
# Linux (standard PC, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Downloads a file from the internet.
- `chmod +x` — Makes the downloaded file executable. If you skip this step, you'll get a "permission denied" error.

### Windows

Open PowerShell (as administrator) and run the following commands:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Add to PATH (takes effect after restarting PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **What is PATH?** It's a list of folders where your computer looks for commands. Adding wall-vault to PATH lets you run `wall-vault` from any directory.

### Building from Source (for developers)

This only applies if you have a Go development environment installed.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build timestamp version**: When you build with `make build`, the version is automatically generated in a format like `v0.1.23.20260406.211004` that includes the date and time. If you build directly with `go build ./...`, the version will simply show `"dev"`.

---

## Getting Started

### Running the setup wizard

After installation, make sure to run the **setup wizard** first. The wizard will guide you through each step, asking for the necessary information.

```bash
wall-vault setup
```

Here are the steps the wizard goes through:

```
1. Language selection (10 languages including English)
2. Theme selection (light / dark / gold / cherry / ocean)
3. Operation mode — standalone (single user) or distributed (multiple machines)
4. Bot name — the name displayed on the dashboard
5. Port settings — defaults: proxy 56244, vault 56243 (press Enter to keep defaults)
6. AI service selection — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool security filter settings
8. Admin token — a password to lock dashboard admin features; can be auto-generated
9. API key encryption password — for extra-secure key storage (optional)
10. Config file save location
```

> ⚠️ **Make sure to remember your admin token.** You'll need it later to add keys or change settings in the dashboard. If you forget it, you'll have to manually edit the config file.

Once the wizard is complete, a `wall-vault.yaml` configuration file is automatically created.

### Starting up

```bash
wall-vault start
```

Two servers start simultaneously:

- **Proxy** (`http://localhost:56244`) — the middleman connecting OpenClaw to AI services
- **Key Vault** (`http://localhost:56243`) — API key management and web dashboard

Open `http://localhost:56243` in your browser to see the dashboard right away.

---

## Registering API Keys

There are four ways to register API keys. **For beginners, Method 1 (environment variables) is recommended.**

### Method 1: Environment Variables (recommended — simplest)

Environment variables are **pre-set values** that a program reads when it starts. Just type the following in your terminal:

```bash
# Register a Google Gemini key
export WV_KEY_GOOGLE=AIzaSy...

# Register an OpenRouter key
export WV_KEY_OPENROUTER=sk-or-v1-...

# Start after registration
wall-vault start
```

If you have multiple keys, separate them with commas. wall-vault will rotate through them automatically (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tip**: The `export` command only applies to the current terminal session. To persist it across reboots, add the line to your `~/.bashrc` or `~/.zshrc` file.

### Method 2: Dashboard UI (point and click)

1. Open `http://localhost:56243` in your browser
2. Click the `[+ Add]` button in the top **🔑 API Keys** card
3. Enter the service type, key value, label (a memo name), and daily limit, then save

### Method 3: REST API (for automation/scripts)

REST API is a way for programs to exchange data over HTTP. It's useful for automated registration via scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Main key",
    "daily_limit": 1000
  }'
```

### Method 4: Proxy flags (for quick testing)

Use this to temporarily inject a key for testing without formal registration. The key disappears when the program is stopped.

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

> 💡 **Easier method**: Click the **🦞 Copy OpenClaw Config** button on the agent card in the dashboard — it copies a snippet with the token and address already filled in. Just paste it.

**Where does the `wall-vault/` prefix in model names route to?**

wall-vault automatically determines which AI service to send the request to based on the model name:

| Model format | Routed to |
|-------------|-----------|
| `wall-vault/gemini-*` | Google Gemini direct |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI direct |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free 1M token context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/model-name`, `openai/model-name`, `anthropic/model-name`, etc. | Direct to that service |
| `custom/google/model-name`, `custom/openai/model-name`, etc. | Strips `custom/` prefix and re-routes |
| `model-name:cloud` | Strips `:cloud` suffix and routes to OpenRouter |

> 💡 **What is context?** It's the amount of conversation an AI can remember at once. 1M (one million tokens) means it can process very long conversations or documents in a single session.

### Direct Gemini API format (for existing tool compatibility)

If you have tools that already use Google's Gemini API directly, just change the address to wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Or if the tool takes a direct URL:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Using with the OpenAI SDK (Python)

You can also connect wall-vault to Python code that uses AI. Just change the `base_url`:

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

### Changing models at runtime

To change the AI model while wall-vault is already running:

```bash
# Change model by sending a request to the proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# In distributed mode (multi-bot), change on the vault server → instantly synced via SSE
curl -X PUT http://localhost:56243/admin/clients/my-bot-id \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Checking available models

```bash
# View full list
curl http://localhost:56244/api/models | python3 -m json.tool

# View Google models only
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
| Ollama | Auto-detects locally installed models |
| LM Studio | Local server (port 1234) |
| vLLM | Local server (port 8000) |

---

## The Key Vault Dashboard

Open `http://localhost:56243` in your browser to see the dashboard.

**Layout:**
- **Top bar (fixed)**: Logo, language/theme selectors, SSE connection status indicator
- **Card grid**: Agent, service, and API key cards arranged in tiles

### API Key Cards

These cards give you an at-a-glance view of your registered API keys.

- Keys are organized by service.
- `today_usage`: Number of tokens (units of text the AI reads/writes) successfully processed today
- `today_attempts`: Total number of calls today (successful + failed)
- Use the `[+ Add]` button to register new keys, and `✕` to delete them.

> 💡 **What is a token?** It's the unit AI uses to process text. Roughly one English word, or 1–2 Korean characters. API pricing is typically based on token count.

### Agent Cards

These cards show the status of bots (agents) connected to the wall-vault proxy.

**Connection status has 4 levels:**

| Indicator | Status | Meaning |
|-----------|--------|---------|
| 🟢 | Running | Proxy is operating normally |
| 🟡 | Delayed | Responding but slow |
| 🔴 | Offline | Proxy is not responding |
| ⚫ | Not connected / Disabled | Proxy has never connected to the vault, or is disabled |

**Buttons at the bottom of agent cards:**

When you register an agent with a specific **agent type**, convenience buttons matching that type automatically appear.

---

#### 🔘 Copy Config Button — automatically generates connection settings

Clicking this button copies a configuration snippet to the clipboard with the agent's token, proxy address, and model info already filled in. Just paste it at the location shown in the table below to complete the connection setup.

| Button | Agent Type | Where to Paste |
|--------|-----------|----------------|
| 🦞 Copy OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copy NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copy Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copy Cursor Config | `cursor` | Cursor → Settings → AI |
| 💻 Copy VSCode Config | `vscode` | `~/.continue/config.json` |

**Example — For Claude Code type, this is what gets copied:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "this-agent's-token"
}
```

**Example — For VSCode (Continue) type:**

```yaml
# ~/.continue/config.yaml  ← paste in config.yaml, NOT config.json
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

**Example — For Cursor type:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : this-agent's-token

// Or as environment variables:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=this-agent's-token
```

> ⚠️ **If clipboard copying doesn't work**: Browser security policies may block copying. If a popup with a textbox appears, use Ctrl+A to select all, then Ctrl+C to copy.

---

#### ⚡ Auto-Apply Button — one click and you're done

For agents of type `cline`, `claude-code`, `openclaw`, or `nanoclaw`, the agent card displays an **⚡ Apply Config** button. Clicking this button automatically updates the agent's local config file.

| Button | Agent Type | Target File |
|--------|-----------|-------------|
| ⚡ Apply Cline Config | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Apply Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| ⚡ Apply OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Apply NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ This button sends a request to **localhost:56244** (local proxy). The proxy must be running on that machine for it to work.

---

#### 🔀 Drag & Drop Card Sorting (v0.1.17)

You can **drag** agent cards on the dashboard to rearrange them in any order.

1. Grab an agent card with your mouse and drag it
2. Drop it on top of another card to swap positions
3. The new order is **saved to the server immediately** and persists after refresh

> 💡 Touch devices (mobile/tablet) are not yet supported. Use a desktop browser.

---

#### 🔄 Bidirectional Model Sync (v0.1.16)

When you change an agent's model in the vault dashboard, the agent's local config is automatically updated.

**For Cline:**
- Model change in vault → SSE event → proxy updates the model field in `globalState.json`
- Updated fields: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` and API key are not touched
- **VS Code reload required (`Ctrl+Alt+R` or `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Because Cline doesn't re-read the config file while running

**For Claude Code:**
- Model change in vault → SSE event → proxy updates the `model` field in `settings.json`
- Automatically searches both WSL and Windows paths (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Reverse direction (agent → vault):**
- When an agent (Cline, Claude Code, etc.) sends a request to the proxy, the proxy includes that client's service/model info in the heartbeat
- The agent card in the vault dashboard shows the currently used service/model in real time

> 💡 **Key point**: The proxy identifies agents by their Authorization token in requests, and auto-routes to the service/model configured in the vault. Even if Cline or Claude Code sends a different model name, the proxy overrides it with the vault's configuration.

---

### Using Cline with VS Code — Detailed Guide

#### Step 1: Install Cline

Install **Cline** (ID: `saoudrizwan.claude-dev`) from the VS Code Extensions Marketplace.

#### Step 2: Register the agent in the vault

1. Open the vault dashboard (`http://vault-IP:56243`)
2. Click **+ Add** in the **Agents** section
3. Fill in the following:

| Field | Value | Description |
|-------|-------|-------------|
| ID | `my_cline` | Unique identifier (alphanumeric, no spaces) |
| Name | `My Cline` | Name displayed on the dashboard |
| Agent Type | `cline` | ← Must select `cline` |
| Service | Select the service to use (e.g., `google`) | |
| Model | Enter the model to use (e.g., `gemini-2.5-flash`) | |

4. Click **Save** — a token is automatically generated

#### Step 3: Connect to Cline

**Method A — Auto-apply (recommended)**

1. Make sure the wall-vault **proxy** is running on that machine (`localhost:56244`)
2. Click the **⚡ Apply Cline Config** button on the agent card in the dashboard
3. If you see "Config applied successfully!" notification, it worked
4. Reload VS Code (`Ctrl+Alt+R`)

**Method B — Manual setup**

Open Settings (⚙️) in the Cline sidebar:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://proxy-address:56244/v1`
  - Same machine: `http://localhost:56244/v1`
  - Different machine (e.g., Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: The token issued by the vault (copy from agent card)
- **Model ID**: The model configured in the vault (e.g., `gemini-2.5-flash`)

#### Step 4: Verify

Send any message in the Cline chat window. If it works:
- The agent card in the vault dashboard shows a **green dot (● Running)**
- The card shows the current service/model (e.g., `google / gemini-2.5-flash`)

#### Changing the model

When you want to change Cline's model, do it from the **vault dashboard**:

1. Change the service/model dropdown on the agent card
2. Click **Apply**
3. Reload VS Code (`Ctrl+Alt+R`) — the model name in Cline's footer will update
4. The new model is used from the next request onward

> 💡 In practice, the proxy identifies Cline's requests by token and routes them to the vault-configured model. Even without a VS Code reload, **the actual model used changes immediately** — the reload is just to update the model display in Cline's UI.

#### Detecting disconnection

When VS Code is closed, the agent card on the vault dashboard turns yellow (delayed) after about **90 seconds**, and red (offline) after **3 minutes**. (From v0.1.18, offline detection is faster thanks to 15-second interval status checks.)

#### Troubleshooting

| Symptom | Cause | Solution |
|---------|-------|----------|
| "Connection failed" error in Cline | Proxy not running or wrong address | Check proxy with `curl http://localhost:56244/health` |
| Green dot doesn't appear in vault | API key (token) not configured | Click the **⚡ Apply Cline Config** button again |
| Cline footer model doesn't change | Cline caches settings | Reload VS Code (`Ctrl+Alt+R`) |
| Wrong model name shown | Old bug (fixed in v0.1.16) | Update proxy to v0.1.16 or later |

---

#### 🟣 Copy Deploy Command Button — for installing on a new machine

Use this when first installing the wall-vault proxy on a new computer and connecting it to the vault. Clicking the button copies the entire installation script. Paste it into the terminal on the new computer and run it to:

1. Install the wall-vault binary (skipped if already installed)
2. Automatically register a systemd user service
3. Start the service and auto-connect to the vault

> 💡 The script already contains this agent's token and vault server address, so you can run it immediately after pasting without any modifications.

---

### Service Cards

These cards let you enable/disable and configure AI services.

- Toggle switch to enable/disable each service
- Enter the address of a local AI server (Ollama, LM Studio, vLLM, etc. running on your computer) to auto-discover available models
- **Local service connection status**: A ● dot next to the service name is **green** if connected, **gray** if not
- **Local service auto-signaling** (v0.1.23+): Local services (Ollama, LM Studio, vLLM) are automatically enabled/disabled based on connection availability. When a service becomes reachable, it turns ● green and the checkbox activates within 15 seconds; when the service goes down, it automatically disables. This works the same way cloud services (Google, OpenRouter, etc.) auto-toggle based on API key availability.

> 💡 **If the local service is running on a different computer**: Enter that computer's IP in the service URL field. Example: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). If the service is only bound to `127.0.0.1` instead of `0.0.0.0`, external IP access won't work — check the service's binding address setting.

### Admin Token Input

When you try to use important features like adding or deleting keys in the dashboard, an admin token input popup appears. Enter the token you set up during the setup wizard. Once entered, it remains valid until you close the browser.

> ⚠️ **If authentication fails more than 10 times within 15 minutes, that IP will be temporarily blocked.** If you've forgotten your token, check the `admin_token` field in the `wall-vault.yaml` file.

---

## Distributed Mode (multi-bot)

When running OpenClaw on multiple computers simultaneously, you can **share a single key vault**. This is convenient because you only need to manage keys in one place.

### Example setup

```
[Key Vault Server]
  wall-vault vault    (key vault :56243, dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]   [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE sync            ↕ SSE sync              ↕ SSE sync
```

All bots point to the central vault server, so when you change a model or add a key in the vault, it's immediately reflected on all bots.

### Step 1: Start the key vault server

Run this on the computer that will serve as the vault server:

```bash
wall-vault vault
```

### Step 2: Register each bot (client)

Register the information for each bot that will connect to the vault server:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Step 3: Start the proxy on each bot's computer

On each computer where a bot is installed, run the proxy with the vault server address and token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Replace **`192.168.x.x`** with the actual internal IP address of the vault server computer. You can find it through your router settings or the `ip addr` command.

---

## Auto-start Setup

If it's tedious to manually start wall-vault every time you reboot your computer, register it as a system service. Once registered, it starts automatically on boot.

### Linux — systemd (most Linux distributions)

systemd is the system that automatically starts and manages programs on Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Check logs:

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

1. Download NSSM from [nssm.cc](https://nssm.cc/download) and add it to your PATH.
2. In an administrator PowerShell:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Self-diagnosis Tool

The `doctor` command is a tool that **diagnoses and repairs** wall-vault's configuration automatically.

```bash
wall-vault doctor check   # Diagnose current state (read-only, changes nothing)
wall-vault doctor fix     # Automatically repair issues
wall-vault doctor all     # Diagnose + auto-repair in one step
```

> 💡 If something seems off, try running `wall-vault doctor all` first. It catches and fixes many issues automatically.

---

## Environment Variables Reference

Environment variables are a way to pass configuration values to a program. Enter them in the terminal using `export VARIABLE=value`, or add them to your auto-start service file for persistent application.

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `WV_LANG` | Dashboard language | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard theme | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API key (comma-separated for multiple) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API key | `sk-or-v1-...` |
| `WV_VAULT_URL` | Vault server address in distributed mode | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client (bot) auth token | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin token | `admin-token-here` |
| `WV_MASTER_PASS` | API key encryption password | `my-password` |
| `WV_AVATAR` | Avatar image file path (relative to `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama local server address | `http://192.168.x.x:11434` |

---

## Troubleshooting

### Proxy won't start

The port is often already in use by another program.

```bash
ss -tlnp | grep 56244   # Check what's using port 56244
wall-vault proxy --port 8080   # Start on a different port
```

### API key errors (429, 402, 401, 403, 582)

| Error Code | Meaning | What to Do |
|-----------|---------|------------|
| **429** | Too many requests (quota exceeded) | Wait a while or add more keys |
| **402** | Payment required or credits exhausted | Top up credits on that service |
| **401 / 403** | Invalid key or no permission | Re-check the key value and re-register |
| **582** | Gateway overload (5-minute cooldown) | Automatically resolves after 5 minutes |

```bash
# Check registered key list and status
curl -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys

# Reset key usage counters
curl -X POST -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys/reset
```

### Agent showing as "Not connected"

"Not connected" means the proxy process is not sending heartbeats to the vault. **It doesn't mean the configuration hasn't been saved.** The proxy needs to be running with the vault server address and token to establish a connection.

```bash
# Start proxy with vault server address, token, and client ID
WV_VAULT_URL=http://vault-server:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Once connected, the dashboard will show 🟢 Running within about 20 seconds.

### Ollama connection issues

Ollama is a program that runs AI directly on your computer. First, make sure Ollama is running.

```bash
curl http://localhost:11434/api/tags   # If a model list appears, it's working
export OLLAMA_URL=http://192.168.x.x:11434   # If running on another computer
```

> ⚠️ If Ollama isn't responding, start it first with `ollama serve`.

> ⚠️ **Large models are slow to respond**: Large models like `qwen3.5:35b` or `deepseek-r1` can take several minutes to generate a response. Even if it seems like nothing is happening, it may still be processing — please be patient.

---

## Recent Changes (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Ollama model change fix**: Fixed an issue where changing the Ollama model in the vault dashboard wasn't reflected in the actual proxy. Previously only the environment variable (`OLLAMA_MODEL`) was used, but now vault settings take priority.
- **Local service auto-signaling**: Ollama, LM Studio, and vLLM are automatically enabled when reachable and disabled when unreachable. This works the same way as key-based auto-toggling for cloud services.

### v0.1.22 (2026-04-05)
- **Empty content field fix**: When thinking models (gemini-3.1-pro, o1, claude thinking, etc.) use up all max_tokens on reasoning and can't produce an actual response, the proxy was omitting the `content`/`text` fields from the response JSON via `omitempty`, causing OpenAI/Anthropic SDK clients to crash with `Cannot read properties of undefined (reading 'trim')`. Fixed to always include the fields per the official API spec.

### v0.1.21 (2026-04-05)
- **Gemma 4 model support**: Gemma family models like `gemma-4-31b-it` and `gemma-4-26b-a4b-it` can now be used via the Google Gemini API.
- **LM Studio / vLLM service support**: Previously these services were missing from proxy routing and always fell back to Ollama. Now properly routed via OpenAI-compatible API.
- **Dashboard service display fix**: Even when fallback occurs, the dashboard always displays the user-configured service.
- **Local service status display**: Shows connection status of local services (Ollama, LM Studio, vLLM, etc.) with ● dot colors when the dashboard loads.
- **Tool filter environment variable**: Use `WV_TOOL_FILTER=passthrough` env var to set the tool pass-through mode.

### v0.1.20 (2026-03-28)
- **Comprehensive security hardening**: XSS prevention (41 locations), constant-time token comparison, CORS restrictions, request size limits, path traversal prevention, SSE authentication, rate limiter hardening, and 12 other security improvements.

### v0.1.19 (2026-03-27)
- **Claude Code online detection**: Claude Code instances not going through the proxy are now shown as online in the dashboard.

### v0.1.18 (2026-03-26)
- **Fallback service sticking fix**: After a temporary error causes an Ollama fallback, it automatically returns to the original service when it recovers.
- **Improved offline detection**: 15-second interval status checks make proxy outage detection faster.

### v0.1.17 (2026-03-25)
- **Drag & drop card sorting**: Agent cards can be dragged and dropped to change their order.
- **Inline apply config button**: The [⚡ Apply Config] button is shown on offline agent cards.
- **cokacdir agent type added**.

### v0.1.16 (2026-03-25)
- **Bidirectional model sync**: Changing a Cline or Claude Code model in the vault dashboard is automatically reflected.

---

*For more detailed API information, see [API.md](API.md).*
