# wall-vault API 手册

本文档详细描述了 wall-vault 的所有 HTTP API 端点。

---

## 目录

- [认证](#认证)
- [代理 API (:56244)](#代理-api-56244)
  - [健康检查](#get-health)
  - [状态查询](#get-status)
  - [模型列表](#get-apimodels)
  - [模型切换](#put-apiconfigmodel)
  - [思考模式](#put-apiconfigthink-mode)
  - [配置刷新](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini 流式传输](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI 兼容 API](#post-v1chatcompletions)
- [密钥金库 API (:56243)](#密钥金库-api-56243)
  - [公开 API](#公开-api无需认证)
  - [SSE 事件流](#get-apievents)
  - [代理专用 API](#代理专用-api客户端令牌)
  - [管理员 API — 密钥](#管理员-api--api-密钥)
  - [管理员 API — 客户端](#管理员-api--客户端)
  - [管理员 API — 服务](#管理员-api--服务)
  - [管理员 API — 模型列表](#管理员-api--模型列表)
  - [管理员 API — 代理状态](#管理员-api--代理状态)
- [SSE 事件类型](#sse-事件类型)
- [提供商·模型路由](#提供商模型路由)
- [数据模式](#数据模式)
- [错误响应](#错误响应)
- [cURL 示例集](#curl-示例集)

---

## 认证

| 范围 | 方式 | 请求头 |
|------|------|--------|
| 管理员 API | Bearer 令牌 | `Authorization: Bearer <admin_token>` |
| 代理 → 金库 | Bearer 令牌 | `Authorization: Bearer <client_token>` |
| 代理 API | 无（本地） | — |

当 `admin_token` 未设置（空字符串）时，所有管理员 API 无需认证即可访问。

### 安全策略

- **速率限制**: 管理员 API 认证失败超过 10 次/15 分钟时，临时封禁该 IP（`429 Too Many Requests`）
- **IP 白名单**: 仅允许在代理（`Client`）的 `ip_whitelist` 字段中注册的 IP/CIDR 访问 `/api/keys`。空数组表示允许所有。
- **theme·lang 保护**: `/admin/theme`、`/admin/lang` 也需要管理员令牌认证

---

## 代理 API (:56244)

代理运行的服务器。默认端口 `56244`。

---

### `GET /health`

健康检查。始终返回 200 OK。

**响应示例:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

代理状态详细查询。

**响应示例:**
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

| 字段 | 类型 | 说明 |
|------|------|------|
| `service` | string | 当前默认服务 |
| `model` | string | 当前默认模型 |
| `sse` | bool | 金库 SSE 连接状态 |
| `filter` | string | 工具过滤模式 |
| `services` | []string | 已启用的服务列表 |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

查询可用模型列表。使用 TTL 缓存（默认 10 分钟）。

**查询参数:**

| 参数 | 说明 | 示例 |
|------|------|------|
| `service` | 服务过滤器 | `?service=google` |
| `q` | 模型 ID/名称搜索 | `?q=gemini` |

**响应示例:**
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

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 模型 ID |
| `name` | string | 模型显示名称 |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` 等 |
| `context_length` | int | 上下文窗口大小 |
| `free` | bool | 是否为免费模型（OpenRouter） |

---

### `PUT /api/config/model`

切换当前服务·模型。

**请求体:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**响应:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **注意:** 在 distributed 模式下，建议使用金库的 `PUT /admin/clients/{id}` 代替此 API。金库中的更改会通过 SSE 在 1-3 秒内自动同步。

---

### `PUT /api/config/think-mode`

思考模式切换（no-op，预留给未来扩展）。

**响应:**
```json
{"status": "ok"}
```

---

### `POST /reload`

立即从金库重新同步客户端配置和密钥。

**响应:**
```json
{"status": "reloading"}
```

重新同步以异步方式执行，收到响应后 1-2 秒内完成。

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API 代理（非流式）。

**路径参数:**
- `{model}`: 模型 ID。带有 `gemini-` 前缀时自动选择 Google 服务。

**请求体:** [Gemini generateContent 请求格式](https://ai.google.dev/api/generate-content)

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

**响应体:** Gemini generateContent 响应格式

**工具过滤:** 设置 `tool_filter: strip_all` 时，请求中的 `tools` 数组会被自动移除。

**回退链:** 指定服务失败 → 按配置的服务顺序回退 → Ollama（最终兜底）。

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API 流式代理。请求格式与非流式相同。响应为 SSE 流:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI 兼容 API。内部会转换为 Gemini 格式后处理。

**请求体:**
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

**`model` 字段的提供商前缀支持（OpenClaw 3.11+）:**

| 模型示例 | 路由 |
|---------|------|
| `gemini-2.5-flash` | 当前配置的服务 |
| `google/gemini-2.5-pro` | 直连 Google |
| `openai/gpt-4o` | 直连 OpenAI |
| `anthropic/claude-opus-4-6` | 经由 OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | 直连 OpenRouter |
| `wall-vault/gemini-2.5-flash` | 自动检测 → Google |
| `wall-vault/claude-opus-4-6` | 自动检测 → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | 自动检测 → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter（免费 1M context） |
| `moonshot/kimi-k2.5` | 经由 OpenRouter |
| `opencode-go/model` | 经由 OpenRouter |
| `kimi-k2.5:cloud` | `:cloud` 后缀 → OpenRouter |

详情请参阅 [提供商·模型路由](#提供商模型路由)。

**响应体:**
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

> **模型控制 token 自动移除:** 如果响应中包含 GLM-5 / DeepSeek / ChatML 分隔符（`<|im_start|>`、`[gMASK]`、`[sop]` 等），会被自动移除。

---

## 密钥金库 API (:56243)

密钥金库运行的服务器。默认端口 `56243`。

---

### 公开 API（无需认证）

#### `GET /`

Web 仪表盘 UI。通过浏览器访问。

---

#### `GET /api/status`

金库状态查询。

**响应示例:**
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

已注册客户端列表（仅公开信息，不含令牌）。

---

### `GET /api/events`

SSE（Server-Sent Events）实时事件流。

**请求头:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**连接后立即接收:**
```
data: {"type":"connected","clients":2}
```

**事件示例:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

详细事件类型请参阅 [SSE 事件类型](#sse-事件类型)。

---

### 代理专用 API（客户端令牌）

需要 `Authorization: Bearer <client_token>` 请求头。也可使用管理员令牌进行认证。

#### `GET /api/keys`

提供给代理的已解密 API 密钥列表。

**查询参数:**

| 参数 | 说明 |
|------|------|
| `service` | 服务过滤器（例: `?service=google`） |

**响应示例:**
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

> **安全提示:** 返回明文密钥。根据客户端的 `allowed_services` 设置，仅返回允许的服务密钥。

---

#### `GET /api/services`

查询代理使用的服务列表。返回 `proxy_enabled=true` 的服务 ID 数组。

**响应示例:**
```json
["google", "ollama"]
```

空数组表示代理可以不受限制地使用所有服务。

---

#### `POST /api/heartbeat`

发送代理状态（每 20 秒自动执行）。

**请求体:**
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

| 字段 | 类型 | 说明 |
|------|------|------|
| `client_id` | string | 客户端 ID |
| `version` | string | 代理版本（含构建时间戳，例: `v0.1.6.20260314.231308`） |
| `service` | string | 当前服务 |
| `model` | string | 当前模型 |
| `sse_connected` | bool | SSE 连接状态 |
| `host` | string | 主机名 |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**响应:**
```json
{"status": "ok"}
```

---

### 管理员 API — API 密钥

需要 `Authorization: Bearer <admin_token>` 请求头。

#### `GET /admin/keys`

所有已注册 API 密钥列表（不含明文密钥）。

**响应示例:**
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

| 字段 | 类型 | 说明 |
|------|------|------|
| `today_usage` | int | 今日成功请求 token 数（不含 429/402/582 错误） |
| `today_attempts` | int | 今日 API 调用总次数（含成功 + rate-limited） |
| `available` | bool | 是否可用（无冷却·未达上限） |
| `usage_pct` | int | 日限额使用百分比（`daily_limit=0` 时为 0） |
| `cooldown_until` | RFC3339 | 冷却结束时间（零值表示无冷却） |
| `last_error` | int | 最后一次 HTTP 错误码 |

---

#### `POST /admin/keys`

注册新 API 密钥。注册后立即广播 SSE `key_added` 事件。

**请求体:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| 自定义 |
| `key` | ✅ | API 密钥明文 |
| `label` | — | 识别标签 |
| `daily_limit` | — | 每日使用上限（0 = 无限制） |

---

#### `DELETE /admin/keys/{id}`

删除 API 密钥。删除后广播 SSE `key_deleted` 事件。

**响应:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

重置所有密钥的每日使用量。广播 SSE `usage_reset` 事件。

**响应:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### 管理员 API — 客户端

#### `GET /admin/clients`

所有客户端列表（含令牌）。

---

#### `POST /admin/clients`

注册新客户端。

**请求体:**
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

| 字段 | 必填 | 说明 |
|------|------|------|
| `id` | ✅ | 客户端唯一 ID |
| `name` | — | 显示名称 |
| `token` | — | 认证令牌（省略时自动生成） |
| `default_service` | — | 默认服务 |
| `default_model` | — | 默认模型 |
| `allowed_services` | — | 允许的服务列表（空数组 = 全部允许） |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | 代理工作目录 |
| `description` | — | 代理描述 |
| `ip_whitelist` | — | 允许的 IP 列表（空数组 = 全部允许，支持 CIDR） |
| `enabled` | — | 是否启用（默认 `true`） |

---

#### `GET /admin/clients/{id}`

查询特定客户端（含令牌）。

---

#### `PUT /admin/clients/{id}`

修改客户端配置。**SSE `config_change` 广播 → 1-3 秒内同步到代理。**

**请求体（仅需变更的字段）:**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**响应:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

删除客户端。

---

### 管理员 API — 服务

#### `GET /admin/services`

已注册服务列表。

**响应示例:**
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

默认提供 8 个服务: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

添加自定义服务。添加后广播 SSE `service_changed` 事件 → **仪表盘下拉菜单即时更新**。

**请求体:**
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

更新服务配置。更改后广播 SSE `service_changed` 事件。

**请求体:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

删除自定义服务。删除后广播 SSE `service_changed` 事件。

尝试删除默认服务（`custom: false`）时:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### 管理员 API — 模型列表

#### `GET /admin/models`

按服务查询模型列表。使用 TTL 缓存（10 分钟）。

**查询参数:**

| 参数 | 说明 | 示例 |
|------|------|------|
| `service` | 服务过滤器 | `?service=google` |
| `q` | 模型搜索 | `?q=gemini` |

**各服务模型查询方式:**

| 服务 | 方式 | 数量 |
|------|------|------|
| `google` | 固定列表 | 8 个（含 embedding） |
| `openai` | 固定列表 | 9 个 |
| `anthropic` | 固定列表 | 6 个 |
| `github-copilot` | 固定列表 | 6 个 |
| `openrouter` | API 动态查询（失败时回退到精选 14 个） | 340+ 个 |
| `ollama` | 本地服务器动态查询（无响应时推荐 7 个） | 可变 |
| `lmstudio` | 本地服务器动态查询 | 可变 |
| `vllm` | 本地服务器动态查询 | 可变 |
| 自定义 | OpenAI 兼容 `/v1/models` | 可变 |

**OpenRouter 回退模型列表（API 无响应时）:**

| 模型 | 备注 |
|------|------|
| `openrouter/hunter-alpha` | 免费，1M context |
| `openrouter/healer-alpha` | 免费，omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### 管理员 API — 代理状态

#### `GET /admin/proxies`

所有已连接代理的最新 Heartbeat 状态。

---

## SSE 事件类型

金库 `/api/events` 流中接收的事件:

| `type` | 触发条件 | `data` 内容 | 仪表盘响应 |
|--------|---------|-------------|-----------|
| `connected` | SSE 连接后立即 | `{"clients": N}` | — |
| `config_change` | 客户端配置变更 | `{"client_id","service","model"}` | 代理卡片模型下拉菜单更新 |
| `key_added` | 新 API 密钥注册 | `{"service": "google"}` | 模型下拉菜单更新 |
| `key_deleted` | API 密钥删除 | `{"service": "google"}` | 模型下拉菜单更新 |
| `service_changed` | 服务添加/修改/删除 | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | 服务 select + 模型下拉菜单即时更新; 代理的 dispatch 服务列表实时更新 |
| `usage_update` | 收到代理 heartbeat 时（每 20 秒） | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | 密钥使用量进度条·数值即时更新, 冷却倒计时开始。无需 fetch，直接使用 SSE 数据。进度条使用 share-of-total 缩放（无限制密钥）。 |
| `usage_reset` | 每日使用量重置 | `{"time": "RFC3339"}` | 页面刷新 |

**代理接收的事件处理:**

```
config_change 接收
  → client_id 与自身匹配时
    → 立即更新 service, model
    → hooksMgr.Fire(EventModelChanged)
```

---

## 提供商·模型路由

在 `/v1/chat/completions` 的 `model` 字段中指定 `provider/model` 格式即可自动路由（兼容 OpenClaw 3.11）。

### 前缀路由规则

| 前缀 | 路由目标 | 示例 |
|------|---------|------|
| `google/` | 直连 Google | `google/gemini-2.5-pro` |
| `openai/` | 直连 OpenAI | `openai/gpt-4o` |
| `anthropic/` | 经由 OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | 直连 Ollama | `ollama/qwen3.5:35b` |
| `custom/` | 递归重解析（去除 `custom/` 后重新路由） | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter（保留 bare 路径） | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter（保留 full path） | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter（full path） | `deepseek/deepseek-r1` |

### `wall-vault/` 前缀自动检测

wall-vault 自有前缀，从模型 ID 自动判断服务。

| 模型 ID 模式 | 路由 |
|-------------|------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter（Anthropic 路径） |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（免费 1M ctx） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| 其他 | OpenRouter |

### `:cloud` 后缀处理

Ollama 标签格式的 `:cloud` 后缀会被自动去除后路由到 OpenRouter。

```
kimi-k2.5:cloud  →  OpenRouter, 模型 ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, 模型 ID: glm-5
```

### OpenClaw openclaw.json 集成示例

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

点击代理卡片上的 **🐾 按钮**，可将该代理的配置代码片段自动复制到剪贴板。

---

## 数据模式

### APIKey

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | UUID 格式唯一 ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| 自定义 |
| `encrypted_key` | string | AES-GCM 加密密钥（Base64） |
| `label` | string | 识别标签 |
| `today_usage` | int | 今日成功请求 token 数（不含 429/402/582 错误） |
| `today_attempts` | int | 今日 API 调用总次数（成功 + rate-limited; 午夜重置） |
| `daily_limit` | int | 每日上限（0 = 无限制） |
| `cooldown_until` | time.Time | 冷却结束时间 |
| `last_error` | int | 最后一次 HTTP 错误码 |
| `created_at` | time.Time | 注册时间 |

**冷却策略:**

| HTTP 错误 | 冷却时间 |
|-----------|---------|
| 429 (Too Many Requests) | 30 分钟 |
| 402 (Payment Required) | 24 小时 |
| 400 / 401 / 403 | 24 小时 |
| 582 (Gateway Overload) | 5 分钟 |
| 网络错误 | 10 分钟 |

> **429·402·582**: 设置冷却 + `today_attempts` 增加。`today_usage` 不变（仅统计成功 token）。
> **Ollama（本地服务）**: `callOllama` 使用 `Timeout: 0`（无限制）的专用 HTTP 客户端。大型模型推理可能需要数十秒到数分钟，因此不应用默认的 60 秒超时。

### Client

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 客户端唯一 ID |
| `name` | string | 显示名称 |
| `token` | string | 认证令牌 |
| `default_service` | string | 默认服务 |
| `default_model` | string | 默认模型（可使用 `provider/model` 格式） |
| `allowed_services` | []string | 允许的服务（空数组 = 全部） |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | 代理工作目录 |
| `description` | string | 描述 |
| `ip_whitelist` | []string | 允许的 IP 列表（支持 CIDR） |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false` 时访问 `/api/keys` 返回 `403` |
| `created_at` | time.Time | 注册时间 |

### ServiceConfig

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 服务唯一 ID |
| `name` | string | 显示名称 |
| `local_url` | string | 本地服务器 URL（Ollama/LMStudio/vLLM/自定义） |
| `enabled` | bool | 是否启用 |
| `custom` | bool | 是否为用户添加的服务 |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| 字段 | 类型 | 说明 |
|------|------|------|
| `client_id` | string | 客户端 ID |
| `version` | string | 代理版本（例: `v0.1.6.20260314.231308`） |
| `service` | string | 当前服务 |
| `model` | string | 当前模型 |
| `sse_connected` | bool | SSE 连接状态 |
| `host` | string | 主机名 |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | 最后更新时间 |
| `vault.today_usage` | int | 今日 token 使用量 |
| `vault.daily_limit` | int | 每日上限 |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## 错误响应

```json
{"error": "오류 메시지"}
```

| 状态码 | 含义 |
|-------|------|
| 200 | 成功 |
| 400 | 请求无效 |
| 401 | 认证失败 |
| 403 | 访问被拒绝（客户端已禁用、IP 被封禁） |
| 404 | 资源不存在 |
| 405 | 方法不允许 |
| 429 | 速率限制超出 |
| 500 | 服务器内部错误 |
| 502 | 上游 API 错误（所有回退均失败） |

---

## cURL 示例集

```bash
# ─── 代理 ─────────────────────────────────────────────────────────────────────

# 健康检查
curl http://localhost:56244/health

# 状态查询
curl http://localhost:56244/status

# 模型列表（全部）
curl http://localhost:56244/api/models

# 仅 Google 模型
curl "http://localhost:56244/api/models?service=google"

# 搜索免费模型
curl "http://localhost:56244/api/models?q=alpha"

# 切换模型（本地）
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# 配置刷新
curl -X POST http://localhost:56244/reload

# 直接调用 Gemini API
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI 兼容（默认模型）
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model 格式
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# 使用免费 1M context 模型
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── 密钥金库（公开） ─────────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── 密钥金库（管理员） ───────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# 密钥列表
curl -H "$ADMIN" http://localhost:56243/admin/keys

# 添加 Google 密钥
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# 添加 OpenAI 密钥
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# 添加 OpenRouter 密钥
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# 删除密钥（SSE key_deleted 广播）
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# 重置每日使用量
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# 客户端列表
curl -H "$ADMIN" http://localhost:56243/admin/clients

# 添加客户端（OpenClaw）
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# 切换客户端模型（SSE 即时同步）
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# 禁用客户端
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# 删除客户端
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# 服务列表
curl -H "$ADMIN" http://localhost:56243/admin/services

# 设置 Ollama 本地 URL（SSE service_changed 广播）
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# 启用 OpenAI 服务
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# 添加自定义服务（SSE service_changed 广播）
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# 删除自定义服务
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# 查询模型列表
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# 代理状态（heartbeat）
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── 分布式模式 — 代理 → 金库 ─────────────────────────────────────────────────

# 查询已解密密钥
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# 发送 Heartbeat
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## 中间件

自动应用于所有请求:

| 中间件 | 功能 |
|-------|------|
| **Logger** | `[method] path status latencyms` 格式日志 |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | panic 恢复，返回 500 响应 |

---

*最后更新: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
