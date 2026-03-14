# wall-vault API 매뉴얼

이 문서는 wall-vault의 모든 HTTP API 엔드포인트를 상세히 기술합니다.

---

## 목차

- [인증](#인증)
- [프록시 API (:56244)](#프록시-api-56244)
  - [헬스체크](#get-health)
  - [상태 조회](#get-status)
  - [모델 목록](#get-apimodels)
  - [모델 변경](#put-apiconfigmodel)
  - [사고 모드](#put-apiconfigthink-mode)
  - [설정 새로고침](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini 스트리밍](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI 호환 API](#post-v1chatcompletions)
- [키 금고 API (:56243)](#키-금고-api-56243)
  - [공개 API](#공개-api-인증-불필요)
  - [SSE 이벤트 스트림](#get-apievents)
  - [프록시 전용 API](#프록시-전용-api-클라이언트-토큰)
  - [관리자 API — 키](#관리자-api--api-키)
  - [관리자 API — 클라이언트](#관리자-api--클라이언트)
  - [관리자 API — 서비스](#관리자-api--서비스)
  - [관리자 API — 모델 목록](#관리자-api--모델-목록)
  - [관리자 API — 프록시 상태](#관리자-api--프록시-상태)
- [SSE 이벤트 타입](#sse-이벤트-타입)
- [프로바이더·모델 라우팅](#프로바이더모델-라우팅)
- [데이터 스키마](#데이터-스키마)
- [오류 응답](#오류-응답)
- [cURL 예제 모음](#curl-예제-모음)

---

## 인증

| 영역 | 방법 | 헤더 |
|------|------|------|
| 관리자 API | Bearer 토큰 | `Authorization: Bearer <admin_token>` |
| 프록시 → 금고 | Bearer 토큰 | `Authorization: Bearer <client_token>` |
| 프록시 API | 없음 (로컬) | — |

`admin_token`이 설정되지 않은 경우(빈 문자열) 모든 관리자 API는 인증 없이 접근 가능합니다.

### 보안 정책

- **Rate Limiting**: 관리자 API 인증 실패 10회/15분 초과 시 해당 IP를 일시 차단 (`429 Too Many Requests`)
- **IP 화이트리스트**: 에이전트(`Client`)의 `ip_whitelist` 필드에 등록된 IP/CIDR만 `/api/keys` 접근 허용. 빈 배열이면 모두 허용.
- **theme·lang 보호**: `/admin/theme`, `/admin/lang`도 관리자 토큰 인증 필요

---

## 프록시 API (:56244)

프록시가 실행되는 서버. 기본 포트 `56244`.

---

### `GET /health`

헬스체크. 항상 200 OK를 반환합니다.

**응답 예시:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

프록시 상태 상세 조회.

**응답 예시:**
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

| 필드 | 타입 | 설명 |
|------|------|------|
| `service` | string | 현재 기본 서비스 |
| `model` | string | 현재 기본 모델 |
| `sse` | bool | 금고 SSE 연결 여부 |
| `filter` | string | 도구 필터 모드 |
| `services` | []string | 활성화된 서비스 목록 |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

사용 가능한 모델 목록 조회. TTL 캐시(기본 10분) 사용.

**쿼리 파라미터:**

| 파라미터 | 설명 | 예시 |
|---------|------|------|
| `service` | 서비스 필터 | `?service=google` |
| `q` | 모델 ID/이름 검색 | `?q=gemini` |

**응답 예시:**
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

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 모델 ID |
| `name` | string | 모델 표시명 |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` 등 |
| `context_length` | int | 컨텍스트 윈도우 크기 |
| `free` | bool | 무료 모델 여부 (OpenRouter) |

---

### `PUT /api/config/model`

현재 서비스·모델 변경.

**요청 바디:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**응답:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **참고:** distributed 모드에서는 이 API 대신 금고의 `PUT /admin/clients/{id}` 사용을 권장합니다. 금고 변경은 SSE를 통해 1–3초 내 자동 반영됩니다.

---

### `PUT /api/config/think-mode`

사고 모드 토글 (no-op, 향후 확장용).

**응답:**
```json
{"status": "ok"}
```

---

### `POST /reload`

금고에서 클라이언트 설정·키를 즉시 재동기화합니다.

**응답:**
```json
{"status": "reloading"}
```

재동기화는 비동기로 실행되므로 응답 수신 후 1–2초 내 완료됩니다.

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API 프록시 (비스트리밍).

**경로 파라미터:**
- `{model}`: 모델 ID. `gemini-` 접두사가 있으면 자동으로 Google 서비스 선택.

**요청 바디:** [Gemini generateContent 요청 형식](https://ai.google.dev/api/generate-content)

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

**응답 바디:** Gemini generateContent 응답 형식

**도구 필터:** `tool_filter: strip_all` 설정 시 요청의 `tools` 배열이 자동 제거됩니다.

**폴백 체인:** 지정 서비스 실패 → 설정된 서비스 순서대로 폴백 → Ollama (최종).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API 스트리밍 프록시. 요청 형식은 비스트리밍과 동일. 응답은 SSE 스트림:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI 호환 API. 내부적으로 Gemini 형식으로 변환 후 처리합니다.

**요청 바디:**
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

**`model` 필드의 프로바이더 접두어 지원 (OpenClaw 3.11+):**

| 모델 예시 | 라우팅 |
|-----------|--------|
| `gemini-2.5-flash` | 현재 설정 서비스 |
| `google/gemini-2.5-pro` | Google 직접 |
| `openai/gpt-4o` | OpenAI 직접 |
| `anthropic/claude-opus-4-6` | OpenRouter 경유 |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter 직접 |
| `wall-vault/gemini-2.5-flash` | 자동 감지 → Google |
| `wall-vault/claude-opus-4-6` | 자동 감지 → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | 자동 감지 → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (무료 1M context) |
| `moonshot/kimi-k2.5` | OpenRouter 경유 |
| `opencode-go/model` | OpenRouter 경유 |
| `kimi-k2.5:cloud` | `:cloud` 접미사 → OpenRouter |

자세한 내용은 [프로바이더·모델 라우팅](#프로바이더모델-라우팅) 참고.

**응답 바디:**
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

> **모델 제어 토큰 자동 제거:** 응답에 GLM-5 / DeepSeek / ChatML 구분자(`<|im_start|>`, `[gMASK]`, `[sop]` 등)가 포함된 경우 자동으로 제거됩니다.

---

## 키 금고 API (:56243)

키 금고가 실행되는 서버. 기본 포트 `56243`.

---

### 공개 API (인증 불필요)

#### `GET /`

웹 대시보드 UI. 브라우저에서 접속.

---

#### `GET /api/status`

금고 상태 조회.

**응답 예시:**
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

등록된 클라이언트 목록 (공개 정보만, 토큰 제외).

---

### `GET /api/events`

SSE(Server-Sent Events) 실시간 이벤트 스트림.

**헤더:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**연결 즉시 수신:**
```
data: {"type":"connected","clients":2}
```

**이벤트 예시:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

자세한 이벤트 타입은 [SSE 이벤트 타입](#sse-이벤트-타입) 참고.

---

### 프록시 전용 API (클라이언트 토큰)

`Authorization: Bearer <client_token>` 헤더 필요. 관리자 토큰으로도 인증 가능.

#### `GET /api/keys`

프록시에게 제공하는 복호화된 API 키 목록.

**쿼리 파라미터:**

| 파라미터 | 설명 |
|---------|------|
| `service` | 서비스 필터 (예: `?service=google`) |

**응답 예시:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000
  }
]
```

> **보안:** 평문 키를 반환합니다. 클라이언트의 `allowed_services` 설정에 따라 허용된 서비스 키만 반환됩니다.

---

#### `GET /api/services`

프록시가 사용할 서비스 목록 조회. `proxy_enabled=true`인 서비스 ID 배열 반환.

**응답 예시:**
```json
["google", "ollama"]
```

빈 배열이면 프록시는 제한 없이 모든 서비스를 사용합니다.

---

#### `POST /api/heartbeat`

프록시 상태 전송 (60초마다 자동 실행).

**요청 바디:**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-wsl",
  "avatar": "data:image/png;base64,...",
  "vault": {
    "today_usage": 42,
    "daily_limit": 1000,
    "key_status": "active"
  }
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `version` | string | 프록시 버전 (build timestamp 포함, e.g. `v0.1.6.20260314.231308`) |
| `service` | string | 현재 서비스 |
| `model` | string | 현재 모델 |
| `sse_connected` | bool | SSE 연결 여부 |
| `host` | string | 호스트명 |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `vault` | object | 키 사용량 상태 |

**응답:**
```json
{"status": "ok"}
```

---

### 관리자 API — API 키

`Authorization: Bearer <admin_token>` 헤더 필요.

#### `GET /admin/keys`

등록된 모든 API 키 목록 (평문 키 제외).

**응답 예시:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `available` | bool | 쿨다운·한도 없이 사용 가능 여부 |
| `usage_pct` | int | 일일 한도 대비 사용량 % |
| `cooldown_until` | RFC3339 | 쿨다운 종료 시각 (제로값이면 없음) |
| `last_error` | int | 마지막 HTTP 오류 코드 |

---

#### `POST /admin/keys`

새 API 키 등록. 등록 즉시 SSE `key_added` 이벤트가 브로드캐스트됩니다.

**요청 바디:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| 필드 | 필수 | 설명 |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| 커스텀 |
| `key` | ✅ | API 키 평문 |
| `label` | — | 식별용 레이블 |
| `daily_limit` | — | 일일 사용 한도 (0 = 무제한) |

---

#### `DELETE /admin/keys/{id}`

API 키 삭제. 삭제 후 SSE `key_deleted` 이벤트가 브로드캐스트됩니다.

**응답:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

모든 키의 일일 사용량 초기화. SSE `usage_reset` 이벤트 브로드캐스트.

**응답:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### 관리자 API — 클라이언트

#### `GET /admin/clients`

모든 클라이언트 목록 (토큰 포함).

---

#### `POST /admin/clients`

새 클라이언트 등록.

**요청 바디:**
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
  "ip_whitelist": ["192.168.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| 필드 | 필수 | 설명 |
|------|------|------|
| `id` | ✅ | 클라이언트 고유 ID |
| `name` | — | 표시명 |
| `token` | — | 인증 토큰 (생략 시 자동 생성) |
| `default_service` | — | 기본 서비스 |
| `default_model` | — | 기본 모델 |
| `allowed_services` | — | 허용 서비스 목록 (빈 배열 = 모두 허용) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | 에이전트 작업 디렉토리 |
| `description` | — | 에이전트 설명 |
| `ip_whitelist` | — | 허용 IP 목록 (빈 배열 = 모두 허용, CIDR 지원) |
| `enabled` | — | 활성화 여부 (기본값 `true`) |

---

#### `GET /admin/clients/{id}`

특정 클라이언트 조회 (토큰 포함).

---

#### `PUT /admin/clients/{id}`

클라이언트 설정 변경. **SSE `config_change` 브로드캐스트 → 프록시에 1–3초 내 반영.**

**요청 바디 (변경할 필드만):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**응답:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

클라이언트 삭제.

---

### 관리자 API — 서비스

#### `GET /admin/services`

등록된 서비스 목록.

**응답 예시:**
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

기본 제공 서비스 8개: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

커스텀 서비스 추가. 추가 후 SSE `service_changed` 이벤트 브로드캐스트 → **대시보드 드롭다운 즉시 갱신**.

**요청 바디:**
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

서비스 설정 업데이트. 변경 후 SSE `service_changed` 이벤트 브로드캐스트.

**요청 바디:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

커스텀 서비스 삭제. 삭제 후 SSE `service_changed` 이벤트 브로드캐스트.

기본 서비스(`custom: false`) 삭제 시도:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### 관리자 API — 모델 목록

#### `GET /admin/models`

서비스별 모델 목록 조회. TTL 캐시(10분) 사용.

**쿼리 파라미터:**

| 파라미터 | 설명 | 예시 |
|---------|------|------|
| `service` | 서비스 필터 | `?service=google` |
| `q` | 모델 검색 | `?q=gemini` |

**서비스별 모델 조회 방식:**

| 서비스 | 방식 | 개수 |
|--------|------|------|
| `google` | 고정 목록 | 8개 (embedding 포함) |
| `openai` | 고정 목록 | 9개 |
| `anthropic` | 고정 목록 | 6개 |
| `github-copilot` | 고정 목록 | 6개 |
| `openrouter` | API 동적 조회 (실패 시 curated 폴백 14개) | 340+개 |
| `ollama` | 로컬 서버 동적 조회 (미응답 시 추천 7개) | 가변 |
| `lmstudio` | 로컬 서버 동적 조회 | 가변 |
| `vllm` | 로컬 서버 동적 조회 | 가변 |
| 커스텀 | OpenAI 호환 `/v1/models` | 가변 |

**OpenRouter 폴백 모델 목록 (API 미응답 시):**

| 모델 | 특이사항 |
|------|----------|
| `openrouter/hunter-alpha` | 무료, 1M context |
| `openrouter/healer-alpha` | 무료, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### 관리자 API — 프록시 상태

#### `GET /admin/proxies`

연결된 모든 프록시의 마지막 Heartbeat 상태.

---

## SSE 이벤트 타입

금고 `/api/events` 스트림에서 수신되는 이벤트:

| `type` | 발생 조건 | `data` 내용 | 대시보드 반응 |
|--------|-----------|-------------|--------------|
| `connected` | SSE 연결 즉시 | `{"clients": N}` | — |
| `config_change` | 클라이언트 설정 변경 | `{"client_id","service","model"}` | 에이전트 카드 모델 드롭다운 갱신 |
| `key_added` | 새 API 키 등록 | `{"service": "google"}` | 모델 드롭다운 갱신 |
| `key_deleted` | API 키 삭제 | `{"service": "google"}` | 모델 드롭다운 갱신 |
| `service_changed` | 서비스 추가/수정/삭제 | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | 서비스 select + 모델 드롭다운 즉시 갱신; 프록시의 dispatch 서비스 목록 실시간 갱신 |
| `usage_reset` | 일일 사용량 초기화 | `{"time": "RFC3339"}` | 페이지 새로고침 |

**프록시가 수신하는 이벤트 처리:**

```
config_change 수신
  → client_id가 자신과 일치하는 경우
    → service, model 즉시 갱신
    → hooksMgr.Fire(EventModelChanged)
```

---

## 프로바이더·모델 라우팅

`/v1/chat/completions`의 `model` 필드에 `provider/model` 형식을 지정하면 자동 라우팅됩니다 (OpenClaw 3.11 호환).

### 접두어 라우팅 규칙

| 접두어 | 라우팅 대상 | 예시 |
|--------|------------|------|
| `google/` | Google 직접 | `google/gemini-2.5-pro` |
| `openai/` | OpenAI 직접 | `openai/gpt-4o` |
| `anthropic/` | OpenRouter 경유 | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama 직접 | `ollama/qwen3.5:35b` |
| `openrouter/` | OpenRouter (bare 경로 유지) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (full path 유지) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### `wall-vault/` 접두어 자동 감지

wall-vault 자체 접두어로 모델 ID에서 서비스를 자동 판별합니다.

| 모델 ID 패턴 | 라우팅 |
|-------------|--------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic 경로) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (무료 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| 기타 | OpenRouter |

### `:cloud` 접미사 처리

Ollama 태그 형식의 `:cloud` 접미사는 자동으로 제거 후 OpenRouter로 라우팅됩니다.

```
kimi-k2.5:cloud  →  OpenRouter, 모델 ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, 모델 ID: glm-5
```

### OpenClaw openclaw.json 연동 예시

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

에이전트 카드의 **🐾 버튼**을 클릭하면 해당 에이전트용 설정 스니펫이 클립보드에 자동 복사됩니다.

---

## 데이터 스키마

### APIKey

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | UUID 형식 고유 ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| 커스텀 |
| `encrypted_key` | string | AES-GCM 암호화된 키 (Base64) |
| `label` | string | 식별용 레이블 |
| `today_usage` | int | 오늘 사용된 토큰 수 |
| `daily_limit` | int | 일일 한도 (0 = 무제한) |
| `cooldown_until` | time.Time | 쿨다운 종료 시각 |
| `last_error` | int | 마지막 HTTP 오류 코드 |
| `created_at` | time.Time | 등록 시각 |

**쿨다운 정책:**

| HTTP 오류 | 쿨다운 |
|-----------|--------|
| 429 (Too Many Requests) | 30분 |
| 400 / 401 / 403 | 24시간 |
| 네트워크 오류 | 10분 |

### Client

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 클라이언트 고유 ID |
| `name` | string | 표시명 |
| `token` | string | 인증 토큰 |
| `default_service` | string | 기본 서비스 |
| `default_model` | string | 기본 모델 (`provider/model` 형식 가능) |
| `allowed_services` | []string | 허용 서비스 (빈 배열 = 모두) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | 에이전트 작업 디렉토리 |
| `description` | string | 설명 |
| `ip_whitelist` | []string | 허용 IP 목록 (CIDR 지원) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false`이면 `/api/keys` 접근 시 `403` |
| `created_at` | time.Time | 등록 시각 |

### ServiceConfig

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 서비스 고유 ID |
| `name` | string | 표시명 |
| `local_url` | string | 로컬 서버 URL (Ollama/LMStudio/vLLM/커스텀) |
| `enabled` | bool | 활성화 여부 |
| `custom` | bool | 사용자 추가 서비스 여부 |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `version` | string | 프록시 버전 (e.g. `v0.1.6.20260314.231308`) |
| `service` | string | 현재 서비스 |
| `model` | string | 현재 모델 |
| `sse_connected` | bool | SSE 연결 여부 |
| `host` | string | 호스트명 |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | 마지막 업데이트 |
| `vault.today_usage` | int | 오늘 토큰 사용량 |
| `vault.daily_limit` | int | 일일 한도 |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## 오류 응답

```json
{"error": "오류 메시지"}
```

| 코드 | 의미 |
|------|------|
| 200 | 성공 |
| 400 | 잘못된 요청 |
| 401 | 인증 실패 |
| 403 | 접근 거부 (비활성 클라이언트, IP 차단) |
| 404 | 리소스 없음 |
| 405 | 허용되지 않는 메서드 |
| 429 | Rate limit 초과 |
| 500 | 서버 내부 오류 |
| 502 | 업스트림 API 오류 (모든 폴백 실패) |

---

## cURL 예제 모음

```bash
# ─── 프록시 ───────────────────────────────────────────────────────────────────

# 헬스체크
curl http://localhost:56244/health

# 상태 조회
curl http://localhost:56244/status

# 모델 목록 (전체)
curl http://localhost:56244/api/models

# Google 모델만
curl "http://localhost:56244/api/models?service=google"

# 무료 모델 검색
curl "http://localhost:56244/api/models?q=alpha"

# 모델 변경 (로컬)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# 설정 새로고침
curl -X POST http://localhost:56244/reload

# Gemini API 직접 호출
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI 호환 (기본 모델)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model 형식
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# 무료 1M context 모델 사용
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── 키 금고 (공개) ───────────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── 키 금고 (관리자) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# 키 목록
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Google 키 추가
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI 키 추가
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter 키 추가
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# 키 삭제 (SSE key_deleted 브로드캐스트)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# 일일 사용량 초기화
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# 클라이언트 목록
curl -H "$ADMIN" http://localhost:56243/admin/clients

# 클라이언트 추가 (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# 클라이언트 모델 변경 (SSE 즉시 반영)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# 클라이언트 비활성화
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# 클라이언트 삭제
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# 서비스 목록
curl -H "$ADMIN" http://localhost:56243/admin/services

# Ollama 로컬 URL 설정 (SSE service_changed 브로드캐스트)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI 서비스 활성화
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# 커스텀 서비스 추가 (SSE service_changed 브로드캐스트)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# 커스텀 서비스 삭제
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# 모델 목록 조회
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# 프록시 상태 (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── 분산 모드 — 프록시 → 금고 ───────────────────────────────────────────────

# 복호화된 키 조회
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat 전송
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## 미들웨어

모든 요청에 자동 적용:

| 미들웨어 | 기능 |
|---------|------|
| **Logger** | `[method] path status latencyms` 형식 로깅 |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | 패닉 복구, 500 응답 반환 |

---

*Last updated: 2026-03-14 — v0.1.6: avatar heartbeat sync, build timestamp versioning, agent save button fix, proxy-only service filter, avatar path support*
