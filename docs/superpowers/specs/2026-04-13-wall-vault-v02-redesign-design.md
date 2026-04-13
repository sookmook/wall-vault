# wall-vault v0.2.0 — 전면 재설계 Design Spec

- **Date**: 2026-04-13
- **Version target**: `v0.2.0` (from v0.1.29)
- **Scope**: Phase 1 (Foundation) + Phase 2 (Admin API + Dashboard UI)
- **Out of scope (next rounds)**: Phase 3 RBAC/multi-user, Phase 4 Webhook/Metrics/Backup UI
- **Status**: Draft, awaiting user review before `writing-plans` transition

---

## 1. Context & Goals

### 1.1 왜 재설계하는가

현 v0.1.x는 **클라이언트(agent)가 model 이름을 직접 소유**하고 dispatch fallback 시에도 **그 model 이름을 다른 provider에 그대로 전달**하는 구조. 이 때문에:

- 작순이 사례: primary `google/gemini-3.1-pro-preview` 가 Google API에서 400 → fallback chain의 Ollama까지 `gemini-3.1-pro-preview` 를 그대로 찾아가 **`HTTP 404 model not found`** 로 종료
- 모토코/라즈/미니: `AnthropicToGemini` 변환 단계에서 tool_use/tool_result content block이 유실되어 "messages too short" 400 연쇄 실패 (v0.1.29 hot-fix로 부분 해소)
- 운영자 관점: "agent가 어느 provider의 어느 model 이름을 쓰는지" 를 두 곳에서 따로 관리해야 해서 setup 시 휴먼 오류 다발

근본 원인은 **서비스와 모델을 결합한 단일 네이밍(`provider/model`)을 dispatch 전체에 공유**하는 데이터 모델. "서비스는 서비스, 모델은 그 서비스의 구현 디테일"이라는 관점으로 엎는다.

### 1.2 Design goals

1. **Service-Model Registry**: 서비스 자체에 default model·allowed models 소유. agent는 "어떤 서비스를 선호"만 말하고 모델은 선택적 override.
2. **Deterministic dispatch**: fallback chain의 각 service는 **자신의 default_model**을 자동 적용. "model name mismatch" 휴리스틱 완전 삭제.
3. **Tool-use first-class**: v0.1.29 convert.go 수정 유지 + 새 dispatch가 tool_use/tool_result를 보존하는 경로로만 라우팅.
4. **Clean, typed UI**: `vault/ui.go` 2,572줄(fmt.Sprintf 서버 렌더) 폐기 → Go templ + HTMX 기반 컴포넌트 트리. 관리자가 한 화면에서 서비스·에이전트·키를 모두 보고 편집할 수 있는 **one-screen hybrid** 레이아웃.
5. **One-shot migration**: 기존 `vault.json` 을 새 스키마로 자동 변환(강제 백업 동반), 구 API 경로 이름은 유지하되 request/response 스키마는 clean break.

### 1.3 Non-goals (이번 라운드)

- RBAC / multi-user / 세션 인증 (Phase 3)
- Webhook / Prometheus metrics / Backup-Restore UI (Phase 4)
- 언어·프레임워크·crypto·SSE broker 재구현
- Key rotation/cooldown 알고리즘 재설계 (호출부만 맞춰 조정)

---

## 2. Architecture Overview

