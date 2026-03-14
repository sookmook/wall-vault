# Changelog

wall-vault의 모든 주요 변경 사항을 기록합니다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따릅니다.

---

## [0.1.5] — 2026-03-14

### Added
- Proxy service selection: "프록시 사용" checkbox per service card → only checked services are used by OpenClaw proxy
- `/api/services` endpoint (client-auth): returns list of proxy-enabled service IDs
- `service_changed` SSE now includes `proxy_services []string` — proxy updates `allowedServices` in real-time
- `proxy/sseconn.go`: `onServiceChange` callback for live proxy service filtering
- `proxy/server.go`: `allowedServices` field + `syncAllowedServices()` on startup

### Changed
- Service card UI: removed redundant ID label below service name (name already identifies the service)
- Documentation reframed around OpenClaw as primary use case (README, MANUAL, API)

### Fixed
- Binary on mini was v0.1.3; redeployed v0.1.4+ darwin/arm64 build

---

## [0.1.3] — 2026-03-13

### Added
- Agent card redesign — per-type icons & config copy buttons:
  - `openclaw` → 🦞 (red lobster) + "OpenClaw 설정 복사" button
  - `claude-code` → 🟠 + "Claude Code 설정 복사" (copies `~/.claude/settings.json` snippet)
  - `cursor` → ⌨ + "Cursor 설정 복사" (copies Cursor AI API settings)
  - `vscode` → 💻 + "VSCode 설정 복사" (copies Continue extension `config.json` snippet)
  - generic/custom → 📋 "설정 복사" (OpenClaw format)
- `copyAgentConfig(clientId, agentType)`: per-type proxy config generator (JS)
- Connection status chip with context hint ("● 프록시 미연결" + heartbeat explanation)
- `💾 저장` button replaces bare "적용" — intent is now explicit
- After save: `✓ 저장됨` inline indicator in status area (3s) + `✓` on button (2s)
- New CSS: `.atbadge`, `.atb-openclaw/claude/cursor/vscode/custom`, `.agent-status`,
  `.status-live/delay/offline/dc/hint/version`, `.btn-cfg`, `.btn-cfg-openclaw/claude`, `.btn-save`

### Changed
- `buildAgentsCard()`: fully rewritten with per-item `strings.Builder` (no more single large `Sprintf`)
- Agent type badge: colored pill per type (red=openclaw, orange=claude-code, blue=cursor/vscode)
- Status display: `미연결` → `● 프록시 미연결 — heartbeat 미수신` with guidance

### Fixed
- "미연결" ambiguity: users can now distinguish proxy connection state from config-save result

---

## [0.1.2] — 2026-03-13

### Added
- `callOpenAI()`: direct OpenAI API handler (separate from OpenRouter)
- `dispatch()`: `openai` case (direct), `anthropic` case (via OpenRouter with `anthropic/model` path)
- `parseProviderModel()` comprehensive rewrite (OpenClaw 3.11 compatibility):
  - `anthropic/` → OpenRouter `anthropic/model` (Anthropic API format differs)
  - `openai/` → direct OpenAI
  - `:cloud` suffix (Ollama cloud tags) → strip + route to OpenRouter
  - New prefixes: `opencode`, `opencode-go`, `opencode-zen`, `moonshot`, `kimi-coding`,
    `groq`, `mistral`, `cohere`, `perplexity`, `minimax`, `together`, `huggingface`,
    `nvidia`, `venice`, `meta-llama`, `qwen`, `deepseek`, `01-ai`
  - `wall-vault/claude-*` → OpenRouter `anthropic/model` (was incorrectly routing to `anthropic` service)
- `stripControlTokens()`: removes GLM-5 / DeepSeek / ChatML control tokens from responses (`<|im_start|>`, `[gMASK]`, `[sop]`, etc.)
- `fetchOpenRouterKnown()`: curated fallback model list — Hunter Alpha (1M ctx, free), Healer Alpha, Kimi K2.5, GLM-5, GLM-4.7 Flash, DeepSeek R1/V3, Qwen 2.5, MiniMax M2.5, Llama 3.3
- `OllamaRecommended()`: Ollama 서버 미응답 시 추천 모델 폴백 (glm-4.7-flash, qwen3.5:35b, deepseek-r1:7b 등)
- Google model list: `gemini-2.5-flash-8b`, `gemini-embedding-2-preview` (OpenClaw 3.11 memorySearch)
- OpenAI model list: `o3` 추가

