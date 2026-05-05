<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

# wall-vault

> **API key vault + AI proxy in a single Go binary.**
> Stores keys locally with AES-GCM, rotates them across providers, falls back when one fails, and ships with a real-time dashboard.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

English · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## What it is

wall-vault sits between an AI agent (OpenClaw, Claude Code, Cursor, Continue, your own script) and the cloud or local AI providers it talks to. Two things in one binary:

- **Vault** — stores API keys encrypted at rest (AES-GCM with a master password), rotates them, tracks per-key usage and cooldowns, broadcasts changes over SSE, and serves a web dashboard at `:56243`.
- **Proxy** — exposes Gemini, Anthropic, and OpenAI-compatible endpoints at `:56244`, picks a key from the vault, dispatches to the upstream you configured, and falls back to the next provider when one fails.

It supports four request shapes (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, and Ollama-native `/api/chat`) and five categories of upstream:

| Provider | Notes |
|----------|-------|
| Google Gemini | Native API; key rotation per project |
| Anthropic | Native `/v1/messages` passthrough |
| OpenAI | Native `/v1/chat/completions` |
| OpenRouter | 340+ models, auto-fallback to `:free` variants |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI-compat local backends; drop-in via plugin yaml |

Adding a new OpenAI-compatible backend is one yaml file under `~/.wall-vault/services/` — no code change.

## Why you might want it

- You're juggling three or four AI services and want one URL the agent talks to.
- You want a free-tier key on a cooldown to step aside for the next one without breaking the session.
- You want the same keys to power multiple bots / IDEs / scripts on the same LAN without copying credentials.
- You want a dashboard, not environment variables, for editing API keys.
- You want a local-first option (Ollama / LM Studio) when cloud limits run out.

## Quick start

### Install (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Or download a pre-built binary directly:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Install (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### First run

```bash
wall-vault setup    # interactive wizard — picks port, services, admin token, master password
wall-vault start    # launches both vault and proxy
```

Open `http://localhost:56243` (or `https://...` once TLS is on — see below) in a browser. The dashboard prompts for the admin token printed by `setup`. From there you add API keys, register clients, and switch models without restarting.

---

## TLS (recommended)

By default `wall-vault setup` writes a config without TLS, so both listeners answer plain HTTP. The example URLs in this README use `https://localhost:56244` because most agents (OpenClaw, Claude Code, Cursor) want a single TLS-fronted endpoint that won't break if you later move the proxy to another host. To match those examples, enable TLS once with the bundled internal CA:

```bash
# 1. Create the wall-vault internal CA (one time, lives in ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Issue a host certificate for THIS machine
#    SANs include hostname, localhost, 127.0.0.1, and any LAN IP detected
wall-vault cert issue $(hostname)

# 3. Trust the CA in the local OS keychain
wall-vault cert install-trust

# 4. Switch the listeners to TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

For another machine on your LAN: copy `~/.wall-vault/ca.crt` over and run `wall-vault cert install-trust --ca <path>` there. Once the CA is trusted everywhere, every machine on the network can reach the proxy over `https://<host>:56244` without certificate warnings.

If you'd rather stay on plain HTTP, leave the config as-is and replace `https://` with `http://` in the client snippets below. Both schemes work; the difference is which port answers a TLS handshake.

**Loopback fallback.** Same-host clients that can't honour the wall-vault CA (notably OpenClaw's bundled Node runtime, which rewrites `NODE_EXTRA_CA_CERTS` on spawn) reach the proxy through a loopback-only plain-HTTP companion on `127.0.0.1:56245`. wall-vault enables it automatically when TLS is on.

---

## Connecting clients

Point any AI client at `https://<host>:56244` (or `http://...` if TLS is off). The proxy answers four shapes:

| Format | Path | Example clients |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, custom scripts, most LLM apps |
| Ollama-native | `/api/chat` | Ollama clients passing through |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

When upstream Anthropic credits run out, dispatch falls back to whichever providers you set in `fallback_services` for this client. To opt in to non-Claude fallback explicitly:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Empty default makes dispatch return an error so misrouting surfaces immediately.)

### Cursor / Continue

In Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # or any model wall-vault knows about
```

Continue (`config.json`):

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

### OpenClaw

OpenClaw is a TUI agent framework that wall-vault was originally built to serve. The dashboard's **Add Agent** modal sets agent type to `openclaw` (or `nanoclaw`); wall-vault then writes `~/.openclaw/openclaw.json` directly, including provider URLs, the vault token, and model entries:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / scripts

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Configuration

`wall-vault setup` writes either `./wall-vault.yaml` or `~/.wall-vault/config.yaml`. Edit by hand for fields the wizard doesn't ask about.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # default: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: client token
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # loopback-only HTTP companion when TLS is on
  ollama_keep_alive: "30m"       # "-1" never unload, "0" unload immediately
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in real backend SSE passthrough
  anthropic_fallback_model: ""   # opt-in non-Claude rewrite on anthropic dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM key encryption password
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # plain-HTTP listener serving only ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # shell command (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variables

Every YAML field has an env override that wins over the file. Common ones:

| Variable | Description |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Language and theme |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Proxy listen address |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Vault listen address |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Distributed-mode endpoints |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Vault credentials |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API keys (comma-separated for multiple) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Loopback HTTP companion (`0` to disable) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA bootstrap listener (`0` to disable) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama tuning |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Local backend overrides |
| `WV_TOKEN_SENTINEL_FALLBACK` | Loopback "proxy-managed" sentinel substitution |
| `WV_OAI_STREAM_FORWARD` | OpenAI-compat real backend SSE passthrough |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Opt-in non-Claude rewrite on anthropic |

---

## Modes

### Standalone (default)

Vault and proxy run in the same process. Best for a single host that hosts both the keys and the agent. Listens on loopback only by default.

```bash
wall-vault start    # runs both
```

### Distributed

The vault runs on one host (the **vault host**) and stores all keys; multiple proxies on other hosts each authenticate with a per-client token. Useful when several machines need the same keys without copying them around.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Each proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

The dashboard's **Add Client** modal mints a token, registers an agent type, and the proxy picks up its config over SSE without restart.

---

## Plugin yaml (drop-in backend)

Any OpenAI-compatible backend can be added as a yaml under `~/.wall-vault/services/`. wall-vault picks it up at start, registers it as a routable service, and dispatch + the OAI-compat detection set + the Gemini-stream bridge all see it without code changes.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp
name: llama.cpp
enabled: true
default_url: http://localhost:8080
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models
auth:
  type: none
request_format: openai
inline_no_think_for_qwen3: false   # opt in if your backend strips the marker
```

Hub topology (one wall-vault fronts another) is supported via `tls_internal_ca: true`, `auth.type: bearer`, and `preserve_model_id: true`.

---

## Build from source

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-compile for the whole supported set:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Versions follow `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` in the Makefile sets the prefix.

### Project layout

```
wall-vault/
├── main.go                     # CLI dispatch (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # interactive setup wizard
│   └── cert/                   # internal CA + per-host TLS certificate issuer
├── internal/
│   ├── config/                 # YAML + env loader, plugin loader
│   ├── proxy/                  # request dispatch, key rotation, format converters
│   ├── vault/                  # AES-GCM store, dashboard, SSE broker
│   ├── doctor/                 # health probe + auto-fix
│   ├── hooks/                  # shell-command event triggers
│   └── i18n/                   # 17-language UI strings
├── configs/services/           # bundled plugin yamls (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API reference, 16 locale variants
```

---

## Documentation

- [User manual](docs/MANUAL.en.md) — installation, dashboard, agents, troubleshooting
- [API reference](docs/API.en.md) — every endpoint with request/response shapes
- [CHANGELOG](CHANGELOG.md)

---

## Tech stack

- Go 1.25, single static binary
- [templ](https://templ.guide) for server-rendered dashboard, [HTMX](https://htmx.org) for partial updates
- AES-GCM (PBKDF2-derived key) for at-rest key encryption
- Server-Sent Events for live config sync between vault and proxies
- Self-signed internal CA + per-host certs (no public DNS / Let's Encrypt required)

## License

GPL-3.0. See [LICENSE](LICENSE).

## Contributing

Pull requests welcome. See [CONTRIBUTING.md](CONTRIBUTING.md). For larger changes please open an issue first to discuss the design.