```
┌────────────────────────── wall-vault v0.2.0 ──────────────────────────┐
│                                                                        │
│   ┌─ vault (56243) ────────────────────┐    ┌─ proxy (56244) ────────┐ │
│   │  HTTP + SSE + encrypted store      │    │  Anthropic/OpenAI/     │ │
│   │  Admin API: /admin/*               │◀───│  Gemini ingress →      │ │
│   │  Dashboard UI: templ + HTMX (α)    │SSE │  dispatch chain →      │ │
│   │  One-shot migration on boot        │    │  upstream provider     │ │
│   │                                    │    │  (key mgr + tool       │ │
│   │  store.json (AES-256-GCM)          │    │   filter + convert)    │ │
│   │    ├ services[] (default_model 등) │    │                        │ │
│   │    └ clients[]  (preferred_service)│    │  Model resolution via  │ │
│   └────────────────────────────────────┘    │  Service.default_model │ │
│                                             │  + Client.override     │ │
│                                             └────────────────────────┘ │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘

Agents (OpenClaw / nanoclaw / Claude Code / Cline):
  POST /v1/messages  —→  wall-vault proxy  —→  resolve service+model
                                             —→  convert (tool_use preserved)
                                             —→  upstream chain with per-service
                                                  default_model fallback
```

### 2.1 변경된 핵심 흐름

**모델 해석 (resolve phase)**:
```
resolveModel(client, service) :=
    if client.model_override != "" and service.allowed_models is empty:
        return client.model_override
    if client.model_override != "" and client.model_override ∈ service.allowed_models:
        return client.model_override
    if client.model_override != "" and client.model_override ∉ service.allowed_models:
        return ERROR_422  (admin-layer) / ERROR_400 (dispatch-layer)
    return service.default_model
```

**Dispatch chain**:
```
dispatch(client, req):
    order := [client.preferred_service] +
             [s ∈ vault.services
                if s.proxy_enabled and
                   (client.allowed_services is empty or s.id ∈ client.allowed_services) and
                   s.id != client.preferred_service
              sorted by s.sort_order asc]

    for s in order:
        model := resolveModel(client, s)
        if model == ERROR: continue  (with warn log)
        try call(s, model, req)
        on success: return
        on failure: record cooldown(s, error_code), continue

    return 502 all-services-failed
```

v0.1.27의 Ollama name-mismatch 휴리스틱은 **삭제**. 현 흐름에서는 각 service가 자기 default_model을 쓰므로 구조적으로 name-mismatch가 발생할 수 없음.

---

## 3. Data Model

### 3.1 Service

```go
type Service struct {
    ID             string   `json:"id"`             // "google", "openrouter", "anthropic", "ollama", "lmstudio", "vllm", "openai", ...
    Name           string   `json:"name"`           // 사람이 읽는 이름
    DefaultModel   string   `json:"default_model"`  // 신규 필수 필드 — 이 service로 dispatch 갈 때의 기본 모델
    LocalURL       string   `json:"local_url,omitempty"` // Ollama/LMStudio/vLLM만
    Enabled        bool     `json:"enabled"`        // 대시보드 표시 여부
    ProxyEnabled   bool     `json:"proxy_enabled"`  // dispatch chain 포함 여부
    SortOrder      int      `json:"sort_order"`     // fallback 순서 (오름차순)
    AllowedModels  []string `json:"allowed_models"` // 선택: 비면 제한 없음, 있으면 strict whitelist
}
```

### 3.2 Client (Agent)

```go
type Client struct {
    ID                string    `json:"id"`                 // "motoko", "mini9", ...
    Name              string    `json:"name"`
    Token             string    `json:"token"`              // agent 측 인증 토큰
    PreferredService  string    `json:"preferred_service"`  // ex: "google"; 이전 default_service
    ModelOverride     string    `json:"model_override"`     // optional; 이전 default_model 의 역할 변경
    AllowedServices   []string  `json:"allowed_services,omitempty"` // 기존 의미 유지
    AgentType         string    `json:"agent_type"`         // "nanoclaw" / "openclaw" / "claude-code" / "cline" / ...
    WorkDir           string    `json:"work_dir,omitempty"`
    IPWhitelist       []string  `json:"ip_whitelist,omitempty"`
    Avatar            string    `json:"avatar,omitempty"`   // base64 또는 상대 경로
    Enabled           bool      `json:"enabled"`
    SortOrder         int       `json:"sort_order"`
    CreatedAt         time.Time `json:"created_at"`
}
```

