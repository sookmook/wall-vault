# wall-vault User Manual
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Table of Contents

1. [What is wall-vault?](#what-is-wall-vault)
2. [Installation](#installation)
3. [Getting Started (Setup Wizard)](#getting-started)
4. [Registering API Keys](#registering-api-keys)
5. [Using the Proxy](#using-the-proxy)
6. [Key Vault Dashboard](#key-vault-dashboard)
7. [Distributed Mode (Multi-Bot)](#distributed-mode-multi-bot)
8. [Auto-Start Configuration](#auto-start-configuration)
9. [Doctor](#doctor)
10. [RTK Token Savings](#rtk-token-savings)
11. [Environment Variables Reference](#environment-variables-reference)
12. [Troubleshooting](#troubleshooting)

---

## v0.2 Upgrade Notes

- `Service` gained `default_model` and `allowed_models`. The per-service default model is now set directly on the service card.
- `Client.default_service` / `default_model` have been renamed and reinterpreted as `preferred_service` / `model_override`. If the override is empty, the service's default model is used.
- On the first v0.2 startup, the existing `vault.json` is auto-migrated, and the pre-migration state is preserved as `vault.json.pre-v02.{timestamp}.bak`.
- The dashboard has been restructured into three zones: a left sidebar, a center card grid, and a right-side edit slideover.
- Admin API paths are unchanged, but request/response body schemas have been updated — old CLI scripts will need to be updated accordingly.

## v0.2.1 New Features

- **Multimodal pass-through (OpenAI → Gemini)**: `/v1/chat/completions` now accepts six content part types in addition to `text` — `input_audio`, `input_video`, `input_image`, `input_file`, and `image_url` (data URIs and external http(s) URLs ≤ 5 MB). The proxy converts each to Gemini's `inlineData`. OpenAI-compatible clients like EconoWorld can stream audio / image / video blobs directly.
- **EconoWorld agent type**: `POST /agent/apply` with `agentType: "econoworld"` writes wall-vault settings into the project's `analyzer/ai_config.json`. `workDir` accepts a comma-separated list of candidate paths and converts Windows drive paths to WSL mount paths.
- **Dashboard keys grid + CRUD**: 11 keys render as compact cards with + add / ✕ delete slideover.
- **Service add + drag-and-drop reorder**: services grid gains a + add button and a drag handle (`⋮⋮`).
- **Header / footer / theme animations / language switcher** restored. The 7 themes (cherry/dark/light/ocean/gold/autumn/winter) play their particle effect on a layer behind cards but above the background.
- **Slideover dismiss UX**: outside click or Esc closes the slideover.
- **SSE status indicator** in the footer (green = connected, orange = reconnecting, grey = disconnected).

## v0.2.2 Stability & UX Improvements

- **Dispatch fast-skip**: cloud services whose keys are all on cooldown or exhausted are no longer force-retried. Dispatch moves to the next fallback immediately. Per-request tail latency dropped from ~15 s to ~1.5 s.
- **Fallback model swap**: each fallback step now applies the target service's own `default_model`. Previously a `gemini-2.5-flash` request would be handed to Anthropic/Ollama verbatim and rejected (400/404).
- **Anthropic credit-balance handling**: when Anthropic returns HTTP 400 with a "credit balance" body, the proxy promotes it to 402-equivalent and sets a 30 min cooldown so subsequent dispatches skip Anthropic automatically.
- **Service edit default_model dropdown polish**:
  - The server now renders the complete model list (Google 15, OpenRouter 345, etc.) into the `<select>` from the first open — no second round-trip required.
  - `↓ Move to Allowed` button demotes the current default into the allowed_models textarea and clears the default.
  - `✕ Clear` empties the default in place.
  - Collapsible `Custom input` details block lets you type a model ID directly when the dropdown is unreachable.
- **Agent edit/create model_override dropdown**: free text replaced by a `<select>` populated from the preferred service's `default_model` + `allowed_models`. Changing the preferred service auto-repopulates the override options.
- **ClientInput v0.2 fields**: POST `/admin/clients` now accepts v0.2 canonical `preferred_service` / `model_override` alongside legacy `default_service` / `default_model` (legacy is a fallback).

---

## What is wall-vault?

**wall-vault = AI Proxy + API Key Vault for OpenClaw**

To use AI services, you need **API keys**. An API key is like a **digital pass** that proves "this person is authorized to use this service." However, these passes have daily usage limits, and there's always a risk of exposure if they're not managed properly.

wall-vault stores these passes in a secure vault and acts as a **proxy** between OpenClaw and AI services. In simple terms, OpenClaw only needs to connect to wall-vault, and wall-vault handles the rest.

Problems wall-vault solves:

- **Automatic API key rotation**: When one key hits its usage limit or gets temporarily blocked (cooldown), it silently switches to the next key. OpenClaw continues working without interruption.
- **Automatic service fallback**: If Google doesn't respond, it switches to OpenRouter. If that fails too, it falls back to locally installed AI (Ollama, LM Studio, vLLM). Your session never drops. When the original service recovers, it automatically switches back on the next request (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Real-time sync (SSE)**: When you change the model in the vault dashboard, it's reflected in the OpenClaw screen within 1-3 seconds. SSE (Server-Sent Events) is a technology where the server pushes changes to clients in real time.
- **Real-time notifications**: Events like key exhaustion or service outages are immediately displayed at the bottom of the OpenClaw TUI (terminal screen).

> :bulb: **Claude Code, Cursor, and VS Code** can also be connected, but wall-vault's primary purpose is to be used with OpenClaw.

```
OpenClaw (TUI terminal screen)
        |
        v
  wall-vault proxy (:56244)   <- Key management, routing, fallback, events
        |
        +-- Google Gemini API
        +-- OpenRouter API (340+ models)
        +-- Ollama / LM Studio / vLLM (local machine, last resort)
        +-- OpenAI / Anthropic API
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

- `curl -L ...` — Downloads the file from the internet.
- `chmod +x` — Makes the downloaded file "executable." If you skip this step, you'll get a "permission denied" error.

### Windows

Open PowerShell (as administrator) and run the following:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Add to PATH (takes effect after restarting PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **What is PATH?** It's a list of folders where your computer looks for commands. You need to add wall-vault to PATH so you can run `wall-vault` from any folder.

### Building from Source (for developers)

Only applicable if you have the Go development environment installed.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> :bulb: **Build timestamp version**: When you build with `make build`, the version is automatically generated in a format like `v0.1.27.20260409` that includes the date and time. If you build directly with `go build ./...`, the version will only show `"dev"`.

---

## Getting Started

### Running the Setup Wizard

After installation, be sure to run the **setup wizard** with the following command. The wizard will guide you through the necessary items one by one.

```bash
wall-vault setup
```

The wizard proceeds through the following steps:

```
1. Language selection (10 languages including English)
2. Theme selection (light / dark / gold / cherry / ocean)
3. Operation mode — standalone (single machine) or distributed (multiple machines)
4. Bot name — the name displayed on the dashboard
5. Port configuration — defaults: proxy 56244, vault 56243 (press Enter to keep defaults)
6. AI service selection — choose from Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool security filter settings
8. Admin token — a password that locks dashboard admin functions. Auto-generation available
9. API key encryption password — for more secure key storage (optional)
10. Config file save path
```

> :warning: **Remember your admin token.** You'll need it later to add keys or change settings in the dashboard. If you lose it, you'll need to edit the config file manually.

Once the wizard completes, a `wall-vault.yaml` config file is automatically created.

### Starting

```bash
wall-vault start
```

The following two servers start simultaneously:

- **Proxy** (`https://localhost:56244`) — The intermediary connecting OpenClaw and AI services
- **Key Vault** (`https://localhost:56243`) — API key management and web dashboard

Open `https://localhost:56243` in your browser to access the dashboard.

---

## Registering API Keys

There are four ways to register API keys. **Method 1 (environment variables) is recommended for beginners.**

### Method 1: Environment Variables (recommended — simplest)

Environment variables are **pre-set values** that programs read when they start. Just type the following in your terminal:

```bash
# Register a Google Gemini key
export WV_KEY_GOOGLE=AIzaSy...

# Register an OpenRouter key
export WV_KEY_OPENROUTER=sk-or-v1-...

# Start after registering
wall-vault start
```

If you have multiple keys, separate them with commas. wall-vault will use them in rotation automatically (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **Tip**: The `export` command only applies to the current terminal session. To persist across reboots, add the lines to your `~/.bashrc` or `~/.zshrc` file.

### Method 2: Dashboard UI (point and click)

1. Open `https://localhost:56243` in your browser
2. Click the `[+ Add]` button in the top **:key: API Keys** card
3. Enter the service type, key value, label (a memo name), and daily limit, then save

### Method 3: REST API (for automation/scripts)

REST API is a method for programs to exchange data via HTTP. Useful for automated registration via scripts.

```bash
curl -X POST https://localhost:56243/admin/keys \
  -H "Authorization: Bearer your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Main key",
    "daily_limit": 1000
  }'
```

### Method 4: Proxy Flags (for quick testing)

Use this for temporary testing without formal registration. Keys disappear when the program exits.

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
        baseUrl: "https://localhost:56244/v1",
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

> :bulb: **Easier method**: Press the **:lobster: Copy OpenClaw Config** button on the dashboard agent card. A snippet with the token and address pre-filled will be copied to your clipboard. Just paste it.

**Where does the `wall-vault/` prefix in the model name route to?**

wall-vault automatically determines which AI service to send the request to based on the model name:

| Model format | Routed service |
|-------------|---------------|
| `wall-vault/gemini-*` | Direct to Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Direct to OpenAI |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free 1M token context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/model-name`, `openai/model-name`, `anthropic/model-name`, etc. | Direct to the respective service |
| `custom/google/model-name`, `custom/openai/model-name`, etc. | Strips `custom/` prefix and re-routes |
| `model-name:cloud` | Strips `:cloud` suffix and routes to OpenRouter |

> :bulb: **What is context?** It's the amount of conversation an AI can remember at once. 1M (one million tokens) means it can process very long conversations or documents in a single pass.

### Direct Connection via Gemini API Format (existing tool compatibility)

If you have tools that use the Google Gemini API directly, just change the URL to wall-vault:

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244/google
```

Or if the tool specifies URLs directly:

```
https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Using with the OpenAI SDK (Python)

You can connect wall-vault to Python code that uses AI. Just change the `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://localhost:56244/v1",
    api_key="not-needed"  # wall-vault manages API keys for you
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Use provider/model format
    messages=[{"role": "user", "content": "Hello"}]
)
```

### Changing Models at Runtime

To change the AI model while wall-vault is already running:

```bash
# Change model via direct request to proxy
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# In distributed mode (multi-bot), change on the vault server -> instantly synced via SSE
curl -X PUT https://localhost:56243/admin/clients/my-bot-id \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Listing Available Models

```bash
# View full list
curl https://localhost:56244/api/models | python3 -m json.tool

# View Google models only
curl "https://localhost:56244/api/models?service=google"

# Search by name (e.g., models containing "claude")
curl "https://localhost:56244/api/models?q=claude"
```

**Key models by service:**

| Service | Key Models |
|---------|-----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context free, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Auto-detected from locally installed server |
| LM Studio | Local server (port 1234) |
| vLLM | Local server (port 8000) |
| llama.cpp | Local server (port 8080) |

---

## Key Vault Dashboard

Access the dashboard by opening `https://localhost:56243` in your browser.

**Screen layout:**
- **Top bar (sticky)**: Logo, language/theme selector, SSE connection status
- **Card grid**: Agent, service, and API key cards arranged in tile format

### API Key Card

A card for managing all registered API keys at a glance.

- Displays the key list grouped by service.
- `today_usage`: Tokens successfully processed today (characters read and written by the AI)
- `today_attempts`: Total calls today (successes + failures combined)
- Use the `[+ Add]` button to register new keys, and `x` to delete keys.

> :bulb: **What are tokens?** Tokens are the units AI uses to process text. Roughly one English word, or 1-2 Korean characters. API pricing is usually calculated based on token count.

### Agent Card

A card showing the status of bots (agents) connected to the wall-vault proxy.

**Connection status is displayed in 4 levels:**

| Indicator | Status | Meaning |
|-----------|--------|---------|
| :green_circle: | Running | Proxy is operating normally |
| :yellow_circle: | Delayed | Responding but slow |
| :red_circle: | Offline | Proxy is not responding |
| :black_circle: | Not connected / Disabled | Proxy has never connected to the vault or is disabled |

**Agent card bottom button guide:**

When you register an agent and specify the **agent type**, convenience buttons automatically appear for that type.

---

#### :radio_button: Copy Config Button — Automatically generates connection settings

Clicking the button copies a config snippet to your clipboard with the agent's token, proxy address, and model info pre-filled. Just paste the copied content into the location shown in the table below to complete the connection setup.

| Button | Agent type | Paste location |
|--------|-----------|---------------|
| :lobster: Copy OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: Copy NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: Copy Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| :keyboard: Copy Cursor Config | `cursor` | Cursor -> Settings -> AI |
| :computer: Copy VSCode Config | `vscode` | `~/.continue/config.json` |

**Example — For Claude Code type, this is what gets copied:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "this-agent's-token"
}
```

**Example — For VSCode (Continue) type:**

```yaml
# ~/.continue/config.yaml  <- paste into config.yaml, NOT config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: this-agent's-token
    roles:
      - chat
      - edit
      - apply
```

> :warning: **Recent versions of Continue use `config.yaml`.** If `config.yaml` exists, `config.json` is completely ignored. Be sure to paste into `config.yaml`.

**Example — For Cursor type:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : this-agent's-token

// Or environment variables:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=this-agent's-token
```

> :warning: **If clipboard copy doesn't work**: Browser security policies may block copying. If a popup text box appears, select all with Ctrl+A and copy with Ctrl+C.

---

#### :zap: Auto-Apply Button — One click and you're configured

For agents of type `cline`, `claude-code`, `openclaw`, or `nanoclaw`, a **:zap: Apply Config** button appears on the agent card. Clicking this button automatically updates the agent's local config file.

| Button | Agent type | Target file |
|--------|-----------|------------|
| :zap: Apply Cline Config | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: Apply Claude Code Config | `claude-code` | `~/.claude/settings.json` |
| :zap: Apply OpenClaw Config | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: Apply NanoClaw Config | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: This button sends a request to **localhost:56244** (local proxy). The proxy must be running on that machine for it to work.

---

#### :twisted_rightwards_arrows: Drag-and-Drop Card Sorting (v0.1.17, improved v0.1.25)

You can **drag** agent cards on the dashboard to rearrange them in any order.

1. Grab the **traffic light (bullet)** area at the top left of a card with your mouse and drag
2. Drop it on top of the card at the desired position to swap their order

> :bulb: The card body (input fields, buttons, etc.) cannot be dragged. You can only grab from the traffic light area.

#### :orange_circle: Agent Process Detection (v0.1.25)

When the proxy is working normally but a local agent process (NanoClaw, OpenClaw) has died, the card's traffic light changes to **orange (blinking)** and displays an "Agent process stopped" message.

- :green_circle: Green: Proxy + agent normal
- :orange_circle: Orange (blinking): Proxy normal, agent dead
- :red_circle: Red: Proxy offline
3. Changed order is **saved to the server immediately** and persists across page refreshes

> :bulb: Touch devices (mobile/tablet) are not yet supported. Use a desktop browser.

---

#### :arrows_counterclockwise: Bidirectional Model Sync (v0.1.16)

When you change an agent's model in the vault dashboard, the agent's local config is automatically updated.

**For Cline:**
- Model change in vault -> SSE event -> proxy updates model fields in `globalState.json`
- Updated fields: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` and API key are not touched
- **VS Code reload required (`Ctrl+Alt+R` or `Ctrl+Shift+P` -> `Developer: Reload Window`)**
  - Because Cline doesn't re-read config files while running

**For Claude Code:**
- Model change in vault -> SSE event -> proxy updates `model` field in `settings.json`
- Automatically searches both WSL and Windows paths (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Reverse direction (agent -> vault):**
- When agents (Cline, Claude Code, etc.) send requests to the proxy, the proxy includes the client's service/model info in the heartbeat
- The vault dashboard's agent card shows the currently used service/model in real time

> :bulb: **Key point**: The proxy identifies agents by the Authorization token in requests and automatically routes to the service/model configured in the vault. Even if Cline or Claude Code sends a different model name, the proxy overrides it with the vault setting.

---

### Using Cline in VS Code — Detailed Guide

#### Step 1: Install Cline

Install **Cline** (ID: `saoudrizwan.claude-dev`) from the VS Code Extension Marketplace.

#### Step 2: Register Agent in Vault

1. Open the vault dashboard (`http://vault-IP:56243`)
2. Click **+ Add** in the **Agents** section
3. Fill in the following:

| Field | Value | Description |
|-------|-------|-------------|
| ID | `my_cline` | Unique identifier (alphanumeric, no spaces) |
| Name | `My Cline` | Name displayed on the dashboard |
| Agent Type | `cline` | <- Must select `cline` |
| Service | Choose service (e.g., `google`) | |
| Model | Enter model (e.g., `gemini-2.5-flash`) | |

4. Click **Save** to auto-generate a token

#### Step 3: Connect Cline

**Method A — Auto-apply (recommended)**

1. Verify the wall-vault **proxy** is running on this machine (`localhost:56244`)
2. Click the **:zap: Apply Cline Config** button on the dashboard agent card
3. Success when you see "Config applied!" notification
4. Reload VS Code (`Ctrl+Alt+R`)

**Method B — Manual setup**

Open settings (:gear:) in the Cline sidebar:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://proxy-address:56244/v1`
  - Same machine: `https://localhost:56244/v1`
  - Different machine (e.g., Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Token issued from the vault (copy from agent card)
- **Model ID**: Model configured in the vault (e.g., `gemini-2.5-flash`)

#### Step 4: Verify

Send any message in the Cline chat window. If working correctly:
- The corresponding agent card on the vault dashboard shows a **green dot (Running)**
- The card displays the current service/model (e.g., `google / gemini-2.5-flash`)

#### Changing Models

When you want to change Cline's model, do it from the **vault dashboard**:

1. Change the service/model dropdown on the agent card
2. Click **Apply**
3. Reload VS Code (`Ctrl+Alt+R`) — The model name in the Cline footer updates
4. The new model is used from the next request

> :bulb: In practice, the proxy identifies Cline's requests by token and routes them to the vault-configured model. Even without a VS Code reload, **the actual model used changes immediately** — the reload is just to update the model display in the Cline UI.

#### Disconnect Detection

When you close VS Code, the agent card on the vault dashboard turns yellow (delayed) after about **90 seconds**, then red (offline) after **3 minutes**. (Since v0.1.18, status checks every 15 seconds have made offline detection faster.)

#### Troubleshooting

| Symptom | Cause | Solution |
|---------|-------|----------|
| "Connection failed" error in Cline | Proxy not running or wrong address | Check proxy with `curl https://localhost:56244/health` |
| Green dot doesn't appear in vault | API key (token) not configured | Click **:zap: Apply Cline Config** button again |
| Model in Cline footer doesn't change | Cline caches settings | Reload VS Code (`Ctrl+Alt+R`) |
| Wrong model name displayed | Old bug (fixed in v0.1.16) | Update proxy to v0.1.16 or later |

---

#### :purple_circle: Copy Deploy Command Button — For installing on new machines

Use this when first installing wall-vault proxy on a new computer and connecting it to the vault. Clicking the button copies the entire installation script. Paste it into the terminal on the new computer and run it — the following is handled all at once:

1. wall-vault binary installation (skipped if already installed)
2. Automatic systemd user service registration
3. Service start and automatic vault connection

> :bulb: The script comes pre-filled with this agent's token and vault server address, so you can run it immediately after pasting without any modifications.

---

### Service Card

A card for enabling/disabling and configuring AI services.

- Per-service enable/disable toggle switch
- Enter the address of local AI servers (Ollama, LM Studio, vLLM, llama.cpp, etc. running on your machine) and available models are automatically discovered.
- **Local service connection status**: The dot next to the service name is **green** if connected, **gray** if not
- **Local service auto traffic light** (v0.1.23+): Local services (Ollama, LM Studio, vLLM, llama.cpp) are automatically enabled/disabled based on connectivity. When a service becomes reachable, it turns green and the checkbox activates within 15 seconds; when it becomes unreachable, it automatically deactivates. This works the same way cloud services (Google, OpenRouter, etc.) auto-toggle based on API key availability.
- **Reasoning mode toggle** (v0.2.17+): A **reasoning mode** checkbox appears at the bottom of the local service edit pane. When enabled, the proxy adds `"reasoning": true` to the chat-completions body sent upstream, so models that expose their thought process — such as DeepSeek R1 or Qwen QwQ — return a `<think>…</think>` block alongside the answer. Servers that don't recognize the field simply ignore it, so it's safe to leave enabled even in mixed workloads.

> :bulb: **If a local service is running on another computer**: Enter that computer's IP in the service URL field. Example: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). If the service is bound to `127.0.0.1` rather than `0.0.0.0`, external IP access won't work — check the binding address in the service settings.

### Admin Token Entry

When you try to use important features like adding or deleting keys in the dashboard, an admin token input popup appears. Enter the token you set during the setup wizard. Once entered, it persists until you close the browser.

> :warning: **If authentication failures exceed 10 within 15 minutes, that IP is temporarily blocked.** If you've forgotten your token, check the `admin_token` field in your `wall-vault.yaml` file.

---

## Distributed Mode (Multi-Bot)

When running OpenClaw simultaneously on multiple computers, this configuration **shares a single key vault**. It's convenient because you only need to manage keys in one place.

### Example Configuration

```
[Key Vault Server]
  wall-vault vault    (key vault :56243, dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]   [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> SSE sync          <-> SSE sync            <-> SSE sync
```

All bots point to the central vault server, so model changes or key additions in the vault are instantly reflected across all bots.

### Step 1: Start the Key Vault Server

Run this on the computer that will serve as the vault server:

```bash
wall-vault vault
```

### Step 2: Register Each Bot (Client)

Pre-register each bot's information that will connect to the vault server:

```bash
curl -X POST https://localhost:56243/admin/clients \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Step 3: Start Proxy on Each Bot Machine

On each bot machine, start the proxy with the vault server address and token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: Replace **`192.168.x.x`** with the actual internal IP address of the vault server machine. You can find it in router settings or via the `ip addr` command.

---

## Auto-Start Configuration

If it's tedious to manually start wall-vault every time you reboot, register it as a system service. Once registered, it starts automatically on boot.

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

1. Download NSSM from [nssm.cc](https://nssm.cc/download) and add it to PATH.
2. In an administrator PowerShell:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor

The `doctor` command is a tool that **self-diagnoses and repairs** wall-vault configuration issues.

```bash
wall-vault doctor check   # Diagnose current state (read-only, changes nothing)
wall-vault doctor fix     # Automatically repair problems
wall-vault doctor all     # Diagnose + auto-repair in one step
```

> :bulb: If something seems off, run `wall-vault doctor all` first. It catches and fixes many problems automatically.

---

## RTK Token Savings

*(v0.1.24+)*

**RTK (Token Reduction Kit)** automatically compresses the output of shell commands executed by AI coding agents (like Claude Code), reducing token usage. For example, 15 lines of `git status` output can be reduced to a 2-line summary.

### Basic Usage

```bash
# Wrap commands with wall-vault rtk to auto-filter output
wall-vault rtk git status          # Shows only changed file list
wall-vault rtk git diff HEAD~1     # Changed lines + minimal context only
wall-vault rtk git log -10         # Hash + one-line message each
wall-vault rtk go test ./...       # Shows only failed tests
wall-vault rtk ls -la              # Unsupported commands get auto-truncated
```

### Supported Commands and Savings

| Command | Filter method | Savings |
|---------|--------------|---------|
| `git status` | Changed file summary only | ~87% |
| `git diff` | Changed lines + 3-line context | ~60-94% |
| `git log` | Hash + first line message | ~90% |
| `git push/pull/fetch` | Progress removed, summary only | ~80% |
| `go test` | Failures only, passes counted | ~88-99% |
| `go build/vet` | Errors only | ~90% |
| All other commands | First 50 + last 50 lines, max 32KB | Variable |

### 3-Stage Filter Pipeline

1. **Command-specific structural filter** — Understands output format of git, go, etc. and extracts meaningful parts
2. **Regex post-processing** — Removes ANSI color codes, collapses blank lines, aggregates duplicate lines
3. **Passthrough + truncation** — Unsupported commands keep only first/last 50 lines

### Claude Code Integration

You can configure Claude Code's `PreToolUse` hook to automatically pass all shell commands through RTK.

```bash
# Install hook (automatically added to Claude Code settings.json)
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

> :bulb: **Exit code preservation**: RTK returns the original command's exit code unchanged. If a command fails (exit code != 0), the AI accurately detects the failure.

> :bulb: **English forced**: RTK runs commands with `LC_ALL=C`, ensuring English output regardless of system language settings. This is necessary for filters to work correctly.

---

## Environment Variables Reference

Environment variables are a way to pass configuration values to programs. Type `export VARIABLE=value` in the terminal, or place them in auto-start service files for permanent application.

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

### Proxy Won't Start

The port is likely already in use by another program.

```bash
ss -tlnp | grep 56244   # Check what's using port 56244
wall-vault proxy --port 8080   # Start on a different port
```

### API Key Errors (429, 402, 401, 403, 582)

| Error Code | Meaning | Resolution |
|------------|---------|------------|
| **429** | Too many requests (quota exceeded) | Wait a moment or add more keys |
| **402** | Payment required or credits depleted | Recharge credits on the service |
| **401 / 403** | Invalid key or no permission | Verify key value and re-register |
| **582** | Gateway overload (5-minute cooldown) | Automatically clears after 5 minutes |

```bash
# Check registered key list and status
curl -H "Authorization: Bearer admin-token" https://localhost:56243/admin/keys

# Reset key usage counters
curl -X POST -H "Authorization: Bearer admin-token" https://localhost:56243/admin/keys/reset
```

### Agent Shows "Not Connected"

"Not connected" means the proxy process is not sending heartbeats to the vault. **It does not mean settings haven't been saved.** The proxy must be running with the vault server address and token to enter a connected state.

```bash
# Start proxy with vault server address, token, and client ID
WV_VAULT_URL=http://vault-server:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Once connected, the dashboard shows :green_circle: Running within about 20 seconds.

### Ollama Won't Connect

Ollama is a program that runs AI directly on your computer. First check if Ollama is running.

```bash
curl http://localhost:11434/api/tags   # If a model list appears, it's working
export OLLAMA_URL=http://192.168.x.x:11434   # If running on another computer
```

> :warning: If Ollama isn't responding, start it first with the `ollama serve` command.

> :warning: **Large models are slow**: Big models like `qwen3.5:35b` or `deepseek-r1` can take several minutes to generate a response. Even if it seems like nothing is happening, it may still be processing — please be patient.

---

## Recent Changes (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Ollama fallback model name fix**: Fixed an issue where provider-prefixed model names (e.g., `google/gemini-3.1-pro-preview`) were passed directly to Ollama during fallback from other services. Now automatically replaced with environment variable/default model.
- **Cooldown duration significantly reduced**: 429 rate limit 30min->5min, 402 payment 1hour->30min, 401/403 24hours->6hours. Prevents total proxy paralysis when all keys enter cooldown simultaneously.
- **Forced retry on full cooldown**: When all keys are in cooldown, the key closest to expiring is force-retried to prevent request rejection.
- **Service list display fix**: The `/status` response now shows the actual service list synced from the vault (prevents anthropic etc. from being omitted).

### v0.1.25 (2026-04-08)
- **Agent process detection**: The proxy detects whether local agents (NanoClaw/OpenClaw) are alive and shows an orange traffic light on the dashboard.
- **Drag handle improvement**: Card sorting now only allows grabbing from the traffic light area. Prevents accidental dragging from input fields or buttons.

### v0.1.24 (2026-04-06)
- **RTK token savings subcommand**: `wall-vault rtk <command>` auto-filters shell command output, reducing AI agent token usage by 60-90%. Includes built-in filters for major commands like git and go, and auto-truncates unsupported commands. Transparently integrates with Claude Code via `PreToolUse` hook.

### v0.1.23 (2026-04-06)
- **Ollama model change fix**: Fixed an issue where changing the Ollama model in the vault dashboard didn't reflect on the actual proxy. Previously only the environment variable (`OLLAMA_MODEL`) was used; now vault settings take priority.
- **Local service auto traffic light**: Ollama, LM Studio, and vLLM are automatically enabled when reachable and disabled when disconnected. Works the same way as key-based auto-toggle for cloud services.

### v0.1.22 (2026-04-05)
- **Empty content field omission fix**: When thinking models (gemini-3.1-pro, o1, claude thinking, etc.) used all max_tokens for reasoning and couldn't produce an actual response, the proxy was omitting `content`/`text` fields from the response JSON via `omitempty`, causing OpenAI/Anthropic SDK clients to crash with `Cannot read properties of undefined (reading 'trim')`. Fixed to always include fields per official API specs.

### v0.1.21 (2026-04-05)
- **Gemma 4 model support**: Gemma series models like `gemma-4-31b-it` and `gemma-4-26b-a4b-it` can now be used through the Google Gemini API.
- **LM Studio / vLLM full service support**: Previously, these services were missing from proxy routing and always fell back to Ollama. Now properly routed via OpenAI-compatible API.
- **Dashboard service display fix**: Even during fallback, the dashboard always shows the user-configured service.
- **Local service status display**: On dashboard load, local service (Ollama, LM Studio, vLLM, etc.) connection status is shown via dot color.
- **Tool filter environment variable**: Set tool pass-through mode with `WV_TOOL_FILTER=passthrough` environment variable.

### v0.1.20 (2026-03-28)
- **Comprehensive security hardening**: XSS prevention (41 locations), constant-time token comparison, CORS restrictions, request size limits, path traversal prevention, SSE authentication, rate limiter hardening, and 12 other security improvements.

### v0.1.19 (2026-03-27)
- **Claude Code online detection**: Claude Code instances not going through the proxy are also shown as online on the dashboard.

### v0.1.18 (2026-03-26)
- **Fallback service sticking fix**: After temporary errors cause Ollama fallback, automatic return to the original service when it recovers.
- **Offline detection improvement**: Status checks every 15 seconds provide faster proxy outage detection.

### v0.1.17 (2026-03-25)
- **Drag-and-drop card sorting**: Agent cards can be dragged to reorder.
- **Inline config apply button**: Offline agents show a [:zap: Apply Config] button.
- **cokacdir agent type added**.

### v0.1.16 (2026-03-25)
- **Bidirectional model sync**: Changing the model for Cline or Claude Code in the vault dashboard is automatically reflected.

---

*For more detailed API information, see [API.md](API.md).*