### Changed
- OpenRouter fetch 실패 시 `fetchOpenRouterKnown()` 폴백 적용
- Ollama 서버 미응답 시 `OllamaRecommended()` 폴백 적용
- Response text in `/v1/chat/completions` now passes through `stripControlTokens()`

### Fixed
- `anthropic` / `openai` 서비스가 `dispatch()`에서 묵묵히 무시되던 버그 수정
- `wall-vault/claude-*` 모델이 실제로 호출되지 않던 버그 수정

---

## [0.1.1] — 2026-03-13

### Added
- Agent card: model dropdown + manual input combo (same as modal) with auto-load on page
- `onAgentServiceChange()`, `onModelSelect()` JS functions for agent service/model combo
- DOMContentLoaded initializer pre-loads model lists for all agent cards on page load
- README: OpenClaw integration section (KO + EN) — socket events, SSE sync, dir layout
- README: multilingual sections (zh, ja, es, fr, de)
- README: copyright/license notice (GPL-3.0)

### Changed
- License: MIT → GPL-3.0
- Theme order unified to light/dark/gold/cherry/ocean across all code and docs
- Agent modal: model field upgraded from datalist to select+input combo
- All commit messages in English going forward

### Fixed
- `setTheme()` / `setLang()` missing `Authorization` header → 401 on theme/lang change
- `server.go` theme error message updated to reflect correct order

---

## [Unreleased]

### 보안 (Security)
- `/admin/theme`, `/admin/lang` 엔드포인트에 `adminAuth` 미들웨어 적용 (기존 무인증 취약점 수정)
- `/api/keys` 핸들러에 IP 화이트리스트 실제 적용 — CIDR 표기법 지원 (`net.ParseCIDR`), `X-Forwarded-For` 헤더 처리
- 관리자 인증 실패 rate limiting 추가: 15분 이내 10회 실패 시 `429 Too Many Requests` 반환
- `realIP()`, `ipAllowed()` 헬퍼 함수 추가

### 추가 (Added)
- 에이전트 모달 기본 모델 선택 UI 개선: 드롭다운으로 서비스별 전체 모델 목록 제공, 직접 입력 병행 지원 (`onAgentServiceChange`, `onModelSelect`)
- 에이전트 상태 4단계 표시: 🟢 실행 중 (<3분) / 🟡 지연 (3-10분) / 🔴 오프라인 (>10분) / ⚫ 비활성·미연결
- `.dot-yellow` CSS 클래스 추가 (+ glow effect)
- `.dot-red` CSS glow effect 추가
- 에이전트 모달 `vscode` 에이전트 종류 옵션 추가
- 에이전트 종류 선택 시 작업 디렉토리 자동 힌트 (`onAgentTypeChange` JS 함수)
  - `openclaw` → `~/.openclaw`
  - `claude-code` → `~/.claude`
  - `cursor` / `vscode` → `~/projects`
- `docs/logo.png` 로고 파일 추가
- README.md 탄생 배경 스토리 및 전면 개편 (MuJaMae 스타일)

### 수정 (Fixed)
- 에이전트 모달 필드 순서 정리: ID → 이름 → 에이전트 종류 → 작업 디렉토리 → 기본 서비스 → 기본 모델 → 설명 → 허용 IP → 토큰 → 활성화
- `buildClientModalBody` `fmt.Sprintf` 인자 수 불일치 수정 (19개 verbs / 20개 args → 20/20)
- 오프라인 상태(`dot-red`) CSS 클래스가 실제로 적용되지 않던 버그 수정

---

## [0.1.0] — 2026-03-11

### v0.1.0 이후 추가 (post-release)
- `cmd/proxy`: `--key-google`, `--key-openrouter`, `--vault`, `--vault-token`, `--filter` 플래그 추가
- `internal/models`: `Registry.NeedsRefresh()`, `Registry.Search(query)` 추가
- `internal/proxy/server_test.go`: 프록시 HTTP 핸들러 테스트 12개
- `internal/vault/server_test.go`: 금고 HTTP 핸들러 테스트 15개
- `internal/middleware/middleware_test.go`: 미들웨어 체인 테스트 8개
- `internal/hooks/hooks_test.go`: 훅 시스템 테스트 7개
- `docs/API.md`: 전체 API 엔드포인트 매뉴얼
- `docs/MANUAL.md`: 사용자 가이드 (설치→분산모드→문제해결)
- `CONTRIBUTING.md`: 기여 가이드
- GitHub Actions CI/Release 워크플로우 (로컬 준비 완료)

---

## [0.1.0] — 2026-03-11

### 초기 릴리스 (단일 Go 바이너리)