### 3.3 Store envelope

```go
type Store struct {
    SchemaVersion int        `json:"schema_version"`  // v0.2.0 → 2
    Services      []Service  `json:"services"`
    Clients       []Client   `json:"clients"`
    APIKeys       []APIKey   `json:"api_keys"`         // 구조 기본 유지
    Theme         string     `json:"theme"`            // 기존 7 테마 그대로
    Lang          string     `json:"lang"`
}
```

실파일은 v0.1.x와 동일하게 `vault.json` (AES-256-GCM + Argon2id).

---

## 4. Dispatch Rules

### 4.1 Preferred service + fallback order

- Primary = `Client.PreferredService`
- Fallback = `allowed_services` (비면 전체) ∩ `proxy_enabled=true` 의 모든 service, `sort_order` 오름차순, primary 제외

### 4.2 Model resolution per step

`resolveModel(client, service)` 규칙은 §2.1.

Strict whitelist:
- `Service.AllowedModels` 비어있음 → model_override 무엇이든 수용
- 비어있지 않음 → model_override가 정확히 그 리스트에 포함되어야 함
- 위반 시:
  - **Admin API (PUT /admin/clients/{id})**: HTTP 422 Unprocessable Entity, body `{"error":"model_override \"X\" not in allowed_models of service \"Y\""}`
  - **Dispatch (POST /v1/messages 등)**: HTTP 400, body Anthropic error shape (비-호환 provider도 proxy가 가공하여 동일 shape 반환)

### 4.3 Cooldown & retry

기존 `internal/proxy/keymgr.go` 정책 유지 (429→5m, 402→30m, 401/403→6h, default→5m). dispatch 함수의 호출부 시그니처만 `(service, model, req)` 로 맞춰 조정.

**Cooldown 스킵**: 특정 service가 cooldown 중이면 해당 라운드에서 chain order 에서 건너뛴다(§2.1 의 `for s in order` 루프가 `keymgr.IsCoolingDown(s)` 체크로 skip). 모든 service가 cooldown이면 v0.1.27의 `force-retry` 로직 동일하게 가장 빨리 풀릴 key를 강제 clear 후 그 service 1회 재시도.

### 4.4 convert.go 유지

v0.1.29의 `AnthropicToGemini = anthropicToOpenAIReq → OpenAIToGemini` 경로 그대로. `doAnthropicRequest` 의 tool-turn JSON 직렬화도 유지. 이번 리팩터는 dispatch 바깥쪽 구조만 바꾸고 content-block 변환은 손대지 않음.

---

## 5. Admin API Spec

URL prefix는 현 v0.1.x 와 동일 `/admin/*`. Body 스키마는 clean break.

### 5.1 Services

| Method | Path | Body | Notes |
|---|---|---|---|
| `GET` | `/admin/services` | — | List all services |
| `GET` | `/admin/services/{id}` | — | Detail |
| `POST` | `/admin/services` | `Service` | Create (id 포함) |
| `PUT` | `/admin/services/{id}` | `Service` partial | **`default_model`, `allowed_models` 편집 가능** |
| `DELETE` | `/admin/services/{id}` | — | Remove (연결된 client 검증 선행) |

### 5.2 Clients

| Method | Path | Body | Notes |
|---|---|---|---|
| `GET` | `/admin/clients` | — | List all clients |
| `GET` | `/admin/clients/{id}` | — | Detail |
| `POST` | `/admin/clients` | `Client` | Create |
| `PUT` | `/admin/clients/{id}` | `Client` partial | **`preferred_service`, `model_override` 편집 시 strict 검증** |
| `DELETE` | `/admin/clients/{id}` | — | Remove |
| `PUT` | `/admin/clients/reorder` | `{order:[id...]}` | sort_order 일괄 갱신 (유지) |

### 5.3 Keys / Themes / Lang / Proxies

