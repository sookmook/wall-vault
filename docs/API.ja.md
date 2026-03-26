# wall-vault API マニュアル

このドキュメントでは、wall-vault のすべての HTTP API エンドポイントを詳細に説明します。

---

## 目次

- [認証](#認証)
- [プロキシ API (:56244)](#プロキシ-api-56244)
  - [ヘルスチェック](#get-health)
  - [ステータス照会](#get-status)
  - [モデル一覧](#get-apimodels)
  - [モデル変更](#put-apiconfigmodel)
  - [思考モード](#put-apiconfigthink-mode)
  - [設定リロード](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini ストリーミング](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI 互換 API](#post-v1chatcompletions)
- [キー金庫 API (:56243)](#キー金庫-api-56243)
  - [公開 API](#公開-api-認証不要)
  - [SSE イベントストリーム](#get-apievents)
  - [プロキシ専用 API](#プロキシ専用-apiクライアントトークン)
  - [管理者 API — キー](#管理者-api--api-キー)
  - [管理者 API — クライアント](#管理者-api--クライアント)
  - [管理者 API — サービス](#管理者-api--サービス)
  - [管理者 API — モデル一覧](#管理者-api--モデル一覧)
  - [管理者 API — プロキシステータス](#管理者-api--プロキシステータス)
- [SSE イベントタイプ](#sse-イベントタイプ)
- [プロバイダ・モデルルーティング](#プロバイダモデルルーティング)
- [データスキーマ](#データスキーマ)
- [エラーレスポンス](#エラーレスポンス)
- [cURL サンプル集](#curl-サンプル集)

---

## 認証

| 領域 | 方法 | ヘッダー |
|------|------|---------|
| 管理者 API | Bearer トークン | `Authorization: Bearer <admin_token>` |
| プロキシ → 金庫 | Bearer トークン | `Authorization: Bearer <client_token>` |
| プロキシ API | なし（ローカル） | — |

`admin_token` が未設定（空文字列）の場合、すべての管理者 API は認証なしでアクセスできます。

### セキュリティポリシー

- **レート制限**: 管理者 API の認証失敗が 15 分間に 10 回を超えると、該当 IP を一時ブロック（`429 Too Many Requests`）
- **IP ホワイトリスト**: エージェント（`Client`）の `ip_whitelist` フィールドに登録された IP/CIDR のみ `/api/keys` へのアクセスを許可。空配列の場合はすべて許可。
- **theme・lang 保護**: `/admin/theme`、`/admin/lang` も管理者トークン認証が必要

---

## プロキシ API (:56244)

プロキシが実行されるサーバー。デフォルトポート `56244`。

---

### `GET /health`

ヘルスチェック。常に 200 OK を返します。

**レスポンス例:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

プロキシステータスの詳細照会。

**レスポンス例:**
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

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `service` | string | 現在のデフォルトサービス |
| `model` | string | 現在のデフォルトモデル |
| `sse` | bool | 金庫 SSE 接続状態 |
| `filter` | string | ツールフィルターモード |
| `services` | []string | 有効なサービス一覧 |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

利用可能なモデルの一覧照会。TTL キャッシュ（デフォルト 10 分）を使用。

**クエリパラメータ:**

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `service` | サービスフィルター | `?service=google` |
| `q` | モデル ID/名前検索 | `?q=gemini` |

**レスポンス例:**
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

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | string | モデル ID |
| `name` | string | モデル表示名 |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` 等 |
| `context_length` | int | コンテキストウィンドウサイズ |
| `free` | bool | 無料モデルかどうか（OpenRouter） |

---

### `PUT /api/config/model`

現在のサービス・モデルを変更。

**リクエストボディ:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**レスポンス:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **注意:** distributed モードでは、この API の代わりに金庫の `PUT /admin/clients/{id}` の使用を推奨します。金庫での変更は SSE を通じて 1〜3 秒以内に自動反映されます。

---

### `PUT /api/config/think-mode`

思考モードの切り替え（no-op、将来の拡張用）。

**レスポンス:**
```json
{"status": "ok"}
```

---

### `POST /reload`

金庫からクライアント設定・キーを即時再同期します。

**レスポンス:**
```json
{"status": "reloading"}
```

再同期は非同期で実行されるため、レスポンス受信後 1〜2 秒以内に完了します。

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API プロキシ（非ストリーミング）。

**パスパラメータ:**
- `{model}`: モデル ID。`gemini-` プレフィックスがある場合、自動的に Google サービスが選択されます。

**リクエストボディ:** [Gemini generateContent リクエスト形式](https://ai.google.dev/api/generate-content)

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

**レスポンスボディ:** Gemini generateContent レスポンス形式

**ツールフィルター:** `tool_filter: strip_all` 設定時、リクエストの `tools` 配列が自動的に除去されます。

**フォールバックチェーン:** 指定サービス失敗 → 設定されたサービス順にフォールバック → Ollama（最終手段）。

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API ストリーミングプロキシ。リクエスト形式は非ストリーミングと同一。レスポンスは SSE ストリーム:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI 互換 API。内部的に Gemini 形式に変換して処理します。

**リクエストボディ:**
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

**`model` フィールドのプロバイダプレフィックスサポート（OpenClaw 3.11+）:**

| モデル例 | ルーティング |
|---------|------------|
| `gemini-2.5-flash` | 現在設定のサービス |
| `google/gemini-2.5-pro` | Google 直接 |
| `openai/gpt-4o` | OpenAI 直接 |
| `anthropic/claude-opus-4-6` | OpenRouter 経由 |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter 直接 |
| `wall-vault/gemini-2.5-flash` | 自動検出 → Google |
| `wall-vault/claude-opus-4-6` | 自動検出 → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | 自動検出 → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter（無料 1M context） |
| `moonshot/kimi-k2.5` | OpenRouter 経由 |
| `opencode-go/model` | OpenRouter 経由 |
| `kimi-k2.5:cloud` | `:cloud` サフィックス → OpenRouter |

詳細は [プロバイダ・モデルルーティング](#プロバイダモデルルーティング) を参照。

**レスポンスボディ:**
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

> **モデル制御トークンの自動除去:** レスポンスに GLM-5 / DeepSeek / ChatML 区切り文字（`<|im_start|>`、`[gMASK]`、`[sop]` 等）が含まれている場合、自動的に除去されます。

---

## キー金庫 API (:56243)

キー金庫が実行されるサーバー。デフォルトポート `56243`。

---

### 公開 API（認証不要）

#### `GET /`

Web ダッシュボード UI。ブラウザでアクセス。

---

#### `GET /api/status`

金庫ステータス照会。

**レスポンス例:**
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

登録済みクライアント一覧（公開情報のみ、トークン除外）。

---

### `GET /api/events`

SSE（Server-Sent Events）リアルタイムイベントストリーム。

**ヘッダー:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**接続直後に受信:**
```
data: {"type":"connected","clients":2}
```

**イベント例:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

イベントタイプの詳細は [SSE イベントタイプ](#sse-イベントタイプ) を参照。

---

### プロキシ専用 API（クライアントトークン）

`Authorization: Bearer <client_token>` ヘッダーが必要。管理者トークンでも認証可能。

#### `GET /api/keys`

プロキシに提供する復号済み API キー一覧。

**クエリパラメータ:**

| パラメータ | 説明 |
|-----------|------|
| `service` | サービスフィルター（例: `?service=google`） |

**レスポンス例:**
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

> **セキュリティ:** 平文キーを返します。クライアントの `allowed_services` 設定に基づき、許可されたサービスのキーのみ返却されます。

---

#### `GET /api/services`

プロキシが使用するサービス一覧照会。`proxy_enabled=true` のサービス ID 配列を返却。

**レスポンス例:**
```json
["google", "ollama"]
```

空配列の場合、プロキシは制限なくすべてのサービスを使用します。

---

#### `POST /api/heartbeat`

プロキシステータス送信（20 秒ごとに自動実行）。

**リクエストボディ:**
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

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `client_id` | string | クライアント ID |
| `version` | string | プロキシバージョン（ビルドタイムスタンプ含む、例: `v0.1.6.20260314.231308`） |
| `service` | string | 現在のサービス |
| `model` | string | 現在のモデル |
| `sse_connected` | bool | SSE 接続状態 |
| `host` | string | ホスト名 |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**レスポンス:**
```json
{"status": "ok"}
```

---

### 管理者 API — API キー

`Authorization: Bearer <admin_token>` ヘッダーが必要。

#### `GET /admin/keys`

登録済みのすべての API キー一覧（平文キー除外）。

**レスポンス例:**
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

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `today_usage` | int | 本日の成功リクエストトークン数（429/402/582 エラーは含まない） |
| `today_attempts` | int | 本日の合計 API 呼び出し回数（成功 + rate-limited 含む） |
| `available` | bool | クールダウン・上限なしで使用可能かどうか |
| `usage_pct` | int | 日次上限に対する使用率 %（`daily_limit=0` の場合は 0） |
| `cooldown_until` | RFC3339 | クールダウン終了時刻（ゼロ値の場合はなし） |
| `last_error` | int | 最後の HTTP エラーコード |

---

#### `POST /admin/keys`

新しい API キーの登録。登録直後に SSE `key_added` イベントがブロードキャストされます。

**リクエストボディ:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| カスタム |
| `key` | ✅ | API キー平文 |
| `label` | — | 識別用ラベル |
| `daily_limit` | — | 日次使用上限（0 = 無制限） |

---

#### `DELETE /admin/keys/{id}`

API キー削除。削除後に SSE `key_deleted` イベントがブロードキャストされます。

**レスポンス:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

すべてのキーの日次使用量をリセット。SSE `usage_reset` イベントがブロードキャストされます。

**レスポンス:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### 管理者 API — クライアント

#### `GET /admin/clients`

すべてのクライアント一覧（トークン含む）。

---

#### `POST /admin/clients`

新しいクライアントの登録。

**リクエストボディ:**
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

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `id` | ✅ | クライアント一意 ID |
| `name` | — | 表示名 |
| `token` | — | 認証トークン（省略時は自動生成） |
| `default_service` | — | デフォルトサービス |
| `default_model` | — | デフォルトモデル |
| `allowed_services` | — | 許可サービス一覧（空配列 = すべて許可） |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | エージェント作業ディレクトリ |
| `description` | — | エージェントの説明 |
| `ip_whitelist` | — | 許可 IP 一覧（空配列 = すべて許可、CIDR サポート） |
| `enabled` | — | 有効化状態（デフォルト `true`） |

---

#### `GET /admin/clients/{id}`

特定クライアントの照会（トークン含む）。

---

#### `PUT /admin/clients/{id}`

クライアント設定の変更。**SSE `config_change` ブロードキャスト → プロキシに 1〜3 秒以内に反映。**

**リクエストボディ（変更するフィールドのみ）:**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**レスポンス:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

クライアント削除。

---

### 管理者 API — サービス

#### `GET /admin/services`

登録済みサービス一覧。

**レスポンス例:**
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

デフォルト提供サービス 8 種: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

カスタムサービスの追加。追加後に SSE `service_changed` イベントがブロードキャスト → **ダッシュボードのドロップダウンが即時更新**。

**リクエストボディ:**
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

サービス設定の更新。変更後に SSE `service_changed` イベントがブロードキャストされます。

**リクエストボディ:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

カスタムサービスの削除。削除後に SSE `service_changed` イベントがブロードキャストされます。

デフォルトサービス（`custom: false`）の削除を試みた場合:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### 管理者 API — モデル一覧

#### `GET /admin/models`

サービス別モデル一覧照会。TTL キャッシュ（10 分）使用。

**クエリパラメータ:**

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `service` | サービスフィルター | `?service=google` |
| `q` | モデル検索 | `?q=gemini` |

**サービス別モデル照会方式:**

| サービス | 方式 | 件数 |
|---------|------|------|
| `google` | 固定リスト | 8 件（embedding 含む） |
| `openai` | 固定リスト | 9 件 |
| `anthropic` | 固定リスト | 6 件 |
| `github-copilot` | 固定リスト | 6 件 |
| `openrouter` | API 動的照会（失敗時は curated フォールバック 14 件） | 340+ 件 |
| `ollama` | ローカルサーバー動的照会（未応答時は推奨 7 件） | 可変 |
| `lmstudio` | ローカルサーバー動的照会 | 可変 |
| `vllm` | ローカルサーバー動的照会 | 可変 |
| カスタム | OpenAI 互換 `/v1/models` | 可変 |

**OpenRouter フォールバックモデル一覧（API 未応答時）:**

| モデル | 特記事項 |
|-------|---------|
| `openrouter/hunter-alpha` | 無料、1M context |
| `openrouter/healer-alpha` | 無料、omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### 管理者 API — プロキシステータス

#### `GET /admin/proxies`

接続されたすべてのプロキシの最新 Heartbeat ステータス。

---

## SSE イベントタイプ

金庫 `/api/events` ストリームで受信されるイベント:

| `type` | 発生条件 | `data` 内容 | ダッシュボード反応 |
|--------|---------|-------------|------------------|
| `connected` | SSE 接続直後 | `{"clients": N}` | — |
| `config_change` | クライアント設定変更 | `{"client_id","service","model"}` | エージェントカードのモデルドロップダウン更新 |
| `key_added` | 新 API キー登録 | `{"service": "google"}` | モデルドロップダウン更新 |
| `key_deleted` | API キー削除 | `{"service": "google"}` | モデルドロップダウン更新 |
| `service_changed` | サービス追加/修正/削除 | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | サービス select + モデルドロップダウン即時更新; プロキシの dispatch サービスリストリアルタイム更新 |
| `usage_update` | プロキシ heartbeat 受信時（20 秒ごと） | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | キー使用量バー・数値の即時更新、クールダウンカウントダウン開始。fetch なしで SSE データを直接使用。バーは share-of-total スケーリング（無制限キー）。 |
| `usage_reset` | 日次使用量リセット | `{"time": "RFC3339"}` | ページリロード |

**プロキシが受信するイベント処理:**

```
config_change 受信
  → client_id が自身と一致する場合
    → service, model を即時更新
    → hooksMgr.Fire(EventModelChanged)
```

---

## プロバイダ・モデルルーティング

`/v1/chat/completions` の `model` フィールドに `provider/model` 形式を指定すると自動ルーティングされます（OpenClaw 3.11 互換）。

### プレフィックスルーティング規則

| プレフィックス | ルーティング先 | 例 |
|-------------|--------------|-----|
| `google/` | Google 直接 | `google/gemini-2.5-pro` |
| `openai/` | OpenAI 直接 | `openai/gpt-4o` |
| `anthropic/` | OpenRouter 経由 | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama 直接 | `ollama/qwen3.5:35b` |
| `custom/` | 再帰的再パース（`custom/` 除去後に再ルーティング） | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter（bare パス維持） | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter（full path 維持） | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter（full path） | `deepseek/deepseek-r1` |

### `wall-vault/` プレフィックス自動検出

wall-vault 独自のプレフィックスでモデル ID からサービスを自動判別します。

| モデル ID パターン | ルーティング |
|------------------|------------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter（Anthropic パス） |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（無料 1M ctx） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| その他 | OpenRouter |

### `:cloud` サフィックス処理

Ollama タグ形式の `:cloud` サフィックスは自動的に除去後、OpenRouter にルーティングされます。

```
kimi-k2.5:cloud  →  OpenRouter, モデル ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, モデル ID: glm-5
```

### OpenClaw openclaw.json 連携例

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

エージェントカードの **🐾 ボタン** をクリックすると、そのエージェント用の設定スニペットがクリップボードに自動コピーされます。

---

## データスキーマ

### APIKey

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | string | UUID 形式の一意 ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| カスタム |
| `encrypted_key` | string | AES-GCM 暗号化キー（Base64） |
| `label` | string | 識別用ラベル |
| `today_usage` | int | 本日の成功リクエストトークン数（429/402/582 エラー非含む） |
| `today_attempts` | int | 本日の合計 API 呼び出し回数（成功 + rate-limited; 深夜リセット） |
| `daily_limit` | int | 日次上限（0 = 無制限） |
| `cooldown_until` | time.Time | クールダウン終了時刻 |
| `last_error` | int | 最後の HTTP エラーコード |
| `created_at` | time.Time | 登録時刻 |

**クールダウンポリシー:**

| HTTP エラー | クールダウン |
|------------|------------|
| 429 (Too Many Requests) | 30 分 |
| 402 (Payment Required) | 24 時間 |
| 400 / 401 / 403 | 24 時間 |
| 582 (Gateway Overload) | 5 分 |
| ネットワークエラー | 10 分 |

> **429・402・582**: クールダウン設定 + `today_attempts` 増加。`today_usage` は変更なし（成功トークンのみ集計）。
> **Ollama（ローカルサービス）**: `callOllama` は `Timeout: 0`（無制限）専用 HTTP クライアントを使用します。大規模モデル推論は数十秒〜数分かかる場合があるため、デフォルトの 60 秒タイムアウトは適用されません。

### Client

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | string | クライアント一意 ID |
| `name` | string | 表示名 |
| `token` | string | 認証トークン |
| `default_service` | string | デフォルトサービス |
| `default_model` | string | デフォルトモデル（`provider/model` 形式可能） |
| `allowed_services` | []string | 許可サービス（空配列 = すべて） |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | エージェント作業ディレクトリ |
| `description` | string | 説明 |
| `ip_whitelist` | []string | 許可 IP 一覧（CIDR サポート） |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false` の場合 `/api/keys` アクセス時に `403` |
| `created_at` | time.Time | 登録時刻 |

### ServiceConfig

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | string | サービス一意 ID |
| `name` | string | 表示名 |
| `local_url` | string | ローカルサーバー URL（Ollama/LMStudio/vLLM/カスタム） |
| `enabled` | bool | 有効化状態 |
| `custom` | bool | ユーザー追加サービスかどうか |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "プロキシ使用" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `client_id` | string | クライアント ID |
| `version` | string | プロキシバージョン（例: `v0.1.6.20260314.231308`） |
| `service` | string | 現在のサービス |
| `model` | string | 現在のモデル |
| `sse_connected` | bool | SSE 接続状態 |
| `host` | string | ホスト名 |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | 最終更新時刻 |
| `vault.today_usage` | int | 本日のトークン使用量 |
| `vault.daily_limit` | int | 日次上限 |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## エラーレスポンス

```json
{"error": "오류 메시지"}
```

| コード | 意味 |
|-------|------|
| 200 | 成功 |
| 400 | 不正なリクエスト |
| 401 | 認証失敗 |
| 403 | アクセス拒否（非アクティブクライアント、IP ブロック） |
| 404 | リソースなし |
| 405 | 許可されていないメソッド |
| 429 | レート制限超過 |
| 500 | サーバー内部エラー |
| 502 | アップストリーム API エラー（すべてのフォールバック失敗） |

---

## cURL サンプル集

```bash
# ─── プロキシ ──────────────────────────────────────────────────────────────────

# ヘルスチェック
curl http://localhost:56244/health

# ステータス照会
curl http://localhost:56244/status

# モデル一覧（全件）
curl http://localhost:56244/api/models

# Google モデルのみ
curl "http://localhost:56244/api/models?service=google"

# 無料モデル検索
curl "http://localhost:56244/api/models?q=alpha"

# モデル変更（ローカル）
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# 設定リロード
curl -X POST http://localhost:56244/reload

# Gemini API 直接呼び出し
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI 互換（デフォルトモデル）
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model 形式
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# 無料 1M context モデル使用
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── キー金庫（公開） ──────────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── キー金庫（管理者） ────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# キー一覧
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Google キー追加
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI キー追加
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter キー追加
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# キー削除（SSE key_deleted ブロードキャスト）
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# 日次使用量リセット
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# クライアント一覧
curl -H "$ADMIN" http://localhost:56243/admin/clients

# クライアント追加（OpenClaw）
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# クライアントのモデル変更（SSE 即時反映）
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# クライアント無効化
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# クライアント削除
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# サービス一覧
curl -H "$ADMIN" http://localhost:56243/admin/services

# Ollama ローカル URL 設定（SSE service_changed ブロードキャスト）
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI サービス有効化
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# カスタムサービス追加（SSE service_changed ブロードキャスト）
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# カスタムサービス削除
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# モデル一覧照会
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# プロキシステータス（heartbeat）
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── 分散モード — プロキシ → 金庫 ──────────────────────────────────────────────

# 復号済みキー照会
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat 送信
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## ミドルウェア

すべてのリクエストに自動適用:

| ミドルウェア | 機能 |
|------------|------|
| **Logger** | `[method] path status latencyms` 形式のロギング |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | パニック復旧、500 レスポンス返却 |

---

*最終更新: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
