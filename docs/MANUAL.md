# wall-vault 사용자 매뉴얼

오픈클로(OpenClaw)와 wall-vault를 함께 쓰는 방법을 중심으로 설명합니다.
*(Last updated: 2026-03-16 — v0.1.7: today_attempts tracking, HTTP 582 cooldown, share-of-total bar scaling, custom/ routing fix, Ollama timeout fix, key_att i18n, logo display fix, avatar heartbeat sync, build timestamp versioning, proxy-only service filter)*

---

## 목차

1. [wall-vault란?](#wall-vault란)
2. [설치](#설치)
3. [처음 시작하기 (setup 마법사)](#처음-시작하기)
4. [API 키 등록](#api-키-등록)
5. [프록시 사용법](#프록시-사용법)
6. [키 금고 대시보드](#키-금고-대시보드)
7. [분산 모드 (멀티 봇)](#분산-모드-멀티-봇)
8. [자동 시작 설정](#자동-시작-설정)
9. [Doctor 주치의](#doctor-주치의)
10. [환경변수 참고](#환경변수-참고)
11. [문제 해결](#문제-해결)

---

## wall-vault란?

**wall-vault = 오픈클로(OpenClaw)를 위한 AI 프록시 + API 키 금고**

오픈클로와 LLM API 사이에 앉아서, 세션을 방해할 모든 요소를 대신 처리합니다:

- **API 키 자동 순환**: 한도 초과·쿨다운 키는 건너뛰고 다음 키로 자동 전환
- **서비스 폴백**: Google 실패 → OpenRouter → Ollama 자동 전환. 오픈클로는 계속 됩니다
- **SSE 실시간 동기화**: 금고 대시보드에서 모델을 바꾸면 1-3초 내 오픈클로 TUI에 반영
- **Unix 소켓 이벤트**: 키 소진·서비스 다운 알림이 TUI 하단에 실시간 표시

Claude Code, Cursor, VS Code도 연결해서 쓸 수 있지만, 오픈클로가 본래 목적입니다.

```
오픈클로 (TUI)
        │
        ▼
  wall-vault 프록시 (:56244)   ← 키 관리, 라우팅, 폴백, 이벤트
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ 모델)
        └─ Ollama (로컬, 최종 폴백)
```

---

## 설치

### Linux / macOS

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

### Windows

PowerShell에서:

```powershell
# 다운로드
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATH 추가 (PowerShell 재시작 후 적용)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

### 소스에서 빌드

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (버전: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> **빌드 타임스탬프 버전**: `make build` 실행 시 버전이 자동으로 `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` 형식으로 생성됩니다 (예: `v0.1.6.20260314.231308`). `go build ./...`로 직접 빌드하면 버전이 `"dev"`로 표시됩니다.

---

## 처음 시작하기

### setup 마법사 실행

```bash
wall-vault setup
```

마법사가 다음 항목을 단계별로 안내합니다:

```
1. 언어 선택 (10개 언어)
2. 테마 선택 (light / dark / gold / cherry / ocean)
3. 운용 모드 (standalone / distributed)
4. 봇 이름 입력
5. 포트 설정 (기본: 프록시 56244, 금고 56243)
6. AI 서비스 선택 (Google / OpenRouter / Ollama)
7. 도구 보안 필터 설정
8. 관리자 토큰 설정 (자동 생성 가능)
9. API 키 암호화 비밀번호 설정 (선택)
10. 설정 파일 저장 경로
```

마법사 완료 후 `wall-vault.yaml` 파일이 생성됩니다.

### 실행

```bash
wall-vault start
```

두 개의 서버가 시작됩니다:
- **프록시**: `http://localhost:56244`
- **키 금고**: `http://localhost:56243`

---

## API 키 등록

### 방법 1: 환경변수 (권장 — 가장 간단)

```bash
# Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# 실행
wall-vault start
```

여러 키는 쉼표로 구분합니다 (라운드 로빈 자동 적용):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

### 방법 2: 대시보드 UI

1. 브라우저에서 `http://localhost:56243` 접속
2. 상단 **🔑 API 키** 카드에서 `[+ 추가]` 클릭
3. 서비스, 키, 레이블, 일일 한도 입력 후 저장

### 방법 3: REST API

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "메인 키",
    "daily_limit": 1000
  }'
```

### 방법 4: proxy 플래그 (임시 사용)

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## 프록시 사용법

### OpenClaw에서 사용 (주목적)

`~/.openclaw/openclaw.json`에 wall-vault 프로바이더 등록:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault 에이전트 토큰
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 무료 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 에이전트 카드의 **🦞 OpenClaw 설정 복사** 버튼을 누르면 이 스니펫이 자동 생성됩니다.

**`wall-vault/` 접두어 라우팅:**

| 모델 형식 | 라우팅 |
|----------|--------|
| `wall-vault/gemini-*` | Google Gemini 직접 |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI 직접 |
| `wall-vault/claude-*` | OpenRouter (Anthropic 경로) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (무료 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/모델명`, `openai/모델명`, `anthropic/모델명` 등 | 해당 서비스 직접 |
| `custom/google/모델명`, `custom/openai/모델명` 등 | `custom/` 제거 후 재라우팅 |
| `모델명:cloud` | `:cloud` 제거 후 OpenRouter |

### Gemini API 형식 (기존 호환)

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

또는 직접 URL:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK에서 사용

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault가 관리
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model 형식
    messages=[{"role": "user", "content": "안녕하세요"}]
)
```

### 모델 변경

실행 중에 모델을 바꾸려면:

```bash
# API로 변경
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 분산 모드에서는 금고에서 변경 (SSE로 즉시 반영)
curl -X PUT http://localhost:56243/admin/clients/내-봇-id \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 사용 가능한 모델 목록

```bash
# 전체 목록
curl http://localhost:56244/api/models | python3 -m json.tool

# Google만
curl "http://localhost:56244/api/models?service=google"

# 검색
curl "http://localhost:56244/api/models?q=claude"
```

**지원 모델 요약 (v0.1.3):**

| 서비스 | 주요 모델 |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash, gemini-embedding-2-preview |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346개+ (Hunter Alpha 1M ctx 무료, Healer Alpha, Kimi K2.5, GLM-5, DeepSeek R1/V3, Qwen 2.5 등) |
| Ollama | 로컬 서버 자동 감지 (폴백: glm-4.7-flash, qwen3.5:35b, deepseek-r1:7b) |

---

## 키 금고 대시보드

`http://localhost:56243` 접속.

### API 키 카드

- 등록된 키 목록, 서비스별 구분
- 각 키의 오늘 사용량·시도 횟수·일일 한도·쿨다운 상태 표시
  - **`today_usage`**: 성공한 요청의 토큰 수 (429/402/582 오류 미포함)
  - **`today_attempts`**: 오늘 총 API 호출 수 (성공 + rate-limited 포함)
  - 대시보드 표시 예: `"42 req (45 att)"` — 42 성공, 45 총 시도
  - 무제한 키(`daily_limit=0`) 막대 그래프: 같은 서비스 내 **share-of-total** 스케일 (각 키의 활동 / 서비스 전체 활동 × 100%)
- `[+ 추가]` — 새 키 등록
- `✕` — 키 삭제 (관리자 토큰 필요)
- 사용량 초기화 버튼

### 에이전트 카드

- 연결된 프록시 에이전트 목록
- **에이전트 상태 4단계 표시:**
  - 🟢 **실행 중** — 3분 이내 heartbeat 수신 (서비스/모델 실시간 표시)
  - 🟡 **지연** — 3-10분 전 마지막 heartbeat
  - 🔴 **오프라인** — 10분 이상 응답 없음
  - ⚫ **미연결** — 활성화되어 있으나 프록시가 아직 heartbeat를 보내지 않음
  - ⚫ **비활성화** — 클라이언트 비활성 상태

> **"미연결"이란?** 클라이언트 설정은 저장되어 있지만, 해당 ID의 프록시 프로세스가 아직 vault에 접속하지 않은 상태입니다. [적용] 버튼으로 저장에 성공하면 **✓ 저장됨** 메시지가 표시됩니다 — "미연결"과는 무관합니다.

- **인라인 서비스·모델 변경 폼:**
  - 서비스 드롭다운 변경 시 → 해당 서비스 모델 목록 자동 로드
  - 서비스/키 변경 이벤트 수신 시 → 드롭다운 실시간 갱신 (페이지 새로고침 없음)
  - `[💾 저장]` 클릭 후 성공 시 → 버튼에 **✓** 표시 + "✓ 저장됨" 인라인 알림 (3초)

**에이전트 아바타 (v0.1.6~):**

에이전트 카드에 아바타 이미지가 표시됩니다. 아바타는 다음 두 가지 방식으로 지정합니다:

| 방법 | 형식 | 예시 |
|------|------|------|
| 상대 경로 | `~/.openclaw/` 기준 경로 | `workspace/avatar.png`, `workspace/avatars/profile.hpg` |
| base64 URI | `data:image/...;base64,...` | 직접 임베드 |

지원 확장자: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`

에이전트 종류가 `openclaw`이거나 지정되지 않은 경우 기본값으로 `~/.openclaw/workspace/avatar.png`를 시도합니다.

**에이전트 종류별 설정 복사 버튼 (v0.1.3~):**

| 종류 | 아이콘 | 버튼 레이블 | 복사 내용 |
|------|--------|------------|----------|
| openclaw | 🦞 | OpenClaw 설정 복사 | `~/.openclaw/openclaw.json` 스니펫 |
| claude-code | 🟠 | Claude Code 설정 복사 | `~/.claude/settings.json` 스니펫 |
| cursor | ⌨ | Cursor 설정 복사 | Cursor AI API Base URL + Key |
| vscode | 💻 | VSCode 설정 복사 | Continue 확장 `config.json` 스니펫 |
| 기타 | 📋 | 설정 복사 | OpenClaw 형식 |

- `[✎]` — 편집, `[✕]` — 삭제

### 에이전트 추가/편집 모달 필드 순서

| 순서 | 필드 | 설명 |
|------|------|------|
| 1 | ID | 영문·숫자·하이픈 |
| 2 | 이름 | 표시 이름 |
| 3 | **에이전트 종류** | openclaw / claude-code / cursor / vscode / custom |
| 4 | **작업 디렉토리** | 에이전트 종류 선택 시 힌트 자동 제안 |
| 5 | 기본 서비스 | google / openai / openrouter 등 |
| 6 | 기본 모델 | **드롭다운** 선택 (서비스 변경 시 자동 로드) + 직접 입력 병행 가능 |
| 7 | 설명 | 선택 사항 |
| 8 | **허용 IP** | 쉼표 구분, CIDR 지원 (`10.0.0.0/24`) |
| 9 | 토큰 | 빈칸이면 자동 생성 |
| 10 | 활성화 | 체크박스 |

> **IP 화이트리스트**: 입력된 IP/CIDR 목록만 `/api/keys` 접근 허용. 빈칸이면 모두 허용.

### 서비스 카드

- 서비스별 활성화·비활성화 토글
- **프록시 사용** 체크박스 — 체크된 서비스만 오픈클로 프록시의 dispatch에 포함됨. SSE로 실시간 반영.
  - 체크된 서비스만 에이전트 카드의 **서비스·모델 드롭다운에도 표시됨** (v0.1.6~)
- 로컬 AI 서버(Ollama/LMStudio/vLLM) URL 입력 및 모델 자동 감지
- 커스텀 서비스 추가·삭제

### 관리자 토큰 입력

브라우저에서 관리자 기능(키 삭제 등) 사용 시 토큰 입력 팝업이 표시됩니다. 한 번 입력하면 브라우저 세션 동안 유지됩니다.

> **보안**: 관리자 토큰 인증 실패가 15분 내 10회 초과하면 해당 IP는 일시 차단됩니다.

---

## 분산 모드 (멀티 봇)

여러 머신에서 오픈클로를 운영할 때, 하나의 키 금고를 공유하는 구성입니다.

### 구성 예시 (실제 운용 환경)

```
[키 금고 서버]
  wall-vault vault    (키 금고 :56243, 대시보드)

[WSL 알파]            [라즈베리파이 감마]    [맥미니 로컬]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE 동기화          ↕ SSE 동기화            ↕ SSE 동기화
```

금고에서 모델 변경 → 세 머신의 오픈클로 TUI에 1-3초 내 반영.

### 키 금고 서버 시작

```bash
wall-vault vault
```

### 클라이언트 등록 (금고 서버에서)

```bash
# 봇A 등록
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "봇A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### 프록시 시작 (각 봇)

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

또는 `wall-vault.yaml`:

```yaml
mode: distributed
proxy:
  vault_url: http://192.168.x.x:56243
  vault_token: bota-secret
  client_id: botA
```

### SSE 실시간 동기화 확인

금고에서 모델을 변경하면 1–3초 내 프록시에 자동 반영됩니다:

```bash
# 금고에서 봇A 모델 변경
curl -X PUT http://192.168.x.x:56243/admin/clients/botA \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "openrouter", "default_model": "anthropic/claude-opus-4"}'

# 봇A 프록시에서 즉시 확인
curl http://봇A주소:56244/status
```

**SSE 이벤트 종류:**

| 이벤트 | 트리거 | 대시보드/프록시 반응 |
|--------|--------|---------------------|
| `config_change` | 클라이언트 모델/서비스 변경 | 모델 드롭다운 갱신 + 오픈클로 TUI 즉시 반영 |
| `service_changed` | 서비스 추가/수정/삭제 | 서비스 select + 모델 드롭다운 갱신 + 프록시 dispatch 서비스 목록 갱신 |
| `key_added` / `key_deleted` | API 키 추가/삭제 | 모델 드롭다운 갱신 + 프록시 키 재동기화 |
| `usage_update` | 프록시 heartbeat 수신 시 (20초마다) | 전체 키 사용량·쿨다운 즉시 갱신 (추가 fetch 없음, SSE 데이터 직접 반영) |
| `usage_reset` | 일일 사용량 초기화 | 페이지 새로고침 |
| `heartbeat` | 프록시 20초마다 | 에이전트 상태 업데이트 |

---

## 자동 시작 설정

### Linux — systemd

```bash
# 서비스 파일 생성
wall-vault doctor deploy

# 등록 및 활성화
systemctl --user daemon-reload
systemctl --user enable --now wall-vault

# 상태 확인
systemctl --user status wall-vault

# 로그 보기
journalctl --user -u wall-vault -f
```

### macOS — launchd

```bash
# plist 파일 생성
wall-vault doctor deploy launchd

# 등록
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist

# 상태 확인
launchctl list | grep wall-vault

# 중지
launchctl unload ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. [nssm.cc](https://nssm.cc/download) 에서 NSSM 다운로드 후 PATH에 추가
2. 서비스 설치 스크립트 생성:
   ```powershell
   .\wall-vault.exe doctor deploy windows
   ```
3. 관리자 권한 PowerShell에서 실행:
   ```powershell
   & "$env:USERPROFILE\install-wall-vault-service.bat"
   ```
4. 서비스 관리:
   ```powershell
   nssm start wall-vault
   nssm stop wall-vault
   nssm status wall-vault
   ```

---

## Doctor 주치의

서비스 헬스체크와 자동복구 기능입니다.

```bash
# 상태 확인만
wall-vault doctor check

# 상세 보고서 (터미널 UI)
wall-vault doctor status

# 자동 복구 실행
wall-vault doctor fix

# 확인 후 필요시 자동 복구
wall-vault doctor all
```

**자동복구 우선순위:**
1. systemd 서비스 재시작 (Linux/WSL)
2. launchd 재시작 (macOS)
3. NSSM 재시작 (Windows)
4. 직접 프로세스 시작

**주기적 헬스체크 (권장):**

```bash
# crontab에 추가 — 5분마다 확인
*/5 * * * * wall-vault doctor all >> /tmp/wv-doctor.log 2>&1
```

---

## 환경변수 참고

| 변수 | 설명 | 예시 |
|------|------|------|
| `WV_LANG` | 언어 | `ko`, `en`, `ja` |
| `WV_THEME` | 테마 | `light`, `dark`, `gold`, `cherry`, `ocean` |
| `WV_KEY_GOOGLE` | Google API 키 (쉼표로 여러 개) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API 키 | `sk-or-v1-...` |
| `WV_VAULT_URL` | 금고 서버 URL (distributed) | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | 클라이언트 인증 토큰 | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 관리자 토큰 | `admin-token-here` |
| `WV_MASTER_PASS` | API 키 암호화 비밀번호 | `my-password` |
| `WV_PROXY_PORT` | 프록시 포트 오버라이드 | `8080` |
| `WV_VAULT_PORT` | 금고 포트 오버라이드 | `8081` |
| `WV_AVATAR` | 프록시 로컬 아바타 파일 경로 (`~/.openclaw/` 기준 상대 경로) — heartbeat마다 base64로 전송, vault가 자동 저장 | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama 서버 URL | `http://192.168.x.x:11434` |
| `VAULT_CLIENT_ID` | 클라이언트 ID (레거시 호환) | `bot-a` |

---

## 문제 해결

### 프록시가 시작되지 않음

```bash
# 포트 점유 확인
ss -tlnp | grep 56244

# 다른 포트로 시작
wall-vault proxy --port 8080

# 로그 확인 (systemd)
journalctl --user -u wall-vault --since "5 minutes ago"
```

### API 키 오류 (429, 402, 401, 403, 582)

```bash
# 현재 키 상태 확인
curl -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys

# 쿨다운 중인 키가 있으면 일일 초기화
curl -X POST -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys/reset
```

### SSE 연결이 안 됨 (distributed 모드)

```bash
# SSE 스트림 직접 테스트
curl -s http://금고서버:56243/api/events --max-time 5

# 방화벽 확인 (ufw)
sudo ufw allow 56243/tcp

# 프록시 SSE 상태
curl http://localhost:56244/status | python3 -c "import sys,json; d=json.load(sys.stdin); print('SSE:', d['sse'])"
```

### 에이전트가 "미연결"로 표시됨

"미연결"은 프록시 프로세스가 vault에 heartbeat를 보내지 않는 상태입니다. 설정 저장([적용] 버튼)과는 **무관**합니다.

프록시를 연결하려면 다음 환경변수로 실행하세요:

```bash
WV_VAULT_URL=http://금고서버:56243 \
WV_VAULT_TOKEN=클라이언트토큰 \
WV_VAULT_CLIENT_ID=클라이언트ID \
wall-vault proxy
```

연결 성공 시 대시보드에서 🟢 실행 중으로 바뀝니다 (약 20초 이내).

> **[적용] 후 ✓ 저장됨이 표시되면** 설정은 정상 저장된 것입니다. "미연결"은 프록시 연결 여부를 뜻할 뿐입니다.

### 모델이 변경되지 않음

```bash
# 현재 서비스·모델 확인
curl http://localhost:56244/status

# 즉시 재동기화
curl -X POST http://localhost:56244/reload
```

### Ollama 연결 실패

```bash
# Ollama 실행 여부 확인
curl http://localhost:11434/api/tags

# Ollama가 다른 IP에 있는 경우
export OLLAMA_URL=http://192.168.x.x:11434
wall-vault start
```

> **대형 모델 추론 타임아웃**: wall-vault의 Ollama 호출은 타임아웃 없이 실행됩니다. 대형 모델(`qwen3.5:35b`, `deepseek-r1` 등)은 응답 생성까지 수 분이 걸릴 수 있으며, 헤더 전송 전까지 응답이 없어도 정상입니다. `stream:false` 모드에서는 생성 완료 후 한 번에 응답하므로 특히 오래 걸릴 수 있습니다.

**HTTP 582 (Gateway Overload) 오류:**

upstream 게이트웨이 과부하 응답. 수신 시 해당 키에 **5분 쿨다운**이 적용됩니다. `today_attempts`는 증가하지만 `today_usage`는 변경되지 않습니다. 5분 후 자동으로 재시도됩니다.

### Windows에서 경로 문제

설정 파일과 데이터가 `%APPDATA%\wall-vault\` 에 저장됩니다:

```powershell
# 설정 파일 위치
echo $env:APPDATA\wall-vault\

# 직접 경로 지정
.\wall-vault.exe start --config C:\Users\나\wall-vault.yaml
```

---

*더 자세한 API 정보는 [API.md](API.md)를 참고하세요.*