기존 v0.1.x 의 엔드포인트 유지 (이번 라운드 스키마 변경 없음): `/admin/keys*`, `/admin/keys/reset`, `/admin/theme`, `/admin/lang`, `/admin/proxies`.

### 5.4 HTMX partial endpoints

UI 부분 렌더용. `/api/ui/*` 대신 **별도 prefix `/hx/*`** 로 구분 (사람-소비 API에 섞이지 않도록).

| Method | Path | Returns | Notes |
|---|---|---|---|
| `GET` | `/hx/sidebar` | HTML fragment | 좌측 섹션 트리 |
| `GET` | `/hx/services/grid` | HTML fragment | 중앙 서비스 카드 그리드 |
| `GET` | `/hx/agents/grid` | HTML fragment | 중앙 에이전트 카드 그리드 |
| `GET` | `/hx/keys/list` | HTML fragment | 중앙 key 리스트 |
| `GET` | `/hx/services/{id}/edit` | HTML fragment | 우측 slideover (서비스 편집 폼) |
| `GET` | `/hx/clients/{id}/edit` | HTML fragment | 우측 slideover (에이전트 편집 폼) |
| `POST/PUT` | `/hx/...` 동일 경로 | HTML fragment | 폼 제출 후 재렌더링 |

관리자 외 API 소비자(OpenClaw, Cline 등)는 `/hx/*` 를 쓰지 않음.

---

## 6. Dashboard UI — α One-Screen Hybrid

### 6.1 Layout

```
┌──────────── 헤더 (로고, 테마 토글, 언어 선택) ──────────────┐
├──────────┬─────────────────────────┬──────────────────────┤
│          │                         │                      │
│ SIDEBAR  │       MAIN              │     SLIDEOVER        │
│ (22%)    │       (flex)            │     (28%, on-demand) │
│          │                         │                      │
│ 🏠 Home  │  Services (카드 그리드) │  ✏︎ ollama            │
│ SERVICES │  ┌────┐ ┌────┐          │  Default Model       │
│  google  │  │svc │ │svc │          │  [gemma4:26b______]  │
│  openro. │  └────┘ └────┘          │  Local URL           │
│  anthro. │                         │  [http://.../11434]  │
│  ollama  │  Agents (카드 그리드)   │  Fallback #: [4]     │
│ AGENTS   │  ┌────┐ ┌────┐          │  Allowed Models      │
│  motoko  │  │agt │ │agt │          │  [gemma4:26b,        │
│  raspi   │  └────┘ └────┘          │   qwen3.5:35b,       │
│  mini    │                         │   gpt-oss:20b]       │
│  작순이  │  Keys (리스트)          │                      │
│ KEYS     │                         │  [Save] [Close]      │
│ MONITOR  │  Monitor (요약 카드)    │                      │
└──────────┴─────────────────────────┴──────────────────────┘
```

### 6.2 컴포넌트 카탈로그 (templ trees)

```
internal/vault/views/
├── layouts/
│   ├── base.templ             // <html>, head, HTMX script, 테마 CSS
│   ├── shell.templ            // sidebar + main + slideover 3-zone
│   └── theme.templ            // 7 테마 CSS 변수 정의
├── sidebar/
│   ├── sidebar.templ          // 전체 트리
│   ├── section_services.templ
│   ├── section_agents.templ
│   ├── section_keys.templ
│   └── section_monitor.templ
├── main/
│   ├── home.templ             // 섹션 요약 + 카드 그리드 렌더
│   ├── services_grid.templ
│   ├── service_card.templ
│   ├── agents_grid.templ
│   ├── agent_card.templ
│   ├── keys_list.templ
│   └── monitor_summary.templ
├── slideover/
│   ├── slideover.templ        // 외피
│   ├── service_edit.templ
│   ├── client_edit.templ
│   └── key_edit.templ
└── shared/
    ├── form_fields.templ      // input/select/textarea wrapper
    ├── badge.templ            // 🟢 proxy, 🔵 selected, 🟠 error, 🔴 cooldown
    └── modal.templ            // 확인 다이얼로그 (reorder, delete)
```

