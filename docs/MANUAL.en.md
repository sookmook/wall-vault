# wall-vault User Manual
*(Last updated: 2026-03-20 — v0.1.15)*

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
- **Automatic service fallback**: If Google doesn't respond, it falls back to OpenRouter. If that fails too, it automatically switches to Ollama running locally on your machine. Your session never drops.
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
        └─ Ollama (your local machine, last resort)
```

---

## Installation

### Linux / macOS

Open a terminal and paste the command for your platform.

```bash
# Linux (standard PC or server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Macs)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — downloads the file from the internet.
- `chmod +x` — marks the downloaded file as executable. Skip this step and you'll get a "permission denied" error.

### Windows

Open PowerShell (as Administrator) and run the following:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Add to PATH (takes effect after restarting PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **What is PATH?** It's the list of folders your computer searches when you type a command. Adding wall-vault to PATH means you can type `wall-vault` from any folder and it will work.

### Build from Source (for developers)

Only applicable if you have the Go development environment installed.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build timestamp versioning**: When you build with `make build`, the version is automatically generated with a date and time, like `v0.1.6.20260314.231308`. If you build directly with `go build ./...`, the version will just show as `"dev"`.

---

## Getting Started

### Running the setup wizard

After installation, always run the **setup wizard** first. It walks you through everything you need to configure, one step at a time.

```bash
wall-vault setup
```

The wizard takes you through these steps:

```
1. Language selection (10 languages including English)
2. Theme selection (light / dark / gold / cherry / ocean)
3. Operating mode — standalone (just you) or distributed (multiple machines)
4. Bot name — the name shown in the dashboard
5. Port settings — defaults: proxy 56244, vault 56243 (just press Enter to keep the defaults)
6. AI service selection — choose from Google / OpenRouter / Ollama
7. Tool security filter settings
8. Admin token setup — a password that locks dashboard admin features. Can be auto-generated.
9. API key encryption password — for extra security when storing keys (optional)
10. Config file save location
```

> ⚠️ **Make sure you remember your admin token.** You'll need it later to add keys or change settings in the dashboard. If you lose it, you'll have to edit the config file manually.

Once the wizard finishes, a `wall-vault.yaml` config file is automatically created.

### Starting wall-vault

```bash
wall-vault start
```

This starts two servers simultaneously:

- **Proxy** (`http://localhost:56244`) — the middleman between OpenClaw and AI services
- **Key vault** (`http://localhost:56243`) — API key management and the web dashboard

Open `http://localhost:56243` in your browser to access the dashboard right away.

---

## Registering API Keys

There are four ways to register API keys. **If you're just getting started, method 1 (environment variables) is recommended.**

### Method 1: Environment Variables (recommended — simplest)

Environment variables are **pre-set values** that a program reads when it starts. Just type these in your terminal:

```bash
# Register a Google Gemini key
export WV_KEY_GOOGLE=AIzaSy...

# Register an OpenRouter key
export WV_KEY_OPENROUTER=sk-or-v1-...

# Start wall-vault after registering
wall-vault start
```

If you have multiple keys, separate them with commas (,). wall-vault will rotate through them automatically (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tip**: The `export` command only applies to the current terminal session. To make it persist across restarts, add those lines to your `~/.bashrc` or `~/.zshrc` file.

### Method 2: Dashboard UI (click-based)

1. Open `http://localhost:56243` in your browser
2. In the **🔑 API Keys** card at the top, click the `[+ Add]` button
3. Fill in the service type, key value, label (a name for your own reference), and daily limit, then save

### Method 3: REST API (for automation and scripts)

REST API is a way for programs to exchange data over HTTP. This is handy for automating key registration with scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "main key",
    "daily_limit": 1000
  }'
```

### Method 4: proxy flag (for quick tests)

Use this when you want to temporarily plug in a key without formally registering it. The key disappears when you stop the program.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Using the Proxy

### Using with OpenClaw (primary use case)

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

> 💡 **Easier way**: Click the **🦞 Copy OpenClaw Config** button on the agent card in the dashboard. It copies a snippet with your token and address already filled in — just paste it.

**Where does `wall-vault/` in the model name route to?**

wall-vault looks at the model name to automatically decide which AI service to send the request to:

| Model format | Routes to |
|---|---|
| `wall-vault/gemini-*` | Google Gemini directly |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI directly |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free 1M token context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/model-name`, `openai/model-name`, `anthropic/model-name`, etc. | That service directly |
| `custom/google/model-name`, `custom/openai/model-name`, etc. | Strips `custom/` prefix and re-routes |
| `model-name:cloud` | Strips `:cloud` suffix and routes to OpenRouter |

> 💡 **What is context?** It's how much conversation an AI can hold in memory at once. 1M (one million tokens) means it can handle very long conversations or large documents in a single session.

### Connecting as a Gemini API endpoint (for existing tools)

If you have a tool that already uses the Google Gemini API directly, just point it at wall-vault instead:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Or if your tool lets you specify a URL directly:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Using with the OpenAI SDK (Python)

You can hook wall-vault into any Python code that uses AI. Just change `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault manages API keys for you
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # use provider/model format
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### Switching models while running

To change the AI model while wall-vault is already running:

```bash
# Change model by calling the proxy directly
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# In distributed mode, change via the vault server → applied instantly via SSE
curl -X PUT http://localhost:56243/admin/clients/your-bot-id \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Listing available models

```bash
# Show all models
curl http://localhost:56244/api/models | python3 -m json.tool

# Show only Google models
curl "http://localhost:56244/api/models?service=google"

# Search by name (e.g., models containing "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Key models by service:**

| Service | Notable models |
|---|---|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ models (Hunter Alpha 1M context free, DeepSeek R1/V3, Qwen 2.5, and more) |
| Ollama | Auto-detects local models installed on your machine |

---

## The Key Vault Dashboard

Open `http://localhost:56243` in your browser to access the dashboard.

**Layout:**
- **Top bar**: Logo, language and theme selectors, SSE connection status indicator
- **Card grid**: Agent, service, and API key cards arranged in a tile layout

### API Keys Card

A card for managing all your registered API keys at a glance.

- Keys are listed grouped by service.
- `today_usage`: tokens successfully processed today (tokens = the units AI uses to read and write text)
- `today_attempts`: total number of calls today (successful + failed)
- Use the `[+ Add]` button to register a new key, and `✕` to delete one.

> 💡 **What is a token?** It's the unit AI uses when processing text. Roughly one English word, or about 1–2 Korean characters. API pricing is usually calculated based on token count.

### Agent Card

A card showing the status of bots (agents) connected to the wall-vault proxy.

**Connection status has four levels:**

| Indicator | Status | Meaning |
|---|---|---|
| 🟢 | Running | Proxy is working normally |
| 🟡 | Delayed | Responding but slow |
| 🔴 | Offline | Proxy is not responding |
| ⚫ | Not connected / Inactive | Proxy has never connected to the vault, or is disabled |

**Buttons at the bottom of the agent card:**

When you register an agent, you specify an **agent type**. The appropriate convenience buttons appear automatically based on that type.

---

#### 🔘 Copy Config Button — auto-generates your connection settings

Click the button and a config snippet with this agent's token, proxy address, and model info already filled in is copied to your clipboard. Just paste it in the location shown in the table below and you're connected.

| Button | Agent type | Where to paste |
|---|---|---|
| 🦞 Copy OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copy NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copy Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copy Cursor Config | `cursor` | Cursor → Settings → AI |
| 💻 Copy VSCode Config | `vscode` | `~/.continue/config.json` |

**Example — Claude Code type copies something like this:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "this-agents-token"
}
```

**Example — VSCode (Continue) type:**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "this-agents-token"
  }]
}
```