#### 아키텍처
- **단일 바이너리** `wall-vault` — 서브커맨드 방식 (start / proxy / vault / doctor / setup)
- **standalone / distributed** 두 가지 운용 모드
- **SSE(Server-Sent Events)** 실시간 설정 동기화 (금고 → 프록시, 1–3초 이내 반영)
- **AES-GCM 암호화** — 마스터 비밀번호 기반 API 키 영속화

#### 서브커맨드

| 커맨드 | 설명 |
|--------|------|
| `wall-vault start` | 프록시 + 키 금고 동시 실행 (standalone) |
| `wall-vault proxy` | 프록시 단독 실행 |
| `wall-vault vault` | 키 금고 단독 실행 |
| `wall-vault doctor` | 헬스체크 및 자동복구 |
| `wall-vault setup` | 대화형 설치 마법사 |

#### 프록시 기능
- **Google Gemini / OpenRouter / Ollama** 동시 지원
- **라운드 로빈 키 관리** — `idx map[string]int`로 서비스별 인덱스 추적
- **쿨다운 관리** — 429: 30분, 400/401/403: 24시간, 네트워크 오류: 10분
- **도구 보안 필터** — strip_all / whitelist / passthrough
- **폴백 체인** — Google → OpenRouter → Ollama
- **훅 시스템** — 모델 변경·키 소진·서비스 다운 시 셸 명령 실행
- **OpenClaw 소켓** 연동 지원

#### 키 금고 (Vault)
- **REST API** — `/api/keys`, `/api/clients`, `/api/status`
- **SSE 브로드캐스트** — `/api/events` 엔드포인트
- **웹 대시보드** — 테마(sakura/dark/light/ocean), 키 CRUD, 클라이언트 관리
- **관리자 토큰** 기반 인증

#### Doctor (헬스체크/자동복구)
- `doctor check` / `fix` / `status` / `all` / `deploy` 서브커맨드
- 자동복구 우선순위: **systemd → launchd → NSSM(Windows) → 직접 프로세스**
- `deploy` — systemd / launchd / NSSM 서비스 파일 자동 생성

#### Setup 마법사
- **세계 10대 언어** — ko/en/zh/es/hi/ar/pt/fr/de/ja
- 테마·모드·포트·서비스·도구 필터·보안 토큰 대화형 구성
- Ollama 서버 자동 연결 및 모델 목록 조회
- `crypto/rand` 기반 안전한 관리자 토큰 자동 생성

#### 다국어 (i18n)
- 세계 10대 언어 지원
- LANG / WV_LANG 환경변수 자동 감지
- 로케일 문자열 파싱 (e.g. `ko_KR.UTF-8` → `ko`)
- 영어 폴백 보장

#### 플랫폼 지원
- **Linux** (amd64 / arm64)
- **macOS** (amd64 / arm64, Apple Silicon)
- **Windows** (amd64, NSSM 서비스 지원)
- **WSL** 완벽 지원

#### 모델 레지스트리
- Google: 6개 고정 모델 (Gemini 1.5/2.0/2.5)
- OpenRouter: 346개+ 동적 조회
- Ollama: 로컬 서버 자동 탐지
- TTL 기반 캐시 (기본 10분)
- 대소문자 무시 모델 ID/이름 검색

#### 서비스 플러그인
- `~/.wall-vault/services/*.yaml` 기반 외부 서비스 플러그인 로더
- `enabled: true/false` 필드로 런타임 활성화 제어

#### 테스트 (39개, 전부 PASS)
- `crypto_test.go` — AES-GCM 암호화·복호화·랜덤 nonce (5개)
- `toolfilter_test.go` — strip_all·whitelist·passthrough (5개)
- `convert_test.go` — Gemini↔OpenAI↔Ollama 포맷 변환 (6개)
- `services_test.go` — 플러그인 로더 엣지 케이스 (5개)
- `keymgr_test.go` — 라운드 로빈·쿨다운·일일한도 (8개)
- `store_test.go` — 키/클라이언트 CRUD·영속화 (10개)

#### CI/CD
- GitHub Actions CI — push/PR 시 vet + test + 4플랫폼 크로스컴파일
- GitHub Actions Release — v* 태그 시 자동 GitHub Release 생성

---

[Unreleased]: https://github.com/sookmook/wall-vault/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/sookmook/wall-vault/compare/v0.1.3...v0.1.5
[0.1.3]: https://github.com/sookmook/wall-vault/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/sookmook/wall-vault/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/sookmook/wall-vault/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/sookmook/wall-vault/releases/tag/v0.1.0