### 6.3 HTMX 상호작용

- 사이드바 항목 클릭 → `GET /hx/{section}/{id}/edit` → slideover 영역 (`#slideover`) swap
- 카드 클릭 → 동일 (card와 sidebar 항목은 같은 target으로 매핑)
- slideover 안 Save 버튼 → `PUT /hx/{section}/{id}` → 응답이 업데이트된 카드 partial + 성공 토스트 trigger
- 테마 토글 → `PUT /admin/theme` (기존 endpoint) → 전체 shell re-render (or CSS 변수만 swap)

### 6.4 기존 7 테마 유지

cherry / dark / light / ocean / gold / autumn / winter. 사용자 선호가 light 기본이라 default=`light`. CSS 변수 기반으로 `layouts/theme.templ` 하나에서 정의.

### 6.5 모바일/좁은 화면

- `@media (max-width: 900px)`: sidebar가 drawer(햄버거로 토글), slideover는 full-screen 모달로 승격
- 카드 그리드는 1열로 전환

---

## 7. Migration (v0.1 → v0.2)

### 7.1 트리거

`wall-vault vault` 기동 시:
1. `vault.json` 읽기
2. `schema_version` 필드 없거나 `1` → migration 대상
3. `vault.json.pre-v02.{ISO8601-UTC}.bak` 로 **강제 복사본 저장** (스킵 옵션 없음)
4. in-place 변환 후 `schema_version: 2` 기록

### 7.2 변환 규칙

**Client**:
- `default_service` → `preferred_service`
- `default_model` → `model_override`
- 나머지 필드 그대로

**Service**:
- 신규 필드 `default_model` = 기존 연결된 client 중 가장 흔한 `default_model` (tie-break: sort_order 낮은 client의 model); 연결 client 없으면 "":
  - 이후 첫 기동 시 admin UI 경고 배너 "`default_model` 이 비어있는 서비스가 있습니다" 노출
- 신규 필드 `sort_order` = 기존 내부 ordering 그대로 반영 (없으면 등록 순서)
- 신규 필드 `allowed_models` = `[]` (제한 없음으로 시작)

### 7.3 Rollback

기동 중 migration 실패:
1. in-memory 롤백 (변환 중 에러는 원본 유지)
2. `vault.json` 은 `.pre-v02.*.bak` 사본으로 수동 덮어쓰기로 복구 가능
3. 에러 로그에 명시적 복구 안내

---

## 8. Deployment & Rollback

### 8.1 브랜치 전략

- `v0.2-redesign` 브랜치에서 개발 완료 → main merge → tag `v0.2.0-rc1`
- rc1으로 mini vault 시범 cutover → 관찰 → `v0.2.0` tag → 전체 cutover

### 8.2 Cutover 순서

wall-vault는 단일 바이너리이고 미니에서는 vault+proxy가 같이 launchd로 기동되므로 (§CLAUDE.md launchd plist) 미니의 단일 프로세스 교체가 vault+proxy 동시 재기동을 의미한다.

1. 미니 vault+proxy stop (1-2분 downtime 시작; `launchctl unload` 양쪽 plist)
2. 새 바이너리 교체 (기존 `Makefile.local` deploy-mini 경로 재사용, 단 v0.2 전용 `deploy-mini` 타겟에 **pre-boot backup 확인** 단계 추가)
3. 미니 vault+proxy 기동 → migration 자동 실행 → 로그로 성공 확인 (startup log에 `migration: v1 → v2 success, backup=vault.json.pre-v02.*.bak`)
4. 라즈·모토코·작순이 proxy 를 순차 교체 (각 머신은 proxy만 돌며 vault 원격 호출, 다운타임 머신당 ~10초)
5. 텔레그램 봇 3개에 스모크 테스트 (§9.2 시나리오)

