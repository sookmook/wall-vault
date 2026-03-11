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
  - [관리자 API — 프록시 상태](#관리자-api--프록시-상태)
- [SSE 이벤트 타입](#sse-이벤트-타입)
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
  "version": "v0.1.0",
  "client": "motoko"
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `status` | string | 항상 `"ok"` |
| `version` | string | 프록시 버전 |
| `client` | string | 클라이언트 ID |

---

### `GET /status`

프록시 상태 상세 조회.

**응답 예시:**
```json
{
  "status": "ok",
  "version": "v0.1.0",
  "client": "motoko",
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
    }
  ],
  "count": 1
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | 모델 ID |
| `name` | string | 모델 표시명 |
| `service` | string | `google` \| `openrouter` \| `ollama` |
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

사고 모드 토글 (현재 no-op, 향후 확장용).

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

```json
{
  "candidates": [
    {
      "content": {
        "parts": [{"text": "안녕하세요! 무엇을 도와드릴까요?"}],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 3,
    "candidatesTokenCount": 15,
    "totalTokenCount": 18
  }
}
```

**도구 필터:** `tool_filter: strip_all` 설정 시 요청의 `tools` 배열이 자동 제거됩니다.

**폴백 체인:** Google 실패 → OpenRouter → Ollama 순으로 자동 폴백.

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API 스트리밍 프록시.

요청 형식은 비스트리밍과 동일. 응답은 SSE(Server-Sent Events) 스트림:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: {"candidates":[{"content":{"parts":[{"text":"녕"}],...},...}]}

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

**응답 바디:**
```json
{
  "id": "",
  "object": "chat.completion",
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

**`stream: true`** 설정 시 OpenAI 스트리밍 형식으로 응답.

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
  "version": "v0.1.0",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `keys` | int | 등록된 API 키 수 |
| `clients` | int | 등록된 클라이언트 수 |
| `sse` | int | 현재 연결된 SSE 클라이언트 수 |

---

#### `GET /api/clients`

등록된 클라이언트 목록 (공개 정보만, 토큰 제외).

**응답 예시:**
```json
[
  {
    "id": "motoko",
    "name": "야마이 모토코",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }
]
```

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

**이벤트 수신 예시:**
```
data: {"type":"config_change","data":{"client_id":"motoko","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"usage_reset","data":{"time":"2026-03-11T00:00:30Z"}}
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

> **보안:** 이 엔드포인트는 평문 키를 반환합니다. 반드시 클라이언트 토큰으로 인증하세요. 클라이언트의 `allowed_services` 설정에 따라 허용된 서비스 키만 반환됩니다.

---

#### `POST /api/heartbeat`

프록시 상태 전송 (5분마다 자동 실행).

**요청 바디:**
```json
{
  "client_id": "motoko",
  "version": "v0.1.0",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "motoko-wsl",
  "vault": {
    "today_usage": 42,
    "daily_limit": 1000,
    "key_status": "active"
  }
}
```

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
    "created_at": "2026-03-11T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `available` | bool | 쿨다운·한도 없이 사용 가능 여부 |
| `usage_pct` | int | 일일 한도 대비 사용량 % |
| `cooldown_until` | RFC3339 | 쿨다운 종료 시각 (제로값이면 쿨다운 없음) |
| `last_error` | int | 마지막 HTTP 오류 코드 |

---

#### `POST /admin/keys`

새 API 키 등록.

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
| `service` | ✅ | `google` \| `openrouter` \| `ollama` \| 커스텀 |
| `key` | ✅ | API 키 평문 |
| `label` | — | 식별용 레이블 |
| `daily_limit` | — | 일일 사용 한도 (0 = 무제한) |

**응답:** 생성된 `APIKey` 객체 (평문 키 포함)

등록 즉시 SSE `key_added` 이벤트가 브로드캐스트됩니다.

---

#### `DELETE /admin/keys/{id}`

API 키 삭제.

**경로 파라미터:**
- `{id}`: 키 ID

**응답:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

모든 키의 일일 사용량 초기화 (자정 자동 초기화와 동일).

**응답:**
```json
{
  "status": "reset",
  "time": "2026-03-11T15:00:00Z"
}
```

초기화 후 SSE `usage_reset` 이벤트가 브로드캐스트됩니다.

---

### 관리자 API — 클라이언트

#### `GET /admin/clients`

모든 클라이언트 목록 (토큰 포함).

**응답 예시:**
```json
[
  {
    "id": "motoko",
    "name": "야마이 모토코",
    "token": "motoko-secret-token",
    "default_service": "google",
    "default_model": "gemini-2.5-flash",
    "allowed_services": ["google", "openrouter"],
    "created_at": "2026-03-11T12:00:00Z"
  }
]
```

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
  "allowed_services": ["google", "openrouter"]
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

**응답:** 생성된 `Client` 객체

---

#### `GET /admin/clients/{id}`

특정 클라이언트 조회.

---

#### `PUT /admin/clients/{id}`

클라이언트 설정 변경. **SSE를 통해 해당 클라이언트 프록시에 즉시 반영됩니다.**

**요청 바디:**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-3.5-sonnet",
  "allowed_services": ["google", "openrouter", "ollama"]
}
```

**응답:**
```json
{"status": "updated"}
```

변경 후 SSE `config_change` 이벤트가 브로드캐스트됩니다:
```json
{
  "type": "config_change",
  "data": {
    "client_id": "motoko",
    "service": "openrouter",
    "model": "anthropic/claude-3.5-sonnet"
  }
}
```

---

#### `DELETE /admin/clients/{id}`

클라이언트 삭제.

**응답:**
```json
{"status": "deleted"}
```

---

### 관리자 API — 프록시 상태

#### `GET /admin/proxies`

연결된 모든 프록시의 마지막 Heartbeat 상태.

**응답 예시:**
```json
[
  {
    "client_id": "motoko",
    "version": "v0.1.0",
    "service": "google",
    "model": "gemini-2.5-flash",
    "sse_connected": true,
    "host": "motoko-wsl",
    "updated_at": "2026-03-11T15:04:05Z",
    "vault": {
      "today_usage": 42,
      "daily_limit": 1000,
      "key_status": "active"
    }
  }
]
```

---

## SSE 이벤트 타입

금고 `/api/events` 스트림에서 수신되는 이벤트:

| `type` | 발생 조건 | `data` 내용 |
|--------|-----------|-------------|
| `connected` | SSE 연결 즉시 | `{"clients": N}` |
| `config_change` | 클라이언트 설정 변경 | `{"client_id","service","model"}` |
| `key_added` | 새 API 키 등록 | `{"service": "google"}` |
| `usage_reset` | 일일 사용량 초기화 | `{"time": "RFC3339"}` |

**프록시가 수신하는 이벤트 처리:**

```
config_change 수신
  → client_id가 자신과 일치하는 경우
    → service, model 즉시 갱신
    → hooksMgr.Fire(EventModelChanged)
```

---

## 데이터 스키마

### APIKey

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | string | UUID 형식 고유 ID |
| `service` | string | `google` \| `openrouter` \| `ollama` |
| `encrypted_key` | string | AES-GCM 암호화된 키 (Base64) |
| `label` | string | 사람이 읽을 수 있는 레이블 |
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
| `default_model` | string | 기본 모델 |
| `allowed_services` | []string | 허용 서비스 (빈 배열 = 모두) |
| `created_at` | time.Time | 등록 시각 |

### ProxyStatus (Heartbeat)

| 필드 | 타입 | 설명 |
|------|------|------|
| `client_id` | string | 클라이언트 ID |
| `version` | string | 프록시 버전 |
| `service` | string | 현재 서비스 |
| `model` | string | 현재 모델 |
| `sse_connected` | bool | SSE 연결 여부 |
| `host` | string | 호스트명 |
| `updated_at` | time.Time | 마지막 업데이트 |
| `vault.today_usage` | int | 오늘 토큰 사용량 |
| `vault.daily_limit` | int | 일일 한도 |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## 오류 응답

모든 오류는 JSON 형식으로 반환됩니다:

```json
{"error": "오류 메시지"}
```

**HTTP 상태 코드:**

| 코드 | 의미 |
|------|------|
| 200 | 성공 |
| 400 | 잘못된 요청 (바디 파싱 실패 등) |
| 401 | 인증 실패 |
| 404 | 리소스 없음 |
| 405 | 허용되지 않는 메서드 |
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

# 모델 목록
curl http://localhost:56244/api/models

# Google 모델만
curl "http://localhost:56244/api/models?service=google"

# 모델 검색
curl "http://localhost:56244/api/models?q=gemini"

# 모델 변경
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# 설정 새로고침
curl -X POST http://localhost:56244/reload

# Gemini API 호출
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI 호환 API
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# ─── 키 금고 (공개) ───────────────────────────────────────────────────────────

# 상태 조회
curl http://localhost:56243/api/status

# 클라이언트 목록 (공개)
curl http://localhost:56243/api/clients

# SSE 스트림 테스트 (3초 후 종료)
curl -s http://localhost:56243/api/events --max-time 3

# ─── 키 금고 (관리자) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# 키 목록
curl -H "$ADMIN" http://localhost:56243/admin/keys

# 키 추가 (Google)
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# 키 삭제
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 \
  -H "$ADMIN"

# 일일 사용량 초기화
curl -X POST http://localhost:56243/admin/keys/reset \
  -H "$ADMIN"

# 클라이언트 목록
curl -H "$ADMIN" http://localhost:56243/admin/clients

# 클라이언트 추가
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-bot","name":"내 봇","token":"my-token","default_service":"google","default_model":"gemini-2.5-flash"}'

# 클라이언트 모델 변경 (SSE 즉시 반영)
curl -X PUT http://localhost:56243/admin/clients/motoko \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"anthropic/claude-3.5-sonnet"}'

# 클라이언트 삭제
curl -X DELETE http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN"

# 프록시 Heartbeat 상태
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── 분산 모드 — 프록시 → 금고 ───────────────────────────────────────────────

# 복호화된 키 조회 (클라이언트 토큰)
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer motoko-secret-token"

# 서비스 필터
curl "http://localhost:56243/api/keys?service=google" \
  -H "Authorization: Bearer motoko-secret-token"

# Heartbeat 전송
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer motoko-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"motoko","version":"v0.1.0","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## 미들웨어

모든 요청에 자동 적용:

| 미들웨어 | 기능 |
|---------|------|
| **Logger** | `[method] path status latencyms` 형식 로깅 |
| **CORS** | `Access-Control-Allow-Origin: *` (모든 출처 허용) |
| **Recovery** | 패닉 복구, 500 응답 반환 |

---

*최종 업데이트: 2026-03-11*
