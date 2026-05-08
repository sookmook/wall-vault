# wall-vault User Manual

English · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

This manual covers installing, configuring, and operating wall-vault. For an at-a-glance overview see the [README](../README.md). For HTTP API details see the [API reference](API.md).

## Contents

1. [What wall-vault does](#what-wall-vault-does)
2. [Installation](#installation)
3. [First run with the setup wizard](#first-run-with-the-setup-wizard)
4. [Enabling TLS](#enabling-tls)
5. [Registering API keys](#registering-api-keys)
6. [Connecting agents](#connecting-agents)
7. [The dashboard](#the-dashboard)
8. [Distributed mode](#distributed-mode)
9. [Auto-start](#auto-start)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Environment variables](#environment-variables)
14. [Troubleshooting](#troubleshooting)

---

## What wall-vault does

wall-vault is a single Go binary that bundles two cooperating services:

- **The vault** stores API keys encrypted at rest (AES-GCM with a master password), tracks usage and cooldowns per key, broadcasts changes over Server-Sent Events (SSE), and serves a web dashboard at `:56243` for human operators.
- **The proxy** exposes Gemini, Anthropic, OpenAI-compatible, and Ollama-native endpoints at `:56244`. Any AI client that points at the proxy is using the keys in the vault — clients never see them. When one upstream fails, dispatch falls back to the next provider in order.

This is useful when:

- You have keys for several providers and want one URL the agent talks to.
- You want a free-tier key on cooldown to step aside without breaking the session.
- You want the same keys to power multiple bots, IDEs, or scripts on the same LAN without copying credentials.
- You want a dashboard, not environment variables, for editing keys and switching models.
- You want a local fallback (Ollama, LM Studio, vLLM) when cloud limits run out.

```
   AI client (OpenClaw, Claude Code, Cursor, …)
            │
            ▼
   wall-vault proxy  :56244
            │  (selects key, dispatches, falls back on failure)
            ├──► Google Gemini
            ├──► Anthropic
            ├──► OpenAI
            ├──► OpenRouter (340+ models, auto :free fallback)
            └──► Local OAI-compat backends (Ollama / LM Studio / vLLM / …)

   vault (AES-GCM key store + dashboard)  :56243
            ▲
            │  SSE broadcast on change
   Multiple proxies on different hosts can share one vault.
```

---

## Installation

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

The script auto-detects OS and architecture, downloads the right binary into `~/.local/bin/wall-vault`, and makes it executable. If `~/.local/bin` is not on your `PATH`, add it:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Manual download

Pre-built binaries are published on every release at `https://github.com/sookmook/wall-vault/releases`.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Intel
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-amd64 \
  -o wall-vault && chmod +x wall-vault
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Build from source

Requires Go 1.25 or newer.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` cross-compiles to all five supported platforms. Binaries land in `bin/`.

---

## First run with the setup wizard

```bash
wall-vault setup
```

The wizard prompts you for, in order:

1. **Language** — picks one of 17 UI locales. Detected automatically from `$LANG`; the wizard offers a list anyway.
2. **Theme** — `light` (default), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Cosmetic only.
3. **Mode** — `standalone` (single host, default) or `distributed` (vault on one host, proxies on others).
4. **Bot name** — a free-form `client_id` slug. The vault uses this to scope per-client config (model overrides, fallback chains).
5. **Proxy port** — default `56244`.
6. **Vault port** — default `56243` (standalone only).
7. **Service selection** — a y/N for each of: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Multiple choices are fine; each one writes its env-var hint at the end.
8. **Tool filter** — `strip_all` (default; blocks all incoming tool definitions for security) or `passthrough` (let any tool through).
9. **Admin token** — leave blank to auto-generate. The dashboard requires this token to log in.
10. **Master password** — leave blank for no encryption (NOT recommended); set a value to AES-GCM encrypt the key store at rest.
11. **Save path** — defaults to `wall-vault.yaml` in the current directory. The loader also looks at `~/.wall-vault/config.yaml`.

After saving, the wizard runs `doctor.FixTrust` so any locally-installed agent (OpenClaw, Claude Code, Cline) gets the wall-vault internal CA added to its trust store automatically. If no such agent is installed, the step prints `SKIP` and writes nothing.

Then start the binary:

```bash
wall-vault start
```

`start` runs both the vault and the proxy in one process (standalone mode). For distributed mode use `wall-vault vault` on the vault host and `wall-vault proxy` on each proxy host.

Open `http://localhost:56243` in a browser. Log in with the admin token the wizard printed.

---

## Enabling TLS

The wizard's defaults leave both listeners on plain HTTP. Most agents (OpenClaw, Claude Code, Cursor) work better against a single HTTPS endpoint, so TLS is recommended in any deployment that spans more than the local machine.

wall-vault ships with its own internal CA so you don't need a public DNS name or Let's Encrypt.

```bash
# 1. Create the internal CA — written to ~/.wall-vault/ca.{crt,key}.
#    The CA is good for 10 years by default; override with --ca-years.
wall-vault cert init

# 2. Issue a host certificate. Subject Alternative Names automatically include:
#       hostname, "localhost", "127.0.0.1", and any non-loopback LAN IP detected.
#    Override the issuer dir with --dir, validity with --host-years.
wall-vault cert issue $(hostname)

# 3. Trust the CA in this machine's OS keychain.
#    Linux: writes to /etc/ssl/certs/ via update-ca-certificates (needs sudo).
#    macOS: adds to the System keychain via security add-trusted-cert (needs sudo).
#    Windows: imports into CurrentUser\Root via certutil (no admin needed).
wall-vault cert install-trust

# 4. Enable TLS on both listeners.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

To extend trust to other LAN machines, copy `~/.wall-vault/ca.crt` over and run `wall-vault cert install-trust --ca <path>` on each one. The vault also exposes `ca.crt` over a tiny plain-HTTP listener at `:56247` (the **bootstrap port**) for the catch-22 case where a fresh client needs the CA to talk HTTPS.

### Loopback HTTP companion

Some agents — notably OpenClaw's bundled Node runtime — rewrite `NODE_EXTRA_CA_CERTS` at process spawn, dropping any operator-supplied CA hint. They cannot honour the wall-vault CA from inside the daemon, even after `cert install-trust`. wall-vault works around this by binding an additional **loopback-only plain-HTTP listener** at `127.0.0.1:56245` whenever TLS is enabled. Same-host clients reach the proxy through that port without TLS at all; LAN clients keep using the TLS listener.

Disable with `WV_PROXY_PLAIN_PORT=0` if you don't need it.

### `wall-vault cert list`

Shows every cert under `~/.wall-vault/` with subject, validity window, and SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Registering API keys

Two ways: the dashboard, or environment variables.

### Dashboard (recommended)

1. Log in at `https://localhost:56243` with the admin token.
2. Click **+ API key** in the keys card.
3. Pick a service (Google, OpenRouter, Anthropic, OpenAI, …).
4. Paste the key. Save.

Multiple keys per service are fine; the proxy round-robins between them and skips ones that hit a per-key cooldown.

### Environment variables (one-shot bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Keys provided this way are written into the encrypted store on first launch. Subsequent starts read them from disk; you can unset the env vars after the first run.

### Cooldowns and rotation

Every successful call increments the key's `usage_count` and refreshes `last_used`. On HTTP 429 / 402 / 403, the proxy puts the key on a **cooldown** (defaults: 60 minutes for 429, 24 hours for 402, 12 hours for 403). The next dispatch picks a different key for that service. When all keys for a service are on cooldown, the proxy fast-skips that service entirely and tries the next provider in the fallback chain.

Cooldowns are visible per-key in the dashboard with a countdown.

---

## Connecting agents

### OpenClaw

OpenClaw is the original target client. Use the dashboard's **+ Add agent** modal:

- Set **Agent type** to `openclaw` or `nanoclaw`.
- Set **Work directory** — for OpenClaw this auto-fills as `~/.openclaw`.
- Choose a **preferred service** and optionally a **model override**.
- Click **Apply**. wall-vault writes `~/.openclaw/openclaw.json` directly (provider URLs, vault token, model entries).

When you change the model from the dashboard, OpenClaw picks up the change over SSE within 1–3 seconds — no restart.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

When upstream Anthropic credits run out, dispatch falls back to whichever services are listed in this client's `fallback_services`. By default, a non-Claude model id sent to the anthropic dispatch returns an error so misrouting surfaces immediately. Opt in to automatic rewrite:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

In Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # or any model wall-vault knows
```

### Continue (VS Code, JetBrains)

`config.json`:

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### Custom HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

The same endpoint accepts streaming (`"stream": true`) when `proxy.oai_stream_forward: true` is set.

---

## The dashboard

`https://localhost:56243`. Five cards on the home grid:

- **Keys** — every API key, grouped by service. Add, edit, delete; see usage and cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, plus any plugin yaml in `~/.wall-vault/services/`. Set per-service `default_model`, `allowed_models`, base URL, reasoning toggle.
- **Clients (agents)** — every registered client (OpenClaw bot, Claude Code session, Cursor instance, …). Assign preferred service, model override, fallback chain.
- **Proxies** — every proxy that has authenticated against this vault. Live status (online/offline), last seen, current model.
- **Settings** — admin token, master password rotation, theme, language.

Each card has an edit slideover (right side). Outside-click or `Esc` closes it. Changes are pushed to all connected proxies over SSE within seconds.

The **footer** carries an SSE indicator (green = connected, orange = reconnecting, grey = disconnected) and the live build version.

---

## Distributed mode

When you have several machines that all need the same keys, run the vault on one host and proxies on each of the others.

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

The dashboard is now reachable at `https://<vault-host>:56243`. Add an agent for each remote proxy in the **Clients** card; each one mints a unique `vault_token`.

### Proxy hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

The proxy authenticates against the vault, opens an SSE stream, and applies any config it receives (preferred service, model override, fallback chain). Subsequent vault edits land in seconds with no restart.

For LAN-spanning installs, enable TLS on the vault host (`WV_VAULT_TLS_ENABLED=1` + the cert/key env vars) and run each proxy host through the same `wall-vault cert install-trust` step so the proxy's HTTPS calls into vault are trusted.

---

## Auto-start

### systemd (Linux)

```ini
# ~/.config/systemd/user/wall-vault-proxy.service
[Unit]
Description=wall-vault proxy
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/wall-vault proxy
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now wall-vault-proxy
loginctl enable-linger $USER       # so the unit keeps running after logout
```

For the vault on the same host, write a parallel `wall-vault-vault.service`. For standalone mode, one unit calling `wall-vault start` is enough.

### launchd (macOS)

```xml
<!-- ~/Library/LaunchAgents/com.wall-vault.proxy.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.wall-vault.proxy</string>
  <key>ProgramArguments</key>
  <array><string>/usr/local/bin/wall-vault</string><string>proxy</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>/tmp/wall-vault.proxy.log</string>
  <key>StandardErrorPath</key><string>/tmp/wall-vault.proxy.err</string>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist
```

### Windows

Use `nssm` to wrap `wall-vault.exe start` as a Windows service, or a `schtasks` entry that runs at user logon.

---

## Plugin yamls

Any OpenAI-compatible backend can be added without code changes by dropping a yaml under `~/.wall-vault/services/`. wall-vault loads it at startup and registers the service for dispatch, the OAI-compat detection set, and the Gemini-stream bridge.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # unique service id
name: llama.cpp              # human label
enabled: true                # disabled plugins are skipped at load

default_url: http://localhost:8080   # operator override; env wins (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # for query_param: the param name (e.g. "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # let the dashboard auto-detect models
  dynamic: true              # re-fetch on every dashboard open
  auto_detect_url: true      # try /v1/models even when not declared

concurrency:
  max: 1                     # max concurrent requests to this backend
  queue_size: 10
  wait_notify: true          # show "queued" hint to TUI agents

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# Opt in to qwen3-family inline /no_think directive when reasoning is off.
# Set true if your backend's chat template strips the marker (LM Studio's
# jinja, Ollama's /v1 layer). Other backends typically echo the literal
# text back, so this stays opt-in per yaml.
inline_no_think_for_qwen3: false

# Hub topology — point at another wall-vault. Required when this plugin
# fronts a remote wall-vault (so the receiving wall-vault sees the
# publisher prefix and routes correctly) and so the bearer token in
# proxy.vault_token is sent as Authorization.
preserve_model_id: false
tls_internal_ca: false       # add ~/.wall-vault/ca.crt to client trust pool
```

The bundled set in `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) ships disabled by default. Copy the one you want into `~/.wall-vault/services/`, set `enabled: true`, restart.

---

## Doctor

`wall-vault doctor` runs a one-shot health probe across the whole install:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Each line is one of:

- `✓` — healthy
- `⚠` — degraded but functioning (one key cooled down, low quota, etc.)
- `✗` — broken
- `SKIP` — not configured / not applicable on this host

A second daemon mode runs the same probe every `doctor.interval` (default 5 minutes) and writes results to `doctor.log_file` (default `/tmp/wall-vault-doctor.log`). When `doctor.auto_fix` is true, it also tries to repair common drift (stale OpenClaw config, missing TLS trust, restartable services).

Trigger one-shot from the dashboard via the **Doctor** card or `wall-vault doctor`.

---

## Hooks

Run a shell command on key events:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Each hook gets event-specific environment variables (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hooks run async with a 5-second timeout — the proxy never blocks on a slow hook.

---

## Environment variables

| Variable | YAML field |
|----------|------------|
| `WV_LANG` | `lang` |
| `WV_THEME` | `theme` |
| `WV_PROXY_PORT` | `proxy.port` |
| `WV_PROXY_HOST` | `proxy.host` |
| `WV_VAULT_PORT` | `vault.port` |
| `WV_VAULT_HOST` | `vault.host` |
| `WV_VAULT_URL` | `proxy.vault_url` (distributed) |
| `WV_VAULT_TOKEN` | `proxy.vault_token` |
| `WV_ADMIN_TOKEN` | `vault.admin_token` |
| `WV_MASTER_PASS` | `vault.master_password` |
| `WV_AVATAR` | `proxy.avatar` |
| `WV_TOOL_FILTER` | `proxy.tool_filter` |
| `WV_CC_CLIENT_ID` | `proxy.claude_code_client_id` |
| `WV_PROXY_TLS_ENABLED` | `proxy.tls.enabled` |
| `WV_PROXY_TLS_CERT` | `proxy.tls.cert_file` |
| `WV_PROXY_TLS_KEY` | `proxy.tls.key_file` |
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (refuse to start with TLS off — fails closed when set) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (comma-separated list, e.g. `192.168.0.0/16,10.0.0.0/8`; loopback always passes) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | One-shot import: comma-separated Google keys |
| `WV_KEY_OPENROUTER` | One-shot import: OpenRouter keys |
| `WV_KEY_ANTHROPIC` | One-shot import: Anthropic keys |
| `WV_KEY_OPENAI` | One-shot import: OpenAI keys |
| `WV_OLLAMA_URL` | Per-host Ollama URL override (single instance) |
| `WV_OLLAMA_URLS` | Comma-separated Ollama URLs (multi-instance dispatch) |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Per-backend URL override (single instance) |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity` (system-message identity guard, off by default) |
| `WV_PROMPT_TOKEN_CAP` | Per-host auto-truncate cap for local OAI-compat prompts (positive int = enable, 0 = off) |
| `WV_DISPATCH_TRACE` | Set to `1` to log every dispatch's resolved service / model with reason (off by default) |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Every env var, when set, wins over the YAML file.

---

## Troubleshooting

### `connection refused` on `:56244`

Either the proxy is not running or it's bound to a different host. Check:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

If it's running on a different port, your config has `proxy.port` overridden — check `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

The client doesn't trust the wall-vault internal CA. Run `wall-vault cert install-trust` on the client machine. For agents whose runtime ignores the OS trust store (e.g. Node with a hardcoded `NODE_EXTRA_CA_CERTS`), use the loopback HTTP companion at `127.0.0.1:56245` (same-host only) or set `WV_PROXY_TLS_ENABLED=0` to fall back to plain HTTP.

### `token not registered with vault`

The client's `Authorization: Bearer <token>` doesn't match any registered client. Verify the token under **Clients** in the dashboard. If you copied a token literal like `proxy-managed`, `dummy`, or `""` from a stale config, replace it with the real client token.

### `Anthropic dispatch needs a Claude model id`

Default behaviour as of v0.2.63: a non-Claude model id sent to anthropic dispatch returns an error. Either fix the routing (don't send `gemini-2.5-flash` to anthropic) or opt in to automatic rewrite via `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

Dispatch saw a service id no plugin yaml claimed. Check:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

If the yaml exists but is `enabled: false`, flip it. If it's missing entirely, copy from `configs/services/` in the source tree.

### Empty response on a reasoning model

`qwen3.6`, `deepseek-r1`, and the GPT-`o1` family sometimes emit only `reasoning_content` and leave `content` empty. As of v0.2.63 wall-vault falls back to the reasoning text automatically — if you still see empty responses, the backend is returning neither field. Check the upstream's logs.

For LM Studio with qwen3 specifically, set `inline_no_think_for_qwen3: true` in the plugin yaml so reasoning gets disabled inline. Built-in lmstudio.yaml and ollama.yaml already do this.

### Dashboard shows "all keys on cooldown" but I just added one

The new key is healthy but the dispatch path may still be in the cooldown for an older key. Try a fresh request — the proxy round-robins per call, and a healthy key will be picked next.

### Vault won't unlock with master password

Wrong password. There's no recovery — wall-vault deliberately doesn't ship a backdoor. If you've genuinely lost the master password, the only path is to delete `~/.wall-vault/data/vault.json`, restart with a new password, and re-add the keys.

### Free-tier OpenRouter limits hit

Set `proxy.services` to include `openrouter` and add at least one OpenRouter key. The proxy auto-falls-back from a paid model to its `:free` variant when the paid path returns 402 / 429.

### `journalctl --user -u wall-vault-proxy` is empty

systemd `--user` logs go to the journal of the user running it. If you started the unit as `root` or via `sudo`, the journal is in the system instance instead — try `journalctl -u wall-vault-proxy` without `--user`.

---

## More

- HTTP API reference — see [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- Bug reports / feature requests — GitHub Issues
- Release history — [CHANGELOG.md](../CHANGELOG.md)