### 8.3 Rollback

v0.2 배포 실패 시 머신별:
- 바이너리: `~/.openclaw/wall-vault.bak.{pre-v02-migration}` 로 복구 (deploy 스크립트가 자동 백업)
- `vault.json`: `vault.json.pre-v02.{ts}.bak` 로 수동 복구

### 8.4 `/admin/services/{id}` 에 대한 하위호환 힌트

Admin API는 clean break라 구 소비자(대시보드 외 CLI 등)는 같이 업데이트 필요. 단, **매뉴얼 MANUAL.md + 16 언어 번역** 에 스키마 변경 표 추가.

---

## 9. Verification

### 9.1 Unit tests (신규)

- `vault/store_test.go`: migration 함수가 v0.1.x 샘플 파일 → v0.2 구조로 정확히 변환
- `proxy/dispatch_test.go`:
  - primary 성공 → 1회 호출
  - primary 실패 → fallback 체인 순서 보장 + **각 서비스가 자기 default_model 사용** 검증
  - model_override 있음 + whitelist 통과 → override 적용
  - model_override 있음 + whitelist 위반 → 400
  - 모든 서비스 실패 → 502 + 상세 에러 체인
- `vault/server_test.go`: admin API PUT 시 whitelist 위반 → 422

### 9.2 End-to-end

- 로컬 샌드박스에서 현 운영 `vault.json` 사본으로 migration 시험. 7 client × 5 service 전부 new schema로 변환됐는지 field-by-field assert
- templ 렌더링 snapshot: 3-zone layout이 모든 섹션에 대해 fragment를 올바로 반환하는지
- 실 텔레그램 봇(모토코·라즈·작순이) 에 단문 프롬프트 + tool-use 요청(`curl 한번 해봐`) 시험 — tool_use 블록 보존·실 실행 결과 응답 확인

### 9.3 성능 기준선

Migration 자체는 수십 ms 이내 (in-memory 변환, 수십 개 레코드). gateway startup grace 60s 안에 여유롭게 완료.

---

## 10. 다음 라운드 예고 (이번 spec 범위 밖)

| Phase | 내용 | 의존 |
|---|---|---|
| 3. Security | admin/operator/viewer 3-role RBAC, 세션 쿠키, audit log, `WV_ADMIN_TOKEN` 단일키 폐기 | v0.2 스키마 확정 후 |
| 4. Ops | webhook fanout, `/metrics` Prometheus, 대시보드에서 1-click backup/restore | v0.2 UI 위에 |

---

## 11. 수정·신규 대상 파일 요약

**수정**
- `internal/vault/models.go` — Service/Client 구조 재정의
- `internal/vault/store.go` — schema_version, migration 함수
- `internal/vault/server.go` — admin API body schema clean break
- `internal/vault/broker.go` — 이벤트 payload 필드 이름만 조정
- `internal/proxy/server.go` — dispatch() 재작성, `resolveModel` 헬퍼
- `internal/proxy/keymgr.go` — 호출부 signature 조정만
- `internal/proxy/convert.go` — v0.1.29 수정 그대로, 호출부 맞춤
- `Makefile` — BASE_VERSION v0.2.0, `templ generate` 타겟 + `build` 전에 실행
- `Makefile.local.example` — deploy 타겟이 pre-boot backup 확인 단계 포함
- `CHANGELOG.md` — 0.2.0 섹션
- `docs/MANUAL.md` + 16 번역 — 스키마 변경 요약

**신규**
- `internal/vault/views/` 전체 templ 트리 (§6.2 구조)
- `internal/vault/views/*_templ.go` — 생성물 커밋
- `internal/vault/migrate.go` — migration 전용 함수 모듈

**삭제**
- `internal/vault/ui.go` (2,572줄 전체) — templ 트리로 이관 완료 후