**Example — Cursor type:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : this-agents-token

// Or as environment variables:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=this-agents-token
```

> ⚠️ **If clipboard copy doesn't work**: Browser security policies can block it in some cases. If a text box pops up instead, press Ctrl+A to select all, then Ctrl+C to copy.

---

#### ⚡ Auto-Apply Button — one click and you're configured

When the agent type is `cline`, `claude-code`, `openclaw`, or `nanoclaw`, an **⚡ Apply Settings** button appears on the agent card. Clicking it automatically updates that agent's local configuration file.

| Button | Agent type | Target file |
|--------|------------|-------------|
| ⚡ Apply Cline Settings | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Apply Claude Code Settings | `claude-code` | `~/.claude/settings.json` |
| ⚡ Apply OpenClaw Settings | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Apply NanoClaw Settings | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ This button sends a request to **localhost:56244** (the local proxy). The proxy must be running on that machine for it to work.

---

#### 🔀 Drag & Drop Card Sorting (v0.1.17)

You can **drag** agent cards on the dashboard to rearrange them in any order you like.

1. Grab an agent card with your mouse and drag it
2. Drop it on another card to swap their positions
3. The new order is **saved to the server immediately** and persists after a page refresh

> 💡 Touch devices (mobile/tablet) are not yet supported. Please use a desktop browser.

---

#### 🔄 Bidirectional Model Sync (v0.1.16)

When you change an agent's model in the vault dashboard, the agent's local settings are automatically updated.

**For Cline:**
- Change a model in the vault → SSE event → proxy updates the model fields in `globalState.json`
- Updated fields: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` and API key are left untouched
- **A VS Code reload is required (`Ctrl+Alt+R` or `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Cline does not re-read its config files while running

**For Claude Code:**
- Change a model in the vault → SSE event → proxy updates the `model` field in `settings.json`
- Both WSL and Windows paths are automatically searched (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Reverse direction (agent → vault):**
- When an agent (Cline, Claude Code, etc.) sends a request through the proxy, the proxy includes that client's service and model information in its heartbeat
- The agent card on the vault dashboard shows the currently active service/model in real time

> 💡 **Key point**: The proxy identifies agents by their Authorization token, and automatically routes requests to the service/model configured in the vault. Even if Cline or Claude Code sends a different model name, the proxy overrides it with the vault's settings.

---

### Using Cline in VS Code — Detailed Guide

#### Step 1: Install Cline

Install **Cline** (ID: `saoudrizwan.claude-dev`) from the VS Code extension marketplace.

#### Step 2: Register the agent in the vault

1. Open the vault dashboard (`http://vault-IP:56243`)
2. In the **Agents** section, click **+ Add**
3. Fill in the following:

| Field | Value | Description |
|-------|-------|-------------|
| ID | `my_cline` | Unique identifier (alphanumeric, no spaces) |
| Name | `My Cline` | Display name shown on the dashboard |
| Agent type | `cline` | ← must select `cline` |
| Service | Select the service to use (e.g. `google`) | |
| Model | Enter the model to use (e.g. `gemini-2.5-flash`) | |

4. Click **Save** — a token is automatically generated

#### Step 3: Connect Cline

**Method A — Auto-apply (recommended)**

1. Make sure the wall-vault **proxy** is running on that machine (`localhost:56244`)
2. Click the **⚡ Apply Cline Settings** button on the agent card in the dashboard
3. If you see a "Settings applied!" notification, it worked
4. Reload VS Code (`Ctrl+Alt+R`)

**Method B — Manual setup**

Open the settings (⚙️) in the Cline sidebar and enter:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://proxy-address:56244/v1`
  - Same machine: `http://localhost:56244/v1`
  - Different machine (e.g. Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: The token issued by the vault (copy from the agent card)
- **Model ID**: The model configured in the vault (e.g. `gemini-2.5-flash`)

#### Step 4: Verify

Send any message in the Cline chat. If everything is working:
- The agent card on the vault dashboard shows a **green dot (● Running)**
- The card displays the current service/model (e.g. `google / gemini-2.5-flash`)

#### Changing the model

When you want to switch Cline's model, change it in the **vault dashboard**:

1. Change the service/model dropdown on the agent card
2. Click **Apply**
3. Reload VS Code (`Ctrl+Alt+R`) — the model name in Cline's footer will update
4. The new model takes effect from the next request

> 💡 The proxy actually identifies Cline's requests by token and routes them to the model configured in the vault. Even without reloading VS Code, **the model actually in use changes immediately** — the reload is just to update Cline's UI display.

#### Detecting disconnection

When you close VS Code, the agent card on the vault dashboard will turn yellow (delayed) after about **2–3 minutes**, and red (offline) after **5 minutes**.

#### Troubleshooting

| Symptom | Cause | Solution |
|---------|-------|----------|
| "Connection failed" error in Cline | Proxy not running or wrong address | Check proxy with `curl http://localhost:56244/health` |
| Green dot doesn't appear in vault | API key (token) not configured | Click the **⚡ Apply Cline Settings** button again |
| Cline footer model doesn't change | Cline caches settings | Reload VS Code (`Ctrl+Alt+R`) |
| Wrong model name displayed | Old bug (fixed in v0.1.16) | Update proxy to v0.1.16 or later |

---

#### 🟣 Copy Deploy Command Button — for installing on a new machine

Use this when you're setting up the wall-vault proxy on a new computer and connecting it to the vault for the first time. Click the button and the full installation script is copied. Paste it into a terminal on the new machine and run it — it handles everything at once:

1. Install the wall-vault binary (skips if already installed)
2. Automatically register a systemd user service
3. Start the service and auto-connect to the vault

> 💡 The script already has this agent's token and vault server address filled in, so you can run it immediately after pasting — no edits needed.

---

### Services Card

A card for enabling/disabling and configuring the AI services you want to use.

- Toggle switches to enable or disable each service
- Enter the address of a local AI server (Ollama, LM Studio, vLLM, etc. running on your machine) and wall-vault will automatically discover its available models.
- **Local service connection indicator**: A ● dot next to the service name turns **green** when connected, **gray** when not connected.
- **Checkbox auto-sync**: When you open the page, if a local service (like Ollama) is already running, its checkbox will automatically be checked.

> 💡 **If your local service is running on a different machine**: Enter that machine's IP in the service URL field. For example: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### Admin Token Prompt

When you try to use sensitive features like adding or deleting keys in the dashboard, a popup will ask for your admin token. Enter the token you set up in the setup wizard. Once entered, it stays active until you close the browser.

> ⚠️ **If authentication fails 10 times within 15 minutes, that IP will be temporarily blocked.** If you've forgotten your token, check the `admin_token` field in your `wall-vault.yaml` file.

---

## Distributed Mode (multi-bot)

When running OpenClaw on multiple machines simultaneously, you can have them all **share a single key vault**. Key management stays in one place, which keeps things simple.

### Example setup

```
[Key Vault Server]
  wall-vault vault    (key vault :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE sync            ↕ SSE sync              ↕ SSE sync
```

All bots point to the same central vault server. Change a model or add a key in the vault, and every bot reflects it instantly.

### Step 1: Start the key vault server

Run this on the machine that will act as the vault server:

```bash
wall-vault vault
```

### Step 2: Register each bot (client)

Register the details for each bot that will connect to the vault:

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

### Step 3: Start the proxy on each bot machine

On each machine running a bot, start the proxy with the vault server address and token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Replace **`192.168.x.x`** with the actual internal IP address of your vault server machine. You can find it in your router settings or by running `ip addr`.

---

## Auto-start Setup

If restarting wall-vault manually every time you reboot is a hassle, register it as a system service. Once registered, it starts automatically on boot.

### Linux — systemd (most Linux distros)

systemd is the Linux system that automatically starts and manages programs:

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

launchd is macOS's system for running programs automatically:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Download NSSM from [nssm.cc](https://nssm.cc/download) and add it to your PATH.
2. In an Administrator PowerShell:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Self-diagnosis Tool

The `doctor` command is a tool that **diagnoses and repairs** wall-vault configuration problems on its own.

```bash
wall-vault doctor check   # diagnose current state (read-only, changes nothing)
wall-vault doctor fix     # automatically fix problems
wall-vault doctor all     # diagnose + auto-fix in one go
```

> 💡 If something seems off, run `wall-vault doctor all` first. It catches and fixes a lot of issues automatically.

---

## Environment Variables Reference

Environment variables are a way to pass configuration values to a program. You can set them in your terminal with `export VARIABLE=value`, or put them in your auto-start service file to apply them permanently.

| Variable | Description | Example value |
|---|---|---|
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

### Proxy won't start

The port is most likely already in use by another program.

```bash
ss -tlnp | grep 56244   # check what's using port 56244
wall-vault proxy --port 8080   # start on a different port
```

### API key errors (429, 402, 401, 403, 582)

| Error code | Meaning | What to do |
|---|---|---|
| **429** | Too many requests (usage limit exceeded) | Wait a bit, or add another key |
| **402** | Payment required or insufficient credits | Top up credits on the service's website |
| **401 / 403** | Invalid key or unauthorized | Double-check the key value and re-register |
| **582** | Gateway overload (cooldown 5 minutes) | Automatically clears after 5 minutes |

```bash
# Check registered keys and their status
curl -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys

# Reset usage counters
curl -X POST -H "Authorization: Bearer your-admin-token" http://localhost:56243/admin/keys/reset
```

### Agent shows as "Not connected"

"Not connected" means the proxy process is not sending heartbeat signals to the vault. **It does not mean your settings were lost.** The proxy needs to be started with the vault server address and token in order to show as connected.

```bash
# Start the proxy with vault address, token, and client ID
WV_VAULT_URL=http://vault-server-address:56243 \
WV_VAULT_TOKEN=your-client-token \
WV_VAULT_CLIENT_ID=your-client-id \
wall-vault proxy
```

Once the connection succeeds, the dashboard will show 🟢 Running within about 20 seconds.

### Ollama won't connect

Ollama is a program that runs AI models directly on your machine. First, make sure Ollama is actually running.

```bash
curl http://localhost:11434/api/tags   # if you see a model list, it's working
export OLLAMA_URL=http://192.168.x.x:11434   # if it's running on a different machine
```

> ⚠️ If Ollama isn't responding, start it first with the `ollama serve` command.

> ⚠️ **Large models are slow**: Models like `qwen3.5:35b` or `deepseek-r1` can take several minutes to generate a response. If it looks like nothing is happening, it may just be processing — give it some time.

---

*For detailed API reference, see [API.md](API.md).*
