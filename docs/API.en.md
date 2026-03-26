# wall-vault API Manual

This document describes all HTTP API endpoints of wall-vault in detail.

---

## Table of Contents

- [Authentication](#authentication)
- [Proxy API (:56244)](#proxy-api-56244)
  - [Health Check](#get-health)
  - [Status](#get-status)
  - [Model List](#get-apimodels)
  - [Change Model](#put-apiconfigmodel)
  - [Think Mode](#put-apiconfigthink-mode)
  - [Reload Config](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini Streaming](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI-Compatible API](#post-v1chatcompletions)
- [Key Vault API (:56243)](#key-vault-api-56243)
  - [Public API](#public-api-no-authentication-required)
  - [SSE Event Stream](#get-apievents)
  - [Proxy-Only API](#proxy-only-api-client-token)
  - [Admin API — Keys](#admin-api--api-keys)
  - [Admin API — Clients](#admin-api--clients)
  - [Admin API — Services](#admin-api--services)
  - [Admin API — Model List](#admin-api--model-list)
  - [Admin API — Proxy Status](#admin-api--proxy-status)
- [SSE Event Types](#sse-event-types)
- [Provider/Model Routing](#providermodel-routing)
- [Data Schema](#data-schema)
- [Error Responses](#error-responses)
- [cURL Examples](#curl-examples)

---

## Authentication

| Scope | Method | Header |
|-------|--------|--------|
| Admin API | Bearer token | `Authorization: Bearer <admin_token>` |
| Proxy → Vault | Bearer token | `Authorization: Bearer <client_token>` |
| Proxy API | None (local) | — |

If `admin_token` is not set (empty string), all admin APIs are accessible without authentication.

### Security Policy

- **Rate Limiting**: If admin API authentication fails more than 10 times within 15 minutes, the IP is temporarily blocked (`429 Too Many Requests`)
- **IP Whitelist**: Only IPs/CIDRs registered in the agent (`Client`) `ip_whitelist` field are allowed to access `/api/keys`. If the array is empty, all IPs are allowed.
- **theme/lang protection**: `/admin/theme` and `/admin/lang` also require admin token authentication

---

## Proxy API (:56244)

The server where the proxy runs. Default port `56244`.

---

### `GET /health`

Health check. Always returns 200 OK.

**Response example:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Detailed proxy status.

**Response example:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse": true,
  "filter": "strip_all",
  "services": ["google", "openrouter", "ollama"],
  "mode": "distributed"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `service` | string | Current default service |
| `model` | string | Current default model |
| `sse` | bool | Whether vault SSE is connected |
| `filter` | string | Tool filter mode |
| `services` | []string | List of active services |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

List available models. Uses TTL cache (default 10 minutes).

**Query parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `service` | Service filter | `?service=google` |
| `q` | Search by model ID/name | `?q=gemini` |

**Response example:**
```json
{
  "models": [
    {
      "id": "gemini-2.5-pro",
      "name": "Gemini 2.5 Pro",
      "service": "google",
      "context_length": 1048576,
      "free": false
    },
    {
      "id": "openrouter/hunter-alpha",
      "name": "Hunter Alpha (1M ctx, free)",
      "service": "openrouter",
      "context_length": 1048576,
      "free": true
    }
  ],
  "count": 2
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Model ID |
| `name` | string | Model display name |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` etc. |
| `context_length` | int | Context window size |
| `free` | bool | Whether it's a free model (OpenRouter) |

---

### `PUT /api/config/model`

Change current service and model.

**Request body:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Response:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Note:** In distributed mode, it is recommended to use the vault's `PUT /admin/clients/{id}` instead of this API. Vault changes are automatically propagated via SSE within 1–3 seconds.

---

### `PUT /api/config/think-mode`

Toggle think mode (no-op, reserved for future expansion).

**Response:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Immediately re-sync client settings and keys from the vault.

**Response:**
```json
{"status": "reloading"}
```

Re-sync runs asynchronously and completes within 1–2 seconds after the response is received.

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API proxy (non-streaming).

**Path parameter:**
- `{model}`: Model ID. If it has a `gemini-` prefix, Google service is automatically selected.

**Request body:** [Gemini generateContent request format](https://ai.google.dev/api/generate-content)

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "안녕하세요"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024
  }
}
```

**Response body:** Gemini generateContent response format

**Tool filter:** When `tool_filter: strip_all` is set, the `tools` array in the request is automatically removed.

**Fallback chain:** If the designated service fails → fallback in configured service order → Ollama (final).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API streaming proxy. Request format is the same as non-streaming. Response is an SSE stream:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI-compatible API. Internally converts to Gemini format for processing.

**Request body:**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    {"role": "system", "content": "당신은 도움이 되는 어시스턴트입니다."},
    {"role": "user", "content": "안녕하세요"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**Provider prefix support in the `model` field (OpenClaw 3.11+):**

| Model example | Routing |
|---------------|---------|
| `gemini-2.5-flash` | Current configured service |
| `google/gemini-2.5-pro` | Google direct |
| `openai/gpt-4o` | OpenAI direct |
| `anthropic/claude-opus-4-6` | Via OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter direct |
| `wall-vault/gemini-2.5-flash` | Auto-detect → Google |
| `wall-vault/claude-opus-4-6` | Auto-detect → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Auto-detect → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (free 1M context) |
| `moonshot/kimi-k2.5` | Via OpenRouter |
| `opencode-go/model` | Via OpenRouter |
| `kimi-k2.5:cloud` | `:cloud` suffix → OpenRouter |

For details, see [Provider/Model Routing](#providermodel-routing).

**Response body:**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "안녕하세요! 무엇을 도와드릴까요?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```

> **Automatic removal of model control tokens:** If the response contains GLM-5 / DeepSeek / ChatML delimiters (`<|im_start|>`, `[gMASK]`, `[sop]`, etc.), they are automatically stripped.

---

## Key Vault API (:56243)

The server where the key vault runs. Default port `56243`.

---

### Public API (No Authentication Required)

#### `GET /`

Web dashboard UI. Access via browser.

---

#### `GET /api/status`

Vault status.

**Response example:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

---

#### `GET /api/clients`

List of registered clients (public information only, tokens excluded).

---

### `GET /api/events`

SSE (Server-Sent Events) real-time event stream.

**Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Received immediately upon connection:**
```
data: {"type":"connected","clients":2}
```

**Event examples:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

For detailed event types, see [SSE Event Types](#sse-event-types).

---

### Proxy-Only API (Client Token)

Requires `Authorization: Bearer <client_token>` header. Admin tokens are also accepted.

#### `GET /api/keys`

Decrypted API key list provided to the proxy.

**Query parameters:**

| Parameter | Description |
|-----------|-------------|
| `service` | Service filter (e.g., `?service=google`) |

**Response example:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000,
    "today_usage": 42,
    "today_attempts": 45
  }
]
```

> **Security:** Returns plaintext keys. Only keys for services allowed by the client's `allowed_services` setting are returned.

---

#### `GET /api/services`

List of services for the proxy to use. Returns an array of service IDs where `proxy_enabled=true`.

**Response example:**
```json
["google", "ollama"]
```

If the array is empty, the proxy uses all services without restriction.

---

#### `POST /api/heartbeat`

Send proxy status (automatically executed every 20 seconds).

**Request body:**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-host",
  "avatar": "data:image/png;base64,...",
  "key_usage":     {"key-abc123": 42, "key-def456": 0},
  "key_attempts":  {"key-abc123": 45, "key-def456": 3},
  "key_cooldowns": {"key-abc123": "2026-03-15T14:30:00Z"}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client_id` | string | Client ID |
| `version` | string | Proxy version (includes build timestamp, e.g. `v0.1.6.20260314.231308`) |
| `service` | string | Current service |
| `model` | string | Current model |
| `sse_connected` | bool | Whether SSE is connected |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Response:**
```json
{"status": "ok"}
```

---

### Admin API — API Keys

Requires `Authorization: Bearer <admin_token>` header.

#### `GET /admin/keys`

List all registered API keys (plaintext keys excluded).

**Response example:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "today_attempts": 45,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `today_usage` | int | Successful request tokens today (does not include 429/402/582 errors) |
| `today_attempts` | int | Total API calls today (success + rate-limited) |
| `available` | bool | Whether available without cooldown or limit |
| `usage_pct` | int | Usage percentage of daily limit (`daily_limit=0` → 0) |
| `cooldown_until` | RFC3339 | Cooldown end time (zero value means none) |
| `last_error` | int | Last HTTP error code |

---

#### `POST /admin/keys`

Register a new API key. An SSE `key_added` event is broadcast immediately upon registration.

**Request body:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| custom |
| `key` | ✅ | API key in plaintext |
| `label` | — | Identification label |
| `daily_limit` | — | Daily usage limit (0 = unlimited) |

---

#### `DELETE /admin/keys/{id}`

Delete an API key. An SSE `key_deleted` event is broadcast after deletion.

**Response:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Reset daily usage for all keys. SSE `usage_reset` event is broadcast.

**Response:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### Admin API — Clients

#### `GET /admin/clients`

List all clients (including tokens).

---

#### `POST /admin/clients`

Register a new client.

**Request body:**
```json
{
  "id": "my-bot",
  "name": "내 봇",
  "token": "my-secret-token",
  "default_service": "google",
  "default_model": "gemini-2.5-flash",
  "allowed_services": ["google", "openrouter"],
  "agent_type": "openclaw",
  "work_dir": "~/.openclaw",
  "description": "OpenClaw 에이전트",
  "ip_whitelist": ["10.0.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `id` | ✅ | Unique client ID |
| `name` | — | Display name |
| `token` | — | Authentication token (auto-generated if omitted) |
| `default_service` | — | Default service |
| `default_model` | — | Default model |
| `allowed_services` | — | Allowed service list (empty array = all allowed) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Agent working directory |
| `description` | — | Agent description |
| `ip_whitelist` | — | Allowed IP list (empty array = all allowed, CIDR supported) |
| `enabled` | — | Whether enabled (default `true`) |

---

#### `GET /admin/clients/{id}`

Get a specific client (including token).

---

#### `PUT /admin/clients/{id}`

Update client settings. **SSE `config_change` broadcast → reflected on proxy within 1–3 seconds.**

**Request body (only include fields to change):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Response:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Delete a client.

---

### Admin API — Services

#### `GET /admin/services`

List registered services.

**Response example:**
```json
[
  {"id": "google",      "name": "Google Gemini",   "enabled": true,  "custom": false},
  {"id": "openai",      "name": "OpenAI",          "enabled": true,  "custom": false},
  {"id": "anthropic",   "name": "Anthropic",       "enabled": false, "custom": false},
  {"id": "openrouter",  "name": "OpenRouter",      "enabled": true,  "custom": false},
  {"id": "ollama",      "name": "Ollama (Local)",  "enabled": true,  "custom": false,
   "local_url": "http://localhost:11434"},
  {"id": "lmstudio",    "name": "LM Studio",       "enabled": false, "custom": false},
  {"id": "vllm",        "name": "vLLM",            "enabled": false, "custom": false},
  {"id": "github-copilot","name":"GitHub Copilot", "enabled": false, "custom": false}
]
```

8 built-in services: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Add a custom service. SSE `service_changed` event is broadcast after addition → **dashboard dropdowns update immediately**.

**Request body:**
```json
{
  "id": "my-llm",
  "name": "사내 LLM 서버",
  "local_url": "http://10.0.0.50:8080",
  "enabled": true
}
```

---

#### `PUT /admin/services/{id}`

Update service settings. SSE `service_changed` event is broadcast after changes.

**Request body:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Delete a custom service. SSE `service_changed` event is broadcast after deletion.

Attempting to delete a built-in service (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### Admin API — Model List

#### `GET /admin/models`

List models by service. Uses TTL cache (10 minutes).

**Query parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `service` | Service filter | `?service=google` |
| `q` | Model search | `?q=gemini` |

**Model retrieval by service:**

| Service | Method | Count |
|---------|--------|-------|
| `google` | Static list | 8 (including embedding) |
| `openai` | Static list | 9 |
| `anthropic` | Static list | 6 |
| `github-copilot` | Static list | 6 |
| `openrouter` | Dynamic API query (falls back to 14 curated models on failure) | 340+ |
| `ollama` | Dynamic local server query (7 recommended if unresponsive) | Variable |
| `lmstudio` | Dynamic local server query | Variable |
| `vllm` | Dynamic local server query | Variable |
| Custom | OpenAI-compatible `/v1/models` | Variable |

**OpenRouter fallback models (when API is unresponsive):**

| Model | Notes |
|-------|-------|
| `openrouter/hunter-alpha` | Free, 1M context |
| `openrouter/healer-alpha` | Free, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### Admin API — Proxy Status

#### `GET /admin/proxies`

Last heartbeat status of all connected proxies.

---

## SSE Event Types

Events received from the vault `/api/events` stream:

| `type` | Trigger | `data` Content | Dashboard Reaction |
|--------|---------|----------------|-------------------|
| `connected` | Immediately on SSE connection | `{"clients": N}` | — |
| `config_change` | Client settings changed | `{"client_id","service","model"}` | Agent card model dropdown refreshed |
| `key_added` | New API key registered | `{"service": "google"}` | Model dropdown refreshed |
| `key_deleted` | API key deleted | `{"service": "google"}` | Model dropdown refreshed |
| `service_changed` | Service added/updated/deleted | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Service select + model dropdown refreshed immediately; proxy's dispatch service list updated in real-time |
| `usage_update` | On proxy heartbeat (every 20s) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Key usage bars and numbers updated instantly, cooldown countdown starts. SSE data used directly without fetch. Bars use share-of-total scaling (for unlimited keys). |
| `usage_reset` | Daily usage reset | `{"time": "RFC3339"}` | Page refresh |

**Event processing on the proxy side:**

```
config_change received
  → If client_id matches own ID
    → service, model updated immediately
    → hooksMgr.Fire(EventModelChanged)
```

---

## Provider/Model Routing

When specifying a `provider/model` format in the `model` field of `/v1/chat/completions`, automatic routing is applied (OpenClaw 3.11 compatible).

### Prefix Routing Rules

| Prefix | Routing Target | Example |
|--------|---------------|---------|
| `google/` | Google direct | `google/gemini-2.5-pro` |
| `openai/` | OpenAI direct | `openai/gpt-4o` |
| `anthropic/` | Via OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama direct | `ollama/qwen3.5:35b` |
| `custom/` | Recursive re-parse (strip `custom/` and re-route) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (bare path preserved) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (full path preserved) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### `wall-vault/` Prefix Auto-Detection

The wall-vault's own prefix automatically determines the service from the model ID.

| Model ID Pattern | Routing |
|-----------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic path) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (free 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Other | OpenRouter |

### `:cloud` Suffix Handling

The Ollama tag-style `:cloud` suffix is automatically removed and routed to OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, model ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, model ID: glm-5
```

### OpenClaw openclaw.json Integration Example

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
          { id: "wall-vault/hunter-alpha" },
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: {
        primary: "wall-vault/gemini-2.5-flash",
        fallbacks: ["wall-vault/hunter-alpha"]
      }
    }
  }
}
```

Click the **🐾 button** on an agent card to automatically copy the configuration snippet for that agent to your clipboard.

---

## Data Schema

### APIKey

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique ID in UUID format |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| custom |
| `encrypted_key` | string | AES-GCM encrypted key (Base64) |
| `label` | string | Identification label |
| `today_usage` | int | Successful request tokens today (does not include 429/402/582 errors) |
| `today_attempts` | int | Total API calls today (success + rate-limited; resets at midnight) |
| `daily_limit` | int | Daily limit (0 = unlimited) |
| `cooldown_until` | time.Time | Cooldown end time |
| `last_error` | int | Last HTTP error code |
| `created_at` | time.Time | Registration time |

**Cooldown policy:**

| HTTP Error | Cooldown |
|------------|----------|
| 429 (Too Many Requests) | 30 minutes |
| 402 (Payment Required) | 24 hours |
| 400 / 401 / 403 | 24 hours |
| 582 (Gateway Overload) | 5 minutes |
| Network error | 10 minutes |

> **429/402/582**: Cooldown is set + `today_attempts` is incremented. `today_usage` is unchanged (only successful tokens are counted).
> **Ollama (local service)**: `callOllama` uses a dedicated HTTP client with `Timeout: 0` (unlimited). Large model inference can take tens of seconds to minutes, so the default 60-second timeout is not applied.

### Client

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique client ID |
| `name` | string | Display name |
| `token` | string | Authentication token |
| `default_service` | string | Default service |
| `default_model` | string | Default model (can be in `provider/model` format) |
| `allowed_services` | []string | Allowed services (empty array = all) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Agent working directory |
| `description` | string | Description |
| `ip_whitelist` | []string | Allowed IP list (CIDR supported) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | If `false`, returns `403` when accessing `/api/keys` |
| `created_at` | time.Time | Registration time |

### ServiceConfig

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique service ID |
| `name` | string | Display name |
| `local_url` | string | Local server URL (Ollama/LMStudio/vLLM/custom) |
| `enabled` | bool | Whether enabled |
| `custom` | bool | Whether it's a user-added service |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Field | Type | Description |
|-------|------|-------------|
| `client_id` | string | Client ID |
| `version` | string | Proxy version (e.g. `v0.1.6.20260314.231308`) |
| `service` | string | Current service |
| `model` | string | Current model |
| `sse_connected` | bool | Whether SSE is connected |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Last update |
| `vault.today_usage` | int | Token usage today |
| `vault.daily_limit` | int | Daily limit |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Error Responses

```json
{"error": "오류 메시지"}
```

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request |
| 401 | Authentication failed |
| 403 | Access denied (disabled client, IP blocked) |
| 404 | Resource not found |
| 405 | Method not allowed |
| 429 | Rate limit exceeded |
| 500 | Internal server error |
| 502 | Upstream API error (all fallbacks failed) |

---

## cURL Examples

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Health check
curl http://localhost:56244/health

# Status
curl http://localhost:56244/status

# Model list (all)
curl http://localhost:56244/api/models

# Google models only
curl "http://localhost:56244/api/models?service=google"

# Search free models
curl "http://localhost:56244/api/models?q=alpha"

# Change model (local)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Reload config
curl -X POST http://localhost:56244/reload

# Direct Gemini API call
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI-compatible (default model)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model format
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Free 1M context model
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Key Vault (Public) ──────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── Key Vault (Admin) ───────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Key list
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Add Google key
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Add OpenAI key
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Add OpenRouter key
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Delete key (SSE key_deleted broadcast)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Reset daily usage
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# Client list
curl -H "$ADMIN" http://localhost:56243/admin/clients

# Add client (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Change client model (SSE instant update)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Disable client
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Delete client
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Service list
curl -H "$ADMIN" http://localhost:56243/admin/services

# Set Ollama local URL (SSE service_changed broadcast)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Enable OpenAI service
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Add custom service (SSE service_changed broadcast)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Delete custom service
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Model list
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# Proxy status (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── Distributed Mode — Proxy → Vault ────────────────────────────────────────

# Get decrypted keys
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Send heartbeat
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Automatically applied to all requests:

| Middleware | Function |
|-----------|----------|
| **Logger** | Logs in `[method] path status latencyms` format |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Panic recovery, returns 500 response |

---

*Last updated: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
